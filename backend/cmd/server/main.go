package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	coreauthbootstrap "goadmin/core/auth/bootstrap"
	corebootstrap "goadmin/core/bootstrap"
	"goadmin/core/config"
	coreevent "goadmin/core/event"
	corelogger "goadmin/core/logger"
	coreregistry "goadmin/core/registry"
	coretenant "goadmin/core/tenant"
	infraDB "goadmin/infrastructure/db"
	authservice "goadmin/modules/auth/application/service"
	authsubscriber "goadmin/modules/auth/application/subscriber"
	authrepo "goadmin/modules/auth/infrastructure/repo"
	menuservice "goadmin/modules/menu/application/service"
	menurepopkg "goadmin/modules/menu/infrastructure/repo"
	userevent "goadmin/modules/user/application/event"
	pluginservice "goadmin/plugin/application/service"
	exampleplugin "goadmin/plugin/builtin/example"
	pluginrepopkg "goadmin/plugin/infrastructure/repo"
	pluginiface "goadmin/plugin/interface"
	pluginloader "goadmin/plugin/loader"
	ginserver "goadmin/transport/http/gin"
	"goadmin/transport/http/gin/router"

	"go.uber.org/zap"
)

func main() {
	projectRoot, err := findProjectRoot()
	if err != nil {
		log.Fatalf("detect project root: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	coretenant.SetEnabled(cfg.Tenant.Enabled)

	logger, err := corelogger.New(corelogger.Config{
		Level:       cfg.Logger.Level,
		Format:      cfg.Logger.Format,
		Output:      cfg.Logger.Output,
		Development: cfg.Logger.Development,
	})
	if err != nil {
		log.Fatalf("init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	dbConn, err := infraDB.Open(cfg.Database)
	if err != nil {
		logger.Fatal("open database", zap.Error(err))
	}

	authBundle, err := coreauthbootstrap.New(cfg, dbConn)
	if err != nil {
		logger.Fatal("init auth bundle", zap.Error(err))
	}

	credentials := authrepo.NewBootstrapStore(cfg.Auth.Bootstrap.Users)
	revocations := authservice.NewMemoryRevocationStore()
	authSvc, err := authservice.New(authBundle.JWT, authBundle.Authorizer, credentials, revocations)
	if err != nil {
		logger.Fatal("init auth service", zap.Error(err))
	}

	eventBus := coreevent.NewLocalBus(logger)
	if err := eventBus.Subscribe(userevent.CreatedTopic, authsubscriber.NewUserCreatedLogger(logger).Handle); err != nil {
		logger.Fatal("register event subscriber", zap.Error(err))
	}

	if err := pluginrepopkg.Migrate(dbConn); err != nil {
		logger.Fatal("migrate plugin repository", zap.Error(err))
	}
	if err := corebootstrap.MigrateAll(dbConn, corebootstrap.Modules()); err != nil {
		logger.Fatal("migrate generated modules", zap.Error(err))
	}
	if err := menurepopkg.SeedDefaults(dbConn); err != nil {
		logger.Fatal("seed default menus", zap.Error(err))
	}

	menuRepo, err := menurepopkg.NewGormRepository(dbConn)
	if err != nil {
		logger.Fatal("init menu repository", zap.Error(err))
	}
	pluginRepo, err := pluginrepopkg.NewGormRepository(dbConn)
	if err != nil {
		logger.Fatal("init plugin repository", zap.Error(err))
	}

	menuSvc, err := menuservice.New(menuRepo)
	if err != nil {
		logger.Fatal("init menu service", zap.Error(err))
	}
	pluginSvc, err := pluginservice.New(pluginRepo)
	if err != nil {
		logger.Fatal("init plugin service", zap.Error(err))
	}

	pluginContainer := coreregistry.New()
	pluginContainer.Register("config", cfg)
	pluginContainer.Register("logger", logger)
	pluginContainer.Register("event_bus", eventBus)

	pluginRegistry, err := pluginloader.Load(&pluginiface.Context{
		Config:    cfg,
		Logger:    logger,
		Container: pluginContainer,
	}, exampleplugin.New())
	if err != nil {
		logger.Fatal("load plugins", zap.Error(err))
	}
	if err := pluginSvc.SeedFromRegistry(context.Background(), pluginRegistry); err != nil {
		logger.Fatal("seed plugin definitions", zap.Error(err))
	}

	server, err := ginserver.New(cfg, logger, router.Dependencies{
		AuthService:    authSvc,
		MenuService:    menuSvc,
		PluginService:  pluginSvc,
		PluginRegistry: pluginRegistry,
		ProjectRoot:    projectRoot,
		// Generated modules use the shared bootstrap registry and only need DB/logger/event bus.
		BootstrapDeps: corebootstrap.Dependencies{
			DB:       dbConn,
			Logger:   logger,
			EventBus: eventBus,
		},
		JWT:         authBundle.JWT,
		Authorizer:  authBundle.Authorizer,
		Revocations: revocations,
	})
	if err != nil {
		logger.Fatal("init server", zap.Error(err))
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	logger.Info("starting server",
		zap.String("app", cfg.App.Name),
		zap.String("env", cfg.App.Env),
		zap.String("addr", cfg.HTTPAddr()),
		zap.Int("bootstrap_users", len(cfg.Auth.Bootstrap.Users)),
	)

	if err := server.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Fatal("server exited with error", zap.Error(err))
	}

	logger.Info("server stopped gracefully")
}

func findProjectRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("detect cwd: %w", err)
	}
	current := cwd
	for {
		if fileExists(filepath.Join(current, "go.work")) {
			return current, nil
		}
		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return cwd, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
