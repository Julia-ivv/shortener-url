// Package main application entry point.
package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"google.golang.org/grpc"

	"github.com/Julia-ivv/shortener-url.git/cmd/certgenerator"
	"github.com/Julia-ivv/shortener-url.git/internal/config"
	"github.com/Julia-ivv/shortener-url.git/internal/grpcserver"
	"github.com/Julia-ivv/shortener-url.git/internal/httpserver"
	"github.com/Julia-ivv/shortener-url.git/internal/interceptors"
	pb "github.com/Julia-ivv/shortener-url.git/internal/proto"
	"github.com/Julia-ivv/shortener-url.git/internal/storage"
	"github.com/Julia-ivv/shortener-url.git/pkg/logger"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Println("Build version:", buildVersion)
	fmt.Println("Build date:", buildDate)
	fmt.Println("Build commit:", buildCommit)

	cfg := config.NewConfig()

	logger.ZapSugar = logger.NewLogger()
	logger.ZapSugar.Infow("Starting http server", "addr", cfg.Host)
	logger.ZapSugar.Infow("Starting gRPC server", "addr", cfg.GRPC)
	logger.ZapSugar.Infow("flags",
		"base url", cfg.URL,
		"filename", cfg.FileName,
		"db dsn", cfg.DBDSN,
		"https enabled", cfg.EnableHTTPS,
		"config file", cfg.ConfigFileName,
	)

	repo, err := storage.NewURLs(*cfg)
	if err != nil {
		logger.ZapSugar.Fatal(err)
	}
	defer repo.Close()

	httpWg := sync.WaitGroup{}
	grpcWg := sync.WaitGroup{}

	var srv = http.Server{
		Addr:    cfg.Host,
		Handler: httpserver.NewURLRouter(repo, *cfg, &httpWg),
	}

	srvGRPC := grpc.NewServer(
		grpc.ChainUnaryInterceptor(interceptors.HandlerWithAuth),
		grpc.ChainUnaryInterceptor(interceptors.HandlerWithLogging))
	pb.RegisterShortUrlServer(srvGRPC, grpcserver.NewShortenerServer(repo, *cfg, &grpcWg))

	idleConnsClosed := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-sigs
		if err := srv.Shutdown(context.Background()); err != nil {
			logger.ZapSugar.Infow("HTTP server Shutdown: %v", err)
		}
		srvGRPC.GracefulStop()
		close(idleConnsClosed)
	}()

	go func() {
		if cfg.EnableHTTPS {
			certFile, privateKeyFile, err := certgenerator.GenCert(4096)
			if err != nil {
				logger.ZapSugar.Fatalw(err.Error(), "event", "create certificate or private key")
			}
			err = srv.ListenAndServeTLS(certFile.Name(), privateKeyFile.Name())
			if err != nil && err != http.ErrServerClosed {
				logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
			}
		} else {
			err = srv.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				logger.ZapSugar.Fatalw(err.Error(), "event", "start server")
			}
		}
	}()

	go func() {
		listen, err := net.Listen("tcp", cfg.GRPC)
		if err != nil {
			logger.ZapSugar.Fatalw(err.Error(), "event", "listen port")
		}
		if err = srvGRPC.Serve(listen); err != nil {
			logger.ZapSugar.Fatalw(err.Error(), "event", "start gRPC server")
		}
	}()

	grpcWg.Wait()
	httpWg.Wait()
	<-idleConnsClosed
}
