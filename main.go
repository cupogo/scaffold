package main

import (
	"context"
	"embed"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/cupogo/andvari/utils/zlog"
	"github.com/cupogo/scaffold/pkg/settings"
	"github.com/cupogo/scaffold/pkg/web"

	_ "github.com/cupogo/scaffold/pkg/web/api_z1"
)

//go:embed htdocs all:htdocs/app
var static embed.FS

func main() {

	var zlogger *zap.Logger
	if settings.InDevelop() {
		zlogger, _ = zap.NewDevelopment()
	} else {
		zlogger, _ = zap.NewProduction()
	}
	sugar := zlogger.Sugar()
	zlog.Set(sugar)

	srv := web.New()
	fsys := fs.FS(static)
	html, _ := fs.Sub(fsys, "htdocs")
	// srv.StaticFS("/", http.FS(html))
	srv.NotFound(http.FileServer(http.FS(html)))

	idleClosed := make(chan struct{})
	ctx := context.Background()
	go func() {
		quit := make(chan os.Signal, 2)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		sugar.Info("shuting down server...")
		if err := srv.Stop(ctx); err != nil {
			sugar.Infow("server shutdown:", "err", err)
		}
		close(idleClosed)
	}()

	if err := srv.Serve(ctx); err != nil {
		sugar.Infow("serve fali", "err", err)
	}

	<-idleClosed
}
