# article-service 对外 API 文档

> 自动生成自 `article.proto`（模式：proto）。
> 网关按 `/api/v1/article/<snake_method>` 反射代理到 gRPC 方法 `article.v1.ArticleService/<Method>`。
> 生成时间：2026-09-21 19:37:19
> Base URL（网关入口）：http://localhost:8081

## 接口列表

| Method | Path | 鉴权 | 说明 |
| --- | --- | --- | --- |
| `POST` | `/api/v1/article/create_article` | 登录 |  |
| `GET` | `/api/v1/article/get_article` | 公开 |  |
| `PUT` | `/api/v1/article/update_article` | 公开 |  |
| `DELETE` | `/api/v1/article/delete_article` | 公开 |  |
| `GET` | `/api/v1/article/list_articles` | 公开 |  |
| `GET` | `/api/v1/article/get_article_by_slug` | 公开 |  |
| `POST` | `/api/v1/article/purchase_article` | 公开 | 付费阅读（二期）：支付积分解锁文章正文，作者获得积分收益 |
| `GET` | `/api/v1/article/list_backgrounds` | 公开 | 文章背景（三期）：列表（含「是否已拥有」）与积分购买 |
| `POST` | `/api/v1/article/buy_background` | 公开 |  |
| `POST` | `/api/v1/article/increment_view_count` | 公开 |  |
| `POST` | `/api/v1/article/like_article` | 公开 |  |
| `POST` | `/api/v1/article/cancel_like_article` | 公开 |  |
| `GET` | `/api/v1/article/get_like_status` | 公开 |  |
| `GET` | `/api/v1/article/list_author_articles` | 登录 | 作者主页（公开）：按作者 user_id 列出其「已发布」文章。 |
| `GET` | `/api/v1/article/search_articles` | 公开 |  |
| `GET` | `/api/v1/article/get_user_articles` | 公开 |  |
| `GET` | `/api/v1/article/get_categories` | 公开 |  |
| `GET` | `/api/v1/article/get_tags` | 公开 |  |
| `GET` | `/api/v1/article/list_articles_for_admin` | 管理员 |  |
| `GET` | `/api/v1/article/get_article_for_admin` | 公开 |  |
| `POST` | `/api/v1/article/approve_article` | 公开 |  |
| `POST` | `/api/v1/article/reject_article` | 公开 |  |
| `POST` | `/api/v1/article/offline_article` | 公开 |  |
| `POST` | `/api/v1/article/publish_article` | 公开 |  |
| `POST` | `/api/v1/article/submit_article` | 公开 |  |
| `PUT` | `/api/v1/article/admin_update_article` | 公开 |  |
| `POST` | `/api/v1/article/admin_delete_article` | 公开 |  |
| `POST` | `/api/v1/article/admin_create_background` | 公开 | 文章背景管理（三期） |
| `PUT` | `/api/v1/article/admin_update_background` | 公开 |  |
| `POST` | `/api/v1/article/admin_delete_background` | 公开 |  |
| `POST` | `/api/v1/article/create_category` | 管理员 |  |
| `PUT` | `/api/v1/article/update_category` | 公开 |  |
| `DELETE` | `/api/v1/article/delete_category` | 公开 |  |

## CreateArticle

- **URL**: `http://localhost:8081/api/v1/article/create_article`
- **Method**: `POST`
- **鉴权**: 登录（需 JWT）

