package models

import (
	"github.com/mysunshines/blog-article/internal/model"
)

// Re-export types from internal/model for backward compatibility
type (
	Article    = model.Article
	Tag        = model.Tag
	Category   = model.Category
	User       = model.User
	ArticleTag = model.ArticleTag
)

const ArticleTableName = "articles"
