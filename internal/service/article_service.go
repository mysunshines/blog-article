package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/mysunshines/blog-article/internal/client"
	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/blog-article/internal/repository"
	"github.com/mysunshines/blog-article/internal/errors"
	"github.com/mysunshines/gocommon/cache"
	"github.com/mysunshines/gocommon/util"

	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

// 浏览计数异步落库相关常量。
// 高频浏览路径先 Redis INCR 记录窗口增量（1 次 Redis 写），由后台协程每
// viewCountFlushInterval 批量写回 MySQL（1 次 UPDATE 合并 N 次浏览），
// 避免"每次浏览一次 UPDATE + 一次 SELECT"的两趟 DB 往返（历史实现）。
const (
	// viewCountIncrKeyFmt Redis 窗口增量 key：自上次落库以来累计的浏览数。
	viewCountIncrKeyFmt = "article:view:incr:%d"
	// viewCountBaseKeyFmt DB 基础浏览数的短 TTL 缓存：读取合并时减少对 DB 的访问。
	viewCountBaseKeyFmt = "article:view:base:%d"
	// viewCountFlushInterval 浏览增量批量落库周期。
	viewCountFlushInterval = 30 * time.Second
	// viewCountBaseTTL 基础浏览数缓存 TTL（落库会同时失效该缓存，TTL 仅兜底）。
	viewCountBaseTTL = 5 * time.Minute
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

	// 浏览计数异步落库：viewPending 累计本实例待写回 DB 的浏览增量（权威值），
	// 由 flushViewCountLoop 每 viewCountFlushInterval 批量落库后清零。
	viewMu      sync.Mutex
	viewPending map[uint]int64
}

