module github.com/webdevelop-pro/migration-service

go 1.20

replace github.com/webdevelop-pro/lib => ./pkg/lib

replace github.com/webdevelop-pro/go-logger => ./pkg/logger

require (
	github.com/jackc/pgconn v1.12.1
	github.com/jackc/pgx/v4 v4.16.1
	github.com/labstack/echo/v4 v4.10.2
	github.com/pkg/errors v0.9.1
	github.com/rs/zerolog v1.27.0
	github.com/webdevelop-pro/go-common v0.0.0-20220929214313-b0e8b0d7638e
	github.com/webdevelop-pro/lib v0.0.0-00010101000000-000000000000
	go.uber.org/fx v1.17.1
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/golang-jwt/jwt v3.2.2+incompatible // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.11.0 // indirect
	github.com/jackc/puddle v1.2.1 // indirect
	github.com/jinzhu/copier v0.3.5 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/labstack/gommon v0.4.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasttemplate v1.2.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/dig v1.14.1 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	go.uber.org/zap v1.21.0 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/net v0.7.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.3.0 // indirect
)
