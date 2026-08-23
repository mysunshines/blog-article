package v1

import (
	"context"

	"github.com/mysunshines/blog-article/internal/errors"
	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/blog-article/internal/service"
	article "github.com/mysunshines/blog-article/proto/pb"
	"github.com/mysunshines/gocommon/constants"
	commonmiddleware "github.com/mysunshines/gocommon/middleware"
	"github.com/mysunshines/gocommon/util"

	"github.com/sony/gobreaker"
)

// errCode 从 service 返回的错误中提取业务码并映射为 proto ArticleErrorCode。
// 若错误为 *errors.AppError，直接取其 Code（与 proto 枚举值对齐）；
// 否则（未预期错误）回退为 ARTICLE_INTERNAL_ERROR。
func errCode(err error) uint32 {
	var ae *errors.AppError
	if errors.As(err, &ae) {
		return uint32(ae.Code)
	}
	return uint32(article.ArticleErrorCode_ARTICLE_INTERNAL_ERROR)
}

// GrpcArticleHandler gRPC 文章处理器适配器（支持熔断）
type GrpcArticleHandler struct {
	article.UnimplementedArticleServiceServer
	Svc service.ArticleService
	Cb  *gobreaker.CircuitBreaker
}

func (h *GrpcArticleHandler) CreateArticle(ctx context.Context, req *article.CreateArticleRequest) (*article.CreateArticleResponse, error) {
	// 强制要求已登录，并使用经 gRPC 拦截器校验过的身份，杜绝伪造/越权 user_id（IDOR）
	uid, err := commonmiddleware.RequireGRPCAuth(ctx)
	if err != nil {
		return nil, err
	}
	createdArticle, err := h.Svc.CreateArticle(ctx, &model.CreateArticleRequest{
		UserID:       uid,
		Title:        req.Title,
		Content:      req.Content,
		Summary:      req.Summary,
		CoverImage:   req.CoverImage,
		CategoryID:   uint(req.CategoryId),
		Tags:         req.Tags,
		IsFeatured:   req.IsFeatured,
		AllowComment: req.AllowComment,
		IsPublished:  req.IsPublished,
	})

	if err != nil {
		return &article.CreateArticleResponse{
			Code:    uint32(article.ArticleErrorCode_ARTICLE_CREATE_FAILED),
			Message: err.Error(),
		}, nil
	}

	return &article.CreateArticleResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
		Article: ConvertToProtoArticle(createdArticle),
	}, nil
}

func (h *GrpcArticleHandler) GetArticle(ctx context.Context, req *article.GetArticleRequest) (*article.GetArticleResponse, error) {
	foundArticle, err := h.Svc.GetArticle(ctx, uint(req.ArticleId))
	if err != nil {
		return &article.GetArticleResponse{
			Code:    uint32(article.ArticleErrorCode_ARTICLE_NOT_FOUND),
			Message: "Article not found",
		}, nil
	}

	return &article.GetArticleResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
		Article: ConvertToProtoArticle(foundArticle),
	}, nil
}

func (h *GrpcArticleHandler) ListArticles(ctx context.Context, req *article.ListArticlesRequest) (*article.ListArticlesResponse, error) {
	result, total, err := h.Svc.ListArticles(ctx, &model.ListArticlesRequest{
		Page:        uint(req.Page),
		Size:        uint(req.PageSize),
		CategoryID:  uint(req.CategoryId),
		Tag:         req.Tag,
		OrderBy:     req.OrderBy,
	})

	if err != nil {
		return &article.ListArticlesResponse{
			Code:    uint32(article.ArticleErrorCode_ARTICLE_LIST_FAILED),
			Message: err.Error(),
		}, nil
	}

	articles := make([]*article.Article, len(result))
	for i, a := range result {
		articles[i] = ConvertToProtoArticle(a)
	}

	return &article.ListArticlesResponse{
		Code:     uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:  "success",
		Articles: articles,
		Total:    uint32(total),
	}, nil
}

func (h *GrpcArticleHandler) UpdateArticle(ctx context.Context, req *article.UpdateArticleRequest) (*article.UpdateArticleResponse, error) {
	uid, err := commonmiddleware.RequireGRPCAuth(ctx)
	if err != nil {
		return nil, err
	}
	updatedArticle, err := h.Svc.UpdateArticle(ctx, uint(req.ArticleId), &model.UpdateArticleRequest{
		UserID:       uid,
		Title:        req.Title,
		Content:      req.Content,
		Summary:      req.Summary,
		CoverImage:   req.CoverImage,
		CategoryID:   uint(req.CategoryId),
		Tags:         req.Tags,
		IsFeatured:   req.IsFeatured,
		AllowComment: req.AllowComment,
		IsPublished:  req.IsPublished,
	})
	if err != nil {
		return &article.UpdateArticleResponse{
			Code:    uint32(article.ArticleErrorCode_ARTICLE_UPDATE_FAILED),
			Message: err.Error(),
		}, nil
	}

	return &article.UpdateArticleResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
		Article: ConvertToProtoArticle(updatedArticle),
	}, nil
}

