# LingDig - DNS查询工具

一个基于Go语言开发的现代化DNS查询工具，提供Web界面进行DNS记录查询。

## 功能特性

- 支持多种DNS记录类型查询 (A, AAAA, CNAME, MX, NS, TXT, SOA, PTR, SRV)
- 支持多个DNS服务器 (Google, Cloudflare, 114DNS, 阿里DNS等)
- 现代化Web界面，响应式设计
- 基于Go语言，性能优异
- 使用embed技术，单文件部署

## 技术栈

- **后端**: Go + Gin框架
- **DNS库**: github.com/miekg/dns
- **前端**: HTML5 + CSS3 + JavaScript
- **部署**: Go embed技术

## 快速开始

### 1. 克隆项目
```bash
git clone <repository-url>
cd LingDig
```

### 2. 安装依赖
```bash
go mod tidy
```

### 3. 构建项目
```bash
go build -o lingdig
```

### 4. 运行应用
```bash
./lingdig
```

### 5. 访问应用
打开浏览器访问: http://localhost:8080

## 使用说明

1. 在域名输入框中输入要查询的域名
2. 选择DNS记录类型 (默认为A记录)
3. 选择DNS服务器 (默认为Google DNS)
4. 点击查询按钮获取结果

## 支持的DNS记录类型

- **A**: IPv4地址记录
- **AAAA**: IPv6地址记录  
- **CNAME**: 别名记录
- **MX**: 邮件交换记录
- **NS**: 域名服务器记录
- **TXT**: 文本记录
- **SOA**: 授权开始记录
- **PTR**: 反向DNS记录
- **SRV**: 服务记录

## 支持的DNS服务器

- Google DNS (8.8.8.8)
- Cloudflare DNS (1.1.1.1)
- 114 DNS (114.114.114.114)
- 阿里 DNS (223.5.5.5)
- 自定义DNS服务器

## API接口

### POST /dig
查询DNS记录

**请求参数:**
```json
{
  "domain": "google.com",
  "record_type": "A",
  "server": "8.8.8.8:53"
}
```

**响应示例:**
```json
{
  "domain": "google.com",
  "record_type": "A",
  "server": "8.8.8.8:53",
  "results": [
    {
      "name": "google.com.",
      "type": "A",
      "ttl": 300,
      "value": "142.250.191.14"
    }
  ]
}
```

## 项目结构

```
LingDig/
├── main.go              # 主程序入口
├── handlers.go          # HTTP处理器
├── dig_service.go       # DNS查询服务
├── templates/           # HTML模板
│   └── index.html
├── static/              # 静态资源
│   ├── css/
│   │   └── style.css
│   └── js/
│       └── app.js
├── go.mod              # Go模块文件
└── README.md           # 项目说明
```

## 开发

### 本地开发
```bash
go run .
```

### 构建发布版本
```bash
go build -ldflags="-s -w" -o lingdig
```

## 许可证

MIT License