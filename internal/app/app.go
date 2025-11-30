package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"comments-system/internal/config"
	comments_h "comments-system/internal/http-server/handler/comments"
	posts_h "comments-system/internal/http-server/handler/posts"
	"comments-system/internal/http-server/middleware"
	"comments-system/internal/http-server/router"
	comments_cache "comments-system/internal/repository/cache/comments/redis"
	posts_cache "comments-system/internal/repository/cache/posts/redis"
	comments_postgres "comments-system/internal/repository/comments/postgres"
	posts_postgres "comments-system/internal/repository/posts/postgres"
	comments_uc "comments-system/internal/usecase/comments"
	posts_uc "comments-system/internal/usecase/posts"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

type App struct {
	cfg    *config.Config
	server *http.Server
	logger *zlog.Zerolog
	db     *dbpg.DB
}

func NewApp(cfg *config.Config, logger *zlog.Zerolog) (*App, error) {
	retries := cfg.DefaultRetryStrategy()

	dbOpts := &dbpg.Options{
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}

	db, err := dbpg.New(cfg.DBDSN(), cfg.DB.Slaves, dbOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	commentsCache := comments_cache.NewCommentsCache(cfg, retries)
	postsCache := posts_cache.NewPostsCache(cfg, retries)

	commentsRepo := comments_postgres.NewCommentsRepository(db, retries)
	postsRepo := posts_postgres.NewPostsRepository(db, retries)

	commentsUsecase := comments_uc.NewCommentsUsecase(commentsRepo, commentsCache, logger)
	postsUsecase := posts_uc.NewPostsUsecase(postsRepo, postsCache, logger)

	commentsHandler := comments_h.NewCommentsHandler(commentsUsecase, logger)
	postsHandler := posts_h.NewPostsHandler(postsUsecase, logger)

	h := &router.Handler{
		CommentsHandler: commentsHandler,
		PostsHandler:    postsHandler,
	}

	mux := router.SetupRouter(h)
	muxWM := middleware.LoggingMiddleware(mux)

	server := &http.Server{
		Addr:         ":" + cfg.Server.Addr,
		Handler:      muxWM,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	return &App{
		cfg:    cfg,
		server: server,
		logger: logger,
		db:     db,
	}, nil
}

func (a *App) Run() error {
	a.logger.Info().Str("addr", a.cfg.Server.Addr).Msg("Starting server")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.handleSignals(cancel)

	serverErr := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		a.logger.Error().Err(err).Msg("Server error")
		return err
	case <-ctx.Done():
		a.logger.Info().Msg("Shutting down server")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), a.cfg.Server.ShutdownTimeout)
		defer shutdownCancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			a.logger.Error().Err(err).Msg("Server shutdown failed")
		}

		a.db.Master.Close()
		a.logger.Info().Msg("Server stopped gracefully")
		return nil
	}
}

func (a *App) handleSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	a.logger.Info().Str("signal", sig.String()).Msg("Received signal")
	cancel()
}
