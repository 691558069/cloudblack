package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	port := flag.String("port", "", "Server port, overrides config.json")
	config := flag.String("config", "config.json", "Config file path")
	dataDir := flag.String("data", "data", "Data directory")
	flag.Parse()

	if runtime.GOOS == "windows" {
		if exePath, err := os.Executable(); err == nil {
			_ = os.Chdir(filepath.Dir(exePath))
		}
	}

	if _, err := os.Stat(*dataDir); os.IsNotExist(err) {
		os.MkdirAll(*dataDir, 0755)
	}

	cfg := LoadConfig(*config)
	AppConfig = cfg
	cfg.DataDir = *dataDir
	if *port != "" {
		cfg.Port = *port
	}

	if err := InitDB(cfg); err != nil {
		log.Fatalf("Database init failed: %v", err)
	}

	InitStatusSampler()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			origin := c.Request().Header.Get("Origin")
			if origin != "" {
				host := c.Request().Host
				if !strings.Contains(origin, host) {
					return c.String(http.StatusForbidden, "Forbidden")
				}
			}
			return next(c)
		}
	})
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.Redirect(302, "/web/")
	})

	RegisterAPIV1Routes(e.Group(""), cfg)
	RegisterAdminRoutes(e, cfg)
	RegisterWebRoutes(e.Group(""), cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Server starting on %s", addr)
	if err := e.Start(addr); err != nil {
		log.Printf("Server stopped: %v", err)
		if runtime.GOOS == "windows" {
			fmt.Println()
			fmt.Println("程序启动失败，常见原因：")
			fmt.Println("1. 端口已被占用，例如默认端口 8080 已有程序在使用")
			fmt.Println("2. 当前目录没有写入 data 数据目录的权限")
			fmt.Println("3. 防火墙或安全软件拦截了程序监听端口")
			fmt.Println()
			fmt.Println("如果是端口占用，可以用以下方式换端口启动：")
			fmt.Println("cloudblack.exe -port 8081")
			fmt.Println()
			fmt.Print("按回车键退出...")
			_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
		}
		os.Exit(1)
	}
}
