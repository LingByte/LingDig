package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

type CurlRequest struct {
	URL            string            `json:"url" form:"url"`
	Method         string            `json:"method" form:"method"`
	Headers        map[string]string `json:"headers" form:"headers"`
	Body           string            `json:"body" form:"body"`
	Timeout        int               `json:"timeout" form:"timeout"`
	FollowRedirect bool              `json:"follow_redirect" form:"follow_redirect"`
	VerifySSL      bool              `json:"verify_ssl" form:"verify_ssl"`
	HeadOnly       bool              `json:"head_only" form:"head_only"`
}

type CurlResponse struct {
	URL           string            `json:"url"`
	Method        string            `json:"method"`
	StatusCode    int               `json:"status_code"`
	StatusText    string            `json:"status_text"`
	Headers       map[string]string `json:"headers"`
	Body          string            `json:"body"`
	BodyPreview   string            `json:"body_preview"`
	BodySize      int64             `json:"body_size"`
	IsBinary      bool              `json:"is_binary"`
	ResponseTime  int64             `json:"response_time_ms"`
	ContentLength int64             `json:"content_length"`
	ContentType   string            `json:"content_type"`
	Error         string            `json:"error,omitempty"`
	RedirectChain []string          `json:"redirect_chain,omitempty"`
	RequestInfo   RequestInfo       `json:"request_info"`
}

type RequestInfo struct {
	FinalURL       string            `json:"final_url"`
	RemoteAddr     string            `json:"remote_addr"`
	Protocol       string            `json:"protocol"`
	TLSVersion     string            `json:"tls_version,omitempty"`
	RequestHeaders map[string]string `json:"request_headers"`
}

type CurlService struct {
	client *http.Client
}

func NewCurlService() *CurlService {
	return &CurlService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *CurlService) Execute(req CurlRequest) (*CurlResponse, error) {
	startTime := time.Now()

	// 验证URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		return nil, fmt.Errorf("无效的URL: %v", err)
	}

	// 如果没有协议，默认添加https
	if parsedURL.Scheme == "" {
		req.URL = "https://" + req.URL
		parsedURL, err = url.Parse(req.URL)
		if err != nil {
			return nil, fmt.Errorf("无效的URL: %v", err)
		}
	}

	// 设置默认值
	if req.Method == "" {
		req.Method = "GET"
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}

	// 如果是HEAD请求，强制设置HeadOnly
	if req.Method == "HEAD" {
		req.HeadOnly = true
	}

	// 创建HTTP客户端
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !req.VerifySSL,
		},
	}

	client := &http.Client{
		Timeout:   time.Duration(req.Timeout) * time.Second,
		Transport: transport,
	}

	// 处理重定向
	var redirectChain []string
	if !req.FollowRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		client.CheckRedirect = func(newReq *http.Request, via []*http.Request) error {
			redirectChain = append(redirectChain, newReq.URL.String())
			if len(via) >= 10 {
				return fmt.Errorf("重定向次数过多")
			}
			return nil
		}
	}

	// 创建请求
	var bodyReader io.Reader
	if req.Body != "" && !req.HeadOnly {
		bodyReader = strings.NewReader(req.Body)
	}

	httpReq, err := http.NewRequest(req.Method, req.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %v", err)
	}

	// 设置请求头
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// 设置默认User-Agent
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "LingDig/1.0 (HTTP Client)")
	}

	// 执行请求
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(req.Timeout)*time.Second)
	defer cancel()

	httpReq = httpReq.WithContext(ctx)
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	// 计算响应时间
	responseTime := time.Since(startTime).Milliseconds()

	// 构建响应头映射
	headers := make(map[string]string)
	for key, values := range resp.Header {
		headers[key] = strings.Join(values, ", ")
	}

	// 构建请求信息
	requestHeaders := make(map[string]string)
	for key, values := range httpReq.Header {
		requestHeaders[key] = strings.Join(values, ", ")
	}

	requestInfo := RequestInfo{
		FinalURL:       resp.Request.URL.String(),
		Protocol:       resp.Proto,
		RequestHeaders: requestHeaders,
	}

	// 获取远程地址（如果可用）
	if resp.Request.RemoteAddr != "" {
		requestInfo.RemoteAddr = resp.Request.RemoteAddr
	}

	// 获取TLS信息
	if resp.TLS != nil {
		switch resp.TLS.Version {
		case 0x0301:
			requestInfo.TLSVersion = "TLS 1.0"
		case 0x0302:
			requestInfo.TLSVersion = "TLS 1.1"
		case 0x0303:
			requestInfo.TLSVersion = "TLS 1.2"
		case 0x0304:
			requestInfo.TLSVersion = "TLS 1.3"
		default:
			requestInfo.TLSVersion = fmt.Sprintf("TLS 0x%04x", resp.TLS.Version)
		}
	}

	// 读取响应体
	var bodyBytes []byte
	var bodyStr, bodyPreview string
	var isBinary bool
	var bodySize int64

	if !req.HeadOnly && req.Method != "HEAD" {
		bodyBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %v", err)
		}

		bodySize = int64(len(bodyBytes))

		// 检查是否为二进制内容
		isBinary = !utf8.Valid(bodyBytes) || c.isBinaryContent(resp.Header.Get("Content-Type"))

		if isBinary {
			bodyStr = fmt.Sprintf("[二进制数据 - %d 字节]", len(bodyBytes))
			bodyPreview = c.generateBinaryPreview(bodyBytes, resp.Header.Get("Content-Type"))
		} else {
			bodyStr = string(bodyBytes)
			if len(bodyStr) > 10000 {
				bodyPreview = bodyStr[:10000] + "\n\n... [内容过长，已截断]"
			} else {
				bodyPreview = bodyStr
			}
		}
	} else {
		bodyStr = "[HEAD请求 - 仅获取响应头]"
		bodyPreview = bodyStr
		bodySize = resp.ContentLength
	}

	// 构建响应
	curlResp := &CurlResponse{
		URL:           req.URL,
		Method:        req.Method,
		StatusCode:    resp.StatusCode,
		StatusText:    resp.Status,
		Headers:       headers,
		Body:          bodyStr,
		BodyPreview:   bodyPreview,
		BodySize:      bodySize,
		IsBinary:      isBinary,
		ResponseTime:  responseTime,
		ContentLength: resp.ContentLength,
		ContentType:   resp.Header.Get("Content-Type"),
		RedirectChain: redirectChain,
		RequestInfo:   requestInfo,
	}

	return curlResp, nil
}

