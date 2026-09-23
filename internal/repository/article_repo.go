package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	apperrors "github.com/mysunshines/blog-article/internal/errors"
	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/gocommon/pool"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 列表/详情转换 ConvertToProtoArticle 仅读取 User.Username / Category.Name / Tags[i].Name，
// 因此关联预加载只取主键 + 所需列，避免把 users/categories/tags 表的多余列（nickname/avatar/description 等）扫进内存。
var (
	preloadUser     = func(db *gorm.DB) *gorm.DB { return db.Select("id, username") }
	preloadCategory = func(db *gorm.DB) *gorm.DB { return db.Select("id, name") }
	preloadTags     = func(db *gorm.DB) *gorm.DB { return db.Select("id, name") }
)

type ArticleRepository interface {
	Create(ctx context.Context, article *model.Article) error
	GetByID(ctx context.Context, id uint) (*model.Article, error)
	GetArticlesByIDs(ctx context.Context, ids []uint) ([]*model.ArticleEnrich, error)
	GetBySlug(ctx context.Context, slug string) (*model.Article, error)
	Update(ctx context.Context, article *model.Article) error
	UpdateStatus(ctx context.Context, id uint, cols map[string]interface{}) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error)
	Search(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error)
	GetByUserID(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error)
	// GetPublishedByUserID 作者主页用：只返回已发布文章（公开可见）
	GetPublishedByUserID(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error)
	IncrementViewCount(ctx context.Context, id uint) (int, error)
	IncrementViewCountBy(ctx context.Context, id uint, delta int64) error
	GetViewCount(ctx context.Context, id uint) (int, error)
	LikeArticle(ctx context.Context, articleID, userID uint) (likeCount int, liked bool, err error)
	CancelLikeArticle(ctx context.Context, articleID, userID uint) (likeCount int, liked bool, err error)
	GetLikeStatus(ctx context.Context, articleID, userID uint) (likeCount int, liked bool, err error)
	AdminList(ctx context.Context, req *model.AdminListArticlesRequest) ([]*model.Article, int64, error)
	GetCategory(ctx context.Context, id uint) (*model.Category, error)
	CountArticlesByCategory(ctx context.Context, categoryID uint, status string) (int64, error)
	DeleteCategory(ctx context.Context, id uint) error

	// 排行榜历史回填
	ListArticlesForRanking(ctx context.Context, lastID, limit uint) ([]model.Article, error)
	CountPublishedArticlesByAuthor(ctx context.Context) (map[uint]int64, error)
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

		// 批量处理标签：将逐标签 FirstOrCreate+Create+Update（N×3 次 SQL，N 为标签数）
		// 压缩为 1 次批量查询 + 1 次批量插入 + 1 次批量关联 + 1 次批量计数。
		// 保持历史"尽力而为"语义：标签环节失败不阻塞文章创建。
		tagNames := extractTagNames(article.Tags)
		if len(tagNames) > 0 {
			// 1) 一次往返批量查询已存在的标签
			var existing []*model.Tag
			queryErr := tx.Where("name IN ?", tagNames).Find(&existing).Error
			existingByName := make(map[string]*model.Tag, len(existing))
			for _, t := range existing {
				existingByName[t.Name] = t
			}

			// 2) 去重（同一文章内重复标签只计一次），并区分新增/已存在
			seen := make(map[string]struct{}, len(tagNames))
			var newTags []*model.Tag
			var tagIDs []uint
			for _, name := range tagNames {
				if _, dup := seen[name]; dup {
					continue
				}
				seen[name] = struct{}{}
				if t, ok := existingByName[name]; ok {
					tagIDs = append(tagIDs, t.ID)
					continue
				}
				newTags = append(newTags, &model.Tag{Name: name, Slug: generateSlug(name)})
			}

			// 3) 批量插入新标签 + 批量建立关联
			if queryErr == nil && len(newTags) > 0 {
				if err := tx.Create(&newTags).Error; err == nil {
					articleTags := make([]*model.ArticleTag, 0, len(newTags))
					for _, t := range newTags {
						tagIDs = append(tagIDs, t.ID)
						articleTags = append(articleTags, &model.ArticleTag{ArticleID: article.ID, TagID: t.ID})
					}
					_ = tx.Create(&articleTags).Error // 尽力而为
				}
			}

			// 4) 一次批量更新所有涉及标签的文章计数
			if len(tagIDs) > 0 {
				_ = tx.Model(&model.Tag{}).
					Where("id IN ?", tagIDs).
					Update("article_count", gorm.Expr("article_count + 1")).Error // 尽力而为
			}
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
		Preload("User", preloadUser).
		Preload("Category", preloadCategory).
		Preload("Tags", preloadTags).
		First(&article, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ArticleNotFound()
		}
		return nil, apperrors.Internal("获取文章失败", result.Error)
	}

	r.fillAuthor(&article)
	return &article, nil
}

