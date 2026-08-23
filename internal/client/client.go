// Package client 是 article-service 统一的「下游 gRPC 客户端适配层」。
//
// 职责边界：
//   - gocommon/grpcclient 是 transport 层，只管连接、调用、鉴权透传、服务发现。
//   - 本包是业务适配层：把「调哪个下游、拼什么请求、错误如何映射」收拢在一处，
//     handler 不直接 import 各下游 proto 后散落调用 SendRequest，而是调用本包提供的
//     语义化函数（如 RecordAudit）。
//
// 服务发现：
//   - 下游地址解析由 consul.UseConsulDiscovery 注入的 grpcclient.ServiceResolver 接管
//     （main.go 启动时启用）。SendRequest 传入 alias（如 user.v1.UserService）后，
//     resolver 会按命名约定推导 Consul 服务名（user.v1.UserService → user-service）
//     并选择健康实例，因此本包无需硬编码下游地址。
//
// 扩展约定：
//   - 新增一个下游依赖时，在本包新增一个 <service>.go 文件，封装该服务的语义化调用。
//   - 例如将来要调 comment-service / notification-service，分别加 client/comment.go、
//     client/notification.go，handler 保持只写业务语义、不碰 proto 拼装与 SendRequest。
package client
