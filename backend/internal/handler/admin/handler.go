package admin

import (
	"net/http"
	"strconv"

	"github.com/My-TuDo/B-B/backend/internal/middleware"
	historymodel "github.com/My-TuDo/B-B/backend/internal/model/history"
	videorepo "github.com/My-TuDo/B-B/backend/internal/repository/video"
	adminservice "github.com/My-TuDo/B-B/backend/internal/service/admin"
	"github.com/My-TuDo/B-B/backend/pkg/errcode"
	"github.com/My-TuDo/B-B/backend/pkg/logger"
	"github.com/My-TuDo/B-B/backend/pkg/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Handler struct {
	db   *gorm.DB
	repo *videorepo.Repository
	svc  *adminservice.Service
}

func NewHandler(db *gorm.DB, repo *videorepo.Repository, svc *adminservice.Service) *Handler {
	return &Handler{db: db, repo: repo, svc: svc}
}

func (h *Handler) AdminVideos(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	var statusFilter int8 = 2
	statusStr := c.DefaultQuery("status", "2")
	if s, err := strconv.ParseInt(statusStr, 10, 8); err == nil {
		statusFilter = int8(s)
	}

	resp, err := h.svc.ListVideos(c.Request.Context(), statusFilter, page, pageSize)
	if err != nil {
		logger.Error("admin list videos failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, resp)
}

func (h *Handler) Review(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	videoIDStr := c.Param("id")
	videoID, err := strconv.ParseUint(videoIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "视频ID格式错误")
		return
	}

	var req historymodel.AdminReviewReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if req.Status != 1 && req.Status != 3 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "status只能为1或3")
		return
	}

	video, err := h.repo.FindByID(c.Request.Context(), uint(videoID))
	if err != nil {
		logger.Error("admin review find failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}
	if video == nil {
		response.Error(c, http.StatusNotFound, errcode.VideoNotFound, errcode.Message(errcode.VideoNotFound))
		return
	}

	video.Status = req.Status
	if err := h.repo.Update(c.Request.Context(), video); err != nil {
		logger.Error("admin review update failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

// Stats returns dashboard aggregate statistics. role >= 3 only.
func (h *Handler) Stats(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	stats, err := getAdminStats(h.db, c.Request.Context())
	if err != nil {
		logger.Error("admin stats failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, stats)
}

// Users lists users with optional search query. role >= 3 only.
func (h *Handler) Users(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	q := c.DefaultQuery("q", "")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	items, total, err := searchUsers(h.db, c.Request.Context(), q, page, pageSize)
	if err != nil {
		logger.Error("admin users search failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	if items == nil {
		items = []UserListItem{}
	}

	response.Success(c, &UsersListResp{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

// System returns server configuration and status info. role >= 3 only.
func (h *Handler) System(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	sysInfo := getSystemInfo(h.db, c.Request.Context())
	response.Success(c, sysInfo)
}

// UpdateUserRole changes a user's role. role >= 3 only.
func (h *Handler) UpdateUserRole(c *gin.Context) {
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "用户ID格式错误")
		return
	}

	var req struct {
		Role uint8 `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	if req.Role < 1 || req.Role > 3 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "role 取值范围为 1-3")
		return
	}

	if err := updateUserRole(h.db, c.Request.Context(), uint(userID), req.Role); err != nil {
		logger.Error("admin update user role failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}
