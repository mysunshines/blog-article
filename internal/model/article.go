package model

import (
	"time"

	"gorm.io/gorm"
)

// 文章状态机
const (
	ArticleStatusDraft     = "draft"     // 草稿
	ArticleStatusPending   = "pending"   // 待审核
	ArticleStatusPublished = "published" // 已发布
	ArticleStatusOffline   = "offline"   // 已下线
	ArticleStatusRejected  = "rejected"  // 已拒绝
)

type Article struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"index" json:"user_id"`
	Title         string         `gorm:"size:256" json:"title"`
	Slug          string         `gorm:"uniqueIndex;size:256" json:"slug"`
	Summary       string         `gorm:"size:512" json:"summary"`
	Content       string         `gorm:"type:text" json:"content"`
	ContentHTML   string         `gorm:"-" json:"content_html"` // 由后端渲染并净化的安全 HTML，不入库
	CoverImage    string         `gorm:"size:256" json:"cover_image"`
	CategoryID    uint           `gorm:"index" json:"category_id"`
	ViewCount     int            `gorm:"default:0" json:"view_count"`
	CommentCount  int            `gorm:"default:0" json:"comment_count"`
	LikeCount     int            `gorm:"default:0" json:"like_count"`
	Status        string         `gorm:"size:16;default:'draft'" json:"status"`
	RejectReason  string         `gorm:"size:256" json:"reject_reason"`
	OfflineReason string         `gorm:"size:256" json:"offline_reason"`
	IsFeatured    bool           `gorm:"default:false" json:"is_featured"`
	AllowComment  bool           `gorm:"default:true" json:"allow_comment"`
	PublishedAt   *time.Time     `json:"published_at"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Tags     []Tag    `gorm:"many2many:article_tags" json:"tags,omitempty"`

	// 作者名（非持久化，由详情接口并行从 user-service 拉取填充）
	AuthorName string `gorm:"-" json:"author_name,omitempty"`
}

func (Article) TableName() string {
	return "articles"
}

type Category struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Name         string         `gorm:"size:64" json:"name"`
	Slug         string         `gorm:"uniqueIndex;size:64" json:"slug"`
	Description  string         `gorm:"size:256" json:"description"`
	ParentID     uint           `gorm:"index" json:"parent_id"`
	Sort         int            `gorm:"default:0" json:"sort"`
	ArticleCount int            `gorm:"default:0" json:"article_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
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
	IsFeatured   bool     `json:"is_featured"`
	AllowComment bool     `json:"allow_comment"`
	// 是否立即发布：true=进入待审核(pending)，false/缺省=存为草稿(draft)
	IsPublished bool `json:"is_published"`
}

type UpdateArticleRequest struct {
	UserID       uint     `json:"user_id"`
	Title        string   `json:"title"`
	Content      string   `json:"content"`
	Summary      string   `json:"summary"`
	CoverImage   string   `json:"cover_image"`
	CategoryID   uint     `json:"category_id"`
	Tags         []string `json:"tags"`
	IsFeatured   bool     `json:"is_featured"`
	AllowComment bool     `json:"allow_comment"`
	// 是否立即发布/提交审核：true=转 pending，false=仅保存草稿态
	IsPublished bool `json:"is_published"`
}

type ListArticlesRequest struct {
	Page       uint   `form:"page"`
	Size       uint   `form:"size"`
	CategoryID uint   `form:"category_id"`
	Tag        string `form:"tag"`
	OrderBy    string `form:"order_by"`
	UserID     uint   `form:"user_id"`
}

type SearchArticlesRequest struct {
	Keyword string `form:"keyword" binding:"required"`
	Page    uint   `form:"page"`
	Size    uint   `form:"size"`
}

// AdminListArticlesRequest 后台审核列表请求（按状态筛选）
type AdminListArticlesRequest struct {
	Status string `form:"status"`
	Page   uint   `form:"page"`
	Size   uint   `form:"size"`
}

// ReviewArticleRequest 审核操作（拒绝/下线）请求，携带原因
type ReviewArticleRequest struct {
	Reason string `json:"reason"`
}
