package client

import (
	"context"
	"fmt"

	v0point "github.com/mysunshines/blog-point/proto/pb/v0"
	point "github.com/mysunshines/blog-point/proto/pb/v1"
	"github.com/mysunshines/gocommon/grpcclient"
)

// ============================================================================
// 积分服务调用（article-service → point-service）
// ----------------------------------------------------------------------------
// 服务间内部调用统一走 point.v0.PointIngestService（与对外 v1 在命名/部署上隔离）：
//   - 公网网关 DeriveGRPCService 仅硬编码 v1，v0 天然不进公网入口；
//   - v0 仅由内网服务经 Consul 直连 gRPC 调用，完全信任内网、不做应用层鉴权，
//     安全性依赖 point 的 gRPC 端口仅对内网开放（不暴露公网）。
// ============================================================================

// EarnPoints 上报积分事件（分值由 point-service 规则引擎决定，本服务不感知具体分值）。
// best-effort：积分是激励侧能力，失败只告警，不阻断主流程。
func EarnPoints(ctx context.Context, userID uint, eventType, contextJSON, relatedType string, relatedID uint) (int64, error) {
	var resp point.EarnPointsResponse
	if err := grpcclient.SendRequest(ctx, v0point.PointIngestService_EarnPoints_FullMethodName,
		&point.EarnPointsRequest{
			UserId:      uint32(userID),
			EventType:   eventType,
			Context:     contextJSON,
			RelatedType: relatedType,
			RelatedId:   uint32(relatedID),
		}, &resp); err != nil {
		return 0, err
	}
	if resp.Code != 0 {
		return 0, fmt.Errorf("earn points failed: event=%s user=%d code=%d message=%s",
			eventType, userID, resp.Code, resp.Message)
	}
	return resp.Points, nil
}

// SpendPoints 代表用户消费积分（购买付费文章 / 背景）
func SpendPoints(ctx context.Context, userID uint, amount int64, itemType string, itemID uint, remark string) (int64, error) {
	var resp point.SpendPointsResponse
	if err := grpcclient.SendRequest(ctx, v0point.PointIngestService_SpendPoints_FullMethodName,
		&point.SpendPointsRequest{
			UserId:   uint32(userID),
			Amount:   amount,
			ItemType: itemType,
			ItemId:   uint32(itemID),
			Remark:   remark,
		}, &resp); err != nil {
		return 0, err
	}
	if resp.Code != 0 {
		return 0, fmt.Errorf("spend points failed: user=%d item=%s/%d code=%d message=%s",
			userID, itemType, itemID, resp.Code, resp.Message)
	}
	return resp.Balance, nil
}

// HasPurchased 查询用户是否已购买某物品（付费文章 / 背景）
func HasPurchased(ctx context.Context, userID uint, itemType string, itemID uint) (bool, error) {
	var resp point.HasPurchasedResponse
	if err := grpcclient.SendRequest(ctx, v0point.PointIngestService_HasPurchased_FullMethodName,
		&point.HasPurchasedRequest{
			UserId:   uint32(userID),
			ItemType: itemType,
			ItemId:   uint32(itemID),
		}, &resp); err != nil {
		return false, err
	}
	if resp.Code != 0 {
		return false, fmt.Errorf("has purchased failed: user=%d item=%s/%d code=%d message=%s",
			userID, itemType, itemID, resp.Code, resp.Message)
	}
	return resp.Purchased, nil
}
