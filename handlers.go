package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type DigRequest struct {
	Domain     string `json:"domain" form:"domain"`
	RecordType string `json:"record_type" form:"record_type"`
	Server     string `json:"server" form:"server"`
}

type DigResponse struct {
	Domain     string      `json:"domain"`
	RecordType string      `json:"record_type"`
	Server     string      `json:"server"`
	Results    []DNSRecord `json:"results"`
	Error      string      `json:"error,omitempty"`
}

func indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "LingDig - 网络工具集",
	})
}

func digHandler(c *gin.Context) {
	var req DigRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, DigResponse{
			Error: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 验证输入
	if strings.TrimSpace(req.Domain) == "" {
		c.JSON(http.StatusBadRequest, DigResponse{
			Error: "域名不能为空",
		})
		return
	}

	// 设置默认值
	if req.RecordType == "" {
		req.RecordType = "A"
	}
	if req.Server == "" {
		req.Server = "8.8.8.8:53"
	}

	// 执行DNS查询
	digService := NewDigService()
	results, err := digService.Query(req.Domain, req.RecordType, req.Server)

	response := DigResponse{
		Domain:     req.Domain,
		RecordType: req.RecordType,
		Server:     req.Server,
		Results:    results,
	}

	if err != nil {
		response.Error = err.Error()
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	c.JSON(http.StatusOK, response)
}

func curlHandler(c *gin.Context) {
	var req CurlRequest

	// 检查是否是curl命令解析
	curlCommand := c.PostForm("curl_command")
	if curlCommand != "" {
		parsedReq, err := NewCurlService().ParseCurlCommand(curlCommand)
		if err != nil {
			c.JSON(http.StatusBadRequest, CurlResponse{
				Error: "解析curl命令失败: " + err.Error(),
			})
			return
		}
		req = *parsedReq
	} else {
		// 普通表单绑定
		req.URL = c.PostForm("url")
		req.Method = c.PostForm("method")
		req.Body = c.PostForm("body")

		// 解析headers JSON
		headersStr := c.PostForm("headers_json")
		if headersStr != "" {
			var headers map[string]string
			if err := json.Unmarshal([]byte(headersStr), &headers); err == nil {
				req.Headers = headers
			} else {
				req.Headers = make(map[string]string)
			}
		} else {
			req.Headers = make(map[string]string)
		}

		// 解析布尔值 - 处理HTML checkbox的"on"值
		followRedirect := c.PostForm("follow_redirect")
		req.FollowRedirect = followRedirect == "on" || followRedirect == "true" || followRedirect == "1"

		verifySSL := c.PostForm("verify_ssl")
		req.VerifySSL = verifySSL == "on" || verifySSL == "true" || verifySSL == "1"

		headOnly := c.PostForm("head_only")
		req.HeadOnly = headOnly == "on" || headOnly == "true" || headOnly == "1"

		// 解析超时时间
		if timeout := c.PostForm("timeout"); timeout != "" {
			if t, err := strconv.Atoi(timeout); err == nil {
				req.Timeout = t
			} else {
				req.Timeout = 30 // 默认值
			}
		} else {
			req.Timeout = 30
		}

		// 设置默认值
		if req.Method == "" {
			req.Method = "GET"
		}
	}

	// 验证输入
	if strings.TrimSpace(req.URL) == "" {
		c.JSON(http.StatusBadRequest, CurlResponse{
			Error: "URL不能为空",
		})
		return
	}

	// 执行HTTP请求
	curlService := NewCurlService()
	response, err := curlService.Execute(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, CurlResponse{
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
