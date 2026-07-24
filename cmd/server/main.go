package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mysunshines/blog-article/internal/config"
	"github.com/mysunshines/blog-article/internal/handler"
	"github.com/mysunshines/blog-article/internal/model"
	"github.com/mysunshines/blog-article/internal/repository"
	"github.com/mysunshines/blog-article/internal/service"
	article "github.com/mysunshines/blog-article/proto/pb"
	"github.com/mysunshines/gocommon/cache"
	"github.com/mysunshines/gocommon/constants"
	common_database "github.com/mysunshines/gocommon/database"
	"github.com/mysunshines/gocommon/log"
	"github.com/mysunshines/gocommon/metrics"
	commonmiddleware "github.com/mysunshines/gocommon/middleware"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"gorm.io/gorm"
)

// Version 由构建脚本通过 -ldflags "-X main.Version=xxx" 注入，未注入时默认 "dev"。
var Version = "dev"

type Server struct {
	cfg          *config.Config
	httpServer   *http.Server
	grpcServer   *grpc.Server
	articleSvc   service.ArticleService
	articleRepo  repository.ArticleRepository
	articleHandl *handler.ArticleHandler
	db           *gorm.DB
	cb           *gobreaker.CircuitBreaker // 熔断器
}

func NewServer(cfg *config.Config) *Server {
	// 初始化数据库（类型别名，直接传递）
	if err := common_database.Init(&cfg.Database, cfg.App.Env); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	db := common_database.GetDB()

	// 初始化 Redis 缓存（必须在 AutoMigrate 之前，用于分布式锁）
	redisCfg := cfg.Redis
	redisCfg.KeyPrefix = constants.RedisKeyPrefixArticle
	if err := cache.Init(&redisCfg); err != nil {
		log.Warnf("Warning: Failed to init Redis: %v", err)
	}

	// 自动迁移（分布式锁保护，多实例只有一个执行）
	const migrationLockKey = "migration:lock:article_service"
	const migrationLockTTL = 60 * time.Second
	hostname, _ := os.Hostname()
	instanceID := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	acquired, err := cache.TryLock(context.Background(), migrationLockKey, instanceID, migrationLockTTL)
	if err != nil {
		log.Warnf("Failed to acquire migration lock (Redis unavailable): %v, proceeding without lock", err)
	} else if acquired {
		log.Infof("Migration lock acquired by instance %s", instanceID)
		defer func() {
			if unlockErr := cache.Unlock(context.Background(), migrationLockKey, instanceID); unlockErr != nil {
				log.Warnf("Failed to release migration lock: %v", unlockErr)
			}
		}()
	} else {
		log.Info("Migration lock held by another instance, skipping AutoMigrate")
		time.Sleep(2 * time.Second)
	}

	if acquired || err != nil {
		if migrateErr := db.AutoMigrate(&model.Article{}, &model.Category{}, &model.Tag{}, &model.ArticleTag{}, &model.User{}); migrateErr != nil {
			log.Fatalf("Failed to migrate database: %v", migrateErr)
		}
		// 兼容性迁移：将历史已发布文章（status 为空）补齐为 published
		if fixErr := db.Model(&model.Article{}).
			Where("status = ? AND is_published = ?", "", true).
			Update("status", model.ArticleStatusPublished).Error; fixErr != nil {
			log.Warnf("Failed to backfill article status: %v", fixErr)
		}
	}

	// 初始化限流器（类型别名，直接传递）
	commonmiddleware.InitRateLimiter(&cfg.RateLimit)

	// 初始化 JWT
	commonmiddleware.InitJWT(cfg.JWT.Secret)

	// 初始化熔断器
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        constants.ServiceNameArticle,
		MaxRequests: constants.DefaultCBMaxRequests,
		Interval:    constants.DefaultCBInterval * time.Second,
		Timeout:     constants.DefaultCBTimeout * time.Second,
	})

	// 初始化仓储层
	articleRepo := repository.NewArticleRepository(db)

	// 初始化服务层
	articleSvc := service.NewArticleService(articleRepo, db)

	// 初始化处理器
	articleHandl := handler.NewArticleHandler(articleSvc)

	return &Server{
		cfg:          cfg,
		articleSvc:   articleSvc,
		articleRepo:  articleRepo,
		articleHandl: articleHandl,
		db:           db,
		cb:           cb,
	}
}

