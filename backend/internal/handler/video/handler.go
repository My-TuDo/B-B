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

type Handler struct {
	svc *videoservice.Service
}

func NewHandler(svc *videoservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Upload(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	// Get file
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请选择视频文件")
		return
	}
	defer file.Close()

	// Validate MIME type
	contentType := header.Header.Get("Content-Type")
	if !isValidVideoMIME(contentType) {
		response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, errcode.Message(errcode.InvalidFileType))
		return
	}

	// Check magic bytes (first 512 bytes)
	magicBuf := make([]byte, 512)
	n, _ := file.Read(magicBuf)
	if !isValidVideoMagic(magicBuf[:n]) {
		response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, errcode.Message(errcode.InvalidFileType))
		return
	}

	// Read form fields
	title := c.PostForm("title")
	description := c.PostForm("description")
	categoryIDStr := c.PostForm("category_id")

	if title == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频标题不能为空")
		return
	}
	if len(title) > 100 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "标题不能超过100个字符")
		return
	}

	var categoryID uint
	if categoryIDStr != "" {
		parsed, err := strconv.ParseUint(categoryIDStr, 10, 64)
		if err == nil {
			categoryID = uint(parsed)
		}
	}

	// Get cover file (optional)
	var coverFile multipart.File
	var coverHeader *multipart.FileHeader
	coverFile, coverHeader, err = c.Request.FormFile("cover")
	if err == nil {
		defer coverFile.Close()

		// Validate cover MIME
		coverContentType := coverHeader.Header.Get("Content-Type")
		if !isValidCoverMIME(coverContentType) {
			response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, "封面图片格式不支持，仅支持 JPEG/PNG/WebP/GIF")
			return
		}

		// Validate cover size (max 5MB)
		const maxCoverSize = 5 * 1024 * 1024
		if coverHeader.Size > maxCoverSize {
			response.Error(c, http.StatusBadRequest, errcode.FileTooLarge, "封面图片大小不能超过 5MB")
			return
		}
	}

	// Use a struct that reads from magicBuf first then from file
	combinedReader := &combinedReader{
		prefix: magicBuf[:n],
		file:   file,
	}

	// SSE - set headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		response.Error(c, http.StatusInternalServerError, errcode.Internal, "SSE not supported")
		return
	}

	// Progress callback sends SSE events
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
		// If resp is not nil, video was saved but cover upload failed
		if resp != nil {
			msg := videomodel.UploadSSEMessage{Error: "封面上传失败，视频已保存"}
			data, _ := json.Marshal(msg)
			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			flusher.Flush()
			// Fall through to send complete event
		} else {
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

	// Send complete event
	completeData, _ := json.Marshal(resp)
	fmt.Fprintf(c.Writer, "event: complete\ndata: %s\n\n", string(completeData))
	flusher.Flush()
}

func (h *Handler) GetVideo(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

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

func (h *Handler) GetPlayURL(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

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

func (h *Handler) ListVideos(c *gin.Context) {
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

func (h *Handler) ListUserVideos(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

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

func (h *Handler) Ranking(c *gin.Context) {
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

// ==================== Helpers ====================

var validVideoMIMEs = map[string]bool{
	"video/mp4":       true,
	"video/webm":      true,
	"video/ogg":       true,
	"video/quicktime": true,
	"video/x-msvideo": true,
	"video/x-matroska": true,
}

var validCoverMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
	"image/gif":  true,
}

func isValidCoverMIME(mime string) bool {
	return validCoverMIMEs[mime]
}

func isValidVideoMIME(mime string) bool {
	return validVideoMIMEs[mime]
}

func isValidVideoMagic(buf []byte) bool {
	if len(buf) < 4 {
		return false
	}
	// Common video file signatures
	// MP4: ftyp (at offset 4)
	// WebM/MKV: 0x1A 0x45 0xDF 0xA3
	// AVI: RIFF
	// QuickTime: ftyp or moov
	// OGG: OggS
	if buf[0] == 0x1A && buf[1] == 0x45 && buf[2] == 0xDF && buf[3] == 0xA3 {
		return true // WebM/MKV
	}
	if buf[0] == 'R' && buf[1] == 'I' && buf[2] == 'F' && buf[3] == 'F' {
		return true // AVI
	}
	if buf[0] == 'O' && buf[1] == 'g' && buf[2] == 'g' && buf[3] == 'S' {
		return true // OGG
	}
	// MP4/QuickTime: check for ftyp at offset 4
	if len(buf) > 8 {
		if (buf[4] == 'f' && buf[5] == 't' && buf[6] == 'y' && buf[7] == 'p') ||
			(buf[4] == 'm' && buf[5] == 'o' && buf[6] == 'o' && buf[7] == 'v') {
			return true
		}
	}
	// Relax for other formats: if it starts with common video container patterns
	return strings.Contains(string(buf), "ftyp") || strings.Contains(string(buf), "moov")
}

// combinedReader reads from a prefix buffer first, then from the underlying reader.
type combinedReader struct {
	prefix []byte
	file   interface {
		Read(p []byte) (n int, err error)
	}
	prefixRead bool
}

func (cr *combinedReader) Read(p []byte) (int, error) {
	if !cr.prefixRead {
		n := copy(p, cr.prefix)
		if n < len(cr.prefix) {
			cr.prefix = cr.prefix[n:]
			return n, nil
		}
		cr.prefixRead = true
		return n, nil
	}
	return cr.file.Read(p)
}
