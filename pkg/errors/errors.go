package errors

import (
	"fmt"
	"net/http"
)

type Code int

const (
	// 成功
	ErrCodeSuccess Code = 0

	// 通用错误码 3000x
	ErrCodeBadRequest   Code = 30001
	ErrCodeUnauthorized Code = 30002
	ErrCodeForbidden    Code = 30003
	ErrCodeNotFound     Code = 30004
	ErrCodeInternal     Code = 30005
	ErrCodeServiceDown  Code = 30006
	ErrCodeTimeout      Code = 30007
	ErrCodeRateLimited  Code = 30008

	// 业务错误码 3001x
	ErrCodeArticleNotFound     Code = 30011
	ErrCodeArticleCreateFailed Code = 30012
	ErrCodeArticleUpdateFailed Code = 30013
	ErrCodeArticleDeleteFailed Code = 30014
	ErrCodePermissionDenied    Code = 30015
	ErrCodeCategoryNotFound    Code = 30016
	ErrCodeTagNotFound         Code = 30017
	ErrCodeInvalidSlug         Code = 30018
)

type AppError struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Err     error  `json:"-"`
	HTTP    int    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
	}
}

func BadRequest(message string) *AppError {
	return New(ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(ErrCodeNotFound, message, http.StatusNotFound)
}

func Internal(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func ServiceDown(message string) *AppError {
	return New(ErrCodeServiceDown, message, http.StatusServiceUnavailable)
}

func Timeout(message string) *AppError {
	return New(ErrCodeTimeout, message, http.StatusGatewayTimeout)
}

func RateLimited(message string) *AppError {
	return New(ErrCodeRateLimited, message, http.StatusTooManyRequests)
}

// 业务错误
func ArticleNotFound() *AppError {
	return New(ErrCodeArticleNotFound, "文章不存在", http.StatusNotFound)
}

func ArticleCreateFailed(err error) *AppError {
	return &AppError{
		Code:    ErrCodeArticleCreateFailed,
		Message: "创建文章失败",
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func ArticleUpdateFailed(err error) *AppError {
	return &AppError{
		Code:    ErrCodeArticleUpdateFailed,
		Message: "更新文章失败",
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func ArticleDeleteFailed(err error) *AppError {
	return &AppError{
		Code:    ErrCodeArticleDeleteFailed,
		Message: "删除文章失败",
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func PermissionDenied() *AppError {
	return New(ErrCodePermissionDenied, "权限不足", http.StatusForbidden)
}

func CategoryNotFound() *AppError {
	return New(ErrCodeCategoryNotFound, "分类不存在", http.StatusNotFound)
}

func TagNotFound() *AppError {
	return New(ErrCodeTagNotFound, "标签不存在", http.StatusNotFound)
}

func InvalidSlug() *AppError {
	return New(ErrCodeInvalidSlug, "URL别名无效", http.StatusBadRequest)
}

func IsNotFound(err error) bool {
	if e, ok := err.(*AppError); ok {
		return e.HTTP == http.StatusNotFound
	}
	return false
}

func IsForbidden(err error) bool {
	if e, ok := err.(*AppError); ok {
		return e.HTTP == http.StatusForbidden
	}
	return false
}
