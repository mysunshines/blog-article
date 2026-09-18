package service

import (
	"context"
	"fmt"

	"github.com/mysunshines/blog-article/internal/client"
	"github.com/mysunshines/blog-article/internal/model"
)

// ============================================================================
// 文章背景（三期）
// ----------------------------------------------------------------------------
// 背景由管理员在后台配置：price=0 免费（所有用户可用），price>0 需用积分购买。
// 「是否已拥有」不在此处单独建表，而是复用 point-service 的已购记录
// （user_purchases，item_type=background），保证与付费文章同一套购买语义。
// ============================================================================

// ListBackgrounds 上架背景列表；owned[i] 表示当前用户是否拥有第 i 个背景。
// 免费背景恒为 true；付费背景查已购记录；userID=0（未登录）付费背景一律 false。
func (s *articleService) ListBackgrounds(ctx context.Context, userID uint) ([]*model.ArticleBackground, []bool, error) {
	list, err := s.bgRepo.ListEnabled(ctx)
	if err != nil {
		return nil, nil, err
	}
	owned := make([]bool, len(list))
	for i, bg := range list {
		if bg.Price <= 0 {
			owned[i] = true // 免费背景人人可用
			continue
		}
		if userID == 0 {
			continue
		}
		if ok, err := client.HasPurchased(ctx, userID, "background", bg.ID); err == nil && ok {
			owned[i] = true
		}
	}
	return list, owned, nil
}

// BuyBackground 用积分购买背景，返回购买后剩余积分。
// 免费背景无需购买；已购买过不重复扣费（point-service 侧幂等）。
func (s *articleService) BuyBackground(ctx context.Context, backgroundID, userID uint) (int64, error) {
	bg, err := s.bgRepo.GetByID(ctx, backgroundID)
	if err != nil {
		return 0, err
	}
	if bg.Price <= 0 {
		return 0, fmt.Errorf("该背景免费，无需购买")
	}
	if ok, err := client.HasPurchased(ctx, userID, "background", bg.ID); err == nil && ok {
		return 0, nil
	}
	balance, err := client.SpendPoints(ctx, userID, bg.Price, "background", bg.ID, "购买背景："+bg.Name)
	if err != nil {
		return 0, err
	}
	return balance, nil
}

// AdminCreateBackground 新增背景（默认上架）
func (s *articleService) AdminCreateBackground(ctx context.Context, req *model.CreateBackgroundRequest) (*model.ArticleBackground, error) {
	if req == nil || req.Name == "" || req.ImageURL == "" {
		return nil, fmt.Errorf("名称与图片地址不能为空")
	}
	bg := &model.ArticleBackground{
		Name:     req.Name,
		ImageURL: req.ImageURL,
		ThumbURL: req.ThumbURL,
		Price:    req.Price,
		Sort:     req.Sort,
		Status:   model.BackgroundStatusOnline,
	}
	if err := s.bgRepo.Create(ctx, bg); err != nil {
		return nil, err
	}
	return bg, nil
}

// AdminUpdateBackground 更新背景（空字段表示不修改）
func (s *articleService) AdminUpdateBackground(ctx context.Context, req *model.UpdateBackgroundRequest) (*model.ArticleBackground, error) {
	if req == nil || req.ID == 0 {
		return nil, fmt.Errorf("背景 ID 不能为空")
	}
	if _, err := s.bgRepo.GetByID(ctx, req.ID); err != nil {
		return nil, err
	}
	// 显式字段更新，避免「只上下架」时把价格/排序清零：
	//   - 空字符串视为「不修改」
	//   - price / sort 传负数（-1）表示「不修改」（用于在上下架时保持原值）
	//   - status 例外：0（下架）必须能写入
	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.ImageURL != "" {
		updates["image_url"] = req.ImageURL
	}
	if req.ThumbURL != "" {
		updates["thumb_url"] = req.ThumbURL
	}
	if req.Price >= 0 {
		updates["price"] = req.Price
	}
	if req.Sort >= 0 {
		updates["sort"] = req.Sort
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if err := s.bgRepo.UpdateFields(ctx, req.ID, updates); err != nil {
		return nil, err
	}
	return s.bgRepo.GetByID(ctx, req.ID)
}

// AdminDeleteBackground 删除背景
func (s *articleService) AdminDeleteBackground(ctx context.Context, id uint) error {
	return s.bgRepo.Delete(ctx, id)
}