### Headers
```http
Authorization: Bearer <token>
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `user_id` | `uint32` |  | `0` |
| `title` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `content_format` | `int32` |  | `0` |
| `summary` | `string` |  | `""` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `tags` | `string[]` |  | `[]` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |
| `is_published` | `bool` |  | `false` |
| `background_id` | `uint32` | 文章背景（三期）：0=默认背景，>0 需已拥有该背景 | `0` |

**Body 示例**：
```json
{"user_id": 0, "title": "", "content": "", "content_format": 0, "summary": "", "cover_image": "", "category_id": 0, "tags": [], "is_featured": false, "allow_comment": false, "is_published": false, "background_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/create_article' \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"user_id": 0, "title": "", "content": "", "content_format": 0, "summary": "", "cover_image": "", "category_id": 0, "tags": [], "is_featured": false, "allow_comment": false, "is_published": false, "background_id": 0}'
```

## GetArticle

- **URL**: `http://localhost:8081/api/v1/article/get_article?article_id=0`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Query 示例**：
```json
article_id=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_article?article_id=0'
```

## UpdateArticle

- **URL**: `http://localhost:8081/api/v1/article/update_article`
- **Method**: `PUT`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |
| `title` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `content_format` | `int32` |  | `0` |
| `summary` | `string` |  | `""` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `tags` | `string[]` |  | `[]` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |
| `is_published` | `bool` |  | `false` |
| `background_id` | `uint32` | 文章背景（三期）：0=默认背景，>0 需已拥有该背景 | `0` |

**Body 示例**：
```json
{"article_id": 0, "user_id": 0, "title": "", "content": "", "content_format": 0, "summary": "", "cover_image": "", "category_id": 0, "tags": [], "is_featured": false, "allow_comment": false, "is_published": false, "background_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X PUT 'http://localhost:8081/api/v1/article/update_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "user_id": 0, "title": "", "content": "", "content_format": 0, "summary": "", "cover_image": "", "category_id": 0, "tags": [], "is_featured": false, "allow_comment": false, "is_published": false, "background_id": 0}'
```

## DeleteArticle

- **URL**: `http://localhost:8081/api/v1/article/delete_article`
- **Method**: `DELETE`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0, "user_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Response 示例**：
```json
{"code": 0, "message": "success"}
```

### curl 示例
```bash
curl -X DELETE 'http://localhost:8081/api/v1/article/delete_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "user_id": 0}'
```

## ListArticles

- **URL**: `http://localhost:8081/api/v1/article/list_articles?page=0&page_size=0&category_id=0&tag=<tag>&order_by=<order_by>`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `category_id` | `uint32` |  | `0` |
| `tag` | `string` |  | `""` |
| `order_by` | `string` |  | `""` |

**Query 示例**：
```json
page=0&page_size=0&category_id=0&tag=<tag>&order_by=<order_by>
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "articles": [], "total": 0}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/list_articles?page=0&page_size=0&category_id=0&tag=<tag>&order_by=<order_by>'
```

## GetArticleBySlug

- **URL**: `http://localhost:8081/api/v1/article/get_article_by_slug?slug=<slug>`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `slug` | `string` |  | `""` |

**Query 示例**：
```json
slug=<slug>
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_article_by_slug?slug=<slug>'
```

## PurchaseArticle

- **URL**: `http://localhost:8081/api/v1/article/purchase_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `balance` | `int64` | 支付后剩余积分 | `0` |
| `content_html` | `string` | 解锁成功直接返回正文 HTML，省一次详情请求 | `""` |

**Response 示例**：
```json
{"code": 0, "message": "success", "balance": 0, "content_html": ""}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/purchase_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0}'
```

## ListBackgrounds

- **URL**: `http://localhost:8081/api/v1/article/list_backgrounds?code=0&message=<message>&background_id=0&code=0&message=<message>&balance=0&name=<name>&image_url=<image_url>&thumb_url=<thumb_url>&price=0&sort=0&code=0&message=<message>&background_id=0&name=<name>&image_url=<image_url>&thumb_url=<thumb_url>&price=0&status=0&sort=0&code=0&message=<message>&background_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&view_count=0&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&user_id=0&page=0&page_size=0&code=0&message=<message>&total=0&keyword=<keyword>&page=0&page_size=0&code=0&message=<message>&total=0&user_id=0&page=0&page_size=0&code=0&message=<message>&total=0&code=0&message=<message>&code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `backgrounds` | `Background[]` |  | [] |
| `background_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `balance` | `int64` | 购买后剩余积分 | `0` |
| `name` | `string` |  | `""` |
| `image_url` | `string` |  | `""` |
| `thumb_url` | `string` |  | `""` |
| `price` | `int64` |  | `0` |
| `sort` | `int32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `background` | `Background` |  | 见 [Background](#background) |
| `background_id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `image_url` | `string` |  | `""` |
| `thumb_url` | `string` |  | `""` |
| `price` | `int64` |  | `0` |
| `status` | `uint32` |  | `0` |
| `sort` | `int32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `background` | `Background` |  | 见 [Background](#background) |
| `background_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `view_count` | `uint32` |  | `0` |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `like_count` | `uint32` |  | `0` |
| `liked` | `bool` |  | `false` |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `like_count` | `uint32` |  | `0` |
| `liked` | `bool` |  | `false` |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `like_count` | `uint32` |  | `0` |
| `liked` | `bool` |  | `false` |
| `user_id` | `uint32` | 作者用户 ID | `0` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |
| `keyword` | `string` |  | `""` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `categories` | `Category[]` |  | [] |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `tags` | `Tag[]` |  | [] |
| `status` | `string` | 按状态筛选：pending/published/offline/rejected/draft，空=全部 | `""` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `title` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `summary` | `string` |  | `""` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |
| `category_id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |
| `category_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Query 示例**：
```json
code=0&message=<message>&background_id=0&code=0&message=<message>&balance=0&name=<name>&image_url=<image_url>&thumb_url=<thumb_url>&price=0&sort=0&code=0&message=<message>&background_id=0&name=<name>&image_url=<image_url>&thumb_url=<thumb_url>&price=0&status=0&sort=0&code=0&message=<message>&background_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&view_count=0&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&user_id=0&page=0&page_size=0&code=0&message=<message>&total=0&keyword=<keyword>&page=0&page_size=0&code=0&message=<message>&total=0&user_id=0&page=0&page_size=0&code=0&message=<message>&total=0&code=0&message=<message>&code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `backgrounds` | `Background[]` |  | [] |

**Response 示例**：
```json
{"code": 0, "message": "success", "backgrounds": []}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/list_backgrounds?code=0&message=<message>&background_id=0&code=0&message=<message>&balance=0&name=<name>&image_url=<image_url>&thumb_url=<thumb_url>&price=0&sort=0&code=0&message=<message>&background_id=0&name=<name>&image_url=<image_url>&thumb_url=<thumb_url>&price=0&status=0&sort=0&code=0&message=<message>&background_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&view_count=0&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&article_id=0&user_id=0&code=0&message=<message>&like_count=0&liked=false&user_id=0&page=0&page_size=0&code=0&message=<message>&total=0&keyword=<keyword>&page=0&page_size=0&code=0&message=<message>&total=0&user_id=0&page=0&page_size=0&code=0&message=<message>&total=0&code=0&message=<message>&code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>'
```

## BuyBackground

- **URL**: `http://localhost:8081/api/v1/article/buy_background`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `background_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"background_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `balance` | `int64` | 购买后剩余积分 | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "balance": 0}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/buy_background' \
  -H 'Content-Type: application/json' \
  -d '{"background_id": 0}'
```

## IncrementViewCount

- **URL**: `http://localhost:8081/api/v1/article/increment_view_count`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `view_count` | `uint32` |  | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "view_count": 0}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/increment_view_count' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0}'
```

## LikeArticle

- **URL**: `http://localhost:8081/api/v1/article/like_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0, "user_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `like_count` | `uint32` |  | `0` |
| `liked` | `bool` |  | `false` |

**Response 示例**：
```json
{"code": 0, "message": "success", "like_count": 0, "liked": false}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/like_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "user_id": 0}'
```

## CancelLikeArticle

- **URL**: `http://localhost:8081/api/v1/article/cancel_like_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0, "user_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `like_count` | `uint32` |  | `0` |
| `liked` | `bool` |  | `false` |

**Response 示例**：
```json
{"code": 0, "message": "success", "like_count": 0, "liked": false}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/cancel_like_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "user_id": 0}'
```

## GetLikeStatus

- **URL**: `http://localhost:8081/api/v1/article/get_like_status?article_id=0&user_id=0`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |

