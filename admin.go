package main

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	RoleSuperAdmin = 1
	RoleReviewer   = 2
	RoleOperator   = 3
)

const adminHTML = `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>{{.Title}} - 云黑系统</title>
<style>
:root{--bg:#f1f3f5;--panel:#fff;--panel-soft:#f8fafc;--text:#111827;--muted:#6b7280;--border:#e5e7eb;--primary:#1f2937;--primary-hover:#111827;--red:#dc2626;--red-dark:#b91c1c;--green:#16a34a;--green-dark:#15803d;--orange:#f59e0b;--shadow:0 1px 4px rgba(15,23,42,.06);--sidebar-w:220px}
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;background:var(--bg);color:var(--text);line-height:1.5;min-height:100vh}
.layout{display:flex;min-height:100vh}
.sidebar{width:var(--sidebar-w);background:var(--primary);color:#fff;display:flex;flex-direction:column;position:fixed;top:0;left:0;bottom:0;z-index:20}
.sidebar-header{padding:20px 20px 18px;font-size:17px;font-weight:800;letter-spacing:.02em;border-bottom:1px solid rgba(255,255,255,.08)}
.sidebar-header span{display:inline-block;width:9px;height:9px;margin-right:10px;border-radius:50%;background:var(--red);box-shadow:0 0 0 4px rgba(220,38,38,.18);vertical-align:2px}
.sidebar-nav{flex:1;overflow-y:auto;padding:12px 0}
.sidebar-nav a{display:block;padding:9px 20px;text-decoration:none;color:rgba(255,255,255,.72);font-size:14px;font-weight:600;transition:background .15s,color .15s}
.sidebar-nav a:hover{background:rgba(255,255,255,.06);color:#fff}
.sidebar-nav a.active{background:rgba(255,255,255,.10);color:#fff;border-left:3px solid var(--red);padding-left:17px}
.sidebar-nav .section{font-size:11px;text-transform:uppercase;letter-spacing:.06em;padding:18px 20px 8px;color:rgba(255,255,255,.36);font-weight:700}
.main{margin-left:var(--sidebar-w);flex:1;min-width:0}
.topbar{position:sticky;top:0;z-index:10;background:rgba(255,255,255,.92);backdrop-filter:blur(12px);padding:10px 24px;display:flex;justify-content:flex-end;align-items:center;border-bottom:1px solid var(--border);gap:14px}
.topbar .user{color:var(--muted);font-size:13px;font-weight:600}
.topbar a{color:var(--text);text-decoration:none;font-weight:700;font-size:13px}
.topbar a:hover{color:var(--red)}
.content{padding:24px;max-width:1360px}
.card{background:var(--panel);border:1px solid var(--border);border-radius:14px;padding:22px;margin-bottom:20px;box-shadow:var(--shadow);overflow-x:auto}
.stats{display:grid;grid-template-columns:repeat(auto-fit,minmax(170px,1fr));gap:14px;margin-bottom:20px}
.stat-box{background:var(--panel);border:1px solid var(--border);border-radius:14px;padding:18px;box-shadow:var(--shadow);position:relative;overflow:hidden}
.stat-box:before{content:"";position:absolute;left:0;top:14px;bottom:14px;width:4px;background:var(--primary);border-radius:0 3px 3px 0}
.stat-box .num{font-size:32px;line-height:1.15;color:var(--text);font-weight:800;letter-spacing:-.03em}
.stat-box .label{color:var(--muted);margin-top:6px;font-size:12px;font-weight:600}
h2{font-size:16px;padding-bottom:12px;margin-bottom:16px;border-bottom:1px solid var(--border);color:var(--text)}
h3{font-size:15px;margin:18px 0 10px;color:var(--text)}
p{color:#374151;margin:6px 0;font-size:14px}
ul{padding-left:20px;margin:8px 0;color:#374151;font-size:14px}
code{background:#f1f5f9;border:1px solid var(--border);border-radius:5px;padding:2px 6px;color:var(--red);font-family:ui-monospace,SFMono-Regular,Consolas,monospace;font-size:13px}
table{width:100%;min-width:700px;border-collapse:collapse}
th,td{padding:11px 13px;text-align:left;border-bottom:1px solid var(--border);vertical-align:middle;font-size:13px}
th{background:var(--panel-soft);color:#4b5563;font-size:12px;letter-spacing:.03em;text-transform:uppercase;white-space:nowrap;font-weight:700}
td{color:#262b33}
tbody tr:hover{background:#fafbfc}
tbody tr:last-child td{border-bottom:0}
.btn{display:inline-flex;align-items:center;gap:6px;padding:7px 12px;background:var(--primary);color:#fff;border:1px solid var(--primary);border-radius:8px;cursor:pointer;text-decoration:none;font-size:13px;font-weight:700;line-height:1.3;transition:background .15s,transform .15s;white-space:nowrap}
.btn:hover{background:var(--primary-hover);transform:translateY(-1px)}
.btn:active{transform:translateY(0)}
.btn-danger{background:var(--red);border-color:var(--red)}
.btn-danger:hover{background:var(--red-dark);border-color:var(--red-dark)}
.btn-success{background:var(--green);border-color:var(--green)}
.btn-success:hover{background:var(--green-dark);border-color:var(--green-dark)}
.btn-sm{padding:5px 9px;font-size:12px}
.form-group{margin-bottom:15px}
.form-group label{display:block;margin-bottom:6px;font-weight:700;color:#374151;font-size:13px}
.form-group input,.form-group select,.form-group textarea{width:100%;padding:10px 12px;border:1px solid #d1d5db;border-radius:8px;background:#fff;color:var(--text);font:inherit;font-size:14px;transition:border .15s,box-shadow .15s}
.form-group textarea{min-height:100px;resize:vertical}
.form-group input:focus,.form-group select:focus,.form-group textarea:focus{outline:none;border-color:#94a3b8;box-shadow:0 0 0 3px rgba(148,163,184,.18)}
.error,.success{padding:11px 13px;border-radius:8px;margin-bottom:14px;border:1px solid;font-weight:600;font-size:13px}
.error{background:#fff1f0;color:#9f1c14;border-color:#fecaca}
.success{background:#ecfdf3;color:#166534;border-color:#bbf7d0}
.badge{display:inline-flex;align-items:center;padding:3px 8px;border-radius:999px;font-size:11px;font-weight:800;border:1px solid transparent}
.badge-success{background:#ecfdf3;color:#166534;border-color:#bbf7d0}
.badge-warning{background:#fff7ed;color:#9a3412;border-color:#fed7aa}
.badge-danger{background:#fff1f0;color:#b42318;border-color:#fecaca}
.pagination{display:flex;gap:6px;margin-top:18px;flex-wrap:wrap}
.pagination a{padding:7px 13px;background:#fff;border-radius:8px;text-decoration:none;color:#374151;border:1px solid var(--border);font-size:13px;font-weight:600}
.pagination a.active{background:var(--primary);color:#fff;border-color:var(--primary)}
@media(max-width:768px){
:root{--sidebar-w:0}
.sidebar{transform:translateX(-100%);transition:transform .2s}
.sidebar.open{transform:translateX(0)}
.main{margin-left:0}
.topbar{padding:10px 16px}
.content{padding:16px}
.menu-btn{display:inline-flex!important}
}
.menu-btn{display:none;padding:6px 8px;background:none;border:1px solid var(--border);border-radius:6px;cursor:pointer;color:var(--text);font-size:18px;align-items:center;justify-content:center}
.sidebar-overlay{display:none;position:fixed;inset:0;background:rgba(0,0,0,.35);z-index:15;opacity:0;transition:opacity .2s}
.sidebar.open~.sidebar-overlay{display:block;opacity:1}
.sys-grid{display:grid;grid-template-columns:repeat(auto-fit,minmax(220px,1fr));gap:14px}
.sys-card{background:var(--panel-soft);border:1px solid var(--border);border-radius:12px;padding:18px}
.sys-card-icon{font-size:20px;line-height:1;margin-bottom:8px}
.sys-card-title{font-size:11px;font-weight:700;color:var(--muted);text-transform:uppercase;letter-spacing:.04em;margin-bottom:6px}
.sys-card-val{font-size:18px;font-weight:800;color:var(--text);margin-bottom:8px}
.sys-card-sub{font-size:11px;color:var(--muted);margin-top:8px}
.gauge-wrap{position:relative;display:flex;justify-content:center;align-items:center;padding:8px 0;cursor:pointer}
.gauge-wrap .progress-tip{display:none;position:absolute;top:-6px;left:50%;transform:translateX(-50%);background:var(--primary);color:#fff;font-size:11px;padding:5px 10px;border-radius:6px;white-space:nowrap;z-index:50;pointer-events:none;font-weight:600;line-height:1.4;text-align:center}
.gauge-wrap .progress-tip:after{content:"";position:absolute;top:100%;left:50%;transform:translateX(-50%);border:5px solid transparent;border-top-color:var(--primary)}
.gauge-wrap:hover .progress-tip{display:block}
.gauge-svg{display:block}
.gauge-track{fill:none;stroke:#e5e7eb;stroke-width:6}
.gauge-fill{fill:none;stroke-width:6;stroke-linecap:round;transition:stroke-dashoffset .6s}
.gauge-center{font-size:14px;font-weight:800;fill:var(--text);font-family:inherit}
.gauge-label{font-size:9px;fill:var(--muted);font-family:inherit}
/* collapsed sidebar */
.sidebar .collapse-btn{position:absolute;right:8px;top:14px;background:none;border:none;color:rgba(255,255,255,.5);cursor:pointer;font-size:14px;padding:4px 6px;border-radius:4px;line-height:1;display:none}
.sidebar .collapse-btn:hover{color:#fff;background:rgba(255,255,255,.08)}
.sidebar.collapsed{width:64px}
.sidebar.collapsed .sidebar-header span{display:none}
.sidebar.collapsed .sidebar-header{padding:12px 10px;font-size:0;position:relative}
.sidebar.collapsed .sidebar-header:after{content:"☁";font-size:20px}
.sidebar.collapsed .collapse-btn{right:18px;top:12px}
.sidebar.collapsed .sidebar-nav a{font-size:0;padding:12px 10px;text-align:center}
.sidebar.collapsed .sidebar-nav a .nav-text{display:none}
.sidebar.collapsed .sidebar-nav a .nav-icon{display:inline;font-size:18px}
.sidebar.collapsed .sidebar-nav .section{font-size:0;height:1px;padding:0;margin:4px 20px;border-bottom:1px solid rgba(255,255,255,.06)}
.sidebar.collapsed~.main{margin-left:64px}
@media(min-width:769px){
.sidebar .collapse-btn{display:block}
.sidebar .sidebar-header{padding-right:36px}
}
@media(max-width:768px){
.sidebar .collapse-btn{display:none!important}
.sidebar .sidebar-header{padding-right:20px}
.sidebar.collapsed{width:var(--sidebar-w);transform:translateX(-100%)}
.sidebar.collapsed .sidebar-header:after{content:none}
.sidebar.collapsed .sidebar-header span{display:inline-block}
.sidebar.collapsed .sidebar-header{font-size:17px;padding:20px 20px 18px}
.sidebar.collapsed~.main{margin-left:0}
}
.nav-icon{display:none;font-style:normal}
@media(min-width:769px){
.sidebar.collapsed .nav-icon{display:inline}
}
</style>
</head>
<body>
<div class="layout">
<aside class="sidebar" id="sidebar">
<div class="sidebar-header"><span></span>云黑系统<button class="collapse-btn" id="collapseBtn" title="折叠/展开">&laquo;</button></div>
<div class="sidebar-nav">
<a href="/admin/" {{if eq .CurrentPage "/admin"}}class="active"{{end}}><i class="nav-icon">&#128202;</i><span class="nav-text">控制台</span></a>
<div class="section">云黑管理</div>
<a href="/admin/list" {{if eq .CurrentPage "/admin/list"}}class="active"{{end}}><i class="nav-icon">&#128203;</i><span class="nav-text">云黑列表</span></a>
<a href="/admin/review" {{if eq .CurrentPage "/admin/review"}}class="active"{{end}}><i class="nav-icon">&#9989;</i><span class="nav-text">审核列表</span></a>
<a href="/admin/add" {{if eq .CurrentPage "/admin/add"}}class="active"{{end}}><i class="nav-icon">&#10133;</i><span class="nav-text">添加云黑</span></a>
<div class="section">系统</div>
<a href="/admin/stats" {{if eq .CurrentPage "/admin/stats"}}class="active"{{end}}><i class="nav-icon">&#128200;</i><span class="nav-text">统计</span></a>
<a href="/admin/admins" {{if eq .CurrentPage "/admin/admins"}}class="active"{{end}}><i class="nav-icon">&#128100;</i><span class="nav-text">管理员</span></a>
<a href="/admin/apikeys" {{if eq .CurrentPage "/admin/apikeys"}}class="active"{{end}}><i class="nav-icon">&#128273;</i><span class="nav-text">API密钥</span></a>
<a href="/admin/settings" {{if eq .CurrentPage "/admin/settings"}}class="active"{{end}}><i class="nav-icon">&#9881;</i><span class="nav-text">系统设置</span></a>
<a href="/admin/ai_settings" {{if eq .CurrentPage "/admin/ai_settings"}}class="active"{{end}}><i class="nav-icon">&#129302;</i><span class="nav-text">AI 设置</span></a>
<a href="/admin/logs" {{if eq .CurrentPage "/admin/logs"}}class="active"{{end}}><i class="nav-icon">&#128240;</i><span class="nav-text">日志</span></a>
<a href="/admin/access_logs" {{if eq .CurrentPage "/admin/access_logs"}}class="active"{{end}}><i class="nav-icon">&#128269;</i><span class="nav-text">访问日志</span></a>
<a href="/admin/ai_review_logs" {{if eq .CurrentPage "/admin/ai_review_logs"}}class="active"{{end}}><i class="nav-icon">&#129302;</i><span class="nav-text">AI 离线记录</span></a>
<a href="/admin/apidoc" {{if eq .CurrentPage "/admin/apidoc"}}class="active"{{end}}><i class="nav-icon">&#128214;</i><span class="nav-text">API文档</span></a>
</div>
</aside>
<div class="sidebar-overlay" id="sidebarOverlay" onclick="document.getElementById('sidebar').classList.remove('open')"></div>
<div class="main">
<header class="topbar">
<button class="menu-btn" id="menuBtn" onclick="document.getElementById('sidebar').classList.toggle('open')">&equiv;</button>
<span class="user">{{.Username}}</span>
<a href="/admin/logout">退出</a>
</header>
<div class="content">
{{.Content}}
</div>
</div>
<script>
(function(){
var sb=document.getElementById('sidebar'),menuBtn=document.getElementById('menuBtn'),overlay=document.getElementById('sidebarOverlay');
if(localStorage.getItem('sidebar_collapsed')==='1')sb.classList.add('collapsed');
function toggleSidebar(){sb.classList.toggle('collapsed');localStorage.setItem('sidebar_collapsed',sb.classList.contains('collapsed')?'1':'0')}
document.getElementById('collapseBtn').addEventListener('click',function(e){e.preventDefault();toggleSidebar()});
menuBtn.addEventListener('click',function(){sb.classList.toggle('open')});
if(overlay){
overlay.addEventListener('click',function(){sb.classList.remove('open')});
}
if(window.innerWidth<=768){
var navLinks=sb.querySelectorAll('.sidebar-nav a');
for(var i=0;i<navLinks.length;i++){
navLinks[i].addEventListener('click',function(){sb.classList.remove('open')});
}
}
})();
</script>
</body>
</html>
`

