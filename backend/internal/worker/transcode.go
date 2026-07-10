package worker

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	mmodel "github.com/My-TuDo/B-B/backend/internal/model/meta"
	qmodel "github.com/My-TuDo/B-B/backend/internal/model/quality"
	tmodel "github.com/My-TuDo/B-B/backend/internal/model/transcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/storage"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const workDir = "/tmp/bb-transcode"

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

func ProcessVideo(videoID uint, db *gorm.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	broker := GetBroker()
	logger.Info("worker: starting transcode", zap.Uint("video_id", videoID))

	// --- Check ffmpeg / ffprobe availability ---
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		logger.Error("worker: ffmpeg not found, skipping transcode", zap.Uint("video_id", videoID))
		updateStatus(db, ctx, videoID, tmodel.StatusFailed, 0, "ffmpeg not available on server")
		return
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		logger.Error("worker: ffprobe not found, skipping transcode", zap.Uint("video_id", videoID))
		updateStatus(db, ctx, videoID, tmodel.StatusFailed, 0, "ffprobe not available on server")
		return
	}

	updateStatus(db, ctx, videoID, tmodel.StatusProcessing, 0, "")

	// --- Get original video presigned URL ---
	var videoURL string
	db.Raw("SELECT video_url FROM videos WHERE id = ?", videoID).Scan(&videoURL)
	if videoURL == "" {
		failTask(db, ctx, videoID, "video not found")
		return
	}
	inputURL, err := storage.GetPresignedURL(ctx, videoURL, time.Hour)
	if err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("presigned URL: %v", err))
		return
	}

	vidDir := filepath.Join(workDir, strconv.FormatUint(uint64(videoID), 10))
	if err := os.MkdirAll(vidDir, 0755); err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("mkdir: %v", err))
		return
	}
	defer os.RemoveAll(vidDir)

	// Download original
	cleanPath := videoURL
	if idx := strings.Index(cleanPath, "?"); idx >= 0 {
		cleanPath = cleanPath[:idx]
	}
	ext := filepath.Ext(cleanPath)
	if ext == "" {
		ext = ".mp4"
	}
	inputFile := filepath.Join(vidDir, "input"+ext)

	// Broadcast download progress
	broker.Publish(ProgressUpdate{VideoID: videoID, Status: tmodel.StatusProcessing, Progress: 2, Quality: "download"})
	if err := downloadFile(ctx, inputURL, inputFile); err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("download: %v", err))
		return
	}

	// --- ffprobe ---
	meta, err := runFFProbe(inputFile)
	if err != nil {
		logger.Warn("worker: ffprobe failed", zap.Uint("video_id", videoID), zap.Error(err))
	} else {
		saveMeta(db, ctx, videoID, meta)
		db.WithContext(ctx).Model(&struct{ ID uint }{}).Table("videos").
			Where("id = ?", videoID).
			Update("duration", uint(meta.Duration))
	}

	publishAndPersist(db, ctx, videoID, tmodel.StatusProcessing, 10, "")

	maxW := meta.Width
	maxH := meta.Height
	if maxW == 0 || maxH == 0 {
		maxW, maxH = 1920, 1080
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
	if len(targets) == 0 {
		targets = append(targets, struct {
			Label  string
			Width  int
			Height int
		}{Label: fmt.Sprintf("%dp", maxH), Width: int(maxW), Height: int(maxH)})
	}

	totalSteps := len(targets)
	// progress ranges: download→2%, ffprobe→10%, each encode→80%/totalSteps
	for i, target := range targets {
		qualityDir := filepath.Join(vidDir, target.Label)
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			logger.Error("worker: mkdir quality dir failed", zap.String("dir", qualityDir), zap.Error(err))
			continue
		}

		outputM3U8 := filepath.Join(qualityDir, "index.m3u8")
		segPattern := filepath.Join(qualityDir, "seg_%03d.ts")

		// Real-time progress callback for ffmpeg
		onFfmpegProgress := func(ffmpegPct float64) {
			// Encode base: 10. Each quality gets 80%/totalSteps width.
			// Real ffmpeg progress fills within that slot.
			base := float64(10 + i*80/totalSteps)
			slotWidth := float64(80) / float64(totalSteps)
			overall := uint8(base + ffmpegPct/100.0*slotWidth)
			if overall > 99 {
				overall = 99
			}
			broker.Publish(ProgressUpdate{
				VideoID:  videoID,
				Status:   tmodel.StatusProcessing,
				Progress: overall,
				Quality:  target.Label,
			})
		}

		err := runFFmpegHLS(ctx, inputFile, outputM3U8, segPattern, target.Width, target.Height, meta.Duration, onFfmpegProgress)
		if err != nil {
			logger.Error("worker: ffmpeg HLS failed", zap.Uint("video_id", videoID), zap.String("quality", target.Label), zap.Error(err))
			continue
		}

		// Upload segments
		entries, _ := os.ReadDir(qualityDir)
		var totalSize uint64
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
			if fileSize > 0 {
				totalSize += uint64(fileSize)
			}
		}

		m3u8ObjName := fmt.Sprintf("videos/%d/%s/index.m3u8", videoID, target.Label)
		saveQuality(db, ctx, videoID, target.Label, m3u8ObjName, totalSize)

		// End-of-quality progress
		qualProgress := uint8(10 + (i+1)*80/totalSteps)
		publishAndPersist(db, ctx, videoID, tmodel.StatusProcessing, qualProgress, "")
	}

	// --- Auto cover ---
	var coverURL string
	db.Raw("SELECT cover_url FROM videos WHERE id = ?", videoID).Scan(&coverURL)
	if coverURL == "" {
		coverPath := filepath.Join(vidDir, "cover.jpg")
		coverObjName := fmt.Sprintf("%d/cover_auto_%d.jpg", videoID, videoID)
		if err := runFFmpegCover(ctx, inputFile, coverPath); err == nil {
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

	publishAndPersist(db, ctx, videoID, tmodel.StatusDone, 100, "")
	broker.Publish(ProgressUpdate{VideoID: videoID, Status: tmodel.StatusDone, Progress: 100})
	logger.Info("worker: transcode complete", zap.Uint("video_id", videoID))
}

// publishAndPersist writes to DB and publishes to SSE broker in one call.
func publishAndPersist(db *gorm.DB, ctx context.Context, videoID uint, status int8, progress uint8, errMsg string) {
	updateStatus(db, ctx, videoID, status, progress, errMsg)
	GetBroker().Publish(ProgressUpdate{
		VideoID:  videoID,
		Status:   status,
		Progress: progress,
		ErrorMsg: errMsg,
	})
}

func updateStatus(db *gorm.DB, ctx context.Context, videoID uint, status int8, progress uint8, errMsg string) {
	existing := &tmodel.TranscodeTask{}
	res := db.WithContext(ctx).Where("video_id = ?", videoID).First(existing)
	if res.Error != nil {
		db.WithContext(ctx).Create(&tmodel.TranscodeTask{
			VideoID:  videoID,
			Status:   status,
			Progress: progress,
			ErrorMsg: errMsg,
		})
		return
	}
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
	GetBroker().Publish(ProgressUpdate{VideoID: videoID, Status: tmodel.StatusFailed, Progress: 0, ErrorMsg: errMsg})
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

// ffprobeOutput types
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

type ffprobeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

type ffprobeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
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

	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("ffprobe parse: %w", err)
	}

	meta := &mmodel.VideoMeta{}
	if d, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		meta.Duration = d
	}
	if b, err := strconv.ParseUint(probe.Format.BitRate, 10, 64); err == nil {
		meta.Bitrate = uint(b / 1000)
	}
	for _, s := range probe.Streams {
		if s.CodecType == "video" {
			meta.Width = uint(s.Width)
			meta.Height = uint(s.Height)
			meta.Codec = s.CodecName
			break
		}
	}
	return meta, nil
}

