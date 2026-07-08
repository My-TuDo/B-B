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

type Handler struct {
	svc *commentservice.Service
}

func NewHandler(svc *commentservice.Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetComments(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

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

	if req.Content == "" {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "评论内容不能为空")
		return
	}

	resp, err := h.svc.CreateComment(c.Request.Context(), uint(videoID), userID, &req)
	if err != nil {
		var svcErr *commentservice.Error
		if errors.As(err, &svcErr) {
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

	// We need the video author ID. For simplicity, pass 0 and let service handle.
	// Actually, we need video ID to get the author. But the route has videoID param.
	videoID, _ := strconv.ParseUint(c.Param("id"), 10, 64)

	// Get video to check author
	// We use the video repo via the service. For now pass 0 for videoAuthorID
	// The service will need to check. We'll pass 0 and let the service not check video auth.
	// Actually we need to implement this properly. Let me just pass videoID and let service check.

	if err := h.svc.DeleteComment(c.Request.Context(), uint(commentID), userID, uint(videoID)); err != nil {
		var svcErr *commentservice.Error
		if errors.As(err, &svcErr) {
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