var adminTmpl = template.Must(template.New("admin").Parse(adminHTML))

type AdminPageData struct {
	Title       string
	Username    string
	CurrentPage string
	Content     template.HTML
}

func renderAdminPage(c echo.Context, title string, content string) error {
	cp := c.Path()
	if strings.HasSuffix(cp, "/") {
		cp = strings.TrimSuffix(cp, "/")
	}
	data := AdminPageData{
		Title:       title,
		Username:    getAdminUsername(c),
		CurrentPage: cp,
		Content:     template.HTML(content),
	}
	return adminTmpl.Execute(c.Response(), data)
}

func getAdminUsername(c echo.Context) string {
	v := c.Get("admin_username")
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func getAdminID(c echo.Context) int {
	v := c.Get("admin_id")
	if v == nil {
		return 0
	}
	if i, ok := v.(int); ok {
		return i
	}
	return 0
}

func esc(s string) string {
	return template.HTMLEscapeString(s)
}

func generateCSRFToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func getCSRFToken(c echo.Context) string {
	token, ok := c.Get("csrf_token").(string)
	if ok && token != "" {
		return token
	}
	return ""
}

func RequireRole(minRole int) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			role, ok := c.Get("admin_role").(int)
			if !ok || role > minRole {
				return c.String(http.StatusForbidden, "权限不足")
			}
			return next(c)
		}
	}
}

func csrfProtectMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if c.Request().Method != "POST" {
			return next(c)
		}
		cookie, err := c.Cookie("admin_session")
		if err != nil || cookie.Value == "" {
			return c.Redirect(302, "/admin/login")
		}
		var storedCSRF string
		err = DB.QueryRow("SELECT COALESCE(csrf_token,'') FROM admin_sessions WHERE token = ? AND datetime(expires_at) > datetime('now')", cookie.Value).Scan(&storedCSRF)
		if err != nil || storedCSRF == "" {
			return c.String(http.StatusForbidden, "CSRF token missing")
		}
		formToken := c.FormValue("csrf_token")
		if formToken == "" {
			formToken = c.QueryParam("csrf_token")
		}
		if subtle.ConstantTimeCompare([]byte(formToken), []byte(storedCSRF)) != 1 {
			return c.String(http.StatusForbidden, "CSRF token mismatch")
		}
		return next(c)
	}
}

func parseSeverity(value string) int {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 5 {
		return 1
	}
	return n
}

func RegisterAdminRoutes(e *echo.Echo, cfg *Config) {
	admin := e.Group("/admin")
	admin.GET("", func(c echo.Context) error {
		return c.Redirect(302, "/admin/")
	})

	loginRL := NewRateLimiter(cfg.RateLimit.Admin, time.Duration(cfg.RateLimit.Window)*time.Second)
	admin.GET("/login", adminLogin)
	admin.POST("/login", RateLimitMiddleware(loginRL, func(c echo.Context) string {
		return GetClientIP(c)
	})(adminLoginPost))
	admin.GET("/logout", adminLogout)

	adminAuth := admin.Group("")
	adminAuth.Use(AdminAuthMiddleware)

	operatorAuth := adminAuth.Group("")
	operatorAuth.Use(RequireRole(RoleOperator))
	operatorAuth.GET("/", adminDashboard)
	operatorAuth.GET("/index", adminDashboard)
	operatorAuth.GET("/status", adminStatusJSON)
	operatorAuth.GET("/password", adminChangePassword)
	operatorAuth.POST("/password", csrfProtectMiddleware(adminChangePasswordPost))
	operatorAuth.GET("/list", adminList)
	operatorAuth.GET("/detail", adminDetail)
	operatorAuth.GET("/add", adminAdd)
	operatorAuth.POST("/add", csrfProtectMiddleware(adminAddPost))
	operatorAuth.GET("/review", adminReview)
	operatorAuth.POST("/review_action", csrfProtectMiddleware(adminReviewAction))
	operatorAuth.GET("/stats", adminStats)
	operatorAuth.GET("/logs", adminLogs)
	operatorAuth.GET("/access_logs", adminAccessLogs)
	operatorAuth.GET("/apidoc", adminAPIDoc)
	operatorAuth.GET("/ai_review_logs", adminAIReviewLogs)

	reviewerAuth := adminAuth.Group("")
	reviewerAuth.Use(RequireRole(RoleReviewer))
	reviewerAuth.POST("/edit", csrfProtectMiddleware(adminEditPost))
	reviewerAuth.POST("/delete", csrfProtectMiddleware(adminDelete))
	reviewerAuth.POST("/subject/add_account", csrfProtectMiddleware(adminSubjectAddAccount))
	reviewerAuth.POST("/subject/delete_account", csrfProtectMiddleware(adminSubjectDeleteAccount))
	reviewerAuth.POST("/subject/merge", csrfProtectMiddleware(adminSubjectMerge))

	superAuth := adminAuth.Group("")
	superAuth.Use(RequireRole(RoleSuperAdmin))
	superAuth.GET("/admins", adminAdmins)
	superAuth.POST("/admins", csrfProtectMiddleware(adminAdminsPost))
	superAuth.POST("/delete_admin", csrfProtectMiddleware(adminDeleteAdmin))
	superAuth.GET("/apikeys", adminAPIKeys)
	superAuth.POST("/apikeys", csrfProtectMiddleware(adminAPIKeysPost))
	superAuth.POST("/toggle_apikey", csrfProtectMiddleware(adminToggleAPIKey))
	superAuth.GET("/settings", adminSettings)
	superAuth.POST("/settings", csrfProtectMiddleware(adminSettingsPost))
	superAuth.GET("/ai_settings", adminAISettings)
	superAuth.POST("/ai_settings", csrfProtectMiddleware(adminAISettingsPost))
	superAuth.GET("/api/models", adminAPIModels)
}

func AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie, err := c.Cookie("admin_session")
		if err != nil || cookie.Value == "" {
			return c.Redirect(302, "/admin/login")
		}

		token := cookie.Value
		var id int
		err = DB.QueryRow("SELECT admin_id FROM admin_sessions WHERE token = ? AND datetime(expires_at) > datetime('now')", token).Scan(&id)
		if err != nil || id <= 0 {
			return c.Redirect(302, "/admin/login")
		}

		var username, nickname string
		var role int
		err = DB.QueryRow("SELECT username, nickname, role FROM admins WHERE id = ?", id).Scan(&username, &nickname, &role)
		if err != nil || username == "" {
			return c.Redirect(302, "/admin/login")
		}

		c.Set("admin_id", id)
		c.Set("admin_username", username)
		c.Set("admin_nickname", nickname)
		c.Set("admin_role", role)

		var csrfToken string
		DB.QueryRow("SELECT COALESCE(csrf_token,'') FROM admin_sessions WHERE token = ?", token).Scan(&csrfToken)
		c.Set("csrf_token", csrfToken)

		return next(c)
	}
}