**Query 示例**：
```json
article_id=0&user_id=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `like_count` | `uint32` |  | `0` |
| `liked` | `bool` |  | `false` |

**Response 示例**：
```json
{"code": 0, "message": "success", "like_count": 0, "liked": false}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_like_status?article_id=0&user_id=0'
```

## ListAuthorArticles

- **URL**: `http://localhost:8081/api/v1/article/list_author_articles?user_id=0&page=0&page_size=0`
- **Method**: `GET`
- **鉴权**: 登录（需 JWT）

### Headers
```http
Authorization: Bearer <token>
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `user_id` | `uint32` | 作者用户 ID | `0` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |

**Query 示例**：
```json
user_id=0&page=0&page_size=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "articles": [], "total": 0}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/list_author_articles?user_id=0&page=0&page_size=0' \
  -H 'Authorization: Bearer <token>'
```

## SearchArticles

- **URL**: `http://localhost:8081/api/v1/article/search_articles?keyword=<keyword>&page=0&page_size=0`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `keyword` | `string` |  | `""` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |

**Query 示例**：
```json
keyword=<keyword>&page=0&page_size=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "articles": [], "total": 0}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/search_articles?keyword=<keyword>&page=0&page_size=0'
```

## GetUserArticles

- **URL**: `http://localhost:8081/api/v1/article/get_user_articles?user_id=0&page=0&page_size=0`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `user_id` | `uint32` |  | `0` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |

