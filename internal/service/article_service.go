package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/blog-article/internal/repository"
	"github.com/mysunshines/blog-article/internal/errors"
	"github.com/mysunshines/gocommon/cache"
	"github.com/mysunshines/gocommon/util"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ArticleService interface {
	CreateArticle(ctx context.Context, req *model.CreateArticleRequest) (*model.Article, error)
	GetArticle(ctx context.Context, id uint) (*model.Article, error)
	GetArticleForAdmin(ctx context.Context, id uint) (*model.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*model.Article, error)
	UpdateArticle(ctx context.Context, id uint, req *model.UpdateArticleRequest) (*model.Article, error)
	DeleteArticle(ctx context.Context, id, userID uint) error
	ListArticles(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error)
	SearchArticles(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error)
	GetUserArticles(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error)
	IncrementViewCount(ctx context.Context, id uint) (int, error)
	LikeArticle(ctx context.Context, articleID, userID uint) (likeCount int, liked bool, err error)
	CancelLikeArticle(ctx context.Context, articleID, userID uint) (likeCount int, liked bool, err error)
	GetLikeStatus(ctx context.Context, articleID, userID uint) (likeCount int, liked bool, err error)
	GetCategories(ctx context.Context) ([]*model.Category, error)
	CreateCategory(ctx context.Context, name, description string, sort int) (*model.Category, error)
	UpdateCategory(ctx context.Context, id uint, name, description string, sort int) (*model.Category, error)
	DeleteCategory(ctx context.Context, id uint) error
	GetTags(ctx context.Context) ([]*model.Tag, error)

	// 后台审核管理（仅管理员调用）
	AdminListArticles(ctx context.Context, req *model.AdminListArticlesRequest) ([]*model.Article, int64, error)
	ApproveArticle(ctx context.Context, id uint) (*model.Article, error)
	RejectArticle(ctx context.Context, id uint, reason string) (*model.Article, error)
	OfflineArticle(ctx context.Context, id uint, reason string) (*model.Article, error)
	PublishArticle(ctx context.Context, id uint) (*model.Article, error)
	SubmitArticle(ctx context.Context, id uint) (*model.Article, error)
	AdminUpdateArticle(ctx context.Context, id uint, req *model.UpdateArticleRequest) (*model.Article, error)
	AdminDeleteArticle(ctx context.Context, id uint) error
}

type articleService struct {
	repo    repository.ArticleRepository
	db      *gorm.DB
	sfGroup singleflight.Group // 高并发：请求合并
}

func NewArticleService(repo repository.ArticleRepository, db *gorm.DB) ArticleService {
	return &articleService{
		repo: repo,
		db:   db,
	}
}

func (s *articleService) CreateArticle(ctx context.Context, req *model.CreateArticleRequest) (*model.Article, error) {
	// 参数校验
	if req.Title == "" {
		return nil, errors.BadRequest("标题不能为空")
	}
	if req.Content == "" {
		return nil, errors.BadRequest("内容不能为空")
	}

	// 生成 slug
	slug := generateSlug(req.Title)

	// 先发后审模式：勾选"立即发布"进入待审核(pending, 前台可见待管理员审核)；
	// 未勾选则存为草稿(draft, 仅作者可见，管理员不可审核)
	var status string
	if req.IsPublished {
		status = model.ArticleStatusPending
	} else {
		status = model.ArticleStatusDraft
	}

	// 构建文章模型
	article := &model.Article{
		UserID:       req.UserID,
		Title:        req.Title,
		Slug:         slug,
		Summary:      req.Summary,
		Content:      req.Content,
		CoverImage:   req.CoverImage,
		CategoryID:   req.CategoryID,
		Status:       status,
		IsFeatured:   req.IsFeatured,
		AllowComment: req.AllowComment,
	}

	// 创建文章
	if err := s.repo.Create(ctx, article); err != nil {
		return nil, err
	}

	// 重新获取完整文章信息
	return s.repo.GetByID(ctx, article.ID)
}

func (s *articleService) GetArticle(ctx context.Context, id uint) (*model.Article, error) {
	// 高并发优化：使用 singleflight 合并并发请求
	key := fmt.Sprintf("article:%d", id)
	result, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		// 先尝试从本地缓存获取
		if cached, found := cache.LocalCacheGet(key); found {
			if article, ok := cached.(*model.Article); ok {
				return article, nil
			}
		}
		// 从数据库获取
		article, err := s.repo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		// 公开接口：已发布(published)与待审核(pending, 先发后审前台可见)可展示；
		// 草稿/已下线/已拒绝一律视为不存在
		if article.Status != model.ArticleStatusPublished && article.Status != model.ArticleStatusPending {
			return nil, errors.ArticleNotFound()
		}
		// 存入本地缓存
		cache.LocalCacheSet(key, article)
		return article, nil
	})
	if err != nil {
		return nil, err
	}
	article := result.(*model.Article)
	// 后端渲染并净化 Markdown，前端只负责展示，杜绝 XSS
	article.ContentHTML = util.RenderMarkdown(article.Content)
	return article, nil
}

