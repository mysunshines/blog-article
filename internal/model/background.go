package model

import "time"

// ArticleBackground 文章背景（三期）
//
// 管理员在后台配置：price=0 为免费背景，所有用户可用；price>0 需用积分购买，
// 购买后（user_purchases 记录 item_type=background）才能在编辑文章时选用。
type ArticleBackground struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:128;not null" json:"name"`
	ImageURL string `gorm:"column:image_url;size:512;not null" json:"image_url"` // 背景图地址
	ThumbURL string `gorm:"column:thumb_url;size:512;not null;default:''" json:"thumb_url"`
	Price    int64  `gorm:"not null;default:0" json:"price"`     // 0=免费，>0=需积分购买
	Status   uint   `gorm:"not null;default:1" json:"status"`    // 1=上架 0=下架
	Sort     int    `gorm:"not null;default:0" json:"sort"`      // 排序（升序）
	CreatedAt time.Time `gorm:"<-:create" json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (ArticleBackground) TableName() string { return "article_backgrounds" }

// 背景状态
const (
	BackgroundStatusOffline uint = 0
	BackgroundStatusOnline  uint = 1
)

// ============ 请求 DTO ============

// CreateBackgroundRequest 新增背景
type CreateBackgroundRequest struct {
	Name     string `json:"name" binding:"required"`
	ImageURL string `json:"image_url" binding:"required"`
	ThumbURL string `json:"thumb_url"`
	Price    int64  `json:"price"`
	Sort     int    `json:"sort"`
}

// UpdateBackgroundRequest 更新背景
type UpdateBackgroundRequest struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url"`
	ThumbURL string `json:"thumb_url"`
	Price    int64  `json:"price"`
	Status   *uint  `json:"status"`
	Sort     int    `json:"sort"`
}