func adminLogin(c echo.Context) error {
	errMsg := c.QueryParam("error")
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>管理员登录 - 云黑系统</title>
<style>
*{box-sizing:border-box}
body{margin:0;background:#111318;min-height:100vh;display:flex;align-items:center;justify-content:center;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;color:#16181d;padding:20px}
.login-box{background:#fff;padding:34px;border-radius:18px;width:100%;max-width:380px;box-shadow:0 24px 60px rgba(0,0,0,.35);border:1px solid #e5e7eb}
h1{text-align:center;color:#111318;margin:0 0 8px;font-size:24px;letter-spacing:.02em}
h1:before{content:"";display:inline-block;width:11px;height:11px;margin-right:10px;border-radius:50%;background:#d92d20;box-shadow:0 0 0 4px rgba(217,45,32,.16);vertical-align:2px}
h1+p{text-align:center;color:#6b7280;margin:0 0 28px;font-weight:600}
.form-group{margin-bottom:18px}
.form-group label{display:block;margin-bottom:7px;color:#2d333d;font-weight:700;font-size:14px}
.form-group input{width:100%;padding:12px;border:1px solid #d6d9df;border-radius:10px;box-sizing:border-box;font-size:15px;transition:border .18s,box-shadow .18s}
.form-group input:focus{outline:none;border-color:#d92d20;box-shadow:0 0 0 3px rgba(217,45,32,.12)}
button{width:100%;padding:12px;background:#111318;color:#fff;border:1px solid #111318;border-radius:10px;cursor:pointer;font-size:15px;font-weight:800;transition:background .18s,transform .18s}
button:hover{background:#2a2f38;transform:translateY(-1px)}
.tips{text-align:center;margin-top:20px;color:#6b7280;font-size:13px}
.error{background:#fff1f0;color:#9f1c14;padding:12px;border-radius:10px;margin-bottom:15px;font-size:14px;border:1px solid #ffc9c3;font-weight:600}
</style>
</head>
<body>
<div class="login-box">
<h1>云黑系统</h1>
<p>管理后台登录</p>`
	if errMsg != "" {
		html += `<div class="error">` + esc(errMsg) + `</div>`
	}
	html += `<form method="POST">
<div class="form-group"><label>用户名</label><input type="text" name="username" required></div>
<div class="form-group"><label>密码</label><input type="password" name="password" required></div>
<button type="submit">登录</button>
</form>
</div>
</body>
</html>`
	c.HTML(http.StatusOK, html)
	return nil
}

func adminLoginPost(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	if username == "" || password == "" {
		return c.Redirect(302, "/admin/login?error=请填写用户名和密码")
	}

	var admin struct {
		ID                 int
		Password           string
		Nickname           string
		Role               int
		MustChangePassword int
	}
	err := DB.QueryRow("SELECT id, password, nickname, role, must_change_password FROM admins WHERE username = ?", username).Scan(&admin.ID, &admin.Password, &admin.Nickname, &admin.Role, &admin.MustChangePassword)
	if err != nil || !CheckPassword(password, admin.Password) {
		return c.Redirect(302, "/admin/login?error=用户名或密码错误")
	}

	token := RandomString(64)
	expire := 7200
	if AppConfig != nil && AppConfig.Admin.Expire > 0 {
		expire = AppConfig.Admin.Expire
	}
	DB.Exec("INSERT INTO admin_sessions (token, admin_id, expires_at, created_at) VALUES (?, ?, datetime('now', ?), datetime('now'))", token, admin.ID, "+"+strconv.Itoa(expire)+" seconds")
	csrfToken := generateCSRFToken()
	DB.Exec("UPDATE admin_sessions SET csrf_token = ? WHERE token = ?", csrfToken, token)
	cookie := &http.Cookie{Name: "admin_session", Value: token, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: expire}
	if AppConfig != nil && AppConfig.Security.SecureCookie {
		cookie.Secure = true
	}
	c.SetCookie(cookie)

	if admin.MustChangePassword == 1 {
		return c.Redirect(302, "/admin/password")
	}

	return c.Redirect(302, "/admin/")
}

func adminLogout(c echo.Context) error {
	cookie, _ := c.Cookie("admin_session")
	if cookie != nil {
		DB.Exec("DELETE FROM admin_sessions WHERE token = ?", cookie.Value)
		cookie.MaxAge = -1
		cookie.Path = "/"
		cookie.HttpOnly = true
		cookie.SameSite = http.SameSiteLaxMode
		c.SetCookie(cookie)
	}
	return c.Redirect(302, "/admin/login")
}

func adminDashboard(c echo.Context) error {
	var total, pending, approved, rejected, today, queryToday int
	var severity [6]int

	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE status = 1").Scan(&total)
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE status = 0").Scan(&pending)
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE status = 1").Scan(&approved)
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE status = 2").Scan(&rejected)
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE DATE(created_at) = DATE('now')").Scan(&today)
	DB.QueryRow("SELECT COUNT(*) FROM stats_log WHERE type = 'query' AND DATE(created_at) = DATE('now')").Scan(&queryToday)

	// AI stats
	var aiToday, aiAutoApprove, aiAutoReject, aiManual int
	DB.QueryRow("SELECT COUNT(*) FROM ai_review_logs WHERE DATE(created_at) = DATE('now')").Scan(&aiToday)
	DB.QueryRow("SELECT COUNT(*) FROM ai_review_logs WHERE ai_result = 'auto_approve' AND DATE(created_at) = DATE('now')").Scan(&aiAutoApprove)
	DB.QueryRow("SELECT COUNT(*) FROM ai_review_logs WHERE ai_result = 'auto_reject' AND DATE(created_at) = DATE('now')").Scan(&aiAutoReject)
	DB.QueryRow("SELECT COUNT(*) FROM ai_review_logs WHERE ai_result = 'manual_review' AND DATE(created_at) = DATE('now')").Scan(&aiManual)

	for i := 1; i <= 5; i++ {
		DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE severity = ? AND status = 1", i).Scan(&severity[i])
	}

	uptime := time.Since(StartTime)
	uptimeStr := fmt.Sprintf("%d天 %d小时 %d分", int(uptime.Hours())/24, int(uptime.Hours())%24, int(uptime.Minutes())%60)
	goVersion := strings.TrimPrefix(runtime.Version(), "go")
	goroutines := runtime.NumGoroutine()
	dbSize := "N/A"
	if fi, err := os.Stat(AppConfig.DataDir + "/cloudblack.db"); err == nil {
		dbSize = fmt.Sprintf("%.1f MB", float64(fi.Size())/1024/1024)
	}

	// CPU
	st := GetLiveStatus()
	cpuPct := st.CPUPercent
	cpuBar := gaugeHTML("gauge-cpu", cpuPct, fmt.Sprintf("CPU 占用: %.1f%%", cpuPct))

	// Memory (system physical, from gopsutil)
	sysMemUsed := st.SysMemUsed
	sysMemTotal := st.SysMemTotal
	var memPct float64
	if sysMemTotal > 0 {
		memPct = math.Round(float64(sysMemUsed)/float64(sysMemTotal)*1000) / 10
	}
	memUsedMB := fmt.Sprintf("%.1f GB", float64(sysMemUsed)/1024/1024/1024)
	memTotalMB := fmt.Sprintf("%.1f GB", float64(sysMemTotal)/1024/1024/1024)
	memBar := gaugeHTML("gauge-mem", memPct, "已用 "+memUsedMB+" / 总量 "+memTotalMB)

	content := `
	<div class="stats">
		<div class="stat-box"><div class="num">` + strconv.Itoa(total) + `</div><div class="label">总云黑数</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(pending) + `</div><div class="label">待审核</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(approved) + `</div><div class="label">已通过</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(rejected) + `</div><div class="label">已拒绝</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(today) + `</div><div class="label">今日提交</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(queryToday) + `</div><div class="label">今日查询</div></div>
	</div>
	<div class="card">
		<h2>严重程度分布</h2>
		<p>轻微: ` + strconv.Itoa(severity[1]) + ` | 一般: ` + strconv.Itoa(severity[2]) + ` | 较重: ` + strconv.Itoa(severity[3]) + ` | 严重: ` + strconv.Itoa(severity[4]) + ` | 极其严重: ` + strconv.Itoa(severity[5]) + `</p>
	</div>
	<div class="card">
		<h2>AI 离线审核统计</h2>
		<div class="sys-grid">
			<div class="sys-card">
				<div class="sys-card-icon">&#129302;</div>
				<div class="sys-card-title">今日 AI 处理</div>
				<div class="sys-card-val">` + strconv.Itoa(aiToday) + `</div>
				<div class="sys-card-sub">自动审核记录</div>
			</div>
			<div class="sys-card">
				<div class="sys-card-icon" style="color:#16a34a">&#9989;</div>
				<div class="sys-card-title">自动通过</div>
				<div class="sys-card-val" style="color:#16a34a">` + strconv.Itoa(aiAutoApprove) + `</div>
				<div class="sys-card-sub">多人举报/高可信度</div>
			</div>
			<div class="sys-card">
				<div class="sys-card-icon" style="color:#dc2626">&#10060;</div>
				<div class="sys-card-title">自动拒绝</div>
				<div class="sys-card-val" style="color:#dc2626">` + strconv.Itoa(aiAutoReject) + `</div>
				<div class="sys-card-sub">垃圾/广告/灌水</div>
			</div>
			<div class="sys-card">
				<div class="sys-card-icon" style="color:#f59e0b">&#128172;</div>
				<div class="sys-card-title">转人工</div>
				<div class="sys-card-val" style="color:#f59e0b">` + strconv.Itoa(aiManual) + `</div>
				<div class="sys-card-sub">AI 不确定</div>
			</div>
		</div>
	</div>
	<div class="card">
		<h2>服务器状态</h2>
		<div class="sys-grid">
			<div class="sys-card">
				<div class="sys-card-icon">&#9881;</div>
				<div class="sys-card-title">运行时间</div>
				<div class="sys-card-val">` + uptimeStr + `</div>
				<div class="sys-card-sub">Go ` + goVersion + ` &middot; ` + strconv.Itoa(goroutines) + ` goroutines</div>
			</div>
			<div class="sys-card">
				<div class="sys-card-icon">&#9889;</div>
				<div class="sys-card-title">CPU</div>
				` + cpuBar + `
			</div>
			<div class="sys-card">
				<div class="sys-card-icon">&#128451;</div>
				<div class="sys-card-title">内存</div>
				` + memBar + `
			</div>
			<div class="sys-card">
				<div class="sys-card-icon">&#128190;</div>
				<div class="sys-card-title">数据库</div>
				<div class="sys-card-val">` + dbSize + `</div>
				<div class="sys-card-sub">SQLite</div>
			</div>
		</div>
	</div>
	<script>
	(function(){
		var C=226.19;
		function gaugeColor(p){return p>80?'#dc2626':p>50?'#f59e0b':'#16a34a'}
		function updateGauge(id,pct,tip){
			var fill=document.getElementById(id+'-fill'),pctEl=document.getElementById(id+'-pct'),tipEl=document.getElementById(id+'-tip');
			if(!fill)return;
			var off=C-(pct/100)*C;
			fill.setAttribute('stroke-dashoffset',off.toFixed(2));
			fill.setAttribute('stroke',gaugeColor(pct));
			if(pctEl)pctEl.textContent=Math.round(pct)+'%';
			if(tipEl)tipEl.textContent=tip;
		}
		function refresh(){
			fetch('/admin/status').then(function(r){return r.json()}).then(function(d){
				updateGauge('gauge-cpu',d.cpu,'CPU \u5360\u7528: '+d.cpu.toFixed(1)+'%');
				updateGauge('gauge-mem',d.mem_pct,'\u5df2\u7528 '+d.mem_used+' / \u603b\u91cf '+d.mem_total);
			}).catch(function(){});
		}
		setInterval(refresh,3000);
	})();
	</script>
	`

	return renderAdminPage(c, "控制台", content)
}

func gaugeHTML(id string, pct float64, tooltip string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	color := "#16a34a"
	if pct > 80 {
		color = "#dc2626"
	} else if pct > 50 {
		color = "#f59e0b"
	}
	circumference := 226.19
	offset := circumference - (pct/100)*circumference
	pctStr := fmt.Sprintf("%.0f", pct)
	return `<div class="gauge-wrap"><div class="progress-tip" id="` + id + `-tip">` + esc(tooltip) + `</div><svg class="gauge-svg" width="100" height="100" viewBox="0 0 90 90"><circle class="gauge-track" cx="45" cy="45" r="36"/><circle class="gauge-fill" id="` + id + `-fill" cx="45" cy="45" r="36" stroke="` + color + `" stroke-dasharray="` + fmt.Sprintf("%.2f", circumference) + `" stroke-dashoffset="` + fmt.Sprintf("%.2f", offset) + `" transform="rotate(-90 45 45)"/><text x="45" y="43" text-anchor="middle" class="gauge-center" id="` + id + `-pct">` + pctStr + `%</text><text x="45" y="58" text-anchor="middle" class="gauge-label">占用</text></svg></div>`
}

func adminStatusJSON(c echo.Context) error {
	st := GetLiveStatus()
	sysMemUsed := st.SysMemUsed
	sysMemTotal := st.SysMemTotal
	var memPct float64
	if sysMemTotal > 0 {
		memPct = math.Round(float64(sysMemUsed)/float64(sysMemTotal)*1000) / 10
	}
	return c.JSON(http.StatusOK, map[string]interface{}{
		"cpu":       st.CPUPercent,
		"mem_pct":   memPct,
		"mem_used":  fmt.Sprintf("%.1f GB", float64(sysMemUsed)/1024/1024/1024),
		"mem_total": fmt.Sprintf("%.1f GB", float64(sysMemTotal)/1024/1024/1024),
	})
}

func adminList(c echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	severity := c.QueryParam("severity")
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 20
	offset := (page - 1) * pageSize

	where := "WHERE 1=1"
	args := []interface{}{}
	if q != "" {
		where += " AND (CAST(qq AS TEXT) LIKE ? OR nickname LIKE ? OR reason LIKE ? OR subject_name LIKE ? OR tags LIKE ? OR accounts LIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like, like, like, like, like)
	}
	if severity != "" {
		where += " AND severity = ?"
		args = append(args, severity)
	}

	var total int
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list "+where, args...).Scan(&total)
	queryArgs := append(args, pageSize, offset)
	rows, err := DB.Query("SELECT id, qq, nickname, reason, severity, status, created_at, COALESCE(tags,'') FROM cloudblack_list "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...)
	if err != nil {
		return renderAdminPage(c, "云黑列表", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var r struct {
			ID        int
			QQ        int64
			Nickname  string
			Reason    string
			Severity  int
			Status    int
			CreatedAt string
			Tags      string
		}
		rows.Scan(&r.ID, &r.QQ, &r.Nickname, &r.Reason, &r.Severity, &r.Status, &r.CreatedAt, &r.Tags)
		records = append(records, map[string]interface{}{
			"id":            r.ID,
			"qq":            r.QQ,
			"nickname":      r.Nickname,
			"reason":        r.Reason,
			"severity":      r.Severity,
			"severity_text": GetSeverityText(r.Severity),
			"status":        r.Status,
			"status_text":   GetStatusText(r.Status),
			"created_at":    r.CreatedAt,
			"tags":          r.Tags,
		})
	}

	content := `<div class="card"><h2>云黑列表</h2><form method="GET" action="/admin/list" style="display:grid;grid-template-columns:1fr 160px auto auto;gap:10px;margin-bottom:16px"><input type="text" name="q" value="` + esc(q) + `" placeholder="搜索 QQ / 主体 / 昵称 / 原因 / 标签 / 账号"><select name="severity"><option value="">全部严重程度</option>`
	for i := 1; i <= 5; i++ {
		selected := ""
		if severity == strconv.Itoa(i) {
			selected = " selected"
		}
		content += `<option value="` + strconv.Itoa(i) + `"` + selected + `>` + GetSeverityText(i) + `</option>`
	}
	content += `</select><button class="btn" type="submit">查询</button><a class="btn" href="/admin/list">重置</a></form><table><thead><tr><th>ID</th><th>QQ号</th><th>昵称</th><th>标签</th><th>原因</th><th>严重程度</th><th>状态</th><th>添加时间</th><th>操作</th></tr></thead><tbody>`
	for _, r := range records {
		id := strconv.Itoa(r["id"].(int))
		content += `<tr><td>` + id + `</td><td>` + strconv.FormatInt(r["qq"].(int64), 10) + `</td><td>` + esc(r["nickname"].(string)) + `</td><td>` + esc(r["tags"].(string)) + `</td><td>` + esc(r["reason"].(string)) + `</td><td>` + esc(r["severity_text"].(string)) + `</td><td>` + esc(r["status_text"].(string)) + `</td><td>` + esc(r["created_at"].(string)) + `</td><td><a href="/admin/detail?id=` + id + `" class="btn">详情</a> <a href="/admin/edit?id=` + id + `" class="btn">编辑</a> <form method="POST" action="/admin/delete" style="display:inline"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="id" value="` + id + `"><button type="submit" class="btn btn-danger" onclick="return confirm('确认删除?')">删除</button></form></td></tr>`
	}
	content += `</tbody></table>`
	maxPage := (total + pageSize - 1) / pageSize
	if maxPage < 1 {
		maxPage = 1
	}
	content += `<div class="pagination">`
	for i := 1; i <= maxPage; i++ {
		active := ""
		if i == page {
			active = ` class="active"`
		}
		content += `<a` + active + ` href="/admin/list?q=` + esc(q) + `&severity=` + esc(severity) + `&page=` + strconv.Itoa(i) + `">` + strconv.Itoa(i) + `</a>`
	}
	content += `</div></div>`

	return renderAdminPage(c, "云黑列表", content)
}

func adminDetail(c echo.Context) error {
	id := c.QueryParam("id")
	var r struct {
		ID                                                                                  int
		QQ                                                                                  int64
		Nickname, Reason, Tags, AccountsRaw, SubjectName, CreatedAt, ReviewedAt, ReviewNote string
		Severity, Status, SubjectID, ReviewedBy                                             int
	}
	err := DB.QueryRow("SELECT id, qq, COALESCE(nickname,''), reason, severity, status, COALESCE(subject_id,0), COALESCE(subject_name,''), COALESCE(tags,''), COALESCE(accounts,''), created_at, COALESCE(reviewed_at,''), COALESCE(reviewed_by,0), COALESCE(review_note,'') FROM cloudblack_list WHERE id = ?", id).Scan(&r.ID, &r.QQ, &r.Nickname, &r.Reason, &r.Severity, &r.Status, &r.SubjectID, &r.SubjectName, &r.Tags, &r.AccountsRaw, &r.CreatedAt, &r.ReviewedAt, &r.ReviewedBy, &r.ReviewNote)
	if err != nil {
		return c.Redirect(302, "/admin/list")
	}
	accounts := DecodeAccounts(r.AccountsRaw)
	content := `<div class="card"><h2>云黑详情 #` + strconv.Itoa(r.ID) + `</h2><p><strong>主体：</strong>` + esc(r.SubjectName) + `（ID: ` + strconv.Itoa(r.SubjectID) + `）</p><p><strong>QQ：</strong>` + strconv.FormatInt(r.QQ, 10) + `</p><p><strong>昵称：</strong>` + esc(r.Nickname) + `</p><p><strong>标签：</strong>` + esc(r.Tags) + `</p><p><strong>严重程度：</strong>` + GetSeverityText(r.Severity) + ` - ` + GetSeverityDesc(r.Severity) + `</p><p><strong>原因：</strong>` + esc(r.Reason) + `</p><p><strong>审核：</strong>管理员ID ` + strconv.Itoa(r.ReviewedBy) + ` / ` + esc(r.ReviewedAt) + `</p><p><strong>审核备注：</strong>` + esc(r.ReviewNote) + `</p></div>`
	content += `<div class="card"><h2>关联账号</h2><table><thead><tr><th>平台</th><th>账号</th><th>昵称</th><th>操作</th></tr></thead><tbody>`
	for _, a := range accounts {
		content += `<tr><td>` + esc(a.Platform) + `</td><td>` + esc(a.Account) + `</td><td>` + esc(a.Nickname) + `</td><td><form method="POST" action="/admin/subject/delete_account"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="record_id" value="` + id + `"><input type="hidden" name="platform" value="` + esc(a.Platform) + `"><input type="hidden" name="account" value="` + esc(a.Account) + `"><button class="btn btn-danger" onclick="return confirm('确认移除该关联账号?')">移除</button></form></td></tr>`
	}
	content += `</tbody></table><h3>添加账号</h3><form method="POST" action="/admin/subject/add_account" style="display:grid;grid-template-columns:1fr 1fr 1fr auto;gap:10px"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="record_id" value="` + id + `"><input name="platform" placeholder="平台" required><input name="account" placeholder="账号" required><input name="nickname" placeholder="昵称"><button class="btn">添加</button></form></div>`
	content += `<div class="card"><h2>合并主体</h2><p>把当前记录的主体合并到目标主体 ID，当前记录和关联账号会归属到目标主体。</p><form method="POST" action="/admin/subject/merge" style="display:grid;grid-template-columns:1fr auto;gap:10px"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="record_id" value="` + id + `"><input name="target_subject_id" placeholder="目标主体ID" required><button class="btn">合并</button></form></div>`
	return renderAdminPage(c, "云黑详情", content)
}

func adminSubjectAddAccount(c echo.Context) error {
	recordID := c.FormValue("record_id")
	platform := strings.TrimSpace(c.FormValue("platform"))
	account := strings.TrimSpace(c.FormValue("account"))
	nickname := strings.TrimSpace(c.FormValue("nickname"))
	if recordID == "" || platform == "" || account == "" {
		return c.Redirect(302, "/admin/detail?id="+recordID)
	}

	var subjectID int
	var accountsRaw string
	DB.QueryRow("SELECT COALESCE(subject_id,0), COALESCE(accounts,'') FROM cloudblack_list WHERE id = ?", recordID).Scan(&subjectID, &accountsRaw)
	accounts := DecodeAccounts(accountsRaw)
	accounts = append(accounts, LinkedAccount{Platform: platform, Account: account, Nickname: nickname})
	encoded := EncodeAccounts(accounts)
	DB.Exec("UPDATE cloudblack_list SET accounts = ? WHERE id = ?", encoded, recordID)
	if subjectID > 0 {
		DB.Exec("INSERT INTO subject_accounts (subject_id, platform, account, nickname, created_at) VALUES (?, ?, ?, ?, datetime('now'))", subjectID, platform, account, nickname)
	}
	return c.Redirect(302, "/admin/detail?id="+recordID)
}

func adminSubjectDeleteAccount(c echo.Context) error {
	recordID := c.FormValue("record_id")
	platform := c.FormValue("platform")
	account := c.FormValue("account")
	var subjectID int
	var accountsRaw string
	DB.QueryRow("SELECT COALESCE(subject_id,0), COALESCE(accounts,'') FROM cloudblack_list WHERE id = ?", recordID).Scan(&subjectID, &accountsRaw)
	var filtered []LinkedAccount
	for _, item := range DecodeAccounts(accountsRaw) {
		if item.Platform == platform && item.Account == account {
			continue
		}
		filtered = append(filtered, item)
	}
	DB.Exec("UPDATE cloudblack_list SET accounts = ? WHERE id = ?", EncodeAccounts(filtered), recordID)
	if subjectID > 0 {
		DB.Exec("DELETE FROM subject_accounts WHERE subject_id = ? AND platform = ? AND account = ?", subjectID, platform, account)
	}
	return c.Redirect(302, "/admin/detail?id="+recordID)
}

func adminSubjectMerge(c echo.Context) error {
	recordID := c.FormValue("record_id")
	targetSubjectID := c.FormValue("target_subject_id")
	if recordID == "" || targetSubjectID == "" {
		return c.Redirect(302, "/admin/detail?id="+recordID)
	}

	var currentSubjectID int
	var accountsRaw string
	DB.QueryRow("SELECT COALESCE(subject_id,0), COALESCE(accounts,'') FROM cloudblack_list WHERE id = ?", recordID).Scan(&currentSubjectID, &accountsRaw)
	for _, account := range DecodeAccounts(accountsRaw) {
		DB.Exec("INSERT INTO subject_accounts (subject_id, platform, account, nickname, created_at) VALUES (?, ?, ?, ?, datetime('now'))", targetSubjectID, account.Platform, account.Account, account.Nickname)
	}
	DB.Exec("UPDATE cloudblack_list SET subject_id = ? WHERE id = ?", targetSubjectID, recordID)
	if currentSubjectID > 0 && strconv.Itoa(currentSubjectID) != targetSubjectID {
		DB.Exec("DELETE FROM subject_accounts WHERE subject_id = ?", currentSubjectID)
		DB.Exec("DELETE FROM black_subjects WHERE id = ?", currentSubjectID)
	}
	return c.Redirect(302, "/admin/detail?id="+recordID)
}

func adminReview(c echo.Context) error {
	rows, err := DB.Query("SELECT id, qq, nickname, reason, severity, created_at, COALESCE(tags,''), COALESCE(accounts,'') FROM cloudblack_records WHERE status = 0 ORDER BY id DESC")
	if err != nil {
		return renderAdminPage(c, "审核列表", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var r struct {
			ID          int
			QQ          int64
			Nickname    string
			Reason      string
			Severity    int
			CreatedAt   string
			Tags        string
			AccountsRaw string
		}
		rows.Scan(&r.ID, &r.QQ, &r.Nickname, &r.Reason, &r.Severity, &r.CreatedAt, &r.Tags, &r.AccountsRaw)
		records = append(records, map[string]interface{}{
			"id":            r.ID,
			"qq":            r.QQ,
			"nickname":      r.Nickname,
			"reason":        r.Reason,
			"severity":      r.Severity,
			"severity_text": GetSeverityText(r.Severity),
			"severity_desc": GetSeverityDesc(r.Severity),
			"tags":          r.Tags,
			"accounts":      DecodeAccounts(r.AccountsRaw),
			"created_at":    r.CreatedAt,
		})
	}

	content := `<div class="card"><table><thead><tr><th>ID</th><th>QQ号</th><th>昵称</th><th>原因</th><th>标签</th><th>关联账号</th><th>严重程度</th><th>提交时间</th><th>审核备注/拒绝原因</th><th>操作</th></tr></thead><tbody>`
	for _, r := range records {
		accountsText := "-"
		if accounts, ok := r["accounts"].([]LinkedAccount); ok && len(accounts) > 0 {
			var items []string
			for _, account := range accounts {
				items = append(items, account.Platform+":"+account.Account)
			}
			accountsText = strings.Join(items, " / ")
		}
		reviewID := strconv.Itoa(r["id"].(int))
		content += `<tr><td>` + reviewID + `</td><td>` + strconv.FormatInt(r["qq"].(int64), 10) + `</td><td>` + esc(r["nickname"].(string)) + `</td><td>` + esc(r["reason"].(string)) + `</td><td>` + esc(r["tags"].(string)) + `</td><td>` + esc(accountsText) + `</td><td>` + esc(r["severity_text"].(string)) + `<br><small style="color:#6b7280">` + esc(r["severity_desc"].(string)) + `</small></td><td>` + esc(r["created_at"].(string)) + `</td><td><textarea form="approve-` + reviewID + `" name="review_note" placeholder="通过备注，可选" style="min-width:180px;min-height:58px"></textarea><textarea form="reject-` + reviewID + `" name="review_note" placeholder="拒绝原因，必填" style="min-width:180px;min-height:58px;margin-top:6px"></textarea></td><td><form id="approve-` + reviewID + `" method="POST" action="/admin/review_action" style="display:inline"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="id" value="` + reviewID + `"><input type="hidden" name="action" value="approve"><button type="submit" class="btn btn-success">通过</button></form> <form id="reject-` + reviewID + `" method="POST" action="/admin/review_action" style="display:inline"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="id" value="` + reviewID + `"><input type="hidden" name="action" value="reject"><button type="submit" class="btn btn-danger">拒绝</button></form></td></tr>`
	}
	content += `</tbody></table></div>`

	return renderAdminPage(c, "审核列表", content)
}

func adminAdd(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	content := `<div class="card"><h2>添加云黑</h2>`
	if errorMsg != "" {
		content += `<div class="error">` + esc(errorMsg) + `</div>`
	}
	content += `<form method="POST" action="/admin/add">
		<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
		<div class="form-group"><label>QQ号 *</label><input type="text" name="qq" required></div>
		<div class="form-group"><label>昵称</label><input type="text" name="nickname"></div>
		<div class="form-group"><label>云黑原因 *</label><textarea name="reason" required></textarea></div>
		<div class="form-group"><label>严重程度</label>
		<select name="severity">
			<option value="1">轻微</option>
			<option value="2">一般</option>
			<option value="3">较重</option>
			<option value="4">严重</option>
			<option value="5">极其严重</option>
		</select>
		</div>
		<button type="submit" class="btn">提交</button>
	</form>
	</div>
	`

	return renderAdminPage(c, "添加云黑", content)
}

func adminAddPost(c echo.Context) error {
	qq := c.FormValue("qq")
	nickname := c.FormValue("nickname")
	reason := c.FormValue("reason")
	severity := c.FormValue("severity")
	subjectName := c.FormValue("subject_name")

	if qq == "" || reason == "" {
		return c.Redirect(302, "/admin/add?error=请填写QQ号和原因")
	}

	if !ValidateQQ(qq) {
		return c.Redirect(302, "/admin/add?error=QQ号格式不正确")
	}
	if len(nickname) > 50 {
		return c.Redirect(302, "/admin/add?error=昵称不能超过50个字符")
	}
	if len(reason) > 2000 {
		return c.Redirect(302, "/admin/add?error=云黑原因不能超过2000个字符")
	}
	if len(subjectName) > 100 {
		return c.Redirect(302, "/admin/add?error=主体名称不能超过100个字符")
	}
	qqNum64, _ := strconv.ParseInt(qq, 10, 64)
	qqNum := int(qqNum64)

	var cnt int
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE qq = ?", qqNum).Scan(&cnt)
	if cnt > 0 {
		return c.Redirect(302, "/admin/add?error=该QQ号已在云黑名单中")
	}
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE qq = ? AND status = 0", qqNum).Scan(&cnt)
	if cnt > 0 {
		return c.Redirect(302, "/admin/add?error=该QQ号已提交待审核")
	}

	severityNum := parseSeverity(severity)

	res, err := DB.Exec("INSERT INTO cloudblack_records (qq, nickname, reason, severity, submitter_id, status, created_at) VALUES (?, ?, ?, ?, 0, 0, datetime('now'))",
		qqNum, nickname, reason, severityNum)
	if err != nil {
		return c.Redirect(302, "/admin/add?error=添加失败")
	}
	recordID64, _ := res.LastInsertId()
	LogAccess("admin_add", int64(qqNum), "admin", "", getAdminID(c), c)

	// 异步 AI 离线审核
	go func() {
		PerformOfflineReview(int(recordID64), int64(qqNum), reason, "", "", nickname, severityNum)
	}()

	return c.Redirect(302, "/admin/review")
}

func adminEdit(c echo.Context) error {
	id := c.QueryParam("id")
	if id == "" {
		return c.Redirect(302, "/admin/list")
	}

	var r struct {
		ID       int
		QQ       int64
		Nickname string
		Reason   string
		Severity int
	}
	DB.QueryRow("SELECT id, qq, nickname, reason, severity FROM cloudblack_list WHERE id = ?", id).Scan(&r.ID, &r.QQ, &r.Nickname, &r.Reason, &r.Severity)

	selected := func(n, s int) string {
		if n == s {
			return " selected"
		}
		return ""
	}

	content := `
	<div class="card">
	<h2>编辑云黑</h2>
	<form method="POST" action="/admin/edit">
		<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
		<input type="hidden" name="id" value="` + id + `">
		<div class="form-group"><label>QQ号</label><input type="text" name="qq" value="` + strconv.FormatInt(r.QQ, 10) + `" required></div>
		<div class="form-group"><label>昵称</label><input type="text" name="nickname" value="` + esc(r.Nickname) + `"></div>
		<div class="form-group"><label>原因</label><textarea name="reason" required>` + esc(r.Reason) + `</textarea></div>
		<div class="form-group"><label>严重程度</label>
		<select name="severity">
			<option value="1"` + selected(1, r.Severity) + `>轻微</option>
			<option value="2"` + selected(2, r.Severity) + `>一般</option>
			<option value="3"` + selected(3, r.Severity) + `>较重</option>
			<option value="4"` + selected(4, r.Severity) + `>严重</option>
			<option value="5"` + selected(5, r.Severity) + `>极其严重</option>
		</select>
		</div>
		<button type="submit" class="btn">保存</button>
	</form>
	</div>
	`

	return renderAdminPage(c, "编辑云黑", content)
}

func adminEditPost(c echo.Context) error {
	id := c.FormValue("id")
	qq := c.FormValue("qq")
	nickname := c.FormValue("nickname")
	reason := c.FormValue("reason")
	severity := c.FormValue("severity")

	if !ValidateQQ(qq) {
		return c.Redirect(302, "/admin/list?error=QQ号格式不正确")
	}
	if len(nickname) > 50 {
		return c.Redirect(302, "/admin/edit?id="+id+"&error=昵称不能超过50个字符")
	}
	if len(reason) > 2000 {
		return c.Redirect(302, "/admin/edit?id="+id+"&error=云黑原因不能超过2000个字符")
	}
	qqNum64, _ := strconv.ParseInt(qq, 10, 64)
	qqNum := int(qqNum64)
	severityNum := parseSeverity(severity)

	DB.Exec("UPDATE cloudblack_list SET qq = ?, nickname = ?, reason = ?, severity = ? WHERE id = ?",
		qqNum, nickname, reason, severityNum, id)

	return c.Redirect(302, "/admin/list")
}

func adminDelete(c echo.Context) error {
	id := c.FormValue("id")
	if id == "" {
		return c.Redirect(302, "/admin/list?error=缺少ID")
	}
	DB.Exec("DELETE FROM cloudblack_list WHERE id = ?", id)
	return c.Redirect(302, "/admin/list")
}

func adminReviewAction(c echo.Context) error {
	id := c.FormValue("id")
	action := c.FormValue("action")
	note := c.FormValue("review_note")
	if err := performReviewAction(id, action, getAdminID(c), note); err != nil {
		return c.Redirect(302, "/admin/review?error=操作失败，请重试")
	}
	qqNum, _ := strconv.ParseInt(id, 10, 64)
	LogAccess("review_"+action, qqNum, "admin", "", getAdminID(c), c)
	return c.Redirect(302, "/admin/review")
}

func performReviewAction(id, action string, reviewerID int, note string) error {
	if id == "" || (action != "approve" && action != "reject") {
		return fmt.Errorf("参数错误")
	}
	note = strings.TrimSpace(note)
	if action == "reject" && note == "" {
		return fmt.Errorf("拒绝时请填写原因")
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("数据库错误")
	}

	var qq int64
	var subjectName, accountsRaw string
	tx.QueryRow("SELECT qq, COALESCE(subject_name,''), COALESCE(accounts,'') FROM cloudblack_records WHERE id = ?", id).Scan(&qq, &subjectName, &accountsRaw)
	if qq > 0 {
		var cnt int
		tx.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE qq = ?", qq).Scan(&cnt)
		if cnt > 0 {
			tx.Rollback()
			return fmt.Errorf("该QQ号已在黑名单中")
		}
	}

	if action == "approve" {
		if subjectName == "" {
			tx.QueryRow("SELECT COALESCE(nickname,'') FROM cloudblack_records WHERE id = ?", id).Scan(&subjectName)
		}
		res, err := tx.Exec("INSERT INTO black_subjects (display_name, created_at) VALUES (?, datetime('now'))", subjectName)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("审批失败")
		}
		subjectID, _ := res.LastInsertId()

		accounts := DecodeAccounts(accountsRaw)
		if len(accounts) == 0 && qq > 0 {
			accounts = append(accounts, LinkedAccount{Platform: "QQ", Account: strconv.FormatInt(qq, 10), Nickname: subjectName})
		}
		for _, account := range accounts {
			tx.Exec("INSERT INTO subject_accounts (subject_id, platform, account, nickname, created_at) VALUES (?, ?, ?, ?, datetime('now'))", subjectID, account.Platform, account.Account, account.Nickname)
		}

		_, err = tx.Exec("INSERT INTO cloudblack_list (qq, nickname, reason, severity, submitter_id, status, subject_id, subject_name, tags, accounts, reviewed_by, review_note, created_at, reviewed_at) SELECT qq, nickname, reason, severity, submitter_id, 1, ?, subject_name, tags, accounts, ?, ?, datetime('now'), datetime('now') FROM cloudblack_records WHERE id = ?", subjectID, reviewerID, note, id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("审批失败")
		}
		_, err = tx.Exec("UPDATE cloudblack_records SET status = 1, reviewed_by = ?, reviewed_at = datetime('now'), review_note = ? WHERE id = ?", reviewerID, note, id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("审批失败")
		}
	} else if action == "reject" {
		if _, err := tx.Exec("UPDATE cloudblack_records SET status = 2, reviewed_by = ?, reviewed_at = datetime('now'), reject_reason = ? WHERE id = ?", reviewerID, note, id); err != nil {
			tx.Rollback()
			return fmt.Errorf("拒绝失败")
		}
	}

	return tx.Commit()
}

func adminStats(c echo.Context) error {
	var total, today, queryToday int
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE status = 1").Scan(&total)
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE DATE(created_at) = DATE('now')").Scan(&today)
	DB.QueryRow("SELECT COUNT(*) FROM stats_log WHERE type = 'query' AND DATE(created_at) = DATE('now')").Scan(&queryToday)

	content := `
	<div class="stats">
		<div class="stat-box"><div class="num">` + strconv.Itoa(total) + `</div><div class="label">总云黑数</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(today) + `</div><div class="label">今日提交</div></div>
		<div class="stat-box"><div class="num">` + strconv.Itoa(queryToday) + `</div><div class="label">今日查询</div></div>
	</div>
	`

	return renderAdminPage(c, "统计", content)
}

func adminAdmins(c echo.Context) error {
	rows, err := DB.Query("SELECT id, username, nickname, role, last_login FROM admins")
	if err != nil {
		return renderAdminPage(c, "管理员", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	var admins []map[string]interface{}
	for rows.Next() {
		var a struct {
			ID        int
			Username  string
			Nickname  string
			Role      int
			LastLogin string
		}
		rows.Scan(&a.ID, &a.Username, &a.Nickname, &a.Role, &a.LastLogin)
		admins = append(admins, map[string]interface{}{
			"id":         a.ID,
			"username":   a.Username,
			"nickname":   a.Nickname,
			"role":       a.Role,
			"last_login": a.LastLogin,
		})
	}

	content := `
	<div class="card">
	<h2>添加管理员</h2>
	<form method="POST" action="/admin/admins">
		<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
		<div class="form-group"><label>用户名</label><input type="text" name="username" required></div>
		<div class="form-group"><label>密码</label><input type="password" name="password" required></div>
		<div class="form-group"><label>昵称</label><input type="text" name="nickname"></div>
		<button type="submit" class="btn">添加</button>
	</form>
	</div>
	<div class="card">
	<h2>管理员列表</h2>
	<table><thead><tr><th>ID</th><th>用户名</th><th>昵称</th><th>最后登录</th><th>操作</th></tr></thead><tbody>
	`
	for _, a := range admins {
		content += `<tr><td>` + strconv.Itoa(a["id"].(int)) + `</td><td>` + esc(a["username"].(string)) + `</td><td>` + esc(a["nickname"].(string)) + `</td><td>` + a["last_login"].(string) + `</td><td><form method="POST" action="/admin/delete_admin" style="display:inline"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="id" value="` + strconv.Itoa(a["id"].(int)) + `"><button type="submit" class="btn btn-danger" onclick="return confirm('确认删除?')">删除</button></form></td></tr>`
	}
	content += `</tbody></table></div>`

	return renderAdminPage(c, "管理员", content)
}

func adminAdminsPost(c echo.Context) error {
	username := c.FormValue("username")
	password := c.FormValue("password")
	nickname := c.FormValue("nickname")

	hash, _ := HashPassword(password)

	DB.Exec("INSERT INTO admins (username, password, nickname, role, must_change_password, created_at) VALUES (?, ?, ?, 1, 0, datetime('now'))",
		username, hash, nickname)

	return c.Redirect(302, "/admin/admins")
}

func adminDeleteAdmin(c echo.Context) error {
	id := c.FormValue("id")
	if id == "" {
		return c.Redirect(302, "/admin/admins?error=缺少ID")
	}
	currentID := getAdminID(c)
	if strconv.Itoa(currentID) == id {
		return c.Redirect(302, "/admin/admins?error=不能删除自己")
	}
	DB.Exec("DELETE FROM admins WHERE id = ?", id)
	return c.Redirect(302, "/admin/admins")
}

func adminAPIKeys(c echo.Context) error {
	rows, err := DB.Query("SELECT id, api_key, admin_id, permissions, status, created_at FROM api_keys")
	if err != nil {
		return renderAdminPage(c, "API密钥", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	var keys []map[string]interface{}
	for rows.Next() {
		var k struct {
			ID          int
			APIKey      string
			AdminID     int
			Permissions string
			Status      int
			CreatedAt   string
		}
		rows.Scan(&k.ID, &k.APIKey, &k.AdminID, &k.Permissions, &k.Status, &k.CreatedAt)
		keys = append(keys, map[string]interface{}{
			"id":          k.ID,
			"api_key":     k.APIKey,
			"admin_id":    k.AdminID,
			"permissions": k.Permissions,
			"status":      k.Status,
			"created_at":  k.CreatedAt,
		})
	}

	content := `
	<div class="card">
	<h2>创建新的API密钥</h2>
	<form method="POST" action="/admin/apikeys">
		<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
		<div class="form-group">
			<label>选择权限</label>
			<div style="margin-top:10px">
				<label><input type="checkbox" name="permissions" value="submit" checked style="width:auto"> 提交云黑</label>
				<label><input type="checkbox" name="permissions" value="query" checked style="width:auto;margin-left:20px"> 查询云黑</label>
				<label><input type="checkbox" name="permissions" value="review" style="width:auto;margin-left:20px"> 审核管理</label>
			</div>
		</div>
		<button type="submit" class="btn">生成密钥</button>
	</form>
	</div>
	<div class="card">
	<h2>已有API密钥列表</h2>
	<table><thead><tr><th>ID</th><th>密钥</th><th>权限</th><th>状态</th><th>创建时间</th><th>操作</th></tr></thead><tbody>
	`
	for _, k := range keys {
		statusText := `<span class="badge badge-success">启用</span>`
		btnText := "禁用"
		if k["status"].(int) == 0 {
			statusText = `<span class="badge badge-danger">禁用</span>`
			btnText = "启用"
		}
		content += `<tr><td>` + strconv.Itoa(k["id"].(int)) + `</td><td style="font-family:monospace;word-break:break-all">` + k["api_key"].(string) + `</td><td>` + k["permissions"].(string) + `</td><td>` + statusText + `</td><td>` + k["created_at"].(string) + `</td><td><form method="POST" action="/admin/toggle_apikey" style="display:inline"><input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `"><input type="hidden" name="id" value="` + strconv.Itoa(k["id"].(int)) + `"><button type="submit" class="btn">` + btnText + `</button></form></td></tr>`
	}
	content += `</tbody></table></div>`

	return renderAdminPage(c, "API密钥", content)
}

func adminAPIKeysPost(c echo.Context) error {
	form, _ := c.FormParams()
	permissionsList := form["permissions"]
	note := c.FormValue("note")
	adminID := getAdminID(c)

	permissions := strings.Join(permissionsList, ",")
	if permissions == "" {
		permissions = "query"
	}

	apiKey := RandomString(32)

	DB.Exec("INSERT INTO api_keys (api_key, admin_id, permissions, note, status, created_at) VALUES (?, ?, ?, ?, 1, datetime('now'))",
		apiKey, adminID, permissions, note)

	return c.Redirect(302, "/admin/apikeys")
}

func adminToggleAPIKey(c echo.Context) error {
	id := c.FormValue("id")
	DB.Exec("UPDATE api_keys SET status = CASE WHEN status = 1 THEN 0 ELSE 1 END WHERE id = ?", id)
	return c.Redirect(302, "/admin/apikeys")
}

func adminSettings(c echo.Context) error {
	queryRPM := GetSettingInt("public_query_rpm", 30)
	submitRPM := GetSettingInt("public_submit_rpm", 5)
	cooldown := GetSettingInt("submit_cooldown", 30)
	minReason := GetSettingInt("submit_min_reason", 10)
	maxHour := GetSettingInt("submit_max_hour", 200)
	feedbackEmail := GetSetting("feedback_email", "")
	successMsg := c.QueryParam("success")
	content := `<div class="card"><h2>系统设置</h2>`
	if successMsg != "" {
		content += `<div class="success">` + esc(successMsg) + `</div>`
	}
	content += `<form method="POST" action="/admin/settings">
		<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
	<div class="form-group"><label>公开查询 RPM</label><input type="number" name="public_query_rpm" min="0" value="` + strconv.Itoa(queryRPM) + `"><p>无 API Key 调用 /api/v1/query、/api/v1/check 的每分钟限制，0 表示关闭公开查询。</p></div>
	<div class="form-group"><label>公开提交 RPM</label><input type="number" name="public_submit_rpm" min="0" value="` + strconv.Itoa(submitRPM) + `"><p>无 API Key 调用 /api/v1/submit 的每分钟限制，0 表示关闭公开提交。</p></div>
	<h2 style="margin-top:24px">提交风控</h2>
	<div class="form-group"><label>同号冷却期（分钟）</label><input type="number" name="submit_cooldown" min="0" value="` + strconv.Itoa(cooldown) + `"><p>同一 IP 对同一账号在冷却期内不可重复提交，0 表示关闭。</p></div>
	<div class="form-group"><label>原因最低字数</label><input type="number" name="submit_min_reason" min="0" value="` + strconv.Itoa(minReason) + `"><p>提交原因不足此字数者拒绝，同时过滤纯数字/符号/重复字符等垃圾内容，0 表示关闭。</p></div>
	<div class="form-group"><label>全局每小时提交上限</label><input type="number" name="submit_max_hour" min="0" value="` + strconv.Itoa(maxHour) + `"><p>整个系统每小时最多接受多少条提交，0 表示不限制。</p></div>
	<h2 style="margin-top:24px">其他</h2>
	<div class="form-group"><label>访问日志</label><select name="enable_access_log"><option value="0"` + func() string { if GetSetting("enable_access_log", "0") == "0" { return " selected" }; return "" }() + `>关闭</option><option value="1"` + func() string { if GetSetting("enable_access_log", "0") == "1" { return " selected" }; return "" }() + `>开启</option></select><p>开启后记录所有查询和提交操作，包括来源、IP、API密钥等信息。</p></div>
	<div class="form-group"><label>反馈邮箱</label><input type="text" name="feedback_email" placeholder="admin@example.com" value="` + esc(feedbackEmail) + `"><p>设置后在查询页面显示联系链接，不设置则不显示。</p></div>
	<button type="submit" class="btn">保存设置</button>
	</form></div>`
	return renderAdminPage(c, "系统设置", content)
}

func adminSettingsPost(c echo.Context) error {
	queryRPM := c.FormValue("public_query_rpm")
	submitRPM := c.FormValue("public_submit_rpm")
	cooldown := c.FormValue("submit_cooldown")
	minReason := c.FormValue("submit_min_reason")
	maxHour := c.FormValue("submit_max_hour")
	if queryRPM == "" {
		queryRPM = "30"
	}
	if submitRPM == "" {
		submitRPM = "5"
	}
	if cooldown == "" {
		cooldown = "30"
	}
	if minReason == "" {
		minReason = "10"
	}
	if maxHour == "" {
		maxHour = "200"
	}
	SetSetting("public_query_rpm", queryRPM)
	SetSetting("public_submit_rpm", submitRPM)
	SetSetting("submit_cooldown", cooldown)
	SetSetting("submit_min_reason", minReason)
	SetSetting("submit_max_hour", maxHour)
	SetSetting("feedback_email", c.FormValue("feedback_email"))
	SetSetting("enable_access_log", c.FormValue("enable_access_log"))
	return c.Redirect(302, "/admin/settings?success=保存成功")
}

func adminAISettings(c echo.Context) error {
	s := GetAISettings()
	errorMsg := c.QueryParam("error")
	successMsg := c.QueryParam("success")

	content := `<div class="card"><h2>AI 离线审核设置</h2>`
	if errorMsg != "" {
		content += `<div class="error">` + esc(errorMsg) + `</div>`
	}
	if successMsg != "" {
		content += `<div class="success">` + esc(successMsg) + `</div>`
	}

	content += `<form method="POST" action="/admin/ai_settings">
	<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
	<h3>离线模式</h3>
	<div class="form-group"><label>开启离线 AI 审核</label><select name="enable_offline_ai"><option value="0"` + func() string { if s.EnableOfflineAI == 0 { return " selected" }; return "" }() + `>关闭</option><option value="1"` + func() string { if s.EnableOfflineAI == 1 { return " selected" }; return "" }() + `>开启</option></select><p>开启后，在设定时段内提交的记录会自动触发 AI 分析。</p></div>
	<div class="form-group"><label>自动启用时段</label><div style="display:flex;gap:10px"><input type="time" name="offline_start" value="` + esc(s.OfflineStart) + `" style="width:120px"><span style="color:#6b7280">至</span><input type="time" name="offline_end" value="` + esc(s.OfflineEnd) + `" style="width:120px"></div><p>例如 23:00 至 08:00 表示夜间无人值班时段。</p></div>
	<h3>决策阈值</h3>
	<div class="form-group"><label>自动拒绝阈值（AI 评分 &le;）</label><input type="number" name="auto_reject_confidence" min="0" max="100" value="` + strconv.Itoa(s.AutoRejectConfidence) + `" style="width:100px"><p>AI 评分低于此值自动拒绝（垃圾/广告/灌水），默认 20。</p></div>
	<div class="form-group"><label>自动通过阈值（AI 评分 &ge;）</label><input type="number" name="auto_approve_confidence" min="0" max="100" value="` + strconv.Itoa(s.AutoApproveConfidence) + `" style="width:100px"><p>AI 评分高于此值且行为证据充分才自动通过，默认 85。</p></div>
	<div class="form-group"><label>多人举报最低 IP 数</label><input type="number" name="min_ip_count" min="1" max="20" value="` + strconv.Itoa(s.MinIPCount) + `" style="width:100px"><p>72 小时内该 QQ 被多少个独立 IP 提交才算"多人举报"，默认 3。</p></div>
	<div class="form-group"><label>高查询热度阈值</label><input type="number" name="min_query_count" min="0" max="9999" value="` + strconv.Itoa(s.MinQueryCount) + `" style="width:100px"><p>30 天内被查询多少次算"高关注度"，默认 50。</p></div>
	<h3>API 配置</h3>
	<div class="form-group"><label>API 提供商</label><select name="api_provider"><option value="openai"` + func() string { if s.APIProvider == "openai" { return " selected" }; return "" }() + `>OpenAI</option><option value="deepseek"` + func() string { if s.APIProvider == "deepseek" { return " selected" }; return "" }() + `>DeepSeek</option><option value="custom"` + func() string { if s.APIProvider == "custom" { return " selected" }; return "" }() + `>自定义</option></select></div>
	<div class="form-group"><label>自定义接口地址</label><input type="text" name="api_endpoint" placeholder="https://api.example.com/v1" value="` + esc(s.APIEndpoint) + `"><p>留空使用官方地址，自定义需填写完整地址（含 /v1）。</p></div>
	<div class="form-group"><label>API 密钥</label><input type="password" name="api_key" placeholder="sk-..." value="` + esc(s.APIKey) + `"><p>密钥仅存储在本地数据库，不会外传。</p></div>
	<div class="form-group"><label>模型</label><div style="display:flex;gap:8px"><select id="modelSelect" name="api_model" style="flex:1"><option value="` + esc(s.APIModel) + `" selected>` + esc(s.APIModel) + `</option></select><input type="text" id="modelInput" name="api_model" value="` + esc(s.APIModel) + `" style="flex:1;display:none" placeholder="手动输入"><button type="button" class="btn btn-sm" onclick="refreshModels()">刷新列表</button></div><p id="modelHint" style="color:#6b7280;font-size:13px">填写 API 地址和密钥后可刷新获取模型列表。</p></div>
	<button type="submit" class="btn">保存设置</button>
	</form></div>
	<script>
	function refreshModels(){
		var btn=document.querySelector('[onclick="refreshModels()"]');
		btn.textContent='获取中...';
		btn.disabled=true;
		fetch('/admin/api/models').then(r=>r.json()).then(res=>{
			btn.textContent='刷新列表';
			btn.disabled=false;
			var sel=document.getElementById('modelSelect');
			var inp=document.getElementById('modelInput');
			if(res.models&&res.models.length>0){
				sel.innerHTML='';
				res.models.forEach(function(m){
					var opt=document.createElement('option');
					opt.value=m;opt.textContent=m;
					if(m==="` + esc(s.APIModel) + `")opt.selected=true;
					sel.appendChild(opt);
				});
				var manual=document.createElement('option');
				manual.value='__manual__';manual.textContent='-- 手动输入 --';
				sel.appendChild(manual);
				sel.style.display='block';
				inp.style.display='none';
				document.getElementById('modelHint').textContent='获取成功，共'+res.models.length+'个模型';
				document.getElementById('modelHint').style.color='#16a34a';
			}else{
				sel.style.display='none';
				inp.style.display='block';
				document.getElementById('modelHint').textContent='该 API 不支持获取模型列表，请手动输入';
				document.getElementById('modelHint').style.color='#b91c1c';
			}
		}).catch(function(){
			btn.textContent='刷新列表';
			btn.disabled=false;
			document.getElementById('modelSelect').style.display='none';
			document.getElementById('modelInput').style.display='block';
			document.getElementById('modelHint').textContent='获取失败，请检查 API 地址和密钥';
			document.getElementById('modelHint').style.color='#b91c1c';
		});
	}
	document.getElementById('modelSelect').addEventListener('change',function(){
		if(this.value==='__manual__'){
			document.getElementById('modelSelect').style.display='none';
			document.getElementById('modelInput').style.display='block';
			document.getElementById('modelInput').name='api_model';
			document.getElementById('modelSelect').name='api_model_unused';
		}else{
			document.getElementById('modelSelect').name='api_model';
			document.getElementById('modelInput').name='api_model_unused';
		}
	});
	if(document.querySelector('[name="api_key"]').value&&document.querySelector('[name="api_endpoint"]').value){
		refreshModels();
	}
	</script>`

	return renderAdminPage(c, "AI 设置", content)
}

func adminAISettingsPost(c echo.Context) error {
	enable := c.FormValue("enable_offline_ai")
	if enable == "" {
		enable = "0"
	}
	start := c.FormValue("offline_start")
	end := c.FormValue("offline_end")
	rejectConf := c.FormValue("auto_reject_confidence")
	approveConf := c.FormValue("auto_approve_confidence")
	minIP := c.FormValue("min_ip_count")
	minQuery := c.FormValue("min_query_count")
	provider := c.FormValue("api_provider")
	endpoint := c.FormValue("api_endpoint")
	apiKey := c.FormValue("api_key")
	model := c.FormValue("api_model")

	if rejectConf == "" {
		rejectConf = "20"
	}
	if approveConf == "" {
		approveConf = "85"
	}
	if minIP == "" {
		minIP = "3"
	}
	if minQuery == "" {
		minQuery = "50"
	}

	DB.Exec(`UPDATE ai_settings SET enable_offline_ai = ?, offline_start = ?, offline_end = ?, auto_reject_confidence = ?, auto_approve_confidence = ?, min_ip_count = ?, min_query_count = ?, api_provider = ?, api_endpoint = ?, api_key = ?, api_model = ?, updated_at = datetime('now') WHERE id = 1`,
		enable, start, end, rejectConf, approveConf, minIP, minQuery, provider, endpoint, apiKey, model)

	return c.Redirect(302, "/admin/ai_settings?success=保存成功")
}

func adminAPIModels(c echo.Context) error {
	s := GetAISettings()
	models, err := GetModelsFromAPI(s.APIEndpoint, s.APIKey)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]interface{}{"models": []string{}, "error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"models": models})
}

func adminAIReviewLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 50
	offset := (page - 1) * pageSize

	aiResultFilter := c.QueryParam("ai_result")
	finalStatus := c.QueryParam("final_status")
	qqFilter := c.QueryParam("qq")

	where := "WHERE 1=1"
	var args []interface{}
	if aiResultFilter != "" {
		where += " AND ai_result = ?"
		args = append(args, aiResultFilter)
	}
	if finalStatus != "" {
		where += " AND final_status = ?"
		args = append(args, finalStatus)
	}
	if qqFilter != "" {
		where += " AND qq = ?"
		args = append(args, qqFilter)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	DB.QueryRow("SELECT COUNT(*) FROM ai_review_logs "+where, countArgs...).Scan(&total)

	queryArgs := append(args, pageSize, offset)
	rows, err := DB.Query("SELECT id, record_id, action, ai_result, ai_score, ai_reason, behavior_ip_count, behavior_query_count, final_status, created_at FROM ai_review_logs "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...)
	if err != nil {
		return renderAdminPage(c, "AI 离线记录", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	content := `<div class="card"><h2>AI 离线审核记录</h2>`
	content += `<form method="GET" action="/admin/ai_review_logs" style="display:flex;gap:10px;flex-wrap:wrap;margin-bottom:16px">`
	content += `<select name="ai_result" style="padding:8px;border:1px solid #d6d9df;border-radius:8px;font-size:13px"><option value="">全部决策</option>`
	for _, r := range []string{"auto_approve", "auto_reject", "manual_review"} {
		selected := ""
		if aiResultFilter == r {
			selected = " selected"
		}
		content += `<option value="` + r + `"` + selected + `>` + r + `</option>`
	}
	content += `</select>`
	content += `<select name="final_status" style="padding:8px;border:1px solid #d6d9df;border-radius:8px;font-size:13px"><option value="">全部状态</option>`
	for _, s := range []string{"pending", "confirmed", "corrected"} {
		selected := ""
		if finalStatus == s {
			selected = " selected"
		}
		content += `<option value="` + s + `"` + selected + `>` + s + `</option>`
	}
	content += `</select>`
	content += `<input type="text" name="qq" placeholder="QQ号" value="` + esc(qqFilter) + `" style="padding:8px;border:1px solid #d6d9df;border-radius:8px;font-size:13px;width:120px">`
	content += `<button class="btn btn-sm" type="submit">筛选</button>`
	content += `</form>`

	content += `<table><thead><tr><th>ID</th><th>QQ</th><th>决策</th><th>AI 评分</th><th>行为数据</th><th>AI 理由</th><th>状态</th><th>时间</th></tr></thead><tbody>`
	for rows.Next() {
		var id, recordID, aiScore, ipCount, queryCount int
		var action, aiResult, aiReason, status, createdAt string
		rows.Scan(&id, &recordID, &action, &aiResult, &aiScore, &aiReason, &ipCount, &queryCount, &status, &createdAt)

		var qqStr string
		var qq int64
		DB.QueryRow("SELECT qq FROM cloudblack_records WHERE id = ?", recordID).Scan(&qq)
		if qq > 0 {
			qqStr = strconv.FormatInt(qq, 10)
		} else {
			qqStr = "-"
		}

		resultClass := ""
		if aiResult == "auto_approve" {
			resultClass = `<span style="color:#16a34a;font-weight:700">自动通过</span>`
		} else if aiResult == "auto_reject" {
			resultClass = `<span style="color:#dc2626;font-weight:700">自动拒绝</span>`
		} else {
			resultClass = `<span style="color:#6b7280">转人工</span>`
		}

		statusClass := status
		if status == "confirmed" {
			statusClass = `<span style="color:#16a34a">已确认</span>`
		} else if status == "corrected" {
			statusClass = `<span style="color:#dc2626">已纠正</span>`
		} else {
			statusClass = `<span style="color:#f59e0b">待确认</span>`
		}

		content += `<tr><td>` + strconv.Itoa(id) + `</td><td>` + qqStr + `</td><td>` + resultClass + `</td><td>` + strconv.Itoa(aiScore) + `</td><td>` + strconv.Itoa(ipCount) + `IP / 查询` + strconv.Itoa(queryCount) + `</td><td>` + esc(aiReason) + `</td><td>` + statusClass + `</td><td>` + esc(createdAt) + `</td></tr>`
	}
	content += `</tbody></table>`

	maxPage := (total + pageSize - 1) / pageSize
	if maxPage < 1 {
		maxPage = 1
	}
	content += `<div class="pagination">`
	for i := 1; i <= maxPage; i++ {
		active := ""
		if i == page {
			active = ` class="active"`
		}
		content += `<a` + active + ` href="/admin/ai_review_logs?page=` + strconv.Itoa(i) + `&ai_result=` + esc(aiResultFilter) + `&final_status=` + esc(finalStatus) + `&qq=` + esc(qqFilter) + `">` + strconv.Itoa(i) + `</a>`
	}
	content += `</div></div>`

	return renderAdminPage(c, "AI 离线记录", content)
}

func adminLogs(c echo.Context) error {
	rows, err := DB.Query("SELECT id, admin_id, action, detail, ip, created_at FROM admin_logs ORDER BY id DESC LIMIT 50")
	if err != nil {
		return renderAdminPage(c, "操作日志", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	content := `<div class="card"><table><thead><tr><th>ID</th><th>管理员ID</th><th>操作</th><th>详情</th><th>IP</th><th>时间</th></tr></thead><tbody>`
	for rows.Next() {
		var id, adminID int
		var action, detail, ip, createdAt string
		rows.Scan(&id, &adminID, &action, &detail, &ip, &createdAt)
		content += `<tr><td>` + strconv.Itoa(id) + `</td><td>` + strconv.Itoa(adminID) + `</td><td>` + esc(action) + `</td><td>` + esc(detail) + `</td><td>` + esc(ip) + `</td><td>` + esc(createdAt) + `</td></tr>`
	}
	content += `</tbody></table></div>`

	return renderAdminPage(c, "操作日志", content)
}

func adminAccessLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	pageSize := 50
	offset := (page - 1) * pageSize

	actionFilter := c.QueryParam("action")
	sourceFilter := c.QueryParam("source")
	qqFilter := c.QueryParam("qq")

	where := "WHERE 1=1"
	var args []interface{}
	if actionFilter != "" {
		where += " AND action = ?"
		args = append(args, actionFilter)
	}
	if sourceFilter != "" {
		where += " AND source = ?"
		args = append(args, sourceFilter)
	}
	if qqFilter != "" {
		where += " AND qq = ?"
		args = append(args, qqFilter)
	}

	var total int
	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)
	DB.QueryRow("SELECT COUNT(*) FROM access_logs "+where, countArgs...).Scan(&total)

	queryArgs := append(args, pageSize, offset)
	rows, err := DB.Query("SELECT id, action, qq, source, api_key, admin_id, ip, user_agent, created_at FROM access_logs "+where+" ORDER BY id DESC LIMIT ? OFFSET ?", queryArgs...)
	if err != nil {
		return renderAdminPage(c, "访问日志", `<div class="card"><p class="error">数据查询失败</p></div>`)
	}
	defer rows.Close()

	enabled := GetSetting("enable_access_log", "0") == "1"
	content := `<div class="card"><h2>访问日志` + func() string {
		if enabled {
			return ` <span style="color:#16a34a;font-size:13px;font-weight:600">已开启</span>`
		}
		return ` <span style="color:#6b7280;font-size:13px;font-weight:600">未开启</span>`
	}() + `</h2>`
	if !enabled {
		content += `<p style="color:#6b7280;font-size:13px">访问日志未开启，请在 <a href="/admin/settings">系统设置</a> 中开启。</p>`
	}

	content += `<form method="GET" action="/admin/access_logs" style="display:flex;gap:10px;flex-wrap:wrap;margin-bottom:16px">`
	content += `<select name="action" style="padding:8px;border:1px solid #d6d9df;border-radius:8px;font-size:13px"><option value="">全部操作</option>`
	for _, a := range []string{"query", "check", "batch_query", "submit", "review_approve", "review_reject", "admin_add"} {
		selected := ""
		if actionFilter == a {
			selected = " selected"
		}
		content += `<option value="` + a + `"` + selected + `>` + a + `</option>`
	}
	content += `</select>`
	content += `<select name="source" style="padding:8px;border:1px solid #d6d9df;border-radius:8px;font-size:13px"><option value="">全部来源</option>`
	for _, s := range []string{"api", "web", "admin"} {
		selected := ""
		if sourceFilter == s {
			selected = " selected"
		}
		content += `<option value="` + s + `"` + selected + `>` + s + `</option>`
	}
	content += `</select>`
	content += `<input type="text" name="qq" placeholder="QQ号" value="` + esc(qqFilter) + `" style="padding:8px;border:1px solid #d6d9df;border-radius:8px;font-size:13px;width:120px">`
	content += `<button class="btn" type="submit">筛选</button>`
	content += `</form>`

	content += `<table><thead><tr><th>ID</th><th>操作</th><th>QQ</th><th>来源</th><th>API密钥</th><th>管理员</th><th>IP</th><th>User-Agent</th><th>时间</th></tr></thead><tbody>`
	for rows.Next() {
		var id, adminID int
		var qq int64
		var action, source, apiKey, ip, userAgent, createdAt string
		rows.Scan(&id, &action, &qq, &source, &apiKey, &adminID, &ip, &userAgent, &createdAt)
		qqStr := "-"
		if qq > 0 {
			qqStr = strconv.FormatInt(qq, 10)
		}
		apiKeyStr := "-"
		if apiKey != "" && apiKey != "public" {
			apiKeyStr = apiKey[:8] + "..."
		}
		adminStr := "-"
		if adminID > 0 {
			adminStr = strconv.Itoa(adminID)
		}
		content += `<tr><td>` + strconv.Itoa(id) + `</td><td>` + esc(action) + `</td><td>` + qqStr + `</td><td>` + esc(source) + `</td><td>` + apiKeyStr + `</td><td>` + adminStr + `</td><td>` + esc(ip) + `</td><td style="max-width:200px;overflow:hidden;text-overflow:ellipsis;white-space:nowrap" title="` + esc(userAgent) + `">` + esc(userAgent) + `</td><td>` + esc(createdAt) + `</td></tr>`
	}
	content += `</tbody></table>`

	maxPage := (total + pageSize - 1) / pageSize
	if maxPage < 1 {
		maxPage = 1
	}
	content += `<div class="pagination">`
	for i := 1; i <= maxPage; i++ {
		active := ""
		if i == page {
			active = ` class="active"`
		}
		content += `<a` + active + ` href="/admin/access_logs?page=` + strconv.Itoa(i) + `&action=` + esc(actionFilter) + `&source=` + esc(sourceFilter) + `&qq=` + esc(qqFilter) + `">` + strconv.Itoa(i) + `</a>`
	}
	content += `</div></div>`

	return renderAdminPage(c, "访问日志", content)
}

func adminChangePassword(c echo.Context) error {
	errorMsg := c.QueryParam("error")

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>修改密码 - 云黑系统</title>
<style>
*{box-sizing:border-box}
body{margin:0;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;background:#f4f5f7;color:#16181d}
.container{max-width:520px;margin:56px auto;padding:20px}
.card{background:#fff;border-radius:16px;padding:24px;box-shadow:0 14px 35px rgba(17,19,24,.08);border:1px solid #e5e7eb}
h2{color:#111318;margin:0 0 20px;padding-bottom:14px;border-bottom:1px solid #e5e7eb;font-size:20px}
.form-group{margin-bottom:16px}
.form-group label{display:block;margin-bottom:7px;font-weight:700;color:#2d333d;font-size:14px}
.form-group input{width:100%;padding:11px 12px;border:1px solid #d6d9df;border-radius:10px;font:inherit;transition:border .18s,box-shadow .18s}
.form-group input:focus{outline:none;border-color:#d92d20;box-shadow:0 0 0 3px rgba(217,45,32,.12)}
.btn{width:100%;padding:12px;background:#111318;color:#fff;border:1px solid #111318;border-radius:10px;cursor:pointer;font-weight:800;font:inherit}
.btn:hover{background:#2a2f38}
.error{background:#fff1f0;color:#9f1c14;padding:12px;border-radius:10px;margin-bottom:15px;border:1px solid #ffc9c3;font-weight:600}
@media(max-width:560px){.container{margin:24px auto;padding:14px}.card{padding:18px;border-radius:14px}}
</style>
</head>
<body>
<div class="container">
<div class="card">
<h2>首次登录必须修改密码</h2>`
	if errorMsg != "" {
		html += `<div class="error">` + esc(errorMsg) + `</div>`
	}
	html += `<form method="POST">
	<input type="hidden" name="csrf_token" value="` + getCSRFToken(c) + `">
<div class="form-group"><label>新密码 (至少6位)</label><input type="password" name="new_password" required minlength="6"></div>
<div class="form-group"><label>确认新密码</label><input type="password" name="confirm_password" required minlength="6"></div>
<button type="submit" class="btn">修改密码</button>
</form>
</div>
</div>
</body>
</html>`
	c.HTML(http.StatusOK, html)
	return nil
}

func adminChangePasswordPost(c echo.Context) error {
	newPassword := c.FormValue("new_password")
	confirmPassword := c.FormValue("confirm_password")

	if len(newPassword) < 6 {
		return c.Redirect(302, "/admin/password?error=密码长度至少6位")
	}

	if newPassword != confirmPassword {
		return c.Redirect(302, "/admin/password?error=两次密码输入不一致")
	}

	adminID := getAdminID(c)
	hash, _ := HashPassword(newPassword)

	DB.Exec("UPDATE admins SET password = ?, must_change_password = 0 WHERE id = ?", hash, adminID)
	DB.Exec("DELETE FROM admin_sessions WHERE admin_id = ?", adminID)

	return c.Redirect(302, "/admin/")
}

func adminAPIDoc(c echo.Context) error {
	content := `
	<div class="card">
	<h2>API接口文档</h2>
	<p>API密钥可以通过以下方式传递：</p>
	<ul>
		<li>Header: <code>X-API-Key: your_api_key</code></li>
		<li>URL参数: <code>?api_key=your_api_key</code></li>
		<li>表单参数: <code>api_key=your_api_key</code></li>
	</ul>
	<p>公开查询、快速检查、批量查询和提交接口可以不带 API 密钥调用，公开访问 RPM 可在 <code>系统设置</code> 中配置。带 API 密钥调用时会校验密钥权限并使用 API 限流。</p>
	
	<h3>1. 提交云黑</h3>
	<div class="card" style="background:#f8f9fb">
		<p><span class="badge badge-success">POST</span> <code>/api/v1/submit</code></p>
		<p><strong>参数:</strong></p>
		<ul>
			<li>qq (必填) - QQ号</li>
			<li>nickname - 昵称</li>
			<li>reason (必填) - 云黑原因</li>
			<li>severity - 严重程度 1-5，默认1</li>
		</ul>
	</div>
	
	<h3>2. 查询云黑</h3>
	<div class="card" style="background:#f8f9fb">
		<p><span class="badge badge-warning">GET</span> <code>/api/v1/query?qq=123456789</code></p>
		<p><strong>参数:</strong></p>
		<ul>
			<li>qq (必填) - QQ号</li>
		</ul>
	</div>
	
	<h3>3. 批量查询</h3>
	<div class="card" style="background:#f8f9fb">
		<p><span class="badge badge-warning">GET</span> <code>/api/v1/batch?qq_list=123,456,789</code></p>
		<p><strong>参数:</strong></p>
		<ul>
			<li>qq_list - QQ号列表，用逗号分隔，最多100个</li>
		</ul>
	</div>
	
	<h3>4. 快速检查</h3>
	<div class="card" style="background:#f8f9fb">
		<p><span class="badge badge-warning">GET</span> <code>/api/v1/check?qq=123456789</code></p>
		<p><strong>参数:</strong></p>
		<ul>
			<li>qq (必填) - QQ号</li>
		</ul>
	</div>

	<h3>5. 获取审核列表</h3>
	<div class="card" style="background:#f8f9fb">
		<p><span class="badge badge-warning">GET</span> <code>/api/v1/review/list</code></p>
		<p>需要 API 密钥拥有 <code>review</code> 权限。</p>
	</div>

	<h3>6. 审核通过/拒绝</h3>
	<div class="card" style="background:#f8f9fb">
		<p><span class="badge badge-success">POST</span> <code>/api/v1/review/action</code></p>
		<p>需要 API 密钥拥有 <code>review</code> 权限。</p>
		<ul>
			<li>id (必填) - 待审核记录 ID</li>
			<li>action (必填) - <code>approve</code> 或 <code>reject</code></li>
		</ul>
	</div>
	
	<h3>响应示例</h3>
	<div class="card" style="background:#f8f9fb;font-family:monospace">
	{"code":200,"message":"success","data":{"id":1,"qq":123456,"nickname":"xxx","reason":"诈骗","severity":3,"severity_text":"较重","status":1,"status_text":"已通过"}}
	</div>
	</div>
	`

	return renderAdminPage(c, "API文档", content)
}