// fillAuthor 处理匿名作者：当作者被软删除（Preload 后 User 为空）时，
// 用占位对象填充，避免前端显示空昵称。
func (r *articleRepository) fillAuthor(article *model.Article) {
	if article == nil {
		return
	}
	if article.User.ID == 0 {
		article.User = model.User{
			ID:       0,
			Username: "anonymous",
			Nickname: "匿名用户",
		}
	}
}

func (r *articleRepository) fillAuthors(articles []*model.Article) {
	for _, a := range articles {
		r.fillAuthor(a)
	}
}

// GetArticlesByIDs 按文章 ID 批量回填展示字段（标题/封面/作者名），供 decorator 使用。
// 仅取 article-service 拥有的字段，不感知 ranking 展示语义；LEFT JOIN users 取作者名。
func (r *articleRepository) GetArticlesByIDs(ctx context.Context, ids []uint) ([]*model.ArticleEnrich, error) {
	if len(ids) == 0 {
		return []*model.ArticleEnrich{}, nil
	}
	var list []*model.ArticleEnrich
	if err := r.db.WithContext(ctx).
		Table("articles a").
		Select("a.id, a.title, a.slug, a.cover_image, a.user_id, u.username as author_name").
		Joins("LEFT JOIN users u ON u.id = a.user_id").
		Where("a.id IN ? AND a.deleted_at IS NULL", ids).
		Find(&list).Error; err != nil {
		return nil, apperrors.Internal("批量获取文章展示信息失败", err)
	}
	return list, nil
}

func (r *articleRepository) GetBySlug(ctx context.Context, slug string) (*model.Article, error) {
	var article model.Article
	result := r.db.WithContext(ctx).
		Preload("User", preloadUser).
		Preload("Category", preloadCategory).
		Preload("Tags", preloadTags).
		Where("slug = ?", slug).
		First(&article)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, apperrors.ArticleNotFound()
		}
		return nil, apperrors.Internal("获取文章失败", result.Error)
	}

	r.fillAuthor(&article)
	return &article, nil
}

func (r *articleRepository) Update(ctx context.Context, article *model.Article) error {
	// 仅更新主表字段，排除预加载的关联对象（User/Category/Tags），
	// 避免 Save 误将半截关联数据写回关联表。
	result := r.db.WithContext(ctx).Model(article).Omit("User", "Category", "Tags").Updates(article)
	if result.Error != nil {
		return apperrors.ArticleUpdateFailed(result.Error)
	}
	return nil
}

// UpdateStatus 仅更新文章主表的指定列（状态机专用），
// 使用 map 可正确写入零值（如清空 reject_reason），且不会触碰关联表。
func (r *articleRepository) UpdateStatus(ctx context.Context, id uint, cols map[string]interface{}) error {
	cols["updated_at"] = time.Now()
	result := r.db.WithContext(ctx).Model(&model.Article{}).Where("id = ?", id).Updates(cols)
	if result.Error != nil {
		return apperrors.ArticleUpdateFailed(result.Error)
	}
	return nil
}

