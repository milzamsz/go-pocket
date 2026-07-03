package app

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	"github.com/milzamsz/go-pocket/internal/config"
	"github.com/milzamsz/go-pocket/internal/server"
	"github.com/milzamsz/go-pocket/internal/services/auth"
	"github.com/milzamsz/go-pocket/internal/services/billing"
	"github.com/milzamsz/go-pocket/internal/services/email"
	"github.com/milzamsz/go-pocket/internal/services/kanban"
	"github.com/milzamsz/go-pocket/internal/services/products"
	"github.com/milzamsz/go-pocket/internal/services/tenancy"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func Bootstrap(pb *pocketbase.PocketBase) error {
	cfg := config.Load()
	emailSvc := email.New(cfg)

	deps := &server.Deps{
		Config:   cfg,
		Auth:     auth.NewWithConfigAndDependencies(pb, cfg, nil, time.Now, emailSvc, nil),
		Billing:  billing.New(cfg),
		Email:    emailSvc,
		Tenancy:  tenancy.New(tenancy.NewPocketBaseRepository(pb)),
		Products: products.New(pb),
		Kanban:   kanban.New(pb),
	}

	pb.OnServe().BindFunc(func(se *core.ServeEvent) error {
		warnIfCSSMissing(se.App)
		logConfigWarnings(se.App, cfg)
		runAuthTokenCleanupScheduler(se.App)
		server.RegisterRoutes(se, deps)
		return se.Next()
	})

	return nil
}

func logConfigWarnings(app core.App, cfg config.Config) {
	for _, warning := range cfg.Validate() {
		app.Logger().Warn("configuration warning", "detail", warning)
	}
}

func warnIfCSSMissing(app core.App) {
	_, err := os.Stat("assets/css/output.css")
	if err == nil {
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		app.Logger().Warn("missing compiled css file", "path", "assets/css/output.css", "hint", "run: task css:build")
		return
	}
	app.Logger().Warn("unable to stat compiled css file", "path", "assets/css/output.css", "error", err)
}

const authTokenCleanupInterval = 15 * time.Minute

func runAuthTokenCleanupScheduler(app core.App) {
	stop := make(chan struct{})
	var closeOnce sync.Once
	app.OnTerminate().BindFunc(func(te *core.TerminateEvent) error {
		closeOnce.Do(func() { close(stop) })
		return te.Next()
	})

	go func() {
		cleanup := func() {
			if _, err := auth.CleanupExpiredOrConsumedTokens(context.Background(), app, time.Now()); err != nil {
				app.Logger().Error("auth token cleanup failed", "error", err)
			}
		}

		cleanup()

		ticker := time.NewTicker(authTokenCleanupInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cleanup()
			case <-stop:
				return
			}
		}
	}()
}