// GetArticleForAdmin 后台获取文章详情（不限状态，供审核/编辑使用）
func (s *articleService) GetArticleForAdmin(ctx context.Context, id uint) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 后端渲染并净化 Markdown，前端只负责展示，杜绝 XSS
	article.ContentHTML = util.RenderMarkdown(article.Content)
	return article, nil
}

func (s *articleService) GetArticleBySlug(ctx context.Context, slug string) (*model.Article, error) {
	if slug == "" {
		return nil, errors.InvalidSlug()
	}

	// 高并发优化：使用 singleflight + 缓存
	key := fmt.Sprintf("article:slug:%s", slug)
	result, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		// 先尝试从本地缓存获取
		if cached, found := cache.LocalCacheGet(key); found {
			if article, ok := cached.(*model.Article); ok {
				return article, nil
			}
		}
		// 从数据库获取
		article, err := s.repo.GetBySlug(ctx, slug)
		if err != nil {
			return nil, err
		}
		// 公开接口：已发布(published)与待审核(pending, 先发后审前台可见)可展示；
		// 草稿/已下线/已拒绝一律视为不存在
		if article.Status != model.ArticleStatusPublished && article.Status != model.ArticleStatusPending {
			return nil, errors.InvalidSlug()
		}
		// 存入本地缓存
		cache.LocalCacheSet(key, article)
		return article, nil
	})
	if err != nil {
		return nil, err
	}
	article := result.(*model.Article)
	// 后端渲染并净化 Markdown，前端只负责展示，杜绝 XSS
	article.ContentHTML = util.RenderMarkdown(article.Content)
	return article, nil
}

func (s *articleService) UpdateArticle(ctx context.Context, id uint, req *model.UpdateArticleRequest) (*model.Article, error) {
	// 获取原文章
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 检查权限
	if article.UserID != req.UserID {
		return nil, errors.PermissionDenied()
	}

	// 更新字段
	if req.Title != "" {
		article.Title = req.Title
		article.Slug = generateSlug(req.Title)
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.Summary != "" {
		article.Summary = req.Summary
	}
	if req.CoverImage != "" {
		article.CoverImage = req.CoverImage
	}
	if req.CategoryID > 0 {
		article.CategoryID = req.CategoryID
	}
	article.IsFeatured = req.IsFeatured
	article.AllowComment = req.AllowComment
	// 发布状态由状态机统一管理：
	// - is_published=true  → 进入待审核(pending, 先发后审前台可见)，并清空历史拒绝/下线原因
	// - is_published=false → 仅当原本是草稿(draft)时保持草稿；非草稿态(已发布/待审等)编辑内容不改变状态
	if req.IsPublished {
		article.Status = model.ArticleStatusPending
		article.RejectReason = ""
		article.OfflineReason = ""
	} else if article.Status == model.ArticleStatusDraft {
		article.Status = model.ArticleStatusDraft
	}

	// 保存更新
	if err := s.repo.Update(ctx, article); err != nil {
		return nil, err
	}

	// 返回更新后的文章
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) DeleteArticle(ctx context.Context, id, userID uint) error {
	// 获取原文章
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 检查权限
	if article.UserID != userID {
		return errors.PermissionDenied()
	}

	return s.repo.Delete(ctx, id)
}

func (s *articleService) ListArticles(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}

	return s.repo.List(ctx, req)
}

func (s *articleService) SearchArticles(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error) {
	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 20
	}

	return s.repo.Search(ctx, req)
}

func (s *articleService) GetUserArticles(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 10
	}

	return s.repo.GetByUserID(ctx, userID, page, size)
}

func (s *articleService) IncrementViewCount(ctx context.Context, id uint) (int, error) {
	return s.repo.IncrementViewCount(ctx, id)
}

func (s *articleService) LikeArticle(ctx context.Context, articleID, userID uint) (int, bool, error) {
	return s.repo.LikeArticle(ctx, articleID, userID)
}

