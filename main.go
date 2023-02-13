package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/cupogo/andvari/utils/zlog"
	"github.com/cupogo/scaffold/pkg/settings"
	"github.com/cupogo/scaffold/pkg/web"
)

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
	idleClosed := make(chan struct{})
	ctx := context.Background()
	go func() {
		quit := make(chan os.Signal, 2)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		sugar.Info("shuting down server...")
		if err := srv.Stop(ctx); err != nil {
			sugar.Warnw("server shutdown:", "err", err)
		}
		close(idleClosed)
	}()

	if err := srv.Serve(ctx); err != nil {
		sugar.Warnw("serve fali", "err", err)
	}

	<-idleClosed
}
