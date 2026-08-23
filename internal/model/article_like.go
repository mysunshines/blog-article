package model

import "time"

// ArticleLike 文章点赞记录（一个用户对一个文章仅能点赞一次）
type ArticleLike struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ArticleID uint      `gorm:"not null;uniqueIndex:idx_article_user" json:"article_id"`
	UserID    uint      `gorm:"not null;uniqueIndex:idx_article_user" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (ArticleLike) TableName() string {
	return "article_likes"
}
