// Package search 提供搜索相关的 HTTP 路由注册。
package search

import (
	searchrepo "github.com/My-TuDo/B-B/backend/internal/repository/search"
	searchservice "github.com/My-TuDo/B-B/backend/internal/service/search"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes 注册搜索相关路由到指定的路由组。
// 初始化 repository → service → handler 依赖链。
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	repo := searchrepo.NewRepository(db)
	svc := searchservice.NewService(repo)
	handler := NewHandler(svc)

	// 搜索路由组
	search := r.Group("/search")
	{
		search.GET("/", handler.Search)
		search.GET("/suggestions", handler.Suggestions)
	}
}
