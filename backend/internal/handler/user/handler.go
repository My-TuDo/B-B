// Package user 提供用户信息相关的 HTTP 处理器，包括用户信息获取、
// 个人信息更新以及头像上传等功能。
package user

import (
	"errors"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	usermodel "github.com/My-TuDo/B-B/backend/internal/model/user"
	"github.com/My-TuDo/B-B/backend/internal/middleware"
	userservice "github.com/My-TuDo/B-B/backend/internal/service/user"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 是用户模块的 HTTP 处理器，持有用户服务实例。
type Handler struct {
	svc *userservice.Service
}

// NewHandler 创建用户处理器实例。
func NewHandler(svc *userservice.Service) *Handler {
	return &Handler{svc: svc}
}

// GetUser 获取指定 ID 的用户公开信息（无需认证）。
func (h *Handler) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	resp, err := h.svc.GetUser(c.Request.Context(), uint(id))
	if err != nil {
		// 判断是否为业务层错误（如用户不存在）
		var svcErr *userservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("get user failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// UpdateUser 更新用户个人信息（需认证，仅允许用户修改自己的信息）。
func (h *Handler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// 用户隔离校验：仅允许用户修改自己的信息
	userID := middleware.GetUserID(c)
	if uint(id) != userID {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	var req usermodel.UpdateUserReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	resp, err := h.svc.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		var svcErr *userservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("update user failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// UploadAvatar 上传用户头像（需认证，仅允许用户修改自己的头像）。
// 支持 JPEG、PNG、WebP 格式，文件大小限制 2MB。
func (h *Handler) UploadAvatar(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	// 用户隔离校验：仅允许用户上传自己的头像
	userID := middleware.GetUserID(c)
	if uint(id) != userID {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	// 从 multipart 表单中读取头像文件
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请选择头像文件")
		return
	}
	defer file.Close()

	// 校验文件大小（最大 2MB）
	const maxSize = 2 * 1024 * 1024
	if header.Size > maxSize {
		response.Error(c, http.StatusBadRequest, errcode.FileTooLarge, errcode.Message(errcode.FileTooLarge))
		return
	}

	// 校验文件扩展名（仅允许 JPEG/PNG/WebP）
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".webp": true,
	}
	if !allowed[ext] {
		response.Error(c, http.StatusBadRequest, errcode.InvalidFileType, errcode.Message(errcode.InvalidFileType))
		return
	}
	// 统一 .jpeg 为 .jpg
	if ext == ".jpeg" {
		ext = ".jpg"
	}

	// 读取文件内容
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		logger.Error("read avatar file failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.AvatarUploadFail, errcode.Message(errcode.AvatarUploadFail))
		return
	}

	// 调用服务层处理头像上传（含存储和数据库更新）
	contentType := header.Header.Get("Content-Type")
	resp, err := h.svc.UploadAvatar(c.Request.Context(), uint(id), strings.NewReader(string(fileBytes)), int64(len(fileBytes)), contentType, ext[1:])
	if err != nil {
		var svcErr *userservice.Error
		if errors.As(err, &svcErr) {
			response.Error(c, http.StatusNotFound, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("upload avatar failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.AvatarUploadFail, errcode.Message(errcode.AvatarUploadFail))
		return
	}

	response.Success(c, resp)
}
