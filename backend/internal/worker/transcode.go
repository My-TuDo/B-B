// Package worker 提供视频转码的后台任务处理功能。
// 负责从对象存储下载原始视频、通过 ffprobe 提取元信息、
// 使用 ffmpeg 生成 HLS 多码率分片、上传转码产物到 MinIO，
// 并通过 ProgressBroker 实时推送进度。
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

// workDir 是转码过程中本地临时文件的存放目录。
// workDir 是转码过程中本地临时文件的存放目录。
const workDir = "/tmp/bb-transcode"

// transcodeTargets 定义了需要转码的目标分辨率档位。
// 每个档位包含标签名（如 "720p"）和对应的宽高。
// transcodeTargets 定义了需要转码的目标分辨率档位。
// 每个档位包含标签名（如"720p"）和对应的宽高。
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

// ProcessVideo 是视频转码的主入口函数。
// 它从数据库读取视频记录，依次完成下载、元信息提取、多码率 HLS 编码、
// 分片上传和自动封面提取，并通过 broker 实时广播进度。
func ProcessVideo(videoID uint, db *gorm.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()
	broker := GetBroker()
	logger.Info("worker: starting transcode", zap.Uint("video_id", videoID))

	// --- 检查 ffmpeg / ffprobe 可用性 ---
	// 检查 ffmpeg 和 ffprobe 是否可用，不可用则直接标记失败
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

	// 将任务状态更新为"处理中"
	updateStatus(db, ctx, videoID, tmodel.StatusProcessing, 0, "")

	// --- 获取原始视频预签名 URL ---
	// 从数据库获取原始视频的存储路径，并生成预签名下载 URL
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

	// 为本次转码创建临时工作目录
	vidDir := filepath.Join(workDir, strconv.FormatUint(uint64(videoID), 10))
	if err := os.MkdirAll(vidDir, 0755); err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("mkdir: %v", err))
		return
	}
	defer os.RemoveAll(vidDir) // 转码完成后清理临时目录

	// 下载原始视频到本地
	// 从预签名 URL 构建本地文件名（去除查询参数后取扩展名）
	cleanPath := videoURL
	if idx := strings.Index(cleanPath, "?"); idx >= 0 {
		cleanPath = cleanPath[:idx]
	}
	ext := filepath.Ext(cleanPath)
	if ext == "" {
		ext = ".mp4" // 默认使用 .mp4 扩展名
	}
	inputFile := filepath.Join(vidDir, "input"+ext)

	// 广播下载进度
	// 广播下载进度（2%）
	broker.Publish(ProgressUpdate{VideoID: videoID, Status: tmodel.StatusProcessing, Progress: 2, Quality: "download"})
	if err := downloadFile(ctx, inputURL, inputFile); err != nil {
		failTask(db, ctx, videoID, fmt.Sprintf("download: %v", err))
		return
	}

	// --- ffprobe 元信息提取 ---
	// 使用 ffprobe 提取视频元信息（分辨率、编码、码率、时长）
	meta, err := runFFProbe(inputFile)
	if err != nil {
		logger.Warn("worker: ffprobe failed", zap.Uint("video_id", videoID), zap.Error(err))
	} else {
		saveMeta(db, ctx, videoID, meta)
		// 同时更新 videos 表中的 duration 字段
		db.WithContext(ctx).Model(&struct{ ID uint }{}).Table("videos").
			Where("id = ?", videoID).
			Update("duration", uint(meta.Duration))
	}

	// 广播 ffprobe 阶段完成（10%）
	publishAndPersist(db, ctx, videoID, tmodel.StatusProcessing, 10, "")

	// 根据源视频分辨率确定有效的转码目标（不放大）
	maxW := meta.Width
	maxH := meta.Height
	if maxW == 0 || maxH == 0 {
		maxW, maxH = 1920, 1080 // 默认兜底
	}

	var targets []struct {
		Label  string
		Width  int
		Height int
	}
	for _, t := range transcodeTargets {
		// 只转码分辨率不高于源视频的档位
		if t.Height <= int(maxH) && t.Width <= int(maxW) {
			targets = append(targets, t)
		}
	}
	// 至少保留一个档位（使用源视频原始分辨率）
	if len(targets) == 0 {
		targets = append(targets, struct {
			Label  string
			Width  int
			Height int
		}{Label: fmt.Sprintf("%dp", maxH), Width: int(maxW), Height: int(maxH)})
	}

	// 依次对每个分辨率档位执行 HLS 转码
	totalSteps := len(targets)
	// progress ranges: download→2%, ffprobe→10%, each encode→80%/totalSteps
	for i, target := range targets {
		// 为当前档位创建子目录
		qualityDir := filepath.Join(vidDir, target.Label)
		if err := os.MkdirAll(qualityDir, 0755); err != nil {
			logger.Error("worker: mkdir quality dir failed", zap.String("dir", qualityDir), zap.Error(err))
			continue
		}

		outputM3U8 := filepath.Join(qualityDir, "index.m3u8")
		segPattern := filepath.Join(qualityDir, "seg_%03d.ts")

		// ffmpeg 实时进度回调
		// ffmpeg 实时进度回调：将 ffmpeg 内部的编码进度映射到总体进度区间
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

		// 执行 HLS 转码
		err := runFFmpegHLS(ctx, inputFile, outputM3U8, segPattern, target.Width, target.Height, meta.Duration, onFfmpegProgress)
		if err != nil {
			logger.Error("worker: ffmpeg HLS failed", zap.Uint("video_id", videoID), zap.String("quality", target.Label), zap.Error(err))
			continue
		}

		// Upload segments
		// 将该档位的所有分片上传到 MinIO
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

		// 保存该档位的质量记录（m3u8 对象名及总大小）
		m3u8ObjName := fmt.Sprintf("videos/%d/%s/index.m3u8", videoID, target.Label)
		saveQuality(db, ctx, videoID, target.Label, m3u8ObjName, totalSize)

		// End-of-quality progress
		// 当前档位编码完成，广播子阶段进度
		qualProgress := uint8(10 + (i+1)*80/totalSteps)
		publishAndPersist(db, ctx, videoID, tmodel.StatusProcessing, qualProgress, "")
	}

	// --- 自动封面提取 ---
	// 自动提取封面：如果视频尚未设置封面，从第 1 秒截取一帧作为封面
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
				// 只在 cover_url 为空时更新，避免覆盖用户手动上传的封面
				db.WithContext(ctx).Table("videos").
					Where("id = ? AND (cover_url = '' OR cover_url IS NULL)", videoID).
					Update("cover_url", coverObjName)
			}
		} else {
			logger.Warn("worker: cover extraction failed", zap.Uint("video_id", videoID), zap.Error(err))
		}
	}

	// 转码完成，广播 100% 进度
	publishAndPersist(db, ctx, videoID, tmodel.StatusDone, 100, "")
	broker.Publish(ProgressUpdate{VideoID: videoID, Status: tmodel.StatusDone, Progress: 100})
	logger.Info("worker: transcode complete", zap.Uint("video_id", videoID))
}

