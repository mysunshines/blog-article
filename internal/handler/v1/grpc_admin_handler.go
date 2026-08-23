package v1

import (
	"context"
	"log"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/mysunshines/blog-article/internal/client"
	"github.com/mysunshines/blog-article/internal/model"
	article "github.com/mysunshines/blog-article/proto/pb"
	"github.com/mysunshines/gocommon/constants"
	commonmiddleware "github.com/mysunshines/gocommon/middleware"

	user "github.com/mysunshines/blog-user/proto/pb"
)

// 以下方法为后台管理 / 分类管理接口（管理员操作）。
// 收敛后只保留 gRPC 这一份实现，经 Gateway /api/v1 反射代理暴露给前端 admin 调用。

// requireGRPCAdmin 校验后台接口访问权限：先校验 JWT 登录，再校验当前用户为管理员
// （role == RoleAdmin）。后台接口经网关 /api/v1 反射代理转发，仅靠 JWT 鉴权不够，
// 必须显式校验角色，否则任意登录用户都能调用后台方法（对应原 HTTP 的 AdminOnlyMiddleware）。
// 所有后台 gRPC 方法统一在入口调用本函数。
func requireGRPCAdmin(ctx context.Context) error {
	if _, err := commonmiddleware.RequireGRPCAuth(ctx); err != nil {
		return err
	}
	raw, ok := commonmiddleware.GetGRPCRole(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "未认证")
	}
	var role uint8
	switch v := raw.(type) {
	case float64:
		role = uint8(v)
	case uint8:
		role = v
	case int:
		role = uint8(v)
	case int64:
		role = uint8(v)
	default:
		return status.Error(codes.PermissionDenied, "无效的角色信息")
	}
	if role != constants.RoleAdmin {
		return status.Error(codes.PermissionDenied, "需要管理员权限")
	}
	return nil
}

// ListArticlesForAdmin 后台文章列表
func (h *GrpcArticleHandler) ListArticlesForAdmin(ctx context.Context, req *article.ListArticlesForAdminRequest) (*article.ListArticlesForAdminResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	articles, total, err := h.Svc.AdminListArticles(ctx, &model.AdminListArticlesRequest{
		Status:   req.GetStatus(),
		Page:     uint(req.GetPage()),
		Size:     uint(req.GetPageSize()),
	})
	if err != nil {
		return nil, err
	}
	return &article.ListArticlesForAdminResponse{
		Code:     0,
		Message:  "success",
		Articles: convertToProtoArticles(articles),
		Total:    uint32(total),
	}, nil
}

// GetArticleForAdmin 后台获取单篇文章（不限状态，供审核/编辑使用）。
// 鉴权：管理员可访问任意文章；作者本人可访问自己的任意状态文章（个人主页/管理后台统一走此接口编辑）。
func (h *GrpcArticleHandler) GetArticleForAdmin(ctx context.Context, req *article.GetArticleForAdminRequest) (*article.GetArticleForAdminResponse, error) {
	// 先校验登录，取出 uid 与角色
	if _, err := commonmiddleware.RequireGRPCAuth(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.GetArticleForAdmin(ctx, uint(req.GetArticleId()))
	if err != nil {
		return nil, err
	}
	// 作者本人可访问自己的文章；否则需管理员权限
	if uid, ok := commonmiddleware.GetGRPCUserID(ctx); ok && uid == art.UserID {
		// 放行：作者本人
	} else if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	return &article.GetArticleForAdminResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// ApproveArticle 审核通过
func (h *GrpcArticleHandler) ApproveArticle(ctx context.Context, req *article.ApproveArticleRequest) (*article.ApproveArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.ApproveArticle(ctx, uint(req.GetArticleId()))
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_APPROVE, "article", uint(art.ID), art.Title, "")
	return &article.ApproveArticleResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// RejectArticle 审核拒绝
func (h *GrpcArticleHandler) RejectArticle(ctx context.Context, req *article.RejectArticleRequest) (*article.RejectArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.RejectArticle(ctx, uint(req.GetArticleId()), req.GetReason())
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_REJECT, "article", uint(art.ID), art.Title, "原因: "+req.GetReason())
	return &article.RejectArticleResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// OfflineArticle 下线文章
func (h *GrpcArticleHandler) OfflineArticle(ctx context.Context, req *article.OfflineArticleRequest) (*article.OfflineArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.OfflineArticle(ctx, uint(req.GetArticleId()), req.GetReason())
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_OFFLINE, "article", uint(art.ID), art.Title, "原因: "+req.GetReason())
	return &article.OfflineArticleResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// PublishArticle 发布文章
func (h *GrpcArticleHandler) PublishArticle(ctx context.Context, req *article.PublishArticleRequest) (*article.PublishArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.PublishArticle(ctx, uint(req.GetArticleId()))
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_PUBLISH, "article", uint(art.ID), art.Title, "")
	return &article.PublishArticleResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// SubmitArticle 提交审核
func (h *GrpcArticleHandler) SubmitArticle(ctx context.Context, req *article.SubmitArticleRequest) (*article.SubmitArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.SubmitArticle(ctx, uint(req.GetArticleId()))
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_SUBMIT, "article", uint(art.ID), art.Title, "")
	return &article.SubmitArticleResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// AdminUpdateArticle 后台更新文章
func (h *GrpcArticleHandler) AdminUpdateArticle(ctx context.Context, req *article.AdminUpdateArticleRequest) (*article.AdminUpdateArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	art, err := h.Svc.AdminUpdateArticle(ctx, uint(req.GetArticleId()), &model.UpdateArticleRequest{
		Title:        req.GetTitle(),
		Content:      req.GetContent(),
		Summary:      req.GetSummary(),
		CoverImage:   req.GetCoverImage(),
		CategoryID:   uint(req.GetCategoryId()),
		IsFeatured:   req.GetIsFeatured(),
		AllowComment: req.GetAllowComment(),
	})
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_UPDATE, "article", uint(art.ID), art.Title, "")
	return &article.AdminUpdateArticleResponse{
		Code:    0,
		Message: "success",
		Article: ConvertToProtoArticle(art),
	}, nil
}

// AdminDeleteArticle 后台删除文章
func (h *GrpcArticleHandler) AdminDeleteArticle(ctx context.Context, req *article.AdminDeleteArticleRequest) (*article.AdminDeleteArticleResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.Svc.AdminDeleteArticle(ctx, uint(req.GetArticleId())); err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_ARTICLE_DELETE, "article", uint(req.GetArticleId()), "", "")
	return &article.AdminDeleteArticleResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// CreateCategory 创建分类
func (h *GrpcArticleHandler) CreateCategory(ctx context.Context, req *article.CreateCategoryRequest) (*article.CreateCategoryResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	category, err := h.Svc.CreateCategory(ctx, req.GetName(), req.GetDescription(), int(req.GetSort()))
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_CATEGORY_CREATE, "category", uint(category.ID), category.Name, "")
	return &article.CreateCategoryResponse{
		Code:     0,
		Message:  "success",
		Category: convertToProtoCategory(category),
	}, nil
}

