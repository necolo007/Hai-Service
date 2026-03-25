package main

import (
	"Hai-Service/config"
	"Hai-Service/core/gin"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/viper"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	handler := gin.GinInit()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", config.GetConfig().App.Port),
		Handler: handler,
	}
	//启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			color.Red("Server Error: %s", err.Error())
		}
	}()
	color.Green("Server Run At: http://localhost:%s", viper.GetString("app.Port"))
	//等待系统信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	color.Blue("Shutdown Server ...")
	//关闭服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		color.Red("Server Shutdown Error: %s", err.Error())
	} else {
		color.Green("Server Shutdown Gracefully")
	}
}
