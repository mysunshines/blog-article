// Package errors 定义 article-service 的应用错误类型。
//
// 设计约定（与「各服务 ErrCode 使用自己 proto 中定义的 enum」架构一致）：
//   - 业务错误码（30001-30018）由各服务的 proto enum 定义，本包仅以常量形式镜像，
//     供 service / repository 构造错误时引用，数值必须与 proto ArticleErrorCode 对齐。
//   - 通用错误码（10001-10008）由 gocommon 定义并透传，不在此重复声明。
//   - handler 层通过 errors.As 取出 *AppError 的 Code，再映射到 proto 枚举值写入响应，
//     避免 handler 硬编码码值、与 service 错误解耦。
package errors

import (
	stderrors "errors"
	"fmt"
	"net/http"

	"github.com/mysunshines/gocommon/constants"
)

// As 是标准库 errors.As 针对 *AppError 的便捷封装，供 handler 层断言应用错误。
func As(err error, target **AppError) bool {
	return stderrors.As(err, target)
}

// Code 是错误码类型，数值与 proto ArticleErrorCode 及 gocommon 通用码一致。
type Code int

// 业务错误码（区间 30000-39999），必须与 proto/article.proto 的 ArticleErrorCode 对齐。
const (
	// 通用码透传 gocommon（10001-10008），此处仅作别名引用，避免散落字面量。
	CodeBadRequest   Code = constants.ErrCodeBadRequest
	CodeUnauthorized Code = constants.ErrCodeUnauthorized
	CodeForbidden    Code = constants.ErrCodeForbidden
	CodeNotFound     Code = constants.ErrCodeNotFound
	CodeInternal     Code = constants.ErrCodeInternal
	CodeServiceDown  Code = constants.ErrCodeServiceUnavailable
	CodeTimeout      Code = constants.ErrCodeTimeout
	CodeRateLimited  Code = constants.ErrCodeRateLimited

	// ---- 文章业务错误码（与 proto ArticleErrorCode 对齐）----
	CodeArticleNotFound     Code = 30001
	CodeArticleCreateFailed Code = 30011
	CodeArticleUpdateFailed Code = 30012
	CodeArticleDeleteFailed Code = 30013
	CodeArticleListFailed   Code = 30014
	CodePermissionDenied    Code = 30015
	CodeCategoryNotFound    Code = 30016
	CodeTagNotFound         Code = 30017
	CodeInvalidSlug         Code = 30018
)

// AppError 是 article-service 的应用错误，携带业务码与可选底层错误，支持 errors.As 断言。
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

// New 构造一个 AppError。
func New(code Code, message string, httpStatus int) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		HTTP:    httpStatus,
	}
}

func BadRequest(message string) *AppError {
	return New(CodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(message string) *AppError {
	return New(CodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(message string) *AppError {
	return New(CodeForbidden, message, http.StatusForbidden)
}

func NotFound(message string) *AppError {
	return New(CodeNotFound, message, http.StatusNotFound)
}

func Internal(message string, err error) *AppError {
	return &AppError{
		Code:    CodeInternal,
		Message: message,
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func ServiceDown(message string) *AppError {
	return New(CodeServiceDown, message, http.StatusServiceUnavailable)
}

func Timeout(message string) *AppError {
	return New(CodeTimeout, message, http.StatusGatewayTimeout)
}

func RateLimited(message string) *AppError {
	return New(CodeRateLimited, message, http.StatusTooManyRequests)
}

// 业务错误
func ArticleNotFound() *AppError {
	return New(CodeArticleNotFound, "文章不存在", http.StatusNotFound)
}

func ArticleCreateFailed(err error) *AppError {
	return &AppError{
		Code:    CodeArticleCreateFailed,
		Message: "创建文章失败",
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func ArticleUpdateFailed(err error) *AppError {
	return &AppError{
		Code:    CodeArticleUpdateFailed,
		Message: "更新文章失败",
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func ArticleDeleteFailed(err error) *AppError {
	return &AppError{
		Code:    CodeArticleDeleteFailed,
		Message: "删除文章失败",
		Err:     err,
		HTTP:    http.StatusInternalServerError,
	}
}

func PermissionDenied() *AppError {
	return New(CodePermissionDenied, "权限不足", http.StatusForbidden)
}

func CategoryNotFound() *AppError {
	return New(CodeCategoryNotFound, "分类不存在", http.StatusNotFound)
}

func TagNotFound() *AppError {
	return New(CodeTagNotFound, "标签不存在", http.StatusNotFound)
}

func InvalidSlug() *AppError {
	return New(CodeInvalidSlug, "URL别名无效", http.StatusBadRequest)
}

// IsNotFound 判断 err 链中是否含 NotFound 语义的 AppError（供 handler 用 errors.As 判断）。
func IsNotFound(err error) bool {
	var ae *AppError
	if As(err, &ae) {
		return ae.HTTP == http.StatusNotFound
	}
	return false
}

// IsForbidden 判断 err 链中是否含 Forbidden 语义的 AppError。
func IsForbidden(err error) bool {
	var ae *AppError
	if As(err, &ae) {
		return ae.HTTP == http.StatusForbidden
	}
	return false
}
