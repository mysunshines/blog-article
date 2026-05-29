package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/blog-article/internal/repository"
	"github.com/mysunshines/blog-article/pkg/errors"
	"github.com/mysunshines/gocommon/cache"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type ArticleService interface {
	CreateArticle(ctx context.Context, req *model.CreateArticleRequest) (*model.Article, error)
	GetArticle(ctx context.Context, id uint) (*model.Article, error)
	GetArticleBySlug(ctx context.Context, slug string) (*model.Article, error)
	UpdateArticle(ctx context.Context, id uint, req *model.UpdateArticleRequest) (*model.Article, error)
	DeleteArticle(ctx context.Context, id, userID uint) error
	ListArticles(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error)
	SearchArticles(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error)
	GetUserArticles(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error)
	IncrementViewCount(ctx context.Context, id uint) (int, error)
	GetCategories(ctx context.Context) ([]*model.Category, error)
	GetTags(ctx context.Context) ([]*model.Tag, error)
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

	// 构建文章模型
	article := &model.Article{
		UserID:       req.UserID,
		Title:        req.Title,
		Slug:         slug,
		Summary:      req.Summary,
		Content:      req.Content,
		CoverImage:   req.CoverImage,
		CategoryID:   req.CategoryID,
		IsPublished:  req.IsPublished,
		IsFeatured:   req.IsFeatured,
		AllowComment: req.AllowComment,
	}

	// 设置发布时间
	if req.IsPublished {
		now := time.Now()
		article.PublishedAt = &now
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
		// 存入本地缓存
		cache.LocalCacheSet(key, article)
		return article, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Article), nil
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
		// 存入本地缓存
		cache.LocalCacheSet(key, article)
		return article, nil
	})
	if err != nil {
		return nil, err
	}
	return result.(*model.Article), nil
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
	article.IsPublished = req.IsPublished
	article.IsFeatured = req.IsFeatured
	article.AllowComment = req.AllowComment

	// 如果是首次发布
	if req.IsPublished && article.PublishedAt == nil {
		now := time.Now()
		article.PublishedAt = &now
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