func (r *articleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 获取文章信息（软删除行需 Unscoped 才能读到，用于清理关联计数）
		var article model.Article
		if err := tx.Unscoped().First(&article, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ArticleNotFound()
			}
			return apperrors.ArticleDeleteFailed(err)
		}

		// 软删除：gorm 自动设置 deleted_at，数据保留可恢复
		if err := tx.Delete(&model.Article{}, id).Error; err != nil {
			return apperrors.ArticleDeleteFailed(err)
		}

		// 清理标签关联（物理删除，避免软删恢复后残留旧标签）
		tx.Where("article_id = ?", id).Delete(&model.ArticleTag{})

		// 更新分类文章数
		if article.CategoryID > 0 {
			tx.Model(&model.Category{}).Where("id = ?", article.CategoryID).
				Update("article_count", gorm.Expr("article_count - 1"))
		}

		return nil
	})
}

func (r *articleRepository) List(ctx context.Context, req *model.ListArticlesRequest) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("status IN ?", []string{model.ArticleStatusPublished, model.ArticleStatusPending})

	// 条件筛选
	if req.CategoryID > 0 {
		baseQuery = baseQuery.Where("category_id = ?", req.CategoryID)
	}
	if req.UserID > 0 {
		baseQuery = baseQuery.Where("user_id = ?", req.UserID)
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

	// 排序：置顶推荐(is_featured)优先，其次按所选排序字段
	orderBy := "created_at DESC"
	switch req.OrderBy {
	case "view_count":
		orderBy = "view_count DESC"
	case "like_count":
		orderBy = "like_count DESC"
	case "published_at":
		orderBy = "published_at DESC"
	}
	// 置顶文章永远排在最前
	orderClause := "is_featured DESC, " + orderBy

	// 并行：COUNT + SELECT 互不依赖，用 Session 克隆查询避免状态污染
	results := pool.Go(ctx,
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).Count(&total).Error
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).
				Preload("User", preloadUser).
				Preload("Category", preloadCategory).
				Preload("Tags", preloadTags).
				Order(orderClause).
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

	r.fillAuthors(articles)
	return articles, total, nil
}

func (r *articleRepository) Search(ctx context.Context, req *model.SearchArticlesRequest) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	keyword := strings.TrimSpace(req.Keyword)
	if keyword == "" {
		return nil, 0, apperrors.BadRequest("搜索关键词不能为空")
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	statuses := []string{model.ArticleStatusPublished, model.ArticleStatusPending}

	// 匹配范围：仅 title / summary，不再扫描正文 content（LONGTEXT，匹配代价高且非搜索诉求）。
	// 多字符（含中文）走 ngram FULLTEXT 索引，快速且按相关度排序；
	// 单字符 ngram 无法分词会无命中，回退到 title/summary 的 LIKE 模糊匹配。
	var where string
	var whereArgs []interface{}
	var orderClause clause.OrderBy
	if utf8.RuneCountInString(keyword) <= 1 {
		pattern := "%" + keyword + "%"
		where = "status IN ? AND (title LIKE ? OR summary LIKE ?)"
		whereArgs = []interface{}{statuses, pattern, pattern}
		orderClause = clause.OrderBy{Expression: clause.Expr{SQL: "created_at DESC"}}
	} else {
		match := "MATCH(title, summary) AGAINST (? IN NATURAL LANGUAGE MODE)"
		where = "status IN ? AND " + match
		whereArgs = []interface{}{statuses, keyword}
		orderClause = clause.OrderBy{Expression: clause.Expr{
			SQL:  "MATCH(title, summary) AGAINST (? IN NATURAL LANGUAGE MODE) DESC, created_at DESC",
			Vars: []interface{}{keyword},
		}}
	}

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{}).
		Where(where, whereArgs...)

	// 并行：COUNT + SELECT
	results := pool.Go(ctx,
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).Count(&total).Error
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).
				Clauses(orderClause).
				Preload("User", preloadUser).
				Preload("Category", preloadCategory).
				Preload("Tags", preloadTags).
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

	r.fillAuthors(articles)
	return articles, total, nil
}

// GetPublishedByUserID 作者主页用：按作者列出「已发布」文章（公开可见）。
// 与 GetByUserID（我的文章，含草稿等全部状态）区分，避免把未发布内容暴露给访客。
func (r *articleRepository) GetPublishedByUserID(ctx context.Context, userID uint, page, size uint) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}
	offset := (page - 1) * size

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("user_id = ? AND status = ?", userID, model.ArticleStatusPublished)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := baseQuery.
		Preload("User", preloadUser).
		Preload("Category", preloadCategory).
		Order("published_at DESC, created_at DESC").
		Offset(int(offset)).
		Limit(int(size)).
		Find(&articles).Error; err != nil {
		return nil, 0, err
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
				Preload("User", preloadUser).
				Preload("Category", preloadCategory).
				Preload("Tags", preloadTags).
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

	r.fillAuthors(articles)
	return articles, total, nil
}

