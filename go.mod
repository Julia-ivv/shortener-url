module github.com/Julia-ivv/shortener-url.git

go 1.20

require (
	github.com/DATA-DOG/go-sqlmock v1.5.2
	github.com/caarlos0/env v3.5.0+incompatible
	github.com/go-chi/chi/v5 v5.0.10
	github.com/golang-jwt/jwt/v4 v4.5.0
	github.com/jackc/pgerrcode v0.0.0-20220416144525-469b46aa5efa
	github.com/jackc/pgx/v5 v5.5.0
	github.com/stretchr/testify v1.8.4
	go.uber.org/zap v1.27.0 // indirect
	google.golang.org/protobuf v1.32.0
)

require (
	github.com/Julia-ivv/shortener-url/pkg/compressing v1.0.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/rogpeppe/go-internal v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/sync v0.6.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/grpc v1.62.1
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require (
	github.com/Julia-ivv/shortener-url/pkg/logger v1.0.0
	github.com/Julia-ivv/shortener-url/pkg/middleware v1.0.0
)

replace (
	github.com/Julia-ivv/shortener-url/pkg/compressing => ./pkg/compressing
	github.com/Julia-ivv/shortener-url/pkg/logger => ./pkg/logger
	github.com/Julia-ivv/shortener-url/pkg/middleware => ./pkg/middleware
)
