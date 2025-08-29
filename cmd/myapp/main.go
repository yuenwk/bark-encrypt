package main

import (
	barkEncrypt "bark-encrypt/internal/bark-encrypt"
	"bark-encrypt/internal/config"
	"log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("could not load config: %v", err)
	}

	// 初始化 Echo 实例
	e := echo.New()
	e.Use(middleware.Logger())  // 使用内置的日志中间件
	e.Use(middleware.Recover()) // 使用 Panic 恢复中间件

	e.GET("/", func(c echo.Context) error {
		return c.String(200, "hello world")
	})

	bark := barkEncrypt.NewPushHandler(cfg.Bark)
	g := e.Group("/bark")
	g.POST("/push-ciphertext", bark.EncryptAndPush)

	// 启动服务器
	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