func (s *Server) Run() error {
	// 启动 HTTP 服务器
	go s.runHTTPServer()

	// 启动 gRPC 服务器
	go s.runGRPCServer()

	// 启动 Prometheus 指标服务器
	if s.cfg.Metrics.Enabled {
		go s.runMetricsServer()
	}

	// 等待信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(ctx); err != nil {
		log.Errorf("HTTP server shutdown error: %v", err)
	}

	s.grpcServer.GracefulStop()

	log.Info("Server exited")
	return nil
}

func (s *Server) runHTTPServer() {
	if s.cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(commonmiddleware.RecoveryMiddleware())
	router.Use(commonmiddleware.LoggingMiddleware())
	router.Use(commonmiddleware.CORSMiddleware())
	router.Use(commonmiddleware.MetricsMiddleware(constants.ServiceNameArticle))
	router.Use(commonmiddleware.TraceMiddleware())

	// 高并发增强：请求超时中间件
	router.Use(commonmiddleware.TimeoutMiddleware(30 * time.Second))

	// 高并发增强：限流中间件
	if s.cfg.RateLimit.Enabled {
		router.Use(commonmiddleware.RateLimitMiddleware())
	}

	// 健康检查（带深度检查）
	router.GET("/health", func(c *gin.Context) {
		// 检查数据库连接
		if sqlDB, _ := s.db.DB(); sqlDB != nil {
			if err := sqlDB.Ping(); err != nil {
				c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "reason": "db"})
				return
			}
		}

		// 检查 Redis 连接
		if err := cache.Ping(context.Background()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy", "reason": "redis"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 就绪探针
	router.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// 版本信息
	router.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": Version})
	})

	// API 路由
	api := router.Group(constants.APIPathPrefix)
	{
		articleGroup := api.Group("/article")
		{
			articleGroup.GET("", s.articleHandl.ListArticles)
			articleGroup.GET("/search", s.articleHandl.SearchArticles)
			articleGroup.GET("/categories", s.articleHandl.GetCategories)
			articleGroup.GET("/tags", s.articleHandl.GetTags)
			articleGroup.GET("/:id", s.articleHandl.GetArticle)
			articleGroup.GET("/slug/:slug", s.articleHandl.GetArticleBySlug)
			articleGroup.GET("/user/:user_id", s.articleHandl.GetUserArticles)
			articleGroup.POST("", commonmiddleware.JWTValidMiddleware(), s.articleHandl.CreateArticle)
			articleGroup.PUT("/:id", commonmiddleware.JWTValidMiddleware(), s.articleHandl.UpdateArticle)
			articleGroup.DELETE("/:id", commonmiddleware.JWTValidMiddleware(), s.articleHandl.DeleteArticle)
			articleGroup.POST("/:id/view", s.articleHandl.IncrementViewCount)
		}

		// 后台审核管理（管理员专属）
		adminGroup := api.Group("/admin/articles")
		adminGroup.Use(commonmiddleware.JWTValidMiddleware(), commonmiddleware.AdminOnlyMiddleware())
		{
			adminGroup.GET("", s.articleHandl.ListArticlesForAdmin)
			adminGroup.GET("/:id", s.articleHandl.GetArticle)
			adminGroup.POST("/:id/approve", s.articleHandl.ApproveArticle)
			adminGroup.POST("/:id/reject", s.articleHandl.RejectArticle)
			adminGroup.POST("/:id/offline", s.articleHandl.OfflineArticle)
			adminGroup.POST("/:id/publish", s.articleHandl.PublishArticle)
			adminGroup.PUT("/:id", s.articleHandl.AdminUpdateArticle)
			adminGroup.DELETE("/:id", s.articleHandl.AdminDeleteArticle)
		}
	}

	addr := s.cfg.HTTP.Addr()

	// 高并发增强：配置 HTTP Server 超时
	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadTimeout:       constants.DefaultReadTimeout * time.Second,
		ReadHeaderTimeout: constants.DefaultReadHeaderTimeout * time.Second,
		WriteTimeout:      constants.DefaultWriteTimeout * time.Second,
		IdleTimeout:       constants.DefaultIdleTimeout * time.Second,
		MaxHeaderBytes:    constants.MaxHeaderBytes,
	}

	log.Infof("HTTP server starting on %s (timeouts: read=%v, write=%v, idle=%v)", addr,
		s.httpServer.ReadTimeout, s.httpServer.WriteTimeout, s.httpServer.IdleTimeout)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start HTTP server: %v", err)
	}
}

