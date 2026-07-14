// Package video 提供视频相关的 HTTP 处理器，包括视频上传（SSE 进度推送）、
// 视频信息获取、播放地址获取、视频更新/删除、列表查询、热门和排行榜等功能。
package video

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	videomodel "github.com/My-TuDo/B-B/backend/internal/model/video"
	videoservice "github.com/My-TuDo/B-B/backend/internal/service/video"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 是视频模块的 HTTP 处理器，持有视频服务实例。
type Handler struct {
	svc *videoservice.Service
}

// NewHandler 创建视频处理器实例。
func NewHandler(svc *videoservice.Service) *Handler {
	return &Handler{svc: svc}
}

// Upload 处理视频上传请求（需认证），支持 SSE（Server-Sent Events）实时进度推送。
// 协程：视频文件 + 可选封面图，校验 MIME 类型与魔数后上传处理。
func (h *Handler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// 从 multipart 表单中读取视频文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请选择视频文件")
		return
	}
	defer file.Close()

	// 校验 MIME 类型
	contentType := header.Header.Get("Content-Type")
	if !isValidVideoMIME(contentType) {
		response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, errcode.Message(errcode.InvalidFileType))
		return
	}

	// 读取文件前 512 字节魔数校验，防止伪造 MIME
	magicBuf := make([]byte, 512)
	n, _ := file.Read(magicBuf)
	if !isValidVideoMagic(magicBuf[:n]) {
		response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, errcode.Message(errcode.InvalidFileType))
		return
	}

	// 读取表单字段
	title := c.PostForm("title")
	description := c.PostForm("description")
	categoryIDStr := c.PostForm("category_id")

	// 标题校验
	if title == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频标题不能为空")
		return
	}
	if len(title) > 100 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "标题不能超过100个字符")
		return
	}

	// 解析分类 ID（可选）
	var categoryID uint
	if categoryIDStr != "" {
		parsed, err := strconv.ParseUint(categoryIDStr, 10, 64)
		if err == nil {
			categoryID = uint(parsed)
		}
	}

	// 获取封面图文件（可选）
	var coverFile multipart.File
	var coverHeader *multipart.FileHeader
	coverFile, coverHeader, err = c.Request.FormFile("cover")
	if err == nil {
		defer coverFile.Close()

		// 校验封面 MIME 类型
		coverContentType := coverHeader.Header.Get("Content-Type")
		if !isValidCoverMIME(coverContentType) {
			response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, "封面图片格式不支持，仅支持 JPEG/PNG/WebP/GIF")
			return
		}

		// 校验封面大小（最大 5MB）
		const maxCoverSize = 5 * 1024 * 1024
		if coverHeader.Size > maxCoverSize {
			response.Error(c, http.StatusBadRequest, errcode.FileTooLarge, "封面图片大小不能超过 5MB")
			return
		}
	}

	// 构造组合读取器：先回放已读取的魔数字节，再继续读文件剩余部分
	combinedReader := &combinedReader{
		prefix: magicBuf[:n],
		file:   file,
	}

	// SSE — 设置响应头，启用服务器推送事件
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, http.StatusInternalServerError, errcode.Internal, "SSE not supported")
		return
	}

	// 进度回调：将上传进度通过 SSE 推送给客户端
	progressFn := func(uploaded, total int64) {
		msg := videomodel.UploadSSEMessage{
			Uploaded: uploaded,
			Total:    total,
		}
		data, _ := json.Marshal(msg)
		fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
		flusher.Flush()
	}

	resp, err := h.svc.UploadVideo(c.Request.Context(), userID, combinedReader, header.Filename, header.Size, contentType, title, description, categoryID, progressFn, coverFile, coverHeader)
	if err != nil {
		// 如果 resp 不为空，说明视频已保存但封面上传失败，仍通知客户端
		if resp != nil {
			msg := videomodel.UploadSSEMessage{Error: "封面上传失败，视频已保存"}
			data, _ := json.Marshal(msg)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			flusher.Flush()
			// 继续执行，发送 complete 事件
		} else {
			// 视频保存失败，区分业务错误与系统错误
			var svcErr *videoservice.Error
			if errors.As(err, &svcErr) {
				msg := videomodel.UploadSSEMessage{Error: svcErr.Msg}
				data, _ := json.Marshal(msg)
				fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
				flusher.Flush()
				return
			}
			logger.Error("upload video failed", zap.Error(err))
			msg := videomodel.UploadSSEMessage{Error: "上传失败"}
			data, _ := json.Marshal(msg)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			flusher.Flush()
			return
		}
	}

	// 发送上传完成事件（包含视频响应数据）
	completeData, _ := json.Marshal(resp)
	fmt.Fprintf(c.Writer, "event: complete\ndata: %s\n\n", string(completeData))
	flusher.Flush()
}

