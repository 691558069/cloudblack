# CloudBlack 云黑系统

基于 Go + SQLite 的轻量级云黑名单查询与管理系统，单文件部署，零依赖。

## 功能特性

- **查询系统** — 按 QQ 号查询云黑记录，支持多平台
- **提交系统** — 公开提交云黑记录，支持严重程度、标签、关联账号
- **审核流程** — 所有提交需管理员审核后才生效
- **API 接口** — RESTful API，支持批量查询、快速检查
- **公开访问** — 查询/提交接口无需 API 密钥，受 RPM 限制保护
- **关联账号** — 支持同一主体下多个平台账号关联展示
- **风控系统** — IP 冷却期、原因字数检查、全局提交频率限制
- **管理后台** — 完整的 Web 管理面板，含仪表盘、日志、API 密钥管理
- **系统监控** — 实时 CPU、内存、数据库大小监控

## 快速开始

### 下载

从 [GitHub Releases](../../releases) 下载对应平台的压缩包，解压后运行。

### 运行

**Linux / macOS:**
```bash
chmod +x cloudblack
./cloudblack
```

**Windows:**
双击 `start-windows.bat` 或直接运行 `cloudblack.exe`

启动后访问 `http://127.0.0.1:8080`

### 默认账号

- 后台地址：`/admin`
- 默认用户名：`admin`
- 默认密码：`123456`
- **首次登录后请立即修改密码**

## 配置

编辑 `config.json`（首次运行自动创建）：

```json
{
  "port": "8080",
  "db": {
    "type": "sqlite",
    "path": "data/cloudblack.db"
  },
  "security": {
    "trust_cloudflare": true,
    "secure_cookie": false
  },
  "rate_limit": {
    "api": 30,
    "web": 5,
    "admin": 10,
    "window": 60
  }
}
```

## API 接口

| 接口 | 方法 | 说明 | 需要密钥 |
|------|------|------|----------|
| `/api/v1/query?qq=123456` | GET | 查询单条记录 | 否 |
| `/api/v1/check?qq=123456` | GET | 快速检查（仅返回是否在黑名单） | 否 |
| `/api/v1/batch?qq_list=123,456` | GET | 批量查询（最多100个） | 否 |
| `/api/v1/submit` | POST | 提交云黑记录 | 否 |
| `/api/v1/review/list` | GET | 待审核列表 | 是（review） |
| `/api/v1/review/action` | POST | 审核通过/拒绝 | 是（review） |

详细文档请访问 `/web/api`

### API 密钥传递方式

- HTTP Header: `X-API-Key: your_key`
- URL 参数: `?api_key=your_key`
- 表单字段: `api_key=your_key`

## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Echo v4
- **数据库**: SQLite（pure-Go 驱动，无需 CGO）
- **密码加密**: bcrypt
- **系统监控**: gopsutil

## 构建

```bash
# 本地构建
go build -ldflags="-s -w" -o cloudblack .

# 交叉编译（Linux amd64）
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o cloudblack .

# 交叉编译（Windows amd64）
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o cloudblack.exe .
```

推送 `v*` 标签将自动触发 GitHub Actions 构建并发布 Release。

```bash
git tag v1.0.0
git push origin v1.0.0
```

## 目录结构

```
├── cloudblack          # 编译后二进制文件
├── config.json         # 配置文件
├── start-windows.bat   # Windows 启动脚本
├── data/
│   └── cloudblack.db   # SQLite 数据库
├── main.go             # 入口、路由注册
├── config.go           # 配置、数据库初始化、工具函数
├── api.go              # API 接口
├── admin.go            # 管理后台
├── web.go              # 前端页面
├── ratelimit.go        # 限流器
└── .github/workflows/
    └── build.yml       # CI/CD 构建配置
```

## 许可证

MIT License