func (s *articleService) CancelLikeArticle(ctx context.Context, articleID, userID uint) (int, bool, error) {
	return s.repo.CancelLikeArticle(ctx, articleID, userID)
}

func (s *articleService) GetLikeStatus(ctx context.Context, articleID, userID uint) (int, bool, error) {
	return s.repo.GetLikeStatus(ctx, articleID, userID)
}

func (s *articleService) CreateCategory(ctx context.Context, name, description string, sort int) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.BadRequest("分类名称不能为空")
	}

	// 查重名
	var cnt int64
	if err := s.db.WithContext(ctx).Model(&model.Category{}).Where("name = ?", name).Count(&cnt).Error; err != nil {
		return nil, errors.Internal("查询分类失败", err)
	}
	if cnt > 0 {
		return nil, errors.BadRequest("分类名称已存在")
	}

	category := &model.Category{
		Name:        name,
		Slug:        generateSlug(name),
		Description: strings.TrimSpace(description),
		Sort:        sort,
	}
	if err := s.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, errors.Internal("创建分类失败", err)
	}

	// 清除分类缓存，使新建分类立即生效
	cache.LocalCacheDelete("categories:all")
	return category, nil
}

func (s *articleService) UpdateCategory(ctx context.Context, id uint, name, description string, sort int) (*model.Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.BadRequest("分类名称不能为空")
	}

	var existing model.Category
	if err := s.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.NotFound("分类不存在")
		}
		return nil, errors.Internal("查询分类失败", err)
	}

	// 查重名（排除自身）
	var cnt int64
	if err := s.db.WithContext(ctx).Model(&model.Category{}).
		Where("name = ? AND id <> ?", name, id).Count(&cnt).Error; err != nil {
		return nil, errors.Internal("查询分类失败", err)
	}
	if cnt > 0 {
		return nil, errors.BadRequest("分类名称已存在")
	}

	cols := map[string]interface{}{
		"name":        name,
		"description": strings.TrimSpace(description),
		"sort":        sort,
	}
	if err := s.db.WithContext(ctx).Model(&model.Category{}).Where("id = ?", id).Updates(cols).Error; err != nil {
		return nil, errors.Internal("更新分类失败", err)
	}

	cache.LocalCacheDelete("categories:all")
	if err := s.db.WithContext(ctx).First(&existing, id).Error; err != nil {
		return nil, errors.Internal("查询更新后的分类失败", err)
	}
	return &existing, nil
}

func (s *articleService) GetCategories(ctx context.Context) ([]*model.Category, error) {
	// 高并发优化：使用 singleflight + 缓存
	key := "categories:all"
	result, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		// 先尝试从本地缓存获取
		if cached, found := cache.LocalCacheGet(key); found {
			if categories, ok := cached.([]*model.Category); ok {
				return categories, nil
			}
		}
		// 从数据库获取
		var categories []*model.Category
		result := s.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&categories)
		if result.Error != nil {
			return nil, errors.Internal("获取分类列表失败", result.Error)
		}
		// 存入本地缓存（分类数据变化不频繁，缓存1小时）
		cache.LocalCacheSet(key, categories)
		return categories, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]*model.Category), nil
}

func (s *articleService) GetTags(ctx context.Context) ([]*model.Tag, error) {
	// 高并发优化：使用 singleflight + 缓存
	key := "tags:all"
	result, err, _ := s.sfGroup.Do(key, func() (interface{}, error) {
		// 先尝试从本地缓存获取
		if cached, found := cache.LocalCacheGet(key); found {
			if tags, ok := cached.([]*model.Tag); ok {
				return tags, nil
			}
		}
		// 从数据库获取
		var tags []*model.Tag
		result := s.db.WithContext(ctx).Order("article_count DESC").Find(&tags)
		if result.Error != nil {
			return nil, errors.Internal("获取标签列表失败", result.Error)
		}
		// 存入本地缓存
		cache.LocalCacheSet(key, tags)
		return tags, nil
	})
	if err != nil {
		return nil, err
	}
	return result.([]*model.Tag), nil
}

// --------------------------- 后台审核管理 ---------------------------

func (s *articleService) AdminListArticles(ctx context.Context, req *model.AdminListArticlesRequest) ([]*model.Article, int64, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Size < 1 {
		req.Size = 10
	}
	return s.repo.AdminList(ctx, req)
}