// GetVideo 获取指定视频的详细信息（公开接口，可选传入 viewerID 用于判断互动状态）。
func (h *Handler) GetVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 获取当前浏览者 ID（未登录则为 0）
	viewerID := middleware.GetUserID(c)

	resp, err := h.svc.GetVideo(c.Request.Context(), uint(id), viewerID)
	if err != nil {
		var svcErr *videoservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get video failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// GetPlayURL 获取视频的播放地址（签名 URL，公开接口）。
func (h *Handler) GetPlayURL(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 获取当前浏览者 ID，用于播放权限校验
	viewerID := middleware.GetUserID(c)

	url, err := h.svc.GetPlayURL(c.Request.Context(), uint(id), viewerID)
	if err != nil {
		var svcErr *videoservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get play url failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, gin.H{"play_url": url})
}

// UpdateVideo 更新视频信息（需认证，仅允许视频作者修改）。
func (h *Handler) UpdateVideo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req videomodel.UpdateVideoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	resp, err := h.svc.UpdateVideo(c.Request.Context(), userID, uint(id), &req)
	if err != nil {
		var svcErr *videoservice.Error
		if errors.As(err, &svcErr) {
			// 权限不足返回 403，不存在返回 404
			status := http.StatusNotFound
			if svcErr.Code == errcode.Forbidden {
				status = http.StatusForbidden
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("update video failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// DeleteVideo 删除视频（需认证，仅允许视频作者删除）。
func (h *Handler) DeleteVideo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	if err := h.svc.DeleteVideo(c.Request.Context(), userID, uint(id)); err != nil {
		var svcErr *videoservice.Error
		if errors.As(err, &svcErr) {
			status := http.StatusNotFound
			if svcErr.Code == errcode.Forbidden {
				status = http.StatusForbidden
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("delete video failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

// ListVideos 获取视频列表（公开接口，支持分页和分类筛选）。
func (h *Handler) ListVideos(c *gin.Context) {
	// 分页参数（带默认值）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))
	categoryIDStr := c.DefaultQuery("category_id", "0")
	categoryID, _ := strconv.ParseUint(categoryIDStr, 10, 64)

	resp, err := h.svc.ListVideos(c.Request.Context(), page, pageSize, uint(categoryID))
	if err != nil {
		logger.Error("list videos failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// ListUserVideos 获取指定用户发布的视频列表（公开接口，支持状态过滤和分页）。
func (h *Handler) ListUserVideos(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	// 可选的状态过滤（如仅查看已审核视频）
	var statusFilter *int8
	statusStr := c.DefaultQuery("status", "")
	if statusStr != "" {
		s, err := strconv.ParseInt(statusStr, 10, 8)
		if err == nil {
			val := int8(s)
			statusFilter = &val
		}
	}

	resp, err := h.svc.ListUserVideos(c.Request.Context(), uint(userID), statusFilter, page, pageSize)
	if err != nil {
		logger.Error("list user videos failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// HotVideos 获取热门视频列表（公开接口，基于热度算法排序）。
func (h *Handler) HotVideos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.HotVideos(c.Request.Context(), page, pageSize)
	if err != nil {
		logger.Error("hot videos failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// Ranking 获取视频排行榜（公开接口，支持日榜/周榜/总榜）。
func (h *Handler) Ranking(c *gin.Context) {
	// 排行榜周期：day（日榜）/ week（周榜）/ total（总榜）
	period := c.DefaultQuery("period", "day")
	if period != "day" && period != "week" && period != "total" {
		period = "day"
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	resp, err := h.svc.Ranking(c.Request.Context(), period, page, pageSize)
	if err != nil {
		logger.Error("ranking failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// ==================== 辅助函数与类型 ====================

// validVideoMIMEs 允许上传的视频 MIME 类型白名单。
var validVideoMIMEs = map[string]bool{
	"video/mp4":        true,
	"video/webm":       true,
	"video/ogg":        true,
	"video/quicktime":  true,
	"video/x-msvideo":  true,
	"video/x-matroska": true,
}

// validCoverMIMEs 允许上传的封面图片 MIME 类型白名单。
var validCoverMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

// isValidCoverMIME 返回封面图片的 MIME 类型是否在白名单中。
func isValidCoverMIME(mime string) bool {
	return validCoverMIMEs[mime]
}

// isValidVideoMIME 返回视频文件的 MIME 类型是否在白名单中。
func isValidVideoMIME(mime string) bool {
	return validVideoMIMEs[mime]
}

// isValidVideoMagic 校验视频文件的魔数字节（Magic Bytes），防止伪造 MIME 类型。
// 支持 MP4、WebM/MKV、AVI、QuickTime、OGG 等常见格式。
func isValidVideoMagic(buf []byte) bool {
	if len(buf) < 4 {
		return false
	}
	// 常见视频文件签名
	// MP4: ftyp (位于偏移 4)
	// WebM/MKV: 0x1A 0x45 0xDF 0xA3
	// AVI: RIFF
	// QuickTime: ftyp 或 moov
	// OGG: OggS

	// WebM / MKV 魔数
	if buf[0] == 0x1A && buf[1] == 0x45 && buf[2] == 0xDF && buf[3] == 0xA3 {
		return true
	}
	// AVI 魔数
	if buf[0] == 'R' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == 'F' {
		return true
	}
	// OGG 魔数
	if buf[0] == 'O' && buf[1] == 'g' && buf[2] == 'g' && buf[3] == 'S' {
		return true
	}
	// MP4 / QuickTime：检测偏移 4 处是否为 ftyp 或 moov
	if len(buf) > 8 {
		if (buf[4] == 'f' && buf[5] == 't' && buf[6] == 'y' && buf[7] == 'p') ||
			(buf[4] == 'm' && buf[5] == 'o' && buf[6] == 'o' && buf[7] == 'v') {
			return true
		}
	}
	// 宽松匹配：兼容其他包含 ftyp/moov 标识的容器格式
	return strings.Contains(string(buf), "ftyp") || strings.Contains(string(buf), "moov")
}

// combinedReader 组合读取器，先回放前缀缓冲区中的魔数字节，
// 再从底层的 multipart.File 读取剩余数据。用于上传时 MIME 校验后无缝读取。
type combinedReader struct {
	prefix     []byte         // 已读取的魔数字节前缀
	file       interface {    // 底层文件读取器
		Read(p []byte) (n int, err error)
	}
	prefixRead bool           // 前缀是否已回放完毕
}

// Read 实现 io.Reader 接口：先读取前缀数据，再继续读文件。
func (cr *combinedReader) Read(p []byte) (int, error) {
	if !cr.prefixRead {
		// 回放前缀中的魔数字节
		n := copy(p, cr.prefix)
		if n < len(cr.prefix) {
			// 前缀尚未完全回放，保留剩余部分
			cr.prefix = cr.prefix[n:]
			return n, nil
		}
		cr.prefixRead = true
		return n, nil
	}
	// 前缀已回放完毕，从文件读取
	return cr.file.Read(p)
}
