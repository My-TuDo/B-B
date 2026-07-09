package worker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	mmodel "github.com/My-TuDo/B-B/backend/internal/model/meta"
	qmodel "github.com/My-TuDo/B-B/backend/internal/model/quality"
	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const workDir = "/tmp/bb-transcode"

// transcodeTargets defines the resolution targets for transcoding.
var transcodeTargets = []struct {
	Label  string
	Width  int
	Height int
}{
	{"360p", 640, 360},
	{"480p", 854, 480},
	{"720p", 1280, 720},
	{"1080p", 1920, 1080},
}

// ProcessVideo handles the full transcode workflow for a single video.
// Returns an error if the entire process fails; individual step failures are logged and handled gracefully.
func ProcessVideo(videoID uint, db *gorm.DB) {
	ctx := context.Background()
	logger.Info("worker: starting transcode", zap.Uint("video_id", videoID))

	// --- Check ffmpeg / ffprobe availability ---
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		logger.Warn("worker: ffmpeg not found, skipping transcode", zap.Uint("video_id", videoID))
		updateStatus(db, ctx, videoID, tmodel.StatusDone, 100, "")
		return
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		logger.Warn("worker: ffprobe not found, skipping transcode", zap.Uint("video_id", videoID))
		updateStatus(db, ctx, videoID, tmodel.StatusDone, 100, "")
		return
	}

	// --- Update status: processing ---
	updateStatus(db, ctx, videoID, tmodel.StatusProcessing, 0, "")

	// --- Get original video presigned URL ---
	var videoURL string
	db.Raw("SELECT video_url FROM videos WHERE id = ?", videoID).Scan(&videoURL)
	if videoURL == "" {
		failTask(db, ctx, videoID, "video not found")
		return
	}
	inputURL, err := storage.GetPresignedURL(ctx, videoURL, 3600)
	if err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("presigned URL: %v", err))
		return
	}

	// --- Create work directory ---
	vidDir := filepath.Join(workDir, strconv.FormatUint(uint64(videoID), 10))
	if err := os.MkdirAll(vidDir, 0755); err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("mkdir: %v", err))
		return
	}
	defer os.RemoveAll(vidDir)

	// Download original
	inputFile := filepath.Join(vidDir, "input"+filepath.Ext(videoURL))
	if filepath.Ext(videoURL) == "" {
		inputFile = filepath.Join(vidDir, "input.mp4")
	}
	if err := downloadFile(ctx, inputURL, inputFile); err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("download: %v", err))
		return
	}

	// --- ffprobe metadata extraction ---
	meta, err := runFFProbe(inputFile)
	if err != nil {
		logger.Warn("worker: ffprobe failed", zap.Uint("video_id", videoID), zap.Error(err))
	} else {
		saveMeta(db, ctx, videoID, meta)
		// Update video duration
		db.WithContext(ctx).Model(&struct{ ID uint }{}).Table("videos").
			Where("id = ?", videoID).
			Update("duration", uint(meta.Duration))
	}

	updateStatus(db, ctx, videoID, tmodel.StatusProcessing, 10, "")

	// --- Determine resolutions based on source ---
	maxW := meta.Width
	maxH := meta.Height
	if maxW == 0 || maxH == 0 {
		maxW = 1920
		maxH = 1080
	}

	var targets []struct {
		Label  string
		Width  int
		Height int
	}
	for _, t := range transcodeTargets {
		if t.Height <= int(maxH) && t.Width <= int(maxW) {
			targets = append(targets, t)
		}
	}
	// If no targets matched (e.g. tiny source), use the source resolution
	if len(targets) == 0 {
		targets = append(targets, struct {
			Label  string
			Width  int
			Height int
		}{Label: fmt.Sprintf("%dp", maxH), Width: int(maxW), Height: int(maxH)})
	}

	totalSteps := len(targets)
	for i, target := range targets {
		qualityDir := filepath.Join(vidDir, target.Label)
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			logger.Error("worker: mkdir quality dir failed", zap.String("dir", qualityDir), zap.Error(err))
			continue
		}

		outputM3U8 := filepath.Join(qualityDir, "index.m3u8")
		segPattern := filepath.Join(qualityDir, "seg_%03d.ts")

		err := runFFmpegHLS(inputFile, outputM3U8, segPattern, target.Width, target.Height)
		if err != nil {
			logger.Error("worker: ffmpeg HLS failed", zap.Uint("video_id", videoID), zap.String("quality", target.Label), zap.Error(err))
			continue
		}

		// Upload all ts segments + m3u8 to MinIO
		entries, _ := os.ReadDir(qualityDir)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			localPath := filepath.Join(qualityDir, entry.Name())
			minioObj := fmt.Sprintf("videos/%d/%s/%s", videoID, target.Label, entry.Name())
			info, _ := entry.Info()
			var fileSize int64
			if info != nil {
				fileSize = info.Size()
			}
			if err := uploadFileToMinio(ctx, localPath, minioObj, fileSize); err != nil {
				logger.Error("worker: upload segment failed", zap.String("obj", minioObj), zap.Error(err))
			}
		}

		// Calculate total size
		var totalSize uint64
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			info, _ := entry.Info()
			if info != nil {
				totalSize += uint64(info.Size())
			}
		}

		// Save quality record
		m3u8ObjName := fmt.Sprintf("videos/%d/%s/index.m3u8", videoID, target.Label)
		saveQuality(db, ctx, videoID, target.Label, m3u8ObjName, totalSize)

		progress := uint8(10 + (i+1)*80/totalSteps)
		updateStatus(db, ctx, videoID, tmodel.StatusProcessing, progress, "")
	}

	// --- Auto cover extraction ---
	var coverURL string
	db.Raw("SELECT cover_url FROM videos WHERE id = ?", videoID).Scan(&coverURL)
	if coverURL == "" {
		coverPath := filepath.Join(vidDir, "cover.jpg")
		coverObjName := fmt.Sprintf("%d/cover_auto_%d.jpg", videoID, videoID)
		// Try with video ID directly since we don't have user ID here
		if err := runFFmpegCover(inputFile, coverPath); err == nil {
			info, _ := os.Stat(coverPath)
			var sz int64
			if info != nil {
				sz = info.Size()
			}
			if err := uploadFileToMinio(ctx, coverPath, coverObjName, sz); err == nil {
				db.WithContext(ctx).Table("videos").
					Where("id = ? AND (cover_url = '' OR cover_url IS NULL)", videoID).
					Update("cover_url", coverObjName)
			}
		} else {
			logger.Warn("worker: cover extraction failed", zap.Uint("video_id", videoID), zap.Error(err))
		}
	}

	updateStatus(db, ctx, videoID, tmodel.StatusDone, 100, "")
	logger.Info("worker: transcode complete", zap.Uint("video_id", videoID))
}