**Query 示例**：
```json
user_id=0&page=0&page_size=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "articles": [], "total": 0}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_user_articles?user_id=0&page=0&page_size=0'
```

## GetCategories

- **URL**: `http://localhost:8081/api/v1/article/get_categories?code=0&message=<message>&code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `categories` | `Category[]` |  | [] |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `tags` | `Tag[]` |  | [] |
| `status` | `string` | 按状态筛选：pending/published/offline/rejected/draft，空=全部 | `""` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `title` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `summary` | `string` |  | `""` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |
| `category_id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |
| `category_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Query 示例**：
```json
code=0&message=<message>&code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `categories` | `Category[]` |  | [] |

**Response 示例**：
```json
{"code": 0, "message": "success", "categories": []}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_categories?code=0&message=<message>&code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>'
```

## GetTags

- **URL**: `http://localhost:8081/api/v1/article/get_tags?code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `tags` | `Tag[]` |  | [] |
| `status` | `string` | 按状态筛选：pending/published/offline/rejected/draft，空=全部 | `""` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `title` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `summary` | `string` |  | `""` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |
| `article_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |
| `category_id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |
| `category_id` | `uint32` |  | `0` |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Query 示例**：
```json
code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `tags` | `Tag[]` |  | [] |

**Response 示例**：
```json
{"code": 0, "message": "success", "tags": []}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_tags?code=0&message=<message>&status=<status>&page=0&page_size=0&code=0&message=<message>&total=0&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&reason=<reason>&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&code=0&message=<message>&article_id=0&title=<title>&content=<content>&summary=<summary>&cover_image=<cover_image>&category_id=0&is_featured=false&allow_comment=false&code=0&message=<message>&article_id=0&code=0&message=<message>&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&name=<name>&description=<description>&sort=0&code=0&message=<message>&category_id=0&code=0&message=<message>'
```

## ListArticlesForAdmin

- **URL**: `http://localhost:8081/api/v1/article/list_articles_for_admin?status=<status>&page=0&page_size=0`
- **Method**: `GET`
- **鉴权**: 管理员（需 JWT + 管理员角色）

### Headers
```http
Authorization: Bearer <token>
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `status` | `string` | 按状态筛选：pending/published/offline/rejected/draft，空=全部 | `""` |
| `page` | `uint32` |  | `0` |
| `page_size` | `uint32` |  | `0` |

**Query 示例**：
```json
status=<status>&page=0&page_size=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `articles` | `Article[]` |  | [] |
| `total` | `uint32` |  | `0` |

**Response 示例**：
```json
{"code": 0, "message": "success", "articles": [], "total": 0}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/list_articles_for_admin?status=<status>&page=0&page_size=0' \
  -H 'Authorization: Bearer <token>'
```

## GetArticleForAdmin

- **URL**: `http://localhost:8081/api/v1/article/get_article_for_admin?article_id=0`
- **Method**: `GET`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Query String

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Query 示例**：
```json
article_id=0
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X GET 'http://localhost:8081/api/v1/article/get_article_for_admin?article_id=0'
```

## ApproveArticle

- **URL**: `http://localhost:8081/api/v1/article/approve_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/approve_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0}'
```

## RejectArticle

