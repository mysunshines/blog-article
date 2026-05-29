package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mysunshines/blog-article/internal/model"
	apperrors "github.com/mysunshines/blog-article/pkg/errors"
	"github.com/mysunshines/gocommon/pool"

	"gorm.io/gorm"
)

type ArticleRepository interface {
	Create(ctx context.Context, article *model.Article) error
	GetByID(ctx context.Context, id uint) (*model.Article, error)
	GetBySlug(ctx context.Context, slug string) (*model.Article, error)
	Update(ctx context.Context, article *model.Article) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error)
	Search(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error)
	GetByUserID(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error)
	IncrementViewCount(ctx context.Context, id uint) (int, error)
}

type articleRepository struct {
	db *gorm.DB
}

func NewArticleRepository(db *gorm.DB) ArticleRepository {
	return &articleRepository{db: db}
}

func (r *articleRepository) Create(ctx context.Context, article *model.Article) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 创建文章
		if err := tx.Create(article).Error; err != nil {
			return apperrors.ArticleCreateFailed(err)
		}

		// 处理标签
		for _, tagName := range extractTagNames(article.Tags) {
			tag := &model.Tag{}
			slug := generateSlug(tagName)

			// 查找或创建标签
			if err := tx.Where("name = ?", tagName).FirstOrCreate(tag, model.Tag{Name: tagName, Slug: slug}).Error; err != nil {
				continue
			}

			// 关联文章和标签
			articleTag := &model.ArticleTag{
				ArticleID: article.ID,
				TagID:     tag.ID,
			}
			tx.Create(articleTag)

			// 更新标签文章数
			tx.Model(&model.Tag{}).Where("id = ?", tag.ID).
				Update("article_count", gorm.Expr("article_count + 1"))
		}

		// 更新分类文章数
		if article.CategoryID > 0 {
			tx.Model(&model.Category{}).Where("id = ?", article.CategoryID).
				Update("article_count", gorm.Expr("article_count + 1"))
		}

		return nil
	})

	return err
}

func (r *articleRepository) GetByID(ctx context.Context, id uint) (*model.Article, error) {
	var article model.Article
	result := r.db.WithContext(ctx).
		Preload("User").
		Preload("Category").
		Preload("Tags").
		First(&article, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ArticleNotFound()
		}
		return nil, apperrors.Internal("获取文章失败", result.Error)
	}

	return &article, nil
}

func (r *articleRepository) GetBySlug(ctx context.Context, slug string) (*model.Article, error) {
	var article model.Article
	result := r.db.WithContext(ctx).
		Preload("User").
		Preload("Category").
		Preload("Tags").
		Where("slug = ?", slug).
		First(&article)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ArticleNotFound()
		}
		return nil, apperrors.Internal("获取文章失败", result.Error)
	}

	return &article, nil
}

func (r *articleRepository) Update(ctx context.Context, article *model.Article) error {
	result := r.db.WithContext(ctx).Save(article)
	if result.Error != nil {
		return apperrors.ArticleUpdateFailed(result.Error)
	}
	return nil
}

func (r *articleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取文章信息
		var article model.Article
		if err := tx.First(&article, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ArticleNotFound()
			}
			return apperrors.ArticleDeleteFailed(err)
		}

		// 删除标签关联
		tx.Where("article_id = ?", id).Delete(&model.ArticleTag{})

		// 更新分类文章数
		if article.CategoryID > 0 {
			tx.Model(&model.Category{}).Where("id = ?", article.CategoryID).
				Update("article_count", gorm.Expr("article_count - 1"))
		}

		// 删除文章
		if err := tx.Delete(&model.Article{}, id).Error; err != nil {
			return apperrors.ArticleDeleteFailed(err)
		}

		return nil
	})
}

func (r *articleRepository) List(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{})

	// 条件筛选
	if req.CategoryID > 0 {
		baseQuery = baseQuery.Where("category_id = ?", req.CategoryID)
	}
	if req.UserID > 0 {
		baseQuery = baseQuery.Where("user_id = ?", req.UserID)
	}
	if req.IsPublished {
		baseQuery = baseQuery.Where("is_published = ?", true)
	}
	if req.Tag != "" {
		baseQuery = baseQuery.Joins("JOIN article_tags ON articles.id = article_tags.article_id").
			Joins("JOIN tags ON article_tags.tag_id = tags.id").
			Where("tags.name = ?", req.Tag)
	}

	// 分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 || size > 100 {
		size = 10
	}
	offset := (page - 1) * size

	// 排序
	orderBy := "created_at DESC"
	switch req.OrderBy {
	case "view_count":
		orderBy = "view_count DESC"
	case "like_count":
		orderBy = "like_count DESC"
	case "published_at":
		orderBy = "published_at DESC"
	}

	// 并行：COUNT + SELECT 互不依赖，用 Session 克隆查询避免状态污染
	results := pool.Go(ctx,
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).Count(&total).Error
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).
				Preload("User").
				Preload("Category").
				Preload("Tags").
				Order(orderBy).
				Offset(int(offset)).
				Limit(int(size)).
				Find(&articles).Error
		},
	)

	if results[0].Err != nil {
		return nil, 0, apperrors.Internal("查询文章总数失败", results[0].Err)
	}
	if results[1].Err != nil {
		return nil, 0, apperrors.Internal("查询文章列表失败", results[1].Err)
	}

	return articles, total, nil
}

func (r *articleRepository) Search(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return nil, 0, apperrors.BadRequest("搜索关键词不能为空")
	}

	searchPattern := "%" + keyword + "%"

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("is_published = ? AND (title LIKE ? OR content LIKE ? OR summary LIKE ?)",
			true, searchPattern, searchPattern, searchPattern)

	// 并行：COUNT + SELECT
	results := pool.Go(ctx,
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).Count(&total).Error
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).
				Preload("User").
				Preload("Category").
				Order("created_at DESC").
				Offset(int(offset)).
				Limit(int(size)).
				Find(&articles).Error
		},
	)

	if results[0].Err != nil {
		return nil, 0, apperrors.Internal("搜索文章总数失败", results[0].Err)
	}
	if results[1].Err != nil {
		return nil, 0, apperrors.Internal("搜索文章失败", results[1].Err)
	}

	return articles, total, nil
}

func (r *articleRepository) GetByUserID(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	offset := (page - 1) * size

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{}).Where("user_id = ?", userID)

	// 并行：COUNT + SELECT
	results := pool.Go(ctx,
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).Count(&total).Error
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).
				Preload("User").
				Preload("Category").
				Preload("Tags").
				Order("created_at DESC").
				Offset(int(offset)).
				Limit(int(size)).
				Find(&articles).Error
		},
	)

	if results[0].Err != nil {
		return nil, 0, apperrors.Internal("查询用户文章总数失败", results[0].Err)
	}
	if results[1].Err != nil {
		return nil, 0, apperrors.Internal("查询用户文章失败", results[1].Err)
	}

	return articles, total, nil
}

func (r *articleRepository) IncrementViewCount(ctx context.Context, id uint) (int, error) {
	result := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1"))

	if result.Error != nil {
		return 0, apperrors.Internal("更新浏览数失败", result.Error)
	}

	var article model.Article
	r.db.WithContext(ctx).First(&article, id)
	return article.ViewCount, nil
}

// 辅助函数
func extractTagNames(tags []model.Tag) []string {
	names := make([]string, len(tags))
	for i, t := range tags {
		names[i] = t.Name
	}
	return names
}

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

	return fmt.Sprintf("%s-%d", slug, time.Now().Unix())
}