// --- internal helpers ---

func updateStatus(db *gorm.DB, ctx context.Context, videoID uint, status int8, progress uint8, errMsg string) {
	db.WithContext(ctx).Model(&tmodel.TranscodeTask{}).
		Where("video_id = ?", videoID).
		Updates(map[string]interface{}{
			"status":    status,
			"progress":  progress,
			"error_msg": errMsg,
		})
}

func failTask(db *gorm.DB, ctx context.Context, videoID uint, errMsg string) {
	logger.Error("worker: transcode failed", zap.Uint("video_id", videoID), zap.String("error", errMsg))
	updateStatus(db, ctx, videoID, tmodel.StatusFailed, 0, errMsg)
}

func saveMeta(db *gorm.DB, ctx context.Context, videoID uint, meta *mmodel.VideoMeta) {
	var count int64
	db.WithContext(ctx).Model(&mmodel.VideoMeta{}).Where("video_id = ?", videoID).Count(&count)
	if count > 0 {
		db.WithContext(ctx).Model(&mmodel.VideoMeta{}).Where("video_id = ?", videoID).Updates(map[string]interface{}{
			"duration": meta.Duration,
			"width":    meta.Width,
			"height":   meta.Height,
			"codec":    meta.Codec,
			"bitrate":  meta.Bitrate,
		})
		return
	}
	meta.VideoID = videoID
	db.WithContext(ctx).Create(meta)
}

func saveQuality(db *gorm.DB, ctx context.Context, videoID uint, quality string, objectName string, fileSize uint64) {
	var count int64
	db.WithContext(ctx).Model(&qmodel.VideoQuality{}).
		Where("video_id = ? AND quality = ?", videoID, quality).
		Count(&count)
	if count > 0 {
		db.WithContext(ctx).Model(&qmodel.VideoQuality{}).
			Where("video_id = ? AND quality = ?", videoID, quality).
			Updates(map[string]interface{}{
				"object_name": objectName,
				"file_size":   fileSize,
			})
		return
	}
	db.WithContext(ctx).Create(&qmodel.VideoQuality{
		VideoID:    videoID,
		Quality:    quality,
		ObjectName: objectName,
		FileSize:   fileSize,
	})
}

