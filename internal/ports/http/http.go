package http

import (
	"net/http"

	"github.com/webdevelop-pro/go-common/server"
	"github.com/webdevelop-pro/migration-service/internal/services"
)

func InitHandlers(srv *server.HttpServer, migration services.Migration) {
	handler := NewHandler(migration)

	srv.AddRoute(server.Route{
		Method: http.MethodPost,
		Path:   "/apply",
		Handle: handler.ApplyMigration,
		NoAuth: true,
	})
}
