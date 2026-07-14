// Package comment 提供评论相关的 HTTP 处理器，包括评论列表获取、
// 创建评论、点赞评论以及删除评论等功能。
package comment

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	commentmodel "github.com/My-TuDo/B-B/backend/internal/model/comment"
	commentservice "github.com/My-TuDo/B-B/backend/internal/service/comment"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// Handler 是评论模块的 HTTP 处理器，持有评论服务实例。
type Handler struct {
	svc *commentservice.Service
}

// NewHandler 创建评论处理器实例。
func NewHandler(svc *commentservice.Service) *Handler {
	return &Handler{svc: svc}
}

// GetComments 获取指定视频的评论列表（公开接口，支持分页和排序）。
// 排序方式：new（最新）或 hot（最热）。
func (h *Handler) GetComments(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	// 分页与排序参数（带默认值）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sort := c.DefaultQuery("sort", "new")
	if sort != "hot" && sort != "new" {
		sort = "new"
	}

	resp, err := h.svc.GetComments(c.Request.Context(), uint(videoID), page, pageSize, sort)
	if err != nil {
		logger.Error("get comments failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// CreateComment 创建一条评论（需认证，支持平级评论和回复子评论）。
func (h *Handler) CreateComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req commentmodel.CommentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 评论内容不能为空
	if req.Content == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "评论内容不能为空")
		return
	}

	resp, err := h.svc.CreateComment(c.Request.Context(), uint(videoID), userID, &req)
	if err != nil {
		var svcErr *commentservice.Error
		if errors.As(err, &svcErr) {
			// 根据错误码返回不同的 HTTP 状态码
			status := http.StatusBadRequest
			if svcErr.Code == errcode.VideoNotFound || svcErr.Code == errcode.CommentNotFound {
				status = http.StatusNotFound
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("create comment failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Created(c, resp)
}

// LikeComment 切换评论点赞状态（需认证），同一个用户重复调用可取消点赞。
func (h *Handler) LikeComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	commentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "评论ID格式错误")
		return
	}

	resp, err := h.svc.LikeComment(c.Request.Context(), uint(commentID), userID)
	if err != nil {
		logger.Error("like comment failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

// DeleteComment 删除评论（需认证，评论作者或视频作者均可删除）。
// 路由中包含视频 ID 和评论 ID，用于校验权限。
func (h *Handler) DeleteComment(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if userID == 0 {
		response.Error(c, http.StatusUnauthorized, errcode.Unauthorized, errcode.Message(errcode.Unauthorized))
		return
	}

	commentID, err := strconv.ParseUint(c.Param("comment_id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "评论ID格式错误")
		return
	}

	// 解析视频 ID（用于服务层校验视频作者是否有权删除评论）
	videoID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	if err := h.svc.DeleteComment(c.Request.Context(), uint(commentID), userID, uint(videoID)); err != nil {
		var svcErr *commentservice.Error
		if errors.As(err, &svcErr) {
			// 权限不足返回 403，评论不存在返回 404
			status := http.StatusForbidden
			if svcErr.Code == errcode.CommentNotFound {
				status = http.StatusNotFound
			}
			response.Error(c, status, svcErr.Code, svcErr.Msg)
			return
		}
		logger.Error("delete comment failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}
