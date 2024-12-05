package main

import (
	"context"
	"encoding/json"
	"fmt"
	"gin-splitwise/internal/account"
	accountDB "gin-splitwise/internal/account/database"
	"gin-splitwise/internal/cache"
	"gin-splitwise/internal/config"
	"gin-splitwise/internal/database"
	"gin-splitwise/internal/group"
	"gin-splitwise/internal/metric"
	"gin-splitwise/internal/middleware"
	"gin-splitwise/pkg/logging"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var serverCmd = &cobra.Command{
	Use: "server",
	Run: func(cmd *cobra.Command, args []string) {
		runApplication()
	},
}

func runApplication() {
	conf, err := config.Load(configFile)
	if err != nil {
		log.Fatal(err)
	}
	logging.SetConfig(&logging.Config{
		Encoding:    conf.LoggingConfig.Encoding,
		Level:       zapcore.Level(conf.LoggingConfig.Level),
		Development: conf.LoggingConfig.Development,
	})
	defer logging.DefaultLogger().Sync()

	app := fx.New(
		fx.Supply(conf),
		fx.Supply(logging.DefaultLogger().Desugar()),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger {
			return &fxevent.ZapLogger{Logger: log.Named("fx")}
		}),
		fx.StopTimeout(conf.ServerConfig.GracefulShutdown+time.Second),
		fx.Invoke(
			printAppInfo,
		),
		fx.Provide(
			metric.NewMetricsProvider,
			// setup database
			database.NewDatabase,
			// setup cache
			cache.NewCacher,
			// setup account packages
			accountDB.NewAccountDB,
			account.NewAuthMiddleware,
			account.NewHandler,
			// setup article packages
			// server
			newServer,
		),
		fx.Invoke(
			account.RouteV1,
			group.RouteV1,
			func(r *gin.Engine) {},
		),
	)
	app.Run()
}

func newServer(lc fx.Lifecycle, cfg *config.Config, mp *metric.MetricsProvider) *gin.Engine {
	gin.SetMode(gin.DebugMode)
	r := gin.New()
	r.Use(middleware.LoggingMiddleware("/metric"), gin.Recovery())

	metric.Route(r)
	r.Use(metric.MetricsMiddleware(mp))

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.ServerConfig.Port),
		Handler: r,
		ReadTimeout: cfg.ServerConfig.ReadTimeout,
		WriteTimeout: cfg.ServerConfig.WriteTimeout,
	}
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			logging.FromContext(ctx).Infof("starting rest api server on port %d", cfg.ServerConfig.Port)
			go func() {
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					logging.DefaultLogger().Errorw("failed to start server", "err", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			logging.FromContext(ctx).Info("Shutting down server...")
			return srv.Shutdown(ctx)
		},
	})
	return r
}

func printAppInfo(cfg *config.Config) {
	b, _ := json.MarshalIndent(&cfg, "", " ")
	logging.DefaultLogger().Infof("APPLICATION INFORMATION \n%s", string(b))
}