package main

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

func main() {
	r := gin.Default()

	// 设置HTML模板
	tmpl := template.Must(template.New("").ParseFS(templateFS, "templates/*.html"))
	r.SetHTMLTemplate(tmpl)

	// 静态文件服务 - 需要创建子文件系统
	staticSub, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticSub))

	// 路由设置
	r.GET("/", indexHandler)
	r.POST("/dig", digHandler)
	r.POST("/curl", curlHandler)

	r.Run(":7788")
}
