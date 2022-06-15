package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/webdevelop-pro/go-common/logger"
	"github.com/webdevelop-pro/migration-service/internal/services"
)

type Handler struct {
	migration services.Migration
	log       logger.Logger
}

func NewHandler(migration services.Migration) *Handler {
	return &Handler{migration: migration, log: logger.NewDefaultComponent("api_handler")}
}

func (h *Handler) ApplyMigration(c echo.Context) error {
	var (
		ctx         = c.Request().Context()
		serviceName = c.QueryParam("service")
	)

	n, err := h.migration.Apply(ctx, serviceName)
	if err != nil {
		// TODO add switcher between type of errors
		return c.String(http.StatusInternalServerError, err.Error())
	}

	h.log.Info().Int("n", n).Str("service", serviceName).Msg("applied migrations")

	return c.JSON(
		http.StatusOK,
		struct {
			MigrationsApplied int
		}{
			MigrationsApplied: n,
		},
	)
}