func (h *GrpcArticleHandler) DeleteArticle(ctx context.Context, req *article.DeleteArticleRequest) (*article.DeleteArticleResponse, error) {
	uid, err := commonmiddleware.RequireGRPCAuth(ctx)
	if err != nil {
		return nil, err
	}
	err = h.Svc.DeleteArticle(ctx, uint(req.ArticleId), uid)
	if err != nil {
		return &article.DeleteArticleResponse{
			Code:    uint32(article.ArticleErrorCode_ARTICLE_DELETE_FAILED),
			Message: err.Error(),
		}, nil
	}

	return &article.DeleteArticleResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
	}, nil
}

func (h *GrpcArticleHandler) GetArticleBySlug(ctx context.Context, req *article.GetArticleBySlugRequest) (*article.GetArticleBySlugResponse, error) {
	foundArticle, err := h.Svc.GetArticleBySlug(ctx, req.Slug)
	if err != nil {
		return &article.GetArticleBySlugResponse{
			Code:    uint32(article.ArticleErrorCode_ARTICLE_NOT_FOUND),
			Message: "Article not found",
		}, nil
	}

	return &article.GetArticleBySlugResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
		Article: ConvertToProtoArticle(foundArticle),
	}, nil
}

func (h *GrpcArticleHandler) IncrementViewCount(ctx context.Context, req *article.IncrementViewCountRequest) (*article.IncrementViewCountResponse, error) {
	viewCount, err := h.Svc.IncrementViewCount(ctx, uint(req.ArticleId))
	if err != nil {
		return &article.IncrementViewCountResponse{
			Code:    errCode(err),
			Message: err.Error(),
		}, nil
	}

	return &article.IncrementViewCountResponse{
		Code:      uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:   "success",
		ViewCount: uint32(viewCount),
	}, nil
}

func (h *GrpcArticleHandler) LikeArticle(ctx context.Context, req *article.LikeArticleRequest) (*article.LikeArticleResponse, error) {
	likeCount, liked, err := h.Svc.LikeArticle(ctx, uint(req.ArticleId), uint(req.UserId))
	if err != nil {
		return &article.LikeArticleResponse{Code: constants.ErrCodeInternal, Message: err.Error()}, nil
	}
	return &article.LikeArticleResponse{
		Code:      uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:   "success",
		LikeCount: uint32(likeCount),
		Liked:     liked,
	}, nil
}

func (h *GrpcArticleHandler) CancelLikeArticle(ctx context.Context, req *article.CancelLikeArticleRequest) (*article.CancelLikeArticleResponse, error) {
	likeCount, liked, err := h.Svc.CancelLikeArticle(ctx, uint(req.ArticleId), uint(req.UserId))
	if err != nil {
		return &article.CancelLikeArticleResponse{Code: constants.ErrCodeInternal, Message: err.Error()}, nil
	}
	return &article.CancelLikeArticleResponse{
		Code:      uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:   "success",
		LikeCount: uint32(likeCount),
		Liked:     liked,
	}, nil
}

func (h *GrpcArticleHandler) GetLikeStatus(ctx context.Context, req *article.GetLikeStatusRequest) (*article.GetLikeStatusResponse, error) {
	likeCount, liked, err := h.Svc.GetLikeStatus(ctx, uint(req.ArticleId), uint(req.UserId))
	if err != nil {
		return &article.GetLikeStatusResponse{Code: constants.ErrCodeInternal, Message: err.Error()}, nil
	}
	return &article.GetLikeStatusResponse{
		Code:      uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:   "success",
		LikeCount: uint32(likeCount),
		Liked:     liked,
	}, nil
}

