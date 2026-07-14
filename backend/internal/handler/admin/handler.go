// Package admin 提供管理后台相关的 HTTP 处理器，包括视频审核、
// 数据统计看板、用户管理、系统信息查询以及用户角色管理等功能。
// 所有接口均要求 role >= 3（管理员权限）。
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

// Handler 是管理后台的 HTTP 处理器，持有数据库实例、视频仓库和管理服务。
type Handler struct {
	db   *gorm.DB
	repo *videorepo.Repository
	svc  *adminservice.Service
}

// NewHandler 创建管理后台处理器实例。
func NewHandler(db *gorm.DB, repo *videorepo.Repository, svc *adminservice.Service) *Handler {
	return &Handler{db: db, repo: repo, svc: svc}
}

// AdminVideos 获取管理后台视频列表（需 role >= 3）。
// 支持按审核状态筛选和分页。
func (h *Handler) AdminVideos(c *gin.Context) {
	// 权限校验：仅管理员可访问
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "12"))

	// 状态过滤器，默认为 2（待审核）
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

// Review 审核视频（需 role >= 3）。
// 将视频状态更新为 1（通过）或 3（驳回）。
func (h *Handler) Review(c *gin.Context) {
	// 权限校验
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

	// 审核状态仅允许 1（通过）或 3（驳回）
	if req.Status != 1 && req.Status != 3 {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "status只能为1或3")
		return
	}

	// 查找目标视频
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

	// 更新视频审核状态并持久化
	video.Status = req.Status
	if err := h.repo.Update(c.Request.Context(), video); err != nil {
		logger.Error("admin review update failed", zap.Error(err))
		response.Error(c, http.StatusInternalServerError, errcode.Internal, errcode.Message(errcode.Internal))
		return
	}

	response.Success(c, nil)
}

// Stats 返回管理后台仪表盘聚合统计数据（需 role >= 3）。
// 包括用户总数、视频总数、总播放量、评论数、弹幕数、今日新增用户和视频。
func (h *Handler) Stats(c *gin.Context) {
	// 权限校验
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

// Users 列出用户列表（需 role >= 3），支持按用户名或昵称搜索。
func (h *Handler) Users(c *gin.Context) {
	// 权限校验
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	// 分页参数与搜索关键词
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	q := c.DefaultQuery("q", "")

	// 分页参数边界保护
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

	// 确保返回空数组而非 nil
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

// System 返回服务器配置和运行状态信息（需 role >= 3）。
// 包括 Go 版本、运行时长、数据库连接状态。
func (h *Handler) System(c *gin.Context) {
	// 权限校验
	role := middleware.GetRole(c)
	if role < 3 {
		response.Error(c, http.StatusForbidden, errcode.Forbidden, errcode.Message(errcode.Forbidden))
		return
	}

	sysInfo := getSystemInfo(h.db, c.Request.Context())
	response.Success(c, sysInfo)
}

// UpdateUserRole 修改用户角色（需 role >= 3）。
// 角色取值范围 1-3：1=普通用户，2=版主，3=管理员。
func (h *Handler) UpdateUserRole(c *gin.Context) {
	// 权限校验
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

	// 解析请求体中的新角色值
	var req struct {
		Role uint8 `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, errcode.BadRequest, "请求参数格式错误")
		return
	}

	// 角色取值范围校验（1-3）
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