// publishAndPersist 将转码状态同时写入数据库并通过 SSE broker 广播。
// 这是保证数据库与实时推送一致性的便捷方法。
func publishAndPersist(db *gorm.DB, ctx context.Context, videoID uint, status int8, progress uint8, errMsg string) {
	updateStatus(db, ctx, videoID, status, progress, errMsg)
	GetBroker().Publish(ProgressUpdate{
		VideoID:  videoID,
		Status:   status,
		Progress: progress,
		ErrorMsg: errMsg,
	})
}

// updateStatus 更新或创建 transcode_tasks 表中的任务状态记录。
// 如果对应 video_id 的记录不存在则创建，存在则更新状态、进度和错误信息。
func updateStatus(db *gorm.DB, ctx context.Context, videoID uint, status int8, progress uint8, errMsg string) {
	existing := &tmodel.TranscodeTask{}
	res := db.WithContext(ctx).Where("video_id = ?", videoID).First(existing)
	if res.Error != nil {
		// 记录不存在，创建新任务
		db.WithContext(ctx).Create(&tmodel.TranscodeTask{
			VideoID:  videoID,
			Status:   status,
			Progress: progress,
			ErrorMsg: errMsg,
		})
		return
	}
	// 记录已存在，更新字段
	db.WithContext(ctx).Model(&tmodel.TranscodeTask{}).
		Where("video_id = ?", videoID).
		Updates(map[string]interface{}{
			"status":    status,
			"progress":  progress,
			"error_msg": errMsg,
		})
}

// failTask 标记转码任务为失败，记录错误信息并广播失败状态。
func failTask(db *gorm.DB, ctx context.Context, videoID uint, errMsg string) {
	logger.Error("worker: transcode failed", zap.Uint("video_id", videoID), zap.String("error", errMsg))
	updateStatus(db, ctx, videoID, tmodel.StatusFailed, 0, errMsg)
	GetBroker().Publish(ProgressUpdate{VideoID: videoID, Status: tmodel.StatusFailed, Progress: 0, ErrorMsg: errMsg})
}

// saveMeta 持久化视频元信息到 video_meta 表。
// 如果已有记录则更新，否则创建新记录。
func saveMeta(db *gorm.DB, ctx context.Context, videoID uint, meta *mmodel.VideoMeta) {
	var count int64
	db.WithContext(ctx).Model(&mmodel.VideoMeta{}).Where("video_id = ?", videoID).Count(&count)
	if count > 0 {
		// 已有记录：更新
		db.WithContext(ctx).Model(&mmodel.VideoMeta{}).Where("video_id = ?", videoID).Updates(map[string]interface{}{
			"duration": meta.Duration,
			"width":    meta.Width,
			"height":   meta.Height,
			"codec":    meta.Codec,
			"bitrate":  meta.Bitrate,
		})
		return
	}
	// 新记录：创建
	meta.VideoID = videoID
	db.WithContext(ctx).Create(meta)
}