- **URL**: `http://localhost:8081/api/v1/article/reject_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |

**Body 示例**：
```json
{"article_id": 0, "reason": ""}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/reject_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "reason": ""}'
```

## OfflineArticle

- **URL**: `http://localhost:8081/api/v1/article/offline_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `reason` | `string` |  | `""` |

**Body 示例**：
```json
{"article_id": 0, "reason": ""}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/offline_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "reason": ""}'
```

## PublishArticle

- **URL**: `http://localhost:8081/api/v1/article/publish_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/publish_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0}'
```

## SubmitArticle

- **URL**: `http://localhost:8081/api/v1/article/submit_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/submit_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0}'
```

## AdminUpdateArticle

- **URL**: `http://localhost:8081/api/v1/article/admin_update_article`
- **Method**: `PUT`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |
| `title` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `summary` | `string` |  | `""` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |

**Body 示例**：
```json
{"article_id": 0, "title": "", "content": "", "summary": "", "cover_image": "", "category_id": 0, "is_featured": false, "allow_comment": false}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `article` | `Article` |  | 见 [Article](#article) |

**Response 示例**：
```json
{"code": 0, "message": "success", "article": {}}
```

### curl 示例
```bash
curl -X PUT 'http://localhost:8081/api/v1/article/admin_update_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0, "title": "", "content": "", "summary": "", "cover_image": "", "category_id": 0, "is_featured": false, "allow_comment": false}'
```

## AdminDeleteArticle

- **URL**: `http://localhost:8081/api/v1/article/admin_delete_article`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `article_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"article_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Response 示例**：
```json
{"code": 0, "message": "success"}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/admin_delete_article' \
  -H 'Content-Type: application/json' \
  -d '{"article_id": 0}'
```

## AdminCreateBackground

- **URL**: `http://localhost:8081/api/v1/article/admin_create_background`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `name` | `string` |  | `""` |
| `image_url` | `string` |  | `""` |
| `thumb_url` | `string` |  | `""` |
| `price` | `int64` |  | `0` |
| `sort` | `int32` |  | `0` |

