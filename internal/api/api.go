package api

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// API is a REST API.
type API struct {
	log        zerolog.Logger
	migrations MigrationService
}

type MigrationService interface {
	ServiceVersion(name string) (int, error)
	ServiceExists(name string) bool
	Apply(name string, priority, minVersion int, isForced, noAutoOnly bool) (int, int, error)
	BumpServiceVersion(name string, ver int) error
}

// NewAPI returns new API insance.
func NewAPI(log zerolog.Logger, migrations MigrationService) *API {
	a := &API{
		log:        log,
		migrations: migrations,
	}
	return a
}

// ApplyMigrations applies 'noAuto' migrations to the specified service
func (a *API) HandleApplyMigrations(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")
	if serviceName == "" || !a.migrations.ServiceExists(serviceName) {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}
	ver, err := a.migrations.ServiceVersion(serviceName)
	if err != nil {
		http.Error(w, errors.Wrap(err, "failed to get current service version").Error(), http.StatusInternalServerError)
		return
	}
	n, lastVersion, err := a.migrations.Apply(serviceName, -1, ver, false, true)
	if err != nil {
		http.Error(w, errors.Wrap(err, "failed to apply migrations").Error(), http.StatusInternalServerError)
		return
	}
	if lastVersion > ver {
		if err := a.migrations.BumpServiceVersion(serviceName, lastVersion); err != nil {
			a.log.Error().Err(err).Msg("failed to bump service version")
		}
	}
	err = json.NewEncoder(w).Encode(struct {
		MigrationsApplied int
	}{
		MigrationsApplied: n,
	})
	if err != nil {
		http.Error(w, errors.Wrap(err, "failed to encode response").Error(), http.StatusInternalServerError)
		return
	}
	a.log.Info().Int("n", n).Str("service", serviceName).Msg("applied migrations")
}
