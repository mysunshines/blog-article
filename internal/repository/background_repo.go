package repository

import (
	"context"

	"github.com/mysunshines/blog-article/internal/model"
	"gorm.io/gorm"
)

// BackgroundRepository 文章背景数据访问（三期）
type BackgroundRepository struct {
	db *gorm.DB
}

// NewBackgroundRepository 构造背景仓储
func NewBackgroundRepository(db *gorm.DB) *BackgroundRepository {
	return &BackgroundRepository{db: db}
}

// ListEnabled 上架背景列表（按 sort 升序，供前端选择）
func (r *BackgroundRepository) ListEnabled(ctx context.Context) ([]*model.ArticleBackground, error) {
	var list []*model.ArticleBackground
	err := r.db.WithContext(ctx).
		Where("status = ?", model.BackgroundStatusOnline).
		Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// ListAll 全量背景（后台管理，含下架）
func (r *BackgroundRepository) ListAll(ctx context.Context) ([]*model.ArticleBackground, error) {
	var list []*model.ArticleBackground
	err := r.db.WithContext(ctx).Order("sort ASC, id ASC").Find(&list).Error
	return list, err
}

// GetByID 按 ID 取背景
func (r *BackgroundRepository) GetByID(ctx context.Context, id uint) (*model.ArticleBackground, error) {
	var bg model.ArticleBackground
	if err := r.db.WithContext(ctx).First(&bg, id).Error; err != nil {
		return nil, err
	}
	return &bg, nil
}

// Create 新增背景
func (r *BackgroundRepository) Create(ctx context.Context, bg *model.ArticleBackground) error {
	return r.db.WithContext(ctx).Create(bg).Error
}

// Update 更新背景（只更新非零值字段）
func (r *BackgroundRepository) Update(ctx context.Context, bg *model.ArticleBackground) error {
	return r.db.WithContext(ctx).Model(bg).Select("*").Omit("created_at").Updates(bg).Error
}

// UpdateFields 按字段名显式更新（可写入零值，如 status=0 下架）
func (r *BackgroundRepository) UpdateFields(ctx context.Context, id uint, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&model.ArticleBackground{}).
		Where("id = ?", id).Updates(updates).Error
}

// Delete 删除背景
func (r *BackgroundRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.ArticleBackground{}, id).Error
}
