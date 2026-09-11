package client

import (
	"context"
	"fmt"

	v0pb "github.com/mysunshines/blog-ranking/proto/pb/v0"
	pb "github.com/mysunshines/blog-ranking/proto/pb/v1"
	"github.com/mysunshines/gocommon/grpcclient"
)

// 榜单键（与 RegisterBoard、各业务推送保持一致）。
//   - 文章维度三榜（浏览/点赞/评论）：member = 文章 ID，装饰器 = article-service
//   - 作者维度榜（发文最多）：member = 用户 ID，装饰器 = user-service
const (
	BoardArticleViews    = "board:article:views"
	BoardArticleLikes    = "board:article:likes"
	BoardArticleComments = "board:article:comments"
	BoardAuthorArticles  = "board:author:articles"
)

// RegisterArticleBoards 声明文章维度三个榜单（装饰器 = article-service）。
// 幂等：重复调用覆盖同一 board 配置；best-effort，失败返回 error 由调用方决定是否告警。
func RegisterArticleBoards(ctx context.Context) error {
	boards := []*pb.BoardConfig{
		{Board: BoardArticleViews, DecoratorType: "remote", DecoratorService: "article-service",
			LinkTemplate: "/article/{member}", CacheTtlSec: 60},
		{Board: BoardArticleLikes, DecoratorType: "remote", DecoratorService: "article-service",
			LinkTemplate: "/article/{member}", CacheTtlSec: 60},
		{Board: BoardArticleComments, DecoratorType: "remote", DecoratorService: "article-service",
			LinkTemplate: "/article/{member}", CacheTtlSec: 60},
	}
	for _, cfg := range boards {
		var resp pb.RegisterBoardResponse
		if err := grpcclient.SendRequest(ctx, v0pb.RankingService_RegisterBoard_FullMethodName,
			&pb.RegisterBoardRequest{Config: cfg}, &resp); err != nil {
			return err
		}
		if resp.Code != 0 {
			return fmt.Errorf("register board %s failed: code=%d message=%s", cfg.Board, resp.Code, resp.Message)
		}
	}
	return nil
}

// RecordScore 推送单成员分数变更（增量或绝对值）。best-effort，失败返回 error。
func RecordScore(ctx context.Context, board, member string, op pb.ScoreOp, delta, score float64) error {
	var resp pb.RecordScoreResponse
	if err := grpcclient.SendRequest(ctx, v0pb.RankingService_RecordScore_FullMethodName, &pb.RecordScoreRequest{
		Board:  board,
		Member: member,
		Op:     op,
		Delta:  delta,
		Score:  score,
	}, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("record score failed: board=%s member=%s code=%d message=%s",
			board, member, resp.Code, resp.Message)
	}
	return nil
}

// BatchSetScore 批量覆盖设置榜单（历史回填）。pruneOthers=true 时移除 ZSET 中不在 items 内的旧成员。
func BatchSetScore(ctx context.Context, board string, items map[string]float64, pruneOthers bool) error {
	if len(items) == 0 {
		return nil
	}
	pbItems := make([]*pb.ScoreItem, 0, len(items))
	for member, score := range items {
		pbItems = append(pbItems, &pb.ScoreItem{Member: member, Score: score})
	}
	var resp pb.BatchSetScoreResponse
	if err := grpcclient.SendRequest(ctx, v0pb.RankingService_BatchSetScore_FullMethodName, &pb.BatchSetScoreRequest{
		Board:       board,
		Items:       pbItems,
		PruneOthers: pruneOthers,
	}, &resp); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("batch set score failed: board=%s code=%d message=%s", board, resp.Code, resp.Message)
	}
	return nil
}

// IncrScore 便捷封装：对 member 在 board 上增量 delta（ZINCRBY）。
func IncrScore(ctx context.Context, board, member string, delta float64) error {
	return RecordScore(ctx, board, member, pb.ScoreOp_SCORE_OP_INCREMENT, delta, 0)
}

// SetScore 便捷封装：对 member 在 board 上设置绝对值 score（ZADD）。
func SetScore(ctx context.Context, board, member string, score float64) error {
	return RecordScore(ctx, board, member, pb.ScoreOp_SCORE_OP_SET, 0, score)
}
