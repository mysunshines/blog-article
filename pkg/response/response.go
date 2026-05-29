package response

import (
	"net/http"

	"github.com/mysunshines/blog-article/pkg/errors"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code    errors.Code `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageResponse struct {
	Code    errors.Code `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Total   uint32      `json:"total,omitempty"`
	Page    uint32      `json:"page,omitempty"`
	Size    uint32      `json:"size,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.ErrCodeSuccess,
		Message: "success",
		Data:    data,
	})
}

func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    errors.ErrCodeSuccess,
		Message: message,
		Data:    data,
	})
}

func SuccessPage(c *gin.Context, data interface{}, total, page, size uint32) {
	c.JSON(http.StatusOK, PageResponse{
		Code:    errors.ErrCodeSuccess,
		Message: "success",
		Data:    data,
		Total:   total,
		Page:    page,
		Size:    size,
	})
}

func Fail(c *gin.Context, err error) {
	if appErr, ok := err.(*errors.AppError); ok {
		c.JSON(appErr.HTTP, Response{
			Code:    appErr.Code,
			Message: appErr.Message,
		})
		return
	}

	c.JSON(http.StatusInternalServerError, Response{
		Code:    errors.ErrCodeInternal,
		Message: err.Error(),
	})
}

func FailWithCode(c *gin.Context, code errors.Code, message string, httpStatus int) {
	c.JSON(httpStatus, Response{
		Code:    code,
		Message: message,
	})
}

func BadRequest(c *gin.Context, message string) {
	FailWithCode(c, errors.ErrCodeBadRequest, message, http.StatusBadRequest)
}

func Unauthorized(c *gin.Context, message string) {
	FailWithCode(c, errors.ErrCodeUnauthorized, message, http.StatusUnauthorized)
}

func Forbidden(c *gin.Context, message string) {
	FailWithCode(c, errors.ErrCodeForbidden, message, http.StatusForbidden)
}

func NotFound(c *gin.Context, message string) {
	FailWithCode(c, errors.ErrCodeNotFound, message, http.StatusNotFound)
}

func InternalError(c *gin.Context, message string) {
	FailWithCode(c, errors.ErrCodeInternal, message, http.StatusInternalServerError)
}

func TooManyRequests(c *gin.Context, message string) {
	FailWithCode(c, errors.ErrCodeRateLimited, message, http.StatusTooManyRequests)
}