func NewArticleService(repo repository.ArticleRepository, db *gorm.DB) ArticleService {
	s := &articleService{
		repo:        repo,
		db:          db,
		viewPending: make(map[uint]int64),
	}
	// 启动浏览计数后台批量落库协程
	go s.flushViewCountLoop()
	return s
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

	// 作者信息聚合：拉取文章作者（独立 gRPC 调用），失败降级（作者名留空），
	// 不阻断详情返回。GetUser 依赖 article.UserID，必须在 GetByID 之后执行。
	if u, err := client.GetUser(ctx, article.UserID); err == nil && u != nil {
		article.AuthorName = u.Username
		article.User = model.User{ID: uint(u.Id), Username: u.Username, Nickname: u.Nickname, Avatar: u.Avatar}
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

	// 返回更新后的文章：内存中的 article 已是更新后的最新值（含 preload），
	// 无需再查一次 DB（历史实现会重复 GetByID，同一请求内两次查询同一数据）
	return article, nil
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

// IncrementViewCount 浏览计数自增。
// 高频路径：先 Redis INCR 记录窗口增量（1 次 Redis 写，fail-fast），本地同时累计
// 待落库增量，由后台协程每 viewCountFlushInterval 批量写回 MySQL；
// 返回值 = DB 基础浏览数 + 窗口增量（近似实时）。Redis 不可用时降级为 DB 直更。
func (s *articleService) IncrementViewCount(ctx context.Context, id uint) (int, error) {
	delta, err := cache.Incr(ctx, fmt.Sprintf(viewCountIncrKeyFmt, id))
	if err != nil {
		// 降级：Redis 不可用（fail-fast 1s 内返回），直接更新 DB，保留原行为
		return s.repo.IncrementViewCount(ctx, id)
	}

	s.viewMu.Lock()
	s.viewPending[id]++
	s.viewMu.Unlock()

	base, err := s.viewBase(ctx, id)
	if err != nil {
		return 0, err
	}
	return base + int(delta), nil
}

// viewBase 读取 DB 基础浏览数，带短 TTL Redis 缓存；缓存 miss 或读取失败时查库一次并回填。
func (s *articleService) viewBase(ctx context.Context, id uint) (int, error) {
	key := fmt.Sprintf(viewCountBaseKeyFmt, id)
	if v, err := cache.Get(ctx, key); err == nil {
		if n, convErr := strconv.Atoi(v); convErr == nil {
			return n, nil
		}
	}
	// miss 或读取失败：查库回填（回填失败仅影响实时性，不影响返回值）
	base, err := s.repo.GetViewCount(ctx, id)
	if err != nil {
		return 0, err
	}
	_ = cache.Set(ctx, key, base, viewCountBaseTTL)
	return base, nil
}

// flushViewCountLoop 后台定时批量落库浏览增量。
func (s *articleService) flushViewCountLoop() {
	ticker := time.NewTicker(viewCountFlushInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.flushViewCounts()
	}
}

// flushViewCounts 把本地累计的待落库增量批量写回 MySQL（1 次 UPDATE 合并 N 次浏览），
// 成功后清理 Redis 窗口增量与基础值缓存，使读取重新以 DB 为基准。
// 落库失败则回填待处理集合，下个周期重试；Redis 清理失败仅影响读取合并实时性，DB 已正确。
func (s *articleService) flushViewCounts() {
	s.viewMu.Lock()
	if len(s.viewPending) == 0 {
		s.viewMu.Unlock()
		return
	}
	pending := s.viewPending
	s.viewPending = make(map[uint]int64, len(pending))
	s.viewMu.Unlock()

	for id, delta := range pending {
		flushCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		err := s.repo.IncrementViewCountBy(flushCtx, id, delta)
		cancel()
		if err != nil {
			// 落库失败：回填待处理集合，下个周期重试
			s.viewMu.Lock()
			s.viewPending[id] += delta
			s.viewMu.Unlock()
			continue
		}
		// 清理 Redis 窗口增量与基础值缓存：读取将重新以 DB 值为基准
		_ = cache.Delete(flushCtx, fmt.Sprintf(viewCountIncrKeyFmt, id), fmt.Sprintf(viewCountBaseKeyFmt, id))
	}
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

// applyStatusAndReturn 执行状态更新 + 失效缓存，并把写入 DB 的状态机列同步回内存中的
// article，直接返回该对象，避免状态机方法"更新后再 GetByID 一次"的重复 DB 往返。
// cols 的 key 需与 model.Article 字段名一致（如 "status"/"published_at"）。
func (s *articleService) applyStatusAndReturn(ctx context.Context, id uint, article *model.Article, cols map[string]interface{}) (*model.Article, error) {
	if err := s.repo.UpdateStatus(ctx, id, cols); err != nil {
		return nil, err
	}
	s.invalidateArticleCache(id, article.Slug)
	if v, ok := cols["status"]; ok {
		if st, ok := v.(string); ok {
			article.Status = st
		}
	}
	if v, ok := cols["published_at"]; ok {
		if pt, ok := v.(*time.Time); ok {
			article.PublishedAt = pt
		}
	}
	if v, ok := cols["reject_reason"]; ok {
		if rr, ok := v.(string); ok {
			article.RejectReason = rr
		}
	}
	if v, ok := cols["offline_reason"]; ok {
		if or, ok := v.(string); ok {
			article.OfflineReason = or
		}
	}
	return article, nil
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
	return s.applyStatusAndReturn(ctx, id, article, map[string]interface{}{
		"status":        model.ArticleStatusPublished,
		"published_at":  &now,
		"reject_reason": "",
	})
}

func (s *articleService) RejectArticle(ctx context.Context, id uint, reason string) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusPending {
		return nil, errors.BadRequest("仅待审核文章可以拒绝")
	}

	return s.applyStatusAndReturn(ctx, id, article, map[string]interface{}{
		"status":        model.ArticleStatusRejected,
		"reject_reason": reason,
	})
}

func (s *articleService) OfflineArticle(ctx context.Context, id uint, reason string) (*model.Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if article.Status != model.ArticleStatusPublished {
		return nil, errors.BadRequest("仅已发布文章可以下线")
	}

	return s.applyStatusAndReturn(ctx, id, article, map[string]interface{}{
		"status":         model.ArticleStatusOffline,
		"offline_reason": reason,
	})
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
	return s.applyStatusAndReturn(ctx, id, article, cols)
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

	return s.applyStatusAndReturn(ctx, id, article, map[string]interface{}{
		"status":         model.ArticleStatusPending,
		"reject_reason":  "",
		"offline_reason": "",
	})
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
	// 内存对象已是最新值（含 preload），无需再查一次 DB
	return article, nil
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
