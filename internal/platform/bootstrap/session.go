package bootstrap

import (
	"errors"
	"log/slog"

	"github.com/go-rio/rio"
	"github.com/libtnb/sessions"

	"github.com/libtnb/chi-skeleton/internal/platform/conf"
)

func NewSession(config *conf.Config, log *slog.Logger, db *rio.DB) (*sessions.Manager, func() error, error) {
	manager, err := sessions.NewManager(&sessions.ManagerOptions{
		Key:                  config.App.Key,
		Lifetime:             config.Session.Lifetime,
		GcInterval:           config.Session.GcInterval,
		DisableDefaultDriver: true,
		// background errors (GC, middleware saves) land in the app log
		Logger: log,
	})
	if err != nil {
		return nil, nil, err
	}

	store, err := newSessionStore(db)
	if err != nil {
		return nil, nil, errors.Join(err, manager.Close())
	}
	if err = manager.Extend("default", store); err != nil {
		return nil, nil, errors.Join(err, store.Close(), manager.Close())
	}

	return manager, manager.Close, nil
}
