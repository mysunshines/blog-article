package v1

import (
	"context"

	"github.com/mysunshines/blog-article/internal/model"
	article "github.com/mysunshines/blog-article/proto/pb/v1"
	commonmiddleware "github.com/mysunshines/gocommon/middleware"
)

// ============================================================================
// 文章背景（三期）：列表 / 购买 / 后台管理
// ============================================================================

// ListBackgrounds 上架背景列表（含「当前用户是否已拥有」，未登录时付费背景为 false）
func (h *GrpcArticleHandler) ListBackgrounds(ctx context.Context, req *article.ListBackgroundsRequest) (*article.ListBackgroundsResponse, error) {
	// 背景列表允许未登录访问（便于展示），未登录时 uid=0
	uid, _ := commonmiddleware.GetGRPCUserID(ctx)
	list, owned, err := h.Svc.ListBackgrounds(ctx, uid)
	if err != nil {
		return &article.ListBackgroundsResponse{Code: errCode(err), Message: err.Error()}, nil
	}
	items := make([]*article.Background, 0, len(list))
	for i, bg := range list {
		b := convertToProtoBackground(bg)
		if i < len(owned) {
			b.Owned = owned[i]
		}
		items = append(items, b)
	}
	return &article.ListBackgroundsResponse{
		Code:        uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:     "success",
		Backgrounds: items,
	}, nil
}

// BuyBackground 用积分购买背景
func (h *GrpcArticleHandler) BuyBackground(ctx context.Context, req *article.BuyBackgroundRequest) (*article.BuyBackgroundResponse, error) {
	uid, err := commonmiddleware.RequireGRPCAuth(ctx)
	if err != nil {
		return nil, err
	}
	balance, err := h.Svc.BuyBackground(ctx, uint(req.BackgroundId), uid)
	if err != nil {
		return &article.BuyBackgroundResponse{Code: errCode(err), Message: err.Error()}, nil
	}
	return &article.BuyBackgroundResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
		Balance: balance,
	}, nil
}

// AdminCreateBackground 新增背景
func (h *GrpcArticleHandler) AdminCreateBackground(ctx context.Context, req *article.AdminCreateBackgroundRequest) (*article.AdminCreateBackgroundResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	bg, err := h.Svc.AdminCreateBackground(ctx, &model.CreateBackgroundRequest{
		Name:     req.Name,
		ImageURL: req.ImageUrl,
		ThumbURL: req.ThumbUrl,
		Price:    req.Price,
		Sort:     int(req.Sort),
	})
	if err != nil {
		return &article.AdminCreateBackgroundResponse{Code: errCode(err), Message: err.Error()}, nil
	}
	return &article.AdminCreateBackgroundResponse{
		Code:       uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:    "success",
		Background: convertToProtoBackground(bg),
	}, nil
}

// AdminUpdateBackground 更新背景
func (h *GrpcArticleHandler) AdminUpdateBackground(ctx context.Context, req *article.AdminUpdateBackgroundRequest) (*article.AdminUpdateBackgroundResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	status := uint(req.Status)
	bg, err := h.Svc.AdminUpdateBackground(ctx, &model.UpdateBackgroundRequest{
		ID:       uint(req.BackgroundId),
		Name:     req.Name,
		ImageURL: req.ImageUrl,
		ThumbURL: req.ThumbUrl,
		Price:    req.Price,
		Status:   &status,
		Sort:     int(req.Sort),
	})
	if err != nil {
		return &article.AdminUpdateBackgroundResponse{Code: errCode(err), Message: err.Error()}, nil
	}
	return &article.AdminUpdateBackgroundResponse{
		Code:       uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message:    "success",
		Background: convertToProtoBackground(bg),
	}, nil
}

// AdminDeleteBackground 删除背景
func (h *GrpcArticleHandler) AdminDeleteBackground(ctx context.Context, req *article.AdminDeleteBackgroundRequest) (*article.AdminDeleteBackgroundResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.Svc.AdminDeleteBackground(ctx, uint(req.BackgroundId)); err != nil {
		return &article.AdminDeleteBackgroundResponse{Code: errCode(err), Message: err.Error()}, nil
	}
	return &article.AdminDeleteBackgroundResponse{
		Code:    uint32(article.ArticleErrorCode_ARTICLE_SUCCESS),
		Message: "success",
	}, nil
}

// convertToProtoBackground 背景模型 → proto
func convertToProtoBackground(bg *model.ArticleBackground) *article.Background {
	if bg == nil {
		return nil
	}
	return &article.Background{
		Id:       uint32(bg.ID),
		Name:     bg.Name,
		ImageUrl: bg.ImageURL,
		ThumbUrl: bg.ThumbURL,
		Price:    bg.Price,
		Status:   uint32(bg.Status),
		Sort:     int32(bg.Sort),
	}
}