// IncrementViewCount 浏览计数降级路径：Redis 不可用时直接更新 DB（保留原行为）。
// 正常路径由 service 走 Redis INCR + 后台定时批量落库（见 IncrementViewCountBy）。
func (r *articleRepository) IncrementViewCount(ctx context.Context, id uint) (int, error) {
	if err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error; err != nil {
		return 0, apperrors.Internal("更新浏览数失败", err)
	}
	return r.GetViewCount(ctx, id)
}

// IncrementViewCountBy 批量落库浏览增量（后台协程每 30s 调用一次，delta 为
// 窗口内累计的浏览数，一次 UPDATE 合并 N 次浏览，避免逐次往返）。
func (r *articleRepository) IncrementViewCountBy(ctx context.Context, id uint, delta int64) error {
	if delta <= 0 {
		return nil
	}
	if err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", delta)).Error; err != nil {
		return apperrors.Internal("批量更新浏览数失败", err)
	}
	return nil
}

// GetViewCount 仅读取单列 view_count（浏览计数高频路径使用，避免全行查询）。
func (r *articleRepository) GetViewCount(ctx context.Context, id uint) (int, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("id = ?", id).
		Pluck("view_count", &cnt).Error; err != nil {
		return 0, apperrors.Internal("读取浏览数失败", err)
	}
	return int(cnt), nil
}

// LikeArticle 点赞：已点赞则幂等返回，未点赞则插入记录并 like_count+1
func (r *articleRepository) LikeArticle(ctx context.Context, articleID, userID uint) (int, bool, error) {
	var existing model.ArticleLike
	err := r.db.WithContext(ctx).Where("article_id = ? AND user_id = ?", articleID, userID).First(&existing).Error
	if err == nil {
		// 已点赞，保持幂等
		return r.countAndReturn(ctx, articleID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, false, apperrors.Internal("查询点赞状态失败", err)
	}

	// 插入点赞记录（唯一索引保证并发下不重复）
	like := &model.ArticleLike{ArticleID: articleID, UserID: userID}
	if err := r.db.WithContext(ctx).Create(like).Error; err != nil {
		// 并发冲突（Duplicate entry）视为已点赞
		if strings.Contains(err.Error(), "Duplicate") || strings.Contains(err.Error(), "1062") {
			return r.countAndReturn(ctx, articleID)
		}
		return 0, false, apperrors.Internal("点赞失败", err)
	}
	if err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error; err != nil {
		return 0, false, apperrors.Internal("更新点赞数失败", err)
	}
	return r.countAndReturn(ctx, articleID)
}

// CancelLikeArticle 取消点赞：删除记录并 like_count-1（下限 0）
func (r *articleRepository) CancelLikeArticle(ctx context.Context, articleID, userID uint) (int, bool, error) {
	res := r.db.WithContext(ctx).Where("article_id = ? AND user_id = ?", articleID, userID).Delete(&model.ArticleLike{})
	if res.Error != nil {
		return 0, false, apperrors.Internal("取消点赞失败", res.Error)
	}
	if res.RowsAffected > 0 {
		r.db.WithContext(ctx).Model(&model.Article{}).
			Where("id = ?", articleID).
			UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)"))
	}
	return r.countAndReturn(ctx, articleID)
}

// GetLikeStatus 查询当前点赞数与登录用户是否已点赞
func (r *articleRepository) GetLikeStatus(ctx context.Context, articleID, userID uint) (int, bool, error) {
	var cnt int64
	if err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("id = ?", articleID).
		Pluck("like_count", &cnt).Error; err != nil {
		return 0, false, apperrors.Internal("查询点赞数失败", err)
	}
	liked := false
	if userID > 0 {
		var n int64
		r.db.WithContext(ctx).Model(&model.ArticleLike{}).
			Where("article_id = ? AND user_id = ?", articleID, userID).
			Count(&n)
		liked = n > 0
	}
	return int(cnt), liked, nil
}