func (s *articleService) ApproveArticle(ctx context.Context, id uint) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusPending {
		return nil, errors.BadRequest("仅待审核文章可以审核通过")
	}

	now := time.Now()
	cols := map[string]interface{}{
		"status":        model.ArticleStatusPublished,
		"published_at":  &now,
		"reject_reason": "",
	}
	if err := s.repo.UpdateStatus(ctx, id, cols); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) RejectArticle(ctx context.Context, id uint, reason string) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusPending {
		return nil, errors.BadRequest("仅待审核文章可以拒绝")
	}

	cols := map[string]interface{}{
		"status":        model.ArticleStatusRejected,
		"reject_reason": reason,
	}
	if err := s.repo.UpdateStatus(ctx, id, cols); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) OfflineArticle(ctx context.Context, id uint, reason string) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusPublished {
		return nil, errors.BadRequest("仅已发布文章可以下线")
	}

	cols := map[string]interface{}{
		"status":         model.ArticleStatusOffline,
		"offline_reason": reason,
	}
	if err := s.repo.UpdateStatus(ctx, id, cols); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) PublishArticle(ctx context.Context, id uint) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusOffline {
		return nil, errors.BadRequest("仅已下线文章可以重新发布")
	}

	cols := map[string]interface{}{
		"status": model.ArticleStatusPublished,
	}
	if article.PublishedAt == nil {
		now := time.Now()
		cols["published_at"] = &now
	}
	if err := s.repo.UpdateStatus(ctx, id, cols); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.GetByID(ctx, id)
}

// SubmitArticle 提交审核（draft/offline/rejected -> pending）
// 作者或管理员将草稿/已下线/已拒绝文章提交进入审核队列，供管理员审核通过/拒绝。
func (s *articleService) SubmitArticle(ctx context.Context, id uint) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusDraft &&
		article.Status != model.ArticleStatusOffline &&
		article.Status != model.ArticleStatusRejected {
		return nil, errors.BadRequest("仅草稿、已下线或已拒绝文章可以提交审核")
	}

	cols := map[string]interface{}{
		"status":         model.ArticleStatusPending,
		"reject_reason":  "",
		"offline_reason": "",
	}
	if err := s.repo.UpdateStatus(ctx, id, cols); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) AdminUpdateArticle(ctx context.Context, id uint, req *model.UpdateArticleRequest) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 管理员编辑：不受作者权限限制，但保持原状态机不变
	if req.Title != "" {
		article.Title = req.Title
		article.Slug = generateSlug(req.Title)
	}
	if req.Content != "" {
		article.Content = req.Content
	}
	if req.Summary != "" {
		article.Summary = req.Summary
	}
	if req.CoverImage != "" {
		article.CoverImage = req.CoverImage
	}
	if req.CategoryID > 0 {
		article.CategoryID = req.CategoryID
	}
	article.IsFeatured = req.IsFeatured
	article.AllowComment = req.AllowComment

	if err := s.repo.Update(ctx, article); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.GetByID(ctx, id)
}

func (s *articleService) AdminDeleteArticle(ctx context.Context, id uint) error {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	s.invalidateArticleCache(id, article.Slug)
	return s.repo.Delete(ctx, id)
}

// DeleteCategory 删除分类（软删除）。
// 仅当该分类下没有任何已发布文章时才允许删除，避免产生悬空文章。
func (s *articleService) DeleteCategory(ctx context.Context, id uint) error {
	// 校验分类存在
	if _, err := s.repo.GetCategory(ctx, id); err != nil {
		return err
	}

	// 统计该分类下已发布文章数
	published, err := s.repo.CountArticlesByCategory(ctx, id, model.ArticleStatusPublished)
	if err != nil {
		return err
	}
	if published > 0 {
		return errors.BadRequest("该分类下还有已发布文章，无法删除")
	}

	// 软删除分类
	if err := s.repo.DeleteCategory(ctx, id); err != nil {
		return err
	}
	cache.LocalCacheDelete("categories:all")
	return nil
}

// invalidateArticleCache 使文章缓存失效（id 与 slug 两个键）
func (s *articleService) invalidateArticleCache(id uint, slug string) {
	cache.LocalCacheDelete(fmt.Sprintf("article:%d", id))
	if slug != "" {
		cache.LocalCacheDelete(fmt.Sprintf("article:slug:%s", slug))
	}
}

// 生成 Slug
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	replacer := strings.NewReplacer(
		" ", "-", "/", "-", "\\", "-",
		":", "", "*", "", "?", "",
		"\"", "", "<", "", ">", "", "|", "",
	)
	slug = replacer.Replace(slug)
	slug = strings.Trim(slug, "-")

	if len(slug) > 100 {
		slug = slug[:100]
	}

	return fmt.Sprintf("%s-%d", slug, time.Now().UnixNano()%100000000)
}
