package web

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sentry"
	"github.com/gin-gonic/gin"

	"github.com/cupogo/scaffold/pkg/settings"
	apis "github.com/cupogo/scaffold/pkg/web/routes"
)

type Service interface {
	Serve(ctx context.Context) error
	http.Handler
	Stop(ctx context.Context) error
	NotFound(h http.Handler)
}

type server struct {
	router *gin.Engine
	hs     *http.Server
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) NotFound(h http.Handler) {
	s.router.NoRoute(gin.WrapH(h))
}

// New return new web server
func New(ahs ...string) Service {

	var (
		router *gin.Engine
	)

	corsCfg := cors.DefaultConfig()
	if settings.AllowAllOrigins() {
		corsCfg.AllowAllOrigins = true
	} else {
		corsCfg.AllowOrigins = settings.Current.AllowOrigins
		corsCfg.AllowAllOrigins = false
		logger().Infow("cors", "allowedOrigins", corsCfg.AllowOrigins)
		corsCfg.AllowWildcard = true
		if len(settings.Current.AllowOrigins) > 0 && "file://" == settings.Current.AllowOrigins[0] {
			corsCfg.AllowFiles = true
		}
	}

	corsCfg.AllowCredentials = true
	corsCfg.AllowHeaders = append(corsCfg.AllowHeaders, "token")

	if settings.InDevelop() {
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
	}
	if err := router.SetTrustedProxies(settings.Current.TrustProxies); err != nil {
		logger().Fatalw("set trust proxy fail", "err", err)
	}

	if err := corsCfg.Validate(); err != nil {
		logger().Fatalw("cors config fail", "err", err)
	}
	router.Use(cors.New(corsCfg))

	if settings.Current.SentryDSN != "" {
		if err := raven.SetDSN(settings.Current.SentryDSN); err != nil {
			logger().Warnw("raven SetDSN fail", "err", err)
		}
		raven.SetTagsContext(map[string]string{"appver": settings.Current.Version})
		onlyCrashes := false
		router.Use(sentry.Recovery(raven.DefaultClient, onlyCrashes))
	}

	router.GET("/ping", handlePing)

	apis.Routers(router, ahs...)

	hs := &http.Server{
		Addr:    settings.Current.HTTPListen,
		Handler: router.Handler(),
	}

	return &server{router: router, hs: hs}
}

func (s *server) Serve(ctx context.Context) error {
	// Run HTTP server
	runErrChan := make(chan error)
	t := time.AfterFunc(time.Millisecond*200, func() {
		runErrChan <- s.hs.ListenAndServe()
	})

	defer t.Stop()
	logger().Infow("Listen on", "addr", s.hs.Addr)

	// Wait
	for {
		select {
		case runErr := <-runErrChan:
			if runErr != nil {
				logger().Errorw("run http server failed",
					"err", runErr,
				)
				return runErr
			}
		case <-ctx.Done():
			//TODO Graceful shutdown
			logger().Info("http server has been stopped")
			return ctx.Err()
		}
	}
}

func (s *server) Stop(ctx context.Context) error {
	if err := s.hs.Shutdown(ctx); err != nil {
		logger().Fatalw("Server Shutdown", "err", err)
		return err
	}
	return nil
}

func handlePing(c *gin.Context) {
	c.String(200, "Pong")
}

// nolint
func handle204(c *gin.Context) {
	c.Status(204)
}