// countAndReturn 读取文章最新点赞数并返回（liked=true 表示已点赞）
func (r *articleRepository) countAndReturn(ctx context.Context, articleID uint) (int, bool, error) {
	var article model.Article
	if err := r.db.WithContext(ctx).First(&article, articleID).Error; err != nil {
		return 0, true, apperrors.Internal("读取点赞数失败", err)
	}
	return article.LikeCount, true, nil
}

func (r *articleRepository) AdminList(ctx context.Context, req *model.AdminListArticlesRequest) ([]*model.Article, int64, error) {
	var articles []*model.Article
	var total int64

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 || size > 100 {
		size = 10
	}
	offset := (page - 1) * size

	baseQuery := r.db.WithContext(ctx).Model(&model.Article{})
	if req.Status != "" {
		baseQuery = baseQuery.Where("status = ?", req.Status)
	}

	results := pool.Go(ctx,
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).Count(&total).Error
		},
		func(ctx context.Context) (interface{}, error) {
			return nil, baseQuery.Session(&gorm.Session{}).
				Preload("User", preloadUser).
				Preload("Category", preloadCategory).
				Preload("Tags", preloadTags).
				Order("created_at DESC").
				Offset(int(offset)).
				Limit(int(size)).
				Find(&articles).Error
		},
	)

	if results[0].Err != nil {
		return nil, 0, apperrors.Internal("查询后台文章列表失败", results[0].Err)
	}
	if results[1].Err != nil {
		return nil, 0, apperrors.Internal("查询后台文章列表失败", results[1].Err)
	}

	r.fillAuthors(articles)
	return articles, total, nil
}

// GetCategory 获取单个分类
func (r *articleRepository) GetCategory(ctx context.Context, id uint) (*model.Category, error) {
	var category model.Category
	if err := r.db.WithContext(ctx).First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NotFound("分类不存在")
		}
		return nil, apperrors.Internal("获取分类失败", err)
	}
	return &category, nil
}

// CountArticlesByCategory 统计某分类下指定状态的文章数量
func (r *articleRepository) CountArticlesByCategory(ctx context.Context, categoryID uint, status string) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Article{}).
		Where("category_id = ? AND status = ?", categoryID, status).
		Count(&count).Error; err != nil {
		return 0, apperrors.Internal("统计分类文章数失败", err)
	}
	return count, nil
}

// DeleteCategory 软删除分类
func (r *articleRepository) DeleteCategory(ctx context.Context, id uint) error {
	if err := r.db.WithContext(ctx).Delete(&model.Category{}, id).Error; err != nil {
		return apperrors.Internal("删除分类失败", err)
	}
	return nil
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

// ListArticlesForRanking 游标分页拉取文章聚合字段（id/user_id/三计数/status/软删标记），
// 用于排行榜历史回填，避免一次性加载整表到内存。
func (r *articleRepository) ListArticlesForRanking(ctx context.Context, lastID, limit uint) ([]model.Article, error) {
	var articles []model.Article
	if err := r.db.WithContext(ctx).
		Select("id, user_id, view_count, like_count, comment_count, status, deleted_at").
		Where("id > ?", lastID).
		Order("id ASC").
		Limit(int(limit)).
		Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

// CountPublishedArticlesByAuthor 统计每个作者的「已发布且未删除」文章数（作者榜权威值）。
func (r *articleRepository) CountPublishedArticlesByAuthor(ctx context.Context) (map[uint]int64, error) {
	type authorCount struct {
		UserID uint
		Cnt    int64
	}
	var rows []authorCount
	if err := r.db.WithContext(ctx).
		Model(&model.Article{}).
		Select("user_id, COUNT(*) as cnt").
		Where("status = ? AND deleted_at IS NULL", model.ArticleStatusPublished).
		Group("user_id").
		Find(&rows).Error; err != nil {
		return nil, err
	}
	m := make(map[uint]int64, len(rows))
	for _, row := range rows {
		m[row.UserID] = row.Cnt
	}
	return m, nil
}