func (c *CurlService) isBinaryContent(contentType string) bool {
	binaryTypes := []string{
		"image/", "audio/", "video/", "application/octet-stream",
		"application/pdf", "application/zip", "application/gzip",
		"application/x-", "font/", "model/",
	}

	contentType = strings.ToLower(contentType)
	for _, binaryType := range binaryTypes {
		if strings.HasPrefix(contentType, binaryType) {
			return true
		}
	}
	return false
}

func (c *CurlService) generateBinaryPreview(data []byte, contentType string) string {
	preview := fmt.Sprintf("Content-Type: %s\n", contentType)
	preview += fmt.Sprintf("Size: %d bytes\n\n", len(data))

	if strings.HasPrefix(contentType, "image/") {
		preview += "📷 图片文件\n"
	} else if strings.HasPrefix(contentType, "audio/") {
		preview += "🎵 音频文件\n"
	} else if strings.HasPrefix(contentType, "video/") {
		preview += "🎬 视频文件\n"
	} else if strings.HasPrefix(contentType, "application/pdf") {
		preview += "📄 PDF文档\n"
	} else if strings.HasPrefix(contentType, "application/zip") {
		preview += "📦 压缩文件\n"
	} else {
		preview += "📁 二进制文件\n"
	}

	// 显示前64字节的十六进制
	preview += "\n十六进制预览 (前64字节):\n"
	maxBytes := 64
	if len(data) < maxBytes {
		maxBytes = len(data)
	}

	for i := 0; i < maxBytes; i += 16 {
		end := i + 16
		if end > maxBytes {
			end = maxBytes
		}

		// 十六进制
		hexPart := ""
		for j := i; j < end; j++ {
			hexPart += fmt.Sprintf("%02x ", data[j])
		}

		// ASCII部分
		asciiPart := ""
		for j := i; j < end; j++ {
			if data[j] >= 32 && data[j] <= 126 {
				asciiPart += string(data[j])
			} else {
				asciiPart += "."
			}
		}

		preview += fmt.Sprintf("%04x: %-48s |%s|\n", i, hexPart, asciiPart)
	}

	if len(data) > maxBytes {
		preview += "...\n"
	}

	return preview
}

func (c *CurlService) ParseCurlCommand(curlCmd string) (*CurlRequest, error) {
	// 简单的curl命令解析
	req := &CurlRequest{
		Method:         "GET",
		Headers:        make(map[string]string),
		FollowRedirect: true,
		VerifySSL:      true,
		Timeout:        30,
	}

	// 移除curl前缀
	curlCmd = strings.TrimSpace(curlCmd)
	if strings.HasPrefix(curlCmd, "curl ") {
		curlCmd = curlCmd[5:]
	}

	// 简单解析（这里可以扩展为更复杂的解析器）
	parts := strings.Fields(curlCmd)

	for i, part := range parts {
		switch {
		case part == "-I" || part == "--head":
			req.Method = "HEAD"
			req.HeadOnly = true
		case part == "-X" || part == "--request":
			if i+1 < len(parts) {
				req.Method = strings.ToUpper(parts[i+1])
			}
		case part == "-H" || part == "--header":
			if i+1 < len(parts) {
				header := parts[i+1]
				if colonIndex := strings.Index(header, ":"); colonIndex > 0 {
					key := strings.TrimSpace(header[:colonIndex])
					value := strings.TrimSpace(header[colonIndex+1:])
					req.Headers[key] = value
				}
			}
		case part == "-d" || part == "--data":
			if i+1 < len(parts) {
				req.Body = parts[i+1]
				if req.Method == "GET" {
					req.Method = "POST"
				}
			}
		case part == "-k" || part == "--insecure":
			req.VerifySSL = false
		case part == "-L" || part == "--location":
			req.FollowRedirect = true
		case strings.HasPrefix(part, "http://") || strings.HasPrefix(part, "https://"):
			req.URL = part
		case !strings.HasPrefix(part, "-") && req.URL == "":
			// 可能是URL
			req.URL = part
		}
	}

	if req.URL == "" {
		return nil, fmt.Errorf("未找到URL")
	}

	return req, nil
}