func (h *GrpcArticleHandler) SearchArticles(ctx context.Context, req *article.SearchArticlesRequest) (*article.SearchArticlesResponse, error) {
	result, total, err := h.Svc.SearchArticles(ctx, &model.SearchArticlesRequest{
		Keyword: req.Keyword,
		Page:    uint(req.Page),
		Size:    uint(req.PageSize),
	})
	if err != nil {
		return &article.SearchArticlesResponse{
			Code:    errCode(err),
			Message: err.Error(),
		}, nil
	}

	articles := make([]*article.Article, len(result))
	for i, a := range result {
		articles[i] = ConvertToProtoArticle(a)
	}

	return &article.SearchArticlesResponse{
		Code:     uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:  "success",
		Articles: articles,
		Total:    uint32(total),
	}, nil
}

func (h *GrpcArticleHandler) GetUserArticles(ctx context.Context, req *article.GetUserArticlesRequest) (*article.GetUserArticlesResponse, error) {
	// 安全：身份取自 gRPC 拦截器校验过的 JWT（RequireGRPCAuth），忽略 req.UserId，
	// 杜绝任意登录用户通过传入他人 user_id 查询其文章（IDOR 越权）。
	uid, err := commonmiddleware.RequireGRPCAuth(ctx)
	if err != nil {
		return nil, err
	}
	result, total, err := h.Svc.GetUserArticles(ctx, uid, uint(req.Page), uint(req.PageSize))
	if err != nil {
		return &article.GetUserArticlesResponse{
			Code:    errCode(err),
			Message: err.Error(),
		}, nil
	}

	articles := make([]*article.Article, len(result))
	for i, a := range result {
		articles[i] = ConvertToProtoArticle(a)
	}

	return &article.GetUserArticlesResponse{
		Code:     uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:  "success",
		Articles: articles,
		Total:    uint32(total),
	}, nil
}

func (h *GrpcArticleHandler) GetCategories(ctx context.Context, req *article.GetCategoriesRequest) (*article.GetCategoriesResponse, error) {
	categories, err := h.Svc.GetCategories(ctx)
	if err != nil {
		return &article.GetCategoriesResponse{
			Code:    errCode(err),
			Message: err.Error(),
		}, nil
	}

	protoCategories := make([]*article.Category, len(categories))
	for i, c := range categories {
		protoCategories[i] = &article.Category{
			Id:           uint32(c.ID),
			Name:         c.Name,
			Slug:         c.Slug,
			Description:  c.Description,
			ParentId:     uint32(c.ParentID),
			ArticleCount: uint32(c.ArticleCount),
			Sort:         uint32(c.Sort),
			CreatedAt:    c.CreatedAt.Format(constants.DateTimeFormat),
		}
	}

	return &article.GetCategoriesResponse{
		Code:       uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:    "success",
		Categories: protoCategories,
	}, nil
}

func (h *GrpcArticleHandler) GetTags(ctx context.Context, req *article.GetTagsRequest) (*article.GetTagsResponse, error) {
	tags, err := h.Svc.GetTags(ctx)
	if err != nil {
		return &article.GetTagsResponse{
			Code:    errCode(err),
			Message: err.Error(),
		}, nil
	}

	protoTags := make([]*article.Tag, len(tags))
	for i, t := range tags {
		protoTags[i] = &article.Tag{
			Id:           uint32(t.ID),
			Name:         t.Name,
			Slug:         t.Slug,
			ArticleCount: uint32(t.ArticleCount),
			CreatedAt:    t.CreatedAt.Format(constants.DateTimeFormat),
		}
	}

	return &article.GetTagsResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
		Tags:    protoTags,
	}, nil
}

// ConvertToProtoArticle 将 model.Article 转换为 proto Article
func ConvertToProtoArticle(a *model.Article) *article.Article {
	tags := make([]string, len(a.Tags))
	for i, t := range a.Tags {
		tags[i] = t.Name
	}

	publishedAt := ""
	if a.PublishedAt != nil {
		publishedAt = util.FormatTime(*a.PublishedAt)
	}

	return &article.Article{
		Id:           uint32(a.ID),
		UserId:       uint32(a.UserID),
		Username:     a.User.Username,
		Title:        a.Title,
		Slug:         a.Slug,
		Summary:      a.Summary,
		Content:      a.Content,
		CoverImage:   a.CoverImage,
		CategoryId:   uint32(a.CategoryID),
		CategoryName: a.Category.Name,
		Tags:         tags,
		ViewCount:    uint32(a.ViewCount),
		CommentCount: uint32(a.CommentCount),
		LikeCount:    uint32(a.LikeCount),
		IsFeatured:   a.IsFeatured,
		AllowComment: a.AllowComment,
		CreatedAt:    util.FormatTime(a.CreatedAt),
		UpdatedAt:    util.FormatTime(a.UpdatedAt),
		PublishedAt:  publishedAt,
		Status:       a.Status,
	}
}