func runFFProbe(inputFile string) (*mmodel.VideoMeta, error) {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputFile,
	)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe: %w", err)
	}

	meta := &mmodel.VideoMeta{}

	// Parse JSON output simplistically (avoid heavy json unmarshaling into complex ffprobe structures)
	raw := string(output)

	// duration
	if idx := strings.Index(raw, `"duration"`); idx >= 0 {
		rest := raw[idx+len(`"duration"`):]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			after := strings.TrimSpace(rest[colon+1:])
			end := strings.IndexAny(after, ",\n\r")
			if end < 0 {
				end = len(after)
			}
			val := strings.Trim(after[:end], `" `)
			if d, err := strconv.ParseFloat(val, 64); err == nil {
				meta.Duration = d
			}
		}
	}

	// width
	if idx := strings.Index(raw, `"width"`); idx >= 0 {
		rest := raw[idx+len(`"width"`):]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			after := strings.TrimSpace(rest[colon+1:])
			end := strings.IndexAny(after, ",\n\r")
			if end < 0 {
				end = len(after)
			}
			val := strings.Trim(after[:end], `" `)
			if w, err := strconv.ParseUint(val, 10, 64); err == nil {
				meta.Width = uint(w)
			}
		}
	}

	// height
	if idx := strings.Index(raw, `"height"`); idx >= 0 {
		rest := raw[idx+len(`"height"`):]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			after := strings.TrimSpace(rest[colon+1:])
			end := strings.IndexAny(after, ",\n\r")
			if end < 0 {
				end = len(after)
			}
			val := strings.Trim(after[:end], `" `)
			if h, err := strconv.ParseUint(val, 10, 64); err == nil {
				meta.Height = uint(h)
			}
		}
	}

	// codec_name
	if idx := strings.Index(raw, `"codec_name"`); idx >= 0 {
		rest := raw[idx+len(`"codec_name"`):]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			after := strings.TrimSpace(rest[colon+1:])
			end := strings.IndexAny(after, ",\n\r")
			if end < 0 {
				end = len(after)
			}
			val := strings.Trim(after[:end], `" `)
			meta.Codec = val
		}
	}

	// bit_rate
	if idx := strings.Index(raw, `"bit_rate"`); idx >= 0 {
		rest := raw[idx+len(`"bit_rate"`):]
		if colon := strings.Index(rest, ":"); colon >= 0 {
			after := strings.TrimSpace(rest[colon+1:])
			end := strings.IndexAny(after, ",\n\r")
			if end < 0 {
				end = len(after)
			}
			val := strings.Trim(after[:end], `" `)
			if b, err := strconv.ParseUint(val, 10, 64); err == nil {
				meta.Bitrate = uint(b / 1000) // Convert to kbps
			}
		}
	}

	return meta, nil
}

func runFFmpegHLS(inputFile, outputM3U8, segPattern string, width, height int) error {
	scaleFilter := fmt.Sprintf("scale=%d:%d", width, height)
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vf", scaleFilter,
		"-c:v", "libx264",
		"-preset", "fast",
		"-crf", "23",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", segPattern,
		outputM3U8,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg HLS: %w (output: %s)", err, string(out))
	}
	return nil
}

func runFFmpegCover(inputFile, outputFile string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-ss", "1",
		"-vframes", "1",
		"-q:v", "2",
		outputFile,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg cover: %w (output: %s)", err, string(out))
	}
	return nil
}

func downloadFile(ctx context.Context, url, destPath string) error {
	// Use ffmpeg's built-in protocol to download, or use a simple approach:
	// For simplicity, pass the URL directly to ffmpeg; no need to download separately
	// since ffmpeg can read URLs directly. But for the local file approach:
	cmd := exec.Command("ffmpeg",
		"-i", url,
		"-c", "copy",
		"-y",
		destPath,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("download via ffmpeg: %w (output: %s)", err, string(out))
	}
	return nil
}

func uploadFileToMinio(ctx context.Context, localPath, objectName string, fileSize int64) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	contentType := "application/octet-stream"
	if strings.HasSuffix(objectName, ".m3u8") {
		contentType = "application/vnd.apple.mpegurl"
	} else if strings.HasSuffix(objectName, ".ts") {
		contentType = "video/mp2t"
	} else if strings.HasSuffix(objectName, ".jpg") || strings.HasSuffix(objectName, ".jpeg") {
		contentType = "image/jpeg"
	}

	return storage.UploadVideo(ctx, objectName, f, fileSize, contentType)
}