**Body 示例**：
```json
{"name": "", "image_url": "", "thumb_url": "", "price": 0, "sort": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `background` | `Background` |  | 见 [Background](#background) |

**Response 示例**：
```json
{"code": 0, "message": "success", "background": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/admin_create_background' \
  -H 'Content-Type: application/json' \
  -d '{"name": "", "image_url": "", "thumb_url": "", "price": 0, "sort": 0}'
```

## AdminUpdateBackground

- **URL**: `http://localhost:8081/api/v1/article/admin_update_background`
- **Method**: `PUT`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `background_id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `image_url` | `string` |  | `""` |
| `thumb_url` | `string` |  | `""` |
| `price` | `int64` |  | `0` |
| `status` | `uint32` |  | `0` |
| `sort` | `int32` |  | `0` |

**Body 示例**：
```json
{"background_id": 0, "name": "", "image_url": "", "thumb_url": "", "price": 0, "status": 0, "sort": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `background` | `Background` |  | 见 [Background](#background) |

**Response 示例**：
```json
{"code": 0, "message": "success", "background": {}}
```

### curl 示例
```bash
curl -X PUT 'http://localhost:8081/api/v1/article/admin_update_background' \
  -H 'Content-Type: application/json' \
  -d '{"background_id": 0, "name": "", "image_url": "", "thumb_url": "", "price": 0, "status": 0, "sort": 0}'
```

## AdminDeleteBackground

- **URL**: `http://localhost:8081/api/v1/article/admin_delete_background`
- **Method**: `POST`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `background_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"background_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Response 示例**：
```json
{"code": 0, "message": "success"}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/admin_delete_background' \
  -H 'Content-Type: application/json' \
  -d '{"background_id": 0}'
```

## CreateCategory

- **URL**: `http://localhost:8081/api/v1/article/create_category`
- **Method**: `POST`
- **鉴权**: 管理员（需 JWT + 管理员角色）

### Headers
```http
Authorization: Bearer <token>
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |

**Body 示例**：
```json
{"name": "", "description": "", "sort": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |

**Response 示例**：
```json
{"code": 0, "message": "success", "category": {}}
```

### curl 示例
```bash
curl -X POST 'http://localhost:8081/api/v1/article/create_category' \
  -H 'Authorization: Bearer <token>' \
  -H 'Content-Type: application/json' \
  -d '{"name": "", "description": "", "sort": 0}'
```

## UpdateCategory

- **URL**: `http://localhost:8081/api/v1/article/update_category`
- **Method**: `PUT`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `category_id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `sort` | `uint32` |  | `0` |

**Body 示例**：
```json
{"category_id": 0, "name": "", "description": "", "sort": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |
| `category` | `Category` |  | 见 [Category](#category) |

**Response 示例**：
```json
{"code": 0, "message": "success", "category": {}}
```

### curl 示例
```bash
curl -X PUT 'http://localhost:8081/api/v1/article/update_category' \
  -H 'Content-Type: application/json' \
  -d '{"category_id": 0, "name": "", "description": "", "sort": 0}'
```

## DeleteCategory

- **URL**: `http://localhost:8081/api/v1/article/delete_category`
- **Method**: `DELETE`
- **鉴权**: 公开（无需鉴权）

### Headers
```http
Content-Type: application/json
```

### Request
**参数位置**：Request Body（JSON）

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `category_id` | `uint32` |  | `0` |

**Body 示例**：
```json
{"category_id": 0}
```

### Response
| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `code` | `uint32` |  | `0` |
| `message` | `string` |  | `""` |

**Response 示例**：
```json
{"code": 0, "message": "success"}
```

### curl 示例
```bash
curl -X DELETE 'http://localhost:8081/api/v1/article/delete_category' \
  -H 'Content-Type: application/json' \
  -d '{"category_id": 0}'
```

---

## 数据结构

> 下列 message / enum 被上述接口的请求或响应引用；结构体字段中的 message 类型可点击跳转到对应定义。

### Article

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `id` | `uint32` |  | `0` |
| `user_id` | `uint32` |  | `0` |
| `username` | `string` |  | `""` |
| `title` | `string` |  | `""` |
| `slug` | `string` |  | `""` |
| `summary` | `string` |  | `""` |
| `content` | `string` |  | `""` |
| `content_format` | `int32` |  | `0` |
| `cover_image` | `string` |  | `""` |
| `category_id` | `uint32` |  | `0` |
| `category_name` | `string` |  | `""` |
| `tags` | `string[]` |  | `[]` |
| `view_count` | `uint32` |  | `0` |
| `comment_count` | `uint32` |  | `0` |
| `like_count` | `uint32` |  | `0` |
| `is_featured` | `bool` |  | `false` |
| `allow_comment` | `bool` |  | `false` |
| `created_at` | `string` |  | `""` |
| `updated_at` | `string` |  | `""` |
| `published_at` | `string` |  | `""` |
| `status` | `string` | 文章状态: draft/pending/published/offline/rejected | `""` |
| `content_html` | `string` |  | `""` |
| `is_paid` | `bool` |  | `false` |
| `price` | `int64` | 阅读所需积分（0=免费） | `0` |
| `background_id` | `uint32` | 文章背景 ID（0=默认背景，三期） | `0` |

### Background

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `image_url` | `string` |  | `""` |
| `thumb_url` | `string` |  | `""` |
| `price` | `int64` | 0=免费，>0=需积分购买 | `0` |
| `status` | `uint32` | 1=上架 0=下架 | `0` |
| `sort` | `int32` |  | `0` |
| `owned` | `bool` | 当前用户是否已拥有（免费背景恒为 true，或已用积分购买） | `false` |

### Category

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `slug` | `string` |  | `""` |
| `description` | `string` |  | `""` |
| `article_count` | `uint32` |  | `0` |
| `parent_id` | `uint32` |  | `0` |
| `sort` | `uint32` |  | `0` |
| `created_at` | `string` |  | `""` |

### Tag

| 字段 | 类型 | 说明 | 示例 |
| --- | --- | --- | --- |
| `id` | `uint32` |  | `0` |
| `name` | `string` |  | `""` |
| `slug` | `string` |  | `""` |
| `article_count` | `uint32` |  | `0` |
| `created_at` | `string` |  | `""` |

