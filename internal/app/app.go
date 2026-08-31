package app

import (
	"context"
	_ "expvar" // registers /debug/vars on the default mux
	"fmt"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // private debug listener only
	"time"

	"github.com/go-rio/migrate"
	"github.com/libtnb/cron"
	"github.com/libtnb/graceful"

	"github.com/libtnb/chi-skeleton/internal/platform/conf"
	"github.com/libtnb/chi-skeleton/internal/shared/registry"
)

type App struct {
	conf     *conf.Config
	server   *http.Server
	migrator *migrate.Migrator
	cron     *cron.Cron
}

func NewApp(
	config *conf.Config,
	server *http.Server,
	migrator *migrate.Migrator,
	scheduler *cron.Cron,
	_ registry.Subscriptions,
) *App {
	return &App{
		conf:     config,
		server:   server,
		migrator: migrator,
		cron:     scheduler,
	}
}

// Run migrates the database, then serves until SIGINT/SIGTERM; SIGHUP hot-upgrades.
func (r *App) Run() error {
	if err := r.migrator.Up(context.Background()); err != nil {
		return err
	}
	fmt.Println("[DB] database migrated")

	g := graceful.New(
		graceful.WithUpgrade(),
		graceful.WithShutdownTimeout(30*time.Second),
	)
	// pprof/expvar live on http.DefaultServeMux, served on a private port
	if addr := r.conf.HTTP.DebugAddress; addr != "" {
		g.Listen("debug", addr, &http.Server{ReadHeaderTimeout: 10 * time.Second})
	}
	g.Add("cron", r.cron.Start, r.cron.Stop)
	g.Listen("http", r.conf.HTTP.Address, r.server)

	fmt.Println("[HTTP] listening and serving on", r.conf.HTTP.Address)
	return g.Run()
}