// runFFmpegHLS runs ffmpeg with -progress pipe:1 to emit real-time progress on stdout.
func runFFmpegHLS(ctx context.Context, inputFile, outputM3U8, segPattern string, width, height int, duration float64, onProgress func(pct float64)) error {
	scaleFilter := fmt.Sprintf("scale=%d:%d", width, height)

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-progress", "pipe:1",
		"-nostats",
		"-i", inputFile,
		"-vf", scaleFilter,
		"-c:v", "libx264",
		"-c:a", "aac",
		"-b:a", "128k",
		"-preset", "fast",
		"-crf", "23",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", segPattern,
		outputM3U8,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	// Parse progress from stdout: each frame emits key=value blocks
	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "out_time=") {
				ts := strings.TrimPrefix(line, "out_time=")
				outTime := parseTimeToSeconds(ts)
				if duration > 0 && outTime > 0 {
					pct := (outTime / duration) * 100.0
					if pct > 100 {
						pct = 100
					}
					onProgress(pct)
				}
			}
		}
	}()

	// Drain stderr so ffmpeg doesn't block
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := stderr.Read(buf)
			if err != nil {
				break
			}
		}
	}()

	err = cmd.Wait()
	<-scanDone

	if err != nil {
		return fmt.Errorf("ffmpeg HLS: %w", err)
	}
	return nil
}

// parseTimeToSeconds converts "HH:MM:SS.microseconds" to float seconds.
func parseTimeToSeconds(ts string) float64 {
	// format: "00:01:23.456789" or "00:01:23.456"
	hmsParts := strings.SplitN(ts, ".", 2)
	hms := hmsParts[0]
	parts := strings.Split(hms, ":")
	if len(parts) != 3 {
		return 0
	}
	h, _ := strconv.Atoi(parts[0])
	m, _ := strconv.Atoi(parts[1])
	s, _ := strconv.Atoi(parts[2])
	return float64(h*3600+m*60) + float64(s)
}

func runFFmpegCover(ctx context.Context, inputFile, outputFile string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputFile,
		"-ss", "1",
		"-vframes", "1",
		"-q:v", "2",
		outputFile,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(out)
		if len(outputStr) > 500 {
			outputStr = outputStr[:500] + "..."
		}
		logger.Debug("ffmpeg cover output", zap.String("output", outputStr))
		return fmt.Errorf("ffmpeg cover: %w", err)
	}
	return nil
}

func downloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("downloadFile request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("downloadFile: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloadFile: HTTP %d", resp.StatusCode)
	}
	f, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("downloadFile create file: %w", err)
	}
	defer f.Close()
	_, err = io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("downloadFile write: %w", err)
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
