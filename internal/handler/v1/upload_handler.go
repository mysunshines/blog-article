package v1

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"

	"github.com/mysunshines/gocommon/response"

	"github.com/gin-gonic/gin"
)

// UploadsDir 上传文件落盘目录（Docker 中已预创建并 chown 给运行用户 /app/uploads）。
const UploadsDir = "/app/uploads"

// allowedImageExt 允许的图片 Content-Type → 扩展名。
var allowedImageExt = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

// UploadCover 接收封面上传（multipart/form-data，字段名 file）。
// 鉴权由路由上的 JWTValidMiddleware 保证（任意登录用户均可上传，非管理员专属）。
// 请求经 /admin-api/ 透传（网关 httpreverse 对原始 HTTP 透传，可携带二进制），
// 返回的图片 URL 同样走该前缀，保证浏览器同源可访问。
//
// 注意：受网关 / 服务的 MaxRequestBody（1MB）限制，单文件建议 ≤ 800KB，
// 更大的封面图片请使用“封面图片链接”输入框。
func (h *ArticleHandler) UploadCover(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "未选择文件或上传失败")
		return
	}

	// 800KB：为 multipart 边界/头预留空间，避免突破网关 1MB 请求体上限（413）。
	if file.Size > 800*1024 {
		response.BadRequest(c, "图片大小不能超过 800KB（过大请用图片链接）")
		return
	}

	contentType := file.Header.Get("Content-Type")
	ext, ok := allowedImageExt[contentType]
	if !ok {
		response.BadRequest(c, "仅支持 JPEG/PNG/GIF/WebP 图片")
		return
	}

	src, err := file.Open()
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer src.Close()

	if err := os.MkdirAll(UploadsDir, 0o755); err != nil {
		response.Fail(c, err)
		return
	}

	name, err := randHex(16)
	if err != nil {
		response.Fail(c, err)
		return
	}
	dst := filepath.Join(UploadsDir, name+ext)

	out, err := os.Create(dst)
	if err != nil {
		response.Fail(c, err)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, src); err != nil {
		response.Fail(c, err)
		return
	}

	// 返回同源可访问的 URL（nginx→gateway→httpreverse 已代理 /admin-api/ 前缀）。
	url := "/admin-api/uploads/" + name + ext
	response.Success(c, map[string]string{"url": url})
}

func randHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
