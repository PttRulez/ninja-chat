package serverdebug

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/pttrulez/ninja-chat/internal/buildinfo"
	"github.com/pttrulez/ninja-chat/internal/logger"
)

const (
	readHeaderTimeout = time.Second
	shutdownTimeout   = 3 * time.Second
)

//go:generate options-gen -out-filename=server_options.gen.go -from-struct=Options
type Options struct {
	addr string `option:"mandatory" validate:"required,hostname_port"`
}

type Server struct {
	lg  *zap.Logger
	srv *http.Server
}

func New(opts Options) (*Server, error) {
	if err := opts.Validate(); err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	lg := zap.L().Named("server-debug")

	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	s := &Server{
		lg: lg,
		srv: &http.Server{
			Addr:              opts.addr,
			Handler:           e,
			ReadHeaderTimeout: readHeaderTimeout,
		},
	}
	index := newIndexPage()

	index.addPage("/version", "Get build information")
	index.addPage("/debug/pprof/", "Get std profiler")
	index.addPage("/debug/pprof/profile?seconds=30", "Take half-min profile")

	e.GET("/", index.handler)

	e.GET("/version", s.Version)

	e.GET("/log/level", echo.WrapHandler(http.HandlerFunc(logger.GetAtomicLevel().ServeHTTP)))
	e.PUT("/log/level", echo.WrapHandler(http.HandlerFunc(logger.GetAtomicLevel().ServeHTTP)))

	e.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
	e.GET("/debug/pprof/profile", echo.WrapHandler(http.HandlerFunc(pprof.Profile)))
	e.GET("/debug/pprof/trace", echo.WrapHandler(http.HandlerFunc(pprof.Trace)))
	e.GET("/debug/pprof/cmdline", echo.WrapHandler(http.HandlerFunc(pprof.Cmdline)))

	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		return s.srv.Shutdown(ctx) //nolint:contextcheck // graceful shutdown with new context
	})

	eg.Go(func() error {
		s.lg.Info("listen and serve", zap.String("addr", s.srv.Addr))

		if err := s.srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("listen and serve: %v", err)
		}
		return nil
	})

	return eg.Wait()
}

func (s *Server) Version(e echo.Context) error {
	return e.JSON(http.StatusOK, buildinfo.BuildInfo)
}
