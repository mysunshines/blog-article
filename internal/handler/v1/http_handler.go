// Package v1 存放 article-service 的 HTTP API 处理器（v1 版本）。
// 后续迭代 v2 接口时，新增 internal/handler/v2 包即可，互不干扰。
package v1

import (
	"strconv"

	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/blog-article/internal/service"
	"github.com/mysunshines/blog-article/pkg/response"

	commonmiddleware "github.com/mysunshines/gocommon/middleware"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	svc service.ArticleService
}

func NewArticleHandler(svc service.ArticleService) *ArticleHandler {
	return &ArticleHandler{svc: svc}
}

// CreateArticle 创建文章
// @Summary 创建文章
// @Tags article
// @Accept json
// @Produce json
// @Param article body model.CreateArticleRequest true "文章信息"
// @Success 200 {object} response.Response
// @Router /api/v1/article [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	var req model.CreateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 获取用户ID（从JWT或上下文）
	userID := getUserID(c)
	req.UserID = userID

	article, err := h.svc.CreateArticle(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// GetArticle 获取文章
// @Summary 获取文章
// @Tags article
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /api/v1/article/{id} [get]
func (h *ArticleHandler) GetArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	article, err := h.svc.GetArticle(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// GetArticleBySlug 通过Slug获取文章
// @Summary 通过Slug获取文章
// @Tags article
// @Produce json
// @Param slug path string true "文章Slug"
// @Success 200 {object} response.Response
// @Router /api/v1/article/slug/{slug} [get]
func (h *ArticleHandler) GetArticleBySlug(c *gin.Context) {
	slug := c.Param("slug")

	article, err := h.svc.GetArticleBySlug(c.Request.Context(), slug)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// UpdateArticle 更新文章
// @Summary 更新文章
// @Tags article
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param article body model.UpdateArticleRequest true "文章信息"
// @Success 200 {object} response.Response
// @Router /api/v1/article/{id} [put]
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req model.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := getUserID(c)
	req.UserID = userID

	article, err := h.svc.UpdateArticle(c.Request.Context(), uint(id), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// DeleteArticle 删除文章
// @Summary 删除文章
// @Tags article
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /api/v1/article/{id} [delete]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	userID := getUserID(c)

	if err := h.svc.DeleteArticle(c.Request.Context(), uint(id), userID); err != nil {
		response.Fail(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// ListArticles 文章列表
// @Summary 文章列表
// @Tags article
// @Produce json
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param category_id query int false "分类ID"
// @Param tag query string false "标签"
// @Param is_published query bool false "是否发布"
// @Param order_by query string false "排序字段" Enums(created_at, view_count, like_count, published_at)
// @Success 200 {object} response.PageResponse
// @Router /api/v1/article [get]
func (h *ArticleHandler) ListArticles(c *gin.Context) {
	var req model.ListArticlesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}

	articles, total, err := h.svc.ListArticles(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.SuccessPage(c, articles, uint32(total), uint32(req.Page), uint32(req.Size))
}

// SearchArticles 搜索文章
// @Summary 搜索文章
// @Tags article
// @Produce json
// @Param keyword query string true "关键词"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} response.PageResponse
// @Router /api/v1/article/search [get]
func (h *ArticleHandler) SearchArticles(c *gin.Context) {
	var req model.SearchArticlesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "关键词不能为空")
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	articles, total, err := h.svc.SearchArticles(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.SuccessPage(c, articles, uint32(total), uint32(req.Page), uint32(req.Size))
}

// GetUserArticles 获取用户文章
// @Summary 获取用户文章
// @Tags article
// @Produce json
// @Param user_id path int true "用户ID"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.PageResponse
// @Router /api/v1/article/user/{user_id} [get]
func (h *ArticleHandler) GetUserArticles(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	page, _ := strconv.ParseUint(c.DefaultQuery("page", "1"), 10, 32)
	size, _ := strconv.ParseUint(c.DefaultQuery("size", "10"), 10, 32)

	articles, total, err := h.svc.GetUserArticles(c.Request.Context(), uint(userID), uint(page), uint(size))
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.SuccessPage(c, articles, uint32(total), uint32(page), uint32(size))
}

// IncrementViewCount 增加浏览数
// @Summary 增加浏览数
// @Tags article
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /api/v1/article/{id}/view [post]
func (h *ArticleHandler) IncrementViewCount(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	viewCount, err := h.svc.IncrementViewCount(c.Request.Context(), uint(id))
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, gin.H{"view_count": viewCount})
}

// GetCategories 获取分类列表
// @Summary 获取分类列表
// @Tags article
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/article/categories [get]
func (h *ArticleHandler) GetCategories(c *gin.Context) {
	categories, err := h.svc.GetCategories(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, categories)
}

// GetTags 获取标签列表
// @Summary 获取标签列表
// @Tags article
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/v1/article/tags [get]
func (h *ArticleHandler) GetTags(c *gin.Context) {
	tags, err := h.svc.GetTags(c.Request.Context())
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, tags)
}

// ===================== 后台审核管理（仅管理员） =====================

// ListArticlesForAdmin 后台文章列表（可按状态筛选）
// @Summary 后台文章列表
// @Tags admin
// @Produce json
// @Param status query string false "状态(draft/pending/published/offline/rejected)"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Success 200 {object} response.PageResponse
// @Router /api/v1/admin/articles [get]
func (h *ArticleHandler) ListArticlesForAdmin(c *gin.Context) {
	var req model.AdminListArticlesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	articles, total, err := h.svc.AdminListArticles(c.Request.Context(), &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.SuccessPage(c, articles, uint32(total), uint32(req.Page), uint32(req.Size))
}

// ApproveArticle 审核通过（pending -> published）
// @Summary 审核通过文章
// @Tags admin
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/articles/{id}/approve [post]
func (h *ArticleHandler) ApproveArticle(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	article, err := h.svc.ApproveArticle(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// RejectArticle 审核拒绝（pending -> rejected）
// @Summary 拒绝文章
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param body body model.ReviewArticleRequest true "拒绝原因"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/articles/{id}/reject [post]
func (h *ArticleHandler) RejectArticle(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req model.ReviewArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	article, err := h.svc.RejectArticle(c.Request.Context(), id, req.Reason)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// OfflineArticle 下线文章（published -> offline）
// @Summary 下线文章
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param body body model.ReviewArticleRequest true "下线原因"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/articles/{id}/offline [post]
func (h *ArticleHandler) OfflineArticle(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req model.ReviewArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	article, err := h.svc.OfflineArticle(c.Request.Context(), id, req.Reason)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// PublishArticle 重新发布（offline -> published）
// @Summary 重新发布已下线文章
// @Tags admin
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/articles/{id}/publish [post]
func (h *ArticleHandler) PublishArticle(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	article, err := h.svc.PublishArticle(c.Request.Context(), id)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// AdminUpdateArticle 管理员编辑（绕过作者权限）
// @Summary 管理员编辑文章
// @Tags admin
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param article body model.UpdateArticleRequest true "文章信息"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/articles/{id} [put]
func (h *ArticleHandler) AdminUpdateArticle(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	var req model.UpdateArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	article, err := h.svc.AdminUpdateArticle(c.Request.Context(), id, &req)
	if err != nil {
		response.Fail(c, err)
		return
	}

	response.Success(c, article)
}

// AdminDeleteArticle 管理员删除（绕过作者权限）
// @Summary 管理员删除文章
// @Tags admin
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} response.Response
// @Router /api/v1/admin/articles/{id} [delete]
func (h *ArticleHandler) AdminDeleteArticle(c *gin.Context) {
	id, err := parseIDParam(c, "id")
	if err != nil {
		response.BadRequest(c, "无效的文章ID")
		return
	}

	if err := h.svc.AdminDeleteArticle(c.Request.Context(), id); err != nil {
		response.Fail(c, err)
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// parseIDParam 从路径参数解析 uint ID
func parseIDParam(c *gin.Context, name string) (uint, error) {
	idStr := c.Param(name)
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// 辅助函数：从上下文获取用户ID
func getUserID(c *gin.Context) uint {
	// 从JWT中间件或上下文获取用户ID
	if userID, exists := c.Get(commonmiddleware.UserIDContextKey); exists {
		switch v := userID.(type) {
		case uint:
			return v
		case uint32:
			return uint(v)
		case uint64:
			return uint(v)
		case float64:
			return uint(v)
		case int:
			return uint(v)
		case int32:
			return uint(v)
		case int64:
			return uint(v)
		}
	}
	return 0
}
