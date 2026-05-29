package model

import (
	"time"
)

type Article struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	UserID       uint       `gorm:"index" json:"user_id"`
	Title        string     `gorm:"size:256" json:"title"`
	Slug         string     `gorm:"uniqueIndex;size:256" json:"slug"`
	Summary      string     `gorm:"size:512" json:"summary"`
	Content      string     `gorm:"type:text" json:"content"`
	CoverImage   string     `gorm:"size:256" json:"cover_image"`
	CategoryID   uint       `gorm:"index" json:"category_id"`
	ViewCount    int        `gorm:"default:0" json:"view_count"`
	CommentCount int        `gorm:"default:0" json:"comment_count"`
	LikeCount    int        `gorm:"default:0" json:"like_count"`
	IsPublished  bool       `gorm:"default:false" json:"is_published"`
	IsFeatured   bool       `gorm:"default:false" json:"is_featured"`
	AllowComment bool       `gorm:"default:true" json:"allow_comment"`
	PublishedAt  *time.Time `json:"published_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`

	// 关联
	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag    `gorm:"many2many:article_tags" json:"tags,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}

type Category struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"size:64" json:"name"`
	Slug         string    `gorm:"uniqueIndex;size:64" json:"slug"`
	Description  string    `gorm:"size:256" json:"description"`
	ParentID     uint      `gorm:"index" json:"parent_id"`
	Sort         int       `gorm:"default:0" json:"sort"`
	ArticleCount int       `gorm:"default:0" json:"article_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Category) TableName() string {
	return "categories"
}

type Tag struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"uniqueIndex;size:64" json:"name"`
	Slug         string    `gorm:"uniqueIndex;size:64" json:"slug"`
	ArticleCount int       `gorm:"default:0" json:"article_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type ArticleTag struct {
	ArticleID uint `gorm:"primaryKey"`
	TagID     uint `gorm:"primaryKey"`
}

func (ArticleTag) TableName() string {
	return "article_tags"
}

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"size:64" json:"username"`
	Nickname  string    `gorm:"size:64" json:"nickname"`
	Avatar    string    `gorm:"size:256" json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
}

func (User) TableName() string {
	return "users"
}

// DTO 请求结构
type CreateArticleRequest struct {
	UserID       uint     `json:"user_id"`
	Title        string   `json:"title" binding:"required"`
	Content      string   `json:"content" binding:"required"`
	Summary      string   `json:"summary"`
	CoverImage   string   `json:"cover_image"`
	CategoryID   uint     `json:"category_id"`
	Tags         []string `json:"tags"`
	IsPublished  bool     `json:"is_published"`
	IsFeatured   bool     `json:"is_featured"`
	AllowComment bool     `json:"allow_comment"`
}

type UpdateArticleRequest struct {
	UserID       uint     `json:"user_id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Summary      string   `json:"summary"`
	CoverImage   string   `json:"cover_image"`
	CategoryID   uint     `json:"category_id"`
	Tags         []string `json:"tags"`
	IsPublished  bool     `json:"is_published"`
	IsFeatured   bool     `json:"is_featured"`
	AllowComment bool     `json:"allow_comment"`
}

type ListArticlesRequest struct {
	Page        uint   `form:"page"`
	Size        uint   `form:"size"`
	CategoryID  uint   `form:"category_id"`
	Tag         string `form:"tag"`
	IsPublished bool   `form:"is_published"`
	OrderBy     string `form:"order_by"`
	UserID      uint   `form:"user_id"`
}

type SearchArticlesRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	Page    uint   `form:"page"`
	Size    uint   `form:"size"`
}
