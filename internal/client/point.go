package client

import (
	"context"
	"fmt"
	"os"

	point "github.com/mysunshines/blog-point/proto/pb/v1"
	"github.com/mysunshines/gocommon/grpcclient"
	"google.golang.org/grpc/metadata"
)

// ============================================================================
// 积分服务调用（article-service → point-service）
// ----------------------------------------------------------------------------
// 服务间调用没有用户 JWT，统一携带内部服务令牌（metadata: x-point-internal），
// 由 point-service 校验后信任请求中的 user_id。令牌通过环境变量
// POINT_INTERNAL_TOKEN 下发，需在两个服务配置一致；未配置时内部通道关闭。
// ============================================================================

// withInternalToken 把内部服务令牌附加到出站 metadata
func withInternalToken(ctx context.Context) context.Context {
	token := os.Getenv("POINT_INTERNAL_TOKEN")
	if token == "" {
		return ctx
	}
	return metadata.AppendToOutgoingContext(ctx, "x-point-internal", token)
}

// EarnPoints 上报积分事件（分值由 point-service 规则引擎决定，本服务不感知具体分值）。
// best-effort：积分是激励侧能力，失败只告警，不阻断主流程。
func EarnPoints(ctx context.Context, userID uint, eventType, contextJSON, relatedType string, relatedID uint) (int64, error) {
	var resp point.EarnPointsResponse
	if err := grpcclient.SendRequest(withInternalToken(ctx), point.PointService_EarnPoints_FullMethodName,
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
	if err := grpcclient.SendRequest(withInternalToken(ctx), point.PointService_SpendPoints_FullMethodName,
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
	if err := grpcclient.SendRequest(withInternalToken(ctx), point.PointService_HasPurchased_FullMethodName,
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