// saveQuality 持久化单个清晰度档位的转码产物信息到 video_qualities 表。
// 包括 object_name（MinIO 中 m3u8 的路径）和 file_size（该档位所有分片的总大小）。
func saveQuality(db *gorm.DB, ctx context.Context, videoID uint, quality string, objectName string, fileSize uint64) {
	var count int64
	db.WithContext(ctx).Model(&qmodel.VideoQuality{}).
		Where("video_id = ? AND quality = ?", videoID, quality).
		Count(&count)
	if count > 0 {
		// 已有记录：更新
		db.WithContext(ctx).Model(&qmodel.VideoQuality{}).
			Where("video_id = ? AND quality = ?", videoID, quality).
			Updates(map[string]interface{}{
				"object_name": objectName,
				"file_size":   fileSize,
			})
		return
	}
	// 新记录：创建
	db.WithContext(ctx).Create(&qmodel.VideoQuality{
		VideoID:    videoID,
		Quality:    quality,
		ObjectName: objectName,
		FileSize:   fileSize,
	})
}

// ffprobeOutput types
// 以下结构体用于解析 ffprobe 的 JSON 输出。

// ffprobeOutput 是 ffprobe JSON 输出的顶层结构。
type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

// ffprobeStream 表示 ffprobe 返回的一条流信息（视频/音频/字幕等）。
type ffprobeStream struct {
	CodecType string `json:"codec_type"`
	CodecName string `json:"codec_name"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
}

// ffprobeFormat 表示 ffprobe 返回的容器格式信息。
type ffprobeFormat struct {
	Duration string `json:"duration"`
	BitRate  string `json:"bit_rate"`
}

// runFFProbe 调用 ffprobe 提取视频文件的元信息，
// 返回包含时长、分辨率、编码和码率的 VideoMeta 结构。
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

	// 解析 ffprobe 的 JSON 输出
	var probe ffprobeOutput
	if err := json.Unmarshal(output, &probe); err != nil {
		return nil, fmt.Errorf("ffprobe parse: %w", err)
	}

	meta := &mmodel.VideoMeta{}
	// 解析时长（秒）
	if d, err := strconv.ParseFloat(probe.Format.Duration, 64); err == nil {
		meta.Duration = d
	}
	// 解析码率（kbps）
	if b, err := strconv.ParseUint(probe.Format.BitRate, 10, 64); err == nil {
		meta.Bitrate = uint(b / 1000)
	}
	// 从视频流中提取分辨率与编码
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

// runFFmpegHLS 调用 ffmpeg 将输入视频转码为 HLS 分片。
// 通过 -progress pipe:1 实时解析编码进度并回调 onProgress。
func runFFmpegHLS(ctx context.Context, inputFile, outputM3U8, segPattern string, width, height int, duration float64, onProgress func(pct float64)) error {
	scaleFilter := fmt.Sprintf("scale=%d:%d", width, height)

	// 构建 ffmpeg 命令：H.264 视频 + AAC 音频，输出 HLS
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-progress", "pipe:1", // 将进度信息输出到 stdout
		"-nostats",
		"-i", inputFile,
		"-vf", scaleFilter, // 缩放滤镜
		"-c:v", "libx264",
		"-c:a", "aac",
		"-b:a", "128k",
		"-preset", "fast",
		"-crf", "23",
		"-hls_time", "10",             // 每个 TS 分片时长 10 秒
		"-hls_list_size", "0",         // m3u8 播放列表包含所有分片
		"-hls_segment_filename", segPattern,
		outputM3U8,
	)

	// 获取 stdout 管道用于解析进度
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stdout pipe: %w", err)
	}
	// 获取 stderr 管道防止阻塞
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("ffmpeg stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	// 从 stdout 解析 ffmpeg 进度：每帧输出 key=value 块，其中 out_time 为当前编码时间戳
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

	// 消费 stderr 防止 ffmpeg 因管道满而阻塞
	go func() {
		buf := make([]byte, 4096)
		for {
			_, err := stderr.Read(buf)
			if err != nil {
				break
			}
		}
	}()

	// 等待 ffmpeg 进程结束
	err = cmd.Wait()
	<-scanDone // 确保进度扫描协程已退出

	if err != nil {
		return fmt.Errorf("ffmpeg HLS: %w", err)
	}
	return nil
}

// parseTimeToSeconds 将 ffmpeg 的 "HH:MM:SS.microseconds" 格式时间戳转换为浮点秒数。
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

// runFFmpegCover 从输入视频的第 1 秒截取一帧作为封面图片。
func runFFmpegCover(ctx context.Context, inputFile, outputFile string) error {
	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", inputFile,
		"-ss", "1",        // 定位到第 1 秒
		"-vframes", "1",   // 只取一帧
		"-q:v", "2",       // 高质量输出
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

// downloadFile 通过 HTTP GET 将远程文件下载到本地路径。
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

// uploadFileToMinio 将本地文件上传到 MinIO 对象存储。
// 根据文件扩展名自动设置 Content-Type。
func uploadFileToMinio(ctx context.Context, localPath, objectName string, fileSize int64) error {
	f, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// 根据文件扩展名设置合适的 MIME 类型
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