func (s *Server) runGRPCServer() {
	lis, err := net.Listen("tcp", s.cfg.GRPC.Addr())
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	// 高并发增强：gRPC 选项配置
	grpcOpts := []grpc.ServerOption{
		// 连接超时
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     constants.DefaultGRPCMaxConnectionIdle * time.Second,
			MaxConnectionAge:      constants.DefaultGRPCMaxConnectionAge * time.Second,
			MaxConnectionAgeGrace: constants.DefaultGRPCMaxConnectionAgeGrace * time.Second,
		}),
		// 超时配置
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             constants.DefaultGRPCMinPingInterval * time.Second,
			PermitWithoutStream: true,
		}),
		// 最大并发连接数
		grpc.MaxConcurrentStreams(constants.DefaultGRPCMaxConcurrentStreams),
	}

	// 高并发增强：添加 unary 拦截器（超时+熔断）
	grpcOpts = append(grpcOpts, grpc.UnaryInterceptor(s.grpcUnaryInterceptor))

	s.grpcServer = grpc.NewServer(grpcOpts...)
	article.RegisterArticleServiceServer(s.grpcServer, &handler.GrpcArticleHandler{
		Svc: s.articleSvc,
		Cb:  s.cb,
	})
	reflection.Register(s.grpcServer)

	log.Infof("gRPC server starting on %s", s.cfg.GRPC.Addr())
	if err := s.grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// grpcUnaryInterceptor gRPC 一元拦截器（超时+熔断）
func (s *Server) grpcUnaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// 超时控制
	ctx, cancel := context.WithTimeout(ctx, constants.DefaultGRPCUnaryTimeout*time.Second)
	defer cancel()

	// 熔断器保护
	if s.cb != nil {
		result, err := s.cb.Execute(func() (interface{}, error) {
			return handler(ctx, req)
		})
		return result, err
	}

	return handler(ctx, req)
}

func (s *Server) runMetricsServer() {
	addr := fmt.Sprintf(":%d", s.cfg.Metrics.Port)
	http.Handle(s.cfg.Metrics.Path, promhttp.Handler())

	log.Infof("Metrics server starting on %s%s", addr, s.cfg.Metrics.Path)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Errorf("Metrics server error: %v", err)
	}
}

func main() {
	// 加载配置（从项目根目录加载）
	configPath := os.Getenv(constants.EnvConfigPath)
	if configPath == "" {
		// 默认从当前目录加载
		configPath = constants.DefaultConfigPath
	}
	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log.Init(cfg.App.LogDir, cfg.App.LogLevel, constants.ServiceNameArticle)

	// 初始化指标
	metrics.Init()

	// 创建并运行服务器
	server := NewServer(cfg)
	defer common_database.Close()
	defer cache.Close()
	defer log.StopRotation()
	if err := server.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
