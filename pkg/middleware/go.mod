module github.com/Julia-ivv/shortener-url/pkg/middleware

go 1.20

replace (
	github.com/Julia-ivv/shortener-url/pkg/compressing => ../compressing
	github.com/Julia-ivv/shortener-url/pkg/logger => ../logger
)

require (
	github.com/Julia-ivv/shortener-url/pkg/compressing v0.0.0-00010101000000-000000000000
	github.com/Julia-ivv/shortener-url/pkg/logger v0.0.0-00010101000000-000000000000
)

require (
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
)