// UpdateCategory 更新分类
func (h *GrpcArticleHandler) UpdateCategory(ctx context.Context, req *article.UpdateCategoryRequest) (*article.UpdateCategoryResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	category, err := h.Svc.UpdateCategory(ctx, uint(req.GetCategoryId()), req.GetName(), req.GetDescription(), int(req.GetSort()))
	if err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_CATEGORY_UPDATE, "category", uint(category.ID), category.Name, "")
	return &article.UpdateCategoryResponse{
		Code:     0,
		Message:  "success",
		Category: convertToProtoCategory(category),
	}, nil
}

// DeleteCategory 删除分类
func (h *GrpcArticleHandler) DeleteCategory(ctx context.Context, req *article.DeleteCategoryRequest) (*article.DeleteCategoryResponse, error) {
	if err := requireGRPCAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.Svc.DeleteCategory(ctx, uint(req.GetCategoryId())); err != nil {
		return nil, err
	}
	operator, _ := commonmiddleware.GetGRPCUsername(ctx)
	h.recordAuditGRPC(ctx, operator, user.AuditAction_AUDIT_ACTION_CATEGORY_DELETE, "category", uint(req.GetCategoryId()), "", "")
	return &article.DeleteCategoryResponse{
		Code:    0,
		Message: "删除成功",
	}, nil
}

// recordAuditGRPC 记录审计日志（gRPC 上下文版本），与 http_handler.go 的 recordAudit 保持等价。
// 失败时仅告警，不影响主业务流程。审计日志经 client 适配层上报 user-service。
func (h *GrpcArticleHandler) recordAuditGRPC(ctx context.Context, operator string, action user.AuditAction, targetType string, targetID uint, targetTitle, detail string) {
	operatorID, _ := commonmiddleware.GetGRPCUserID(ctx)
	if err := client.RecordAudit(ctx, operatorID, operator, action, targetType, targetID, targetTitle, detail); err != nil {
		// 审计日志写入失败不应阻断主流程，仅记录告警。
		log.Printf("[audit][warn] record audit failed: %v", err)
	}
}

// convertToProtoArticles 批量转换文章列表（grpc_handler.go 仅提供单条 ConvertToProtoArticle）。
func convertToProtoArticles(list []*model.Article) []*article.Article {
	out := make([]*article.Article, len(list))
	for i, a := range list {
		out[i] = ConvertToProtoArticle(a)
	}
	return out
}

// convertToProtoCategory 将 model.Category 转换为 proto Category（与 GetCategories 保持一致的字段映射）。
func convertToProtoCategory(c *model.Category) *article.Category {
	return &article.Category{
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
