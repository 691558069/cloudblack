package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func RegisterWebRoutes(e *echo.Group, cfg *Config) {
	e.Static("/web", "web")
	e.GET("/web/", webIndex)
	e.GET("/web/submit", webSubmit)
	e.GET("/web/api", webAPIDoc)

	webRL := NewRateLimiter(cfg.RateLimit.Web, time.Duration(cfg.RateLimit.Window)*time.Second)
	e.GET("/api/web/query", webQuery, RateLimitMiddleware(webRL, func(c echo.Context) string {
		return GetClientIP(c)
	}))
	e.POST("/api/web/submit", webSubmitPost, RateLimitMiddleware(webRL, func(c echo.Context) string {
		return GetClientIP(c)
	}))
}

func webIndex(c echo.Context) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>云黑查询</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#f5f7fb;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;min-height:100vh;color:#111827;line-height:1.5;overflow-x:hidden}
body:before{content:"";position:fixed;inset:0;background:radial-gradient(circle at 50% -120px,rgba(37,99,235,.10),transparent 460px);pointer-events:none;z-index:-2}
body:after{content:"";position:fixed;inset:0;background-image:linear-gradient(rgba(15,23,42,.035) 1px,transparent 1px),linear-gradient(90deg,rgba(15,23,42,.035) 1px,transparent 1px);background-size:44px 44px;mask-image:linear-gradient(to bottom,rgba(0,0,0,.45),transparent 70%);pointer-events:none;z-index:-1}
.container{max-width:760px;margin:0 auto;padding:56px 22px}
.hero{position:relative;background:rgba(255,255,255,.86);border:1px solid rgba(229,231,235,.9);border-radius:26px;padding:28px;box-shadow:0 24px 70px rgba(15,23,42,.10);backdrop-filter:blur(18px)}
.header{margin-bottom:24px;display:flex;justify-content:space-between;gap:18px;align-items:flex-end}
.header h1{font-size:38px;line-height:1.08;margin-bottom:10px;letter-spacing:-.045em;color:#111827;font-weight:900}
.header p{color:#6b7280;font-size:15px;max-width:420px}
.status-chip{display:inline-flex;align-items:center;gap:8px;padding:9px 12px;border:1px solid #e5e7eb;border-radius:999px;color:#374151;background:#f8fafc;font-size:13px;font-weight:800;white-space:nowrap}
.status-chip:before{content:"";width:8px;height:8px;border-radius:50%;background:#1f2937;box-shadow:0 0 0 5px rgba(31,41,55,.10)}
.tab-nav{display:flex;gap:6px;margin-bottom:20px;background:#f3f4f6;border:1px solid #e5e7eb;border-radius:14px;padding:5px}
.tab-nav a{flex:1;text-align:center;padding:12px 14px;text-decoration:none;color:#6b7280;border-radius:10px;font-weight:800;transition:background .2s,color .2s,transform .2s}
.tab-nav a.active{background:#111827;color:#fff}
.tab-nav a:not(.active):hover{background:#fff;color:#111827;transform:translateY(-1px)}
.search-box{display:grid;grid-template-columns:140px 1fr auto;gap:12px;margin-bottom:18px;background:#fff;border:1px solid #e5e7eb;border-radius:18px;padding:10px;box-shadow:0 16px 42px rgba(15,23,42,.08)}
.search-box input,.search-box select{min-width:0;padding:14px 16px;border:1px solid transparent;border-radius:12px;font-size:16px;background:#f9fafb;color:#111827;transition:border .2s,box-shadow .2s}
.search-box input:focus,.search-box select:focus{outline:none;border-color:#94a3b8;box-shadow:0 0 0 3px rgba(148,163,184,.22)}
.search-box input::placeholder{color:#9ca3af}
.search-box button{padding:14px 28px;background:#111827;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:800;cursor:pointer;transition:background .2s,transform .2s}
.search-box button:hover{background:#1f2937;transform:translateY(-1px)}
.result{background:#fff;border:1px solid #e5e7eb;border-radius:18px;padding:24px;margin-top:18px;box-shadow:0 16px 42px rgba(15,23,42,.08);animation:rise .22s ease-out}
.result.in-blacklist{border-color:rgba(220,38,38,.45);background:#fffafa}
.result.not-in-blacklist{border-color:rgba(22,163,74,.28);background:#fbfffc}
.result h2{font-size:19px;line-height:1.35;margin-bottom:16px;color:#111827}
.result .info{display:grid;grid-template-columns:96px 1fr;gap:8px 14px;color:#374151;padding:9px 0;border-top:1px solid #eef2f7}
.result .info-label{color:#6b7280;font-weight:700}
.tag-list{display:flex;flex-wrap:wrap;gap:6px}.tag{display:inline-flex;padding:3px 8px;border-radius:999px;background:#fee2e2;color:#991b1b;font-size:12px;font-weight:800}.account-list{display:grid;gap:6px}.account-item{padding:7px 9px;border:1px solid #e5e7eb;border-radius:10px;background:#f9fafb;color:#374151}
.tips{background:#fff;border:1px solid #e5e7eb;border-radius:18px;padding:22px;margin-top:18px;box-shadow:0 12px 30px rgba(15,23,42,.05)}
.tips h3{color:#111827;margin-bottom:10px;font-size:17px}
.tips ul{color:#6b7280;padding-left:18px;font-size:14px}
.tips li{margin-bottom:6px}
.quick-grid{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-top:18px}.quick-card{background:#fff;border:1px solid #e5e7eb;border-radius:14px;padding:14px;color:#6b7280;font-size:13px;box-shadow:0 10px 26px rgba(15,23,42,.05)}.quick-card strong{display:block;color:#111827;font-size:15px;margin-bottom:4px}
.error{background:#fff1f0;color:#b91c1c;padding:14px 15px;border-radius:12px;margin-top:15px;border:1px solid #fecaca}
.loading{text-align:center;padding:34px;margin-top:18px;color:#6b7280}
.spinner{width:42px;height:42px;border:3px solid #e5e7eb;border-top:3px solid #111827;border-radius:50%;animation:spin .8s linear infinite;margin:0 auto 14px}
@keyframes rise{from{opacity:0;transform:translateY(8px)}to{opacity:1;transform:translateY(0)}}
@keyframes spin{0%{transform:rotate(0deg)}100%{transform:rotate(360deg)}}
@media (max-width:720px){.container{padding:28px 14px}.hero{padding:18px;border-radius:20px}.header{display:block}.status-chip{margin-top:14px}.header h1{font-size:30px}.quick-grid{grid-template-columns:1fr}.search-box{grid-template-columns:1fr;padding:9px}.search-box input,.search-box select{width:100%}.search-box button{width:100%;padding:14px}.result,.tips{padding:18px;border-radius:16px}.result .info{grid-template-columns:1fr;gap:2px}.tab-nav a{padding:11px 10px}}
</style>
</head>
<body>
<div class="container">
<div class="hero">
<div class="header">
<div>
<h1>云黑查询</h1>
<p>查询QQ是否在云黑名单中</p>
</div>
<div class="status-chip">实时查询</div>
</div>
<div class="tab-nav">
<a href="/web/" class="active">查询</a>
<a href="/web/submit">提交</a>
<a href="/web/api">API</a>
</div>
<div class="search-box">
<select id="platformInput"><option value="QQ">QQ</option><option value="微信">微信</option><option value="Telegram">Telegram</option><option value="抖音">抖音</option><option value="其他">其他</option></select>
<input type="text" id="qqInput" placeholder="请输入账号" maxlength="64">
<button onclick="search()">查询</button>
</div>
<div id="loading" class="loading" style="display:none;">
<div class="spinner"></div>
<p>正在查询...</p>
</div>
<div id="result"></div>
<div class="tips">
<h3>使用说明</h3>
<ul>
<li>在上方输入QQ号并点击查询</li>
<li>系统会显示平台是否已收录该QQ的云黑记录</li>
<li>如已在云黑，会显示详细信息和严重程度</li>
<li>暂未收录不代表绝对安全，仅表示当前平台暂无相关记录</li>
</ul>
</div>
<div class="quick-grid"><div class="quick-card"><strong>快速响应</strong>网页查询接口即时返回</div><div class="quick-card"><strong>重复拦截</strong>提交前检查现有记录</div><div class="quick-card"><strong>人工审核</strong>新增记录进入后台审核</div></div>
</div>`

	feedbackEmail := GetSetting("feedback_email", "")
	if feedbackEmail != "" {
		html += `<div style="text-align:center;margin-top:16px"><a href="mailto:` + esc(feedbackEmail) + `" style="color:#6b7280;text-decoration:none;font-size:13px;font-weight:600;transition:color .15s" onmouseover="this.style.color='#111827'" onmouseout="this.style.color='#6b7280'">&#9993; 反馈问题</a></div>`
	}

	html += `
</div>
<script>
function escHtml(v){return String(v??'').replace(/[&<>'"]/g,function(c){return {'&':'&amp;','<':'&lt;','>':'&gt;',"'":'&#39;','"':'&quot;'}[c]})}
function search(){
var qq=document.getElementById('qqInput').value.trim();
var platform=document.getElementById('platformInput').value;
if(!qq){alert('请输入账号');return;}
document.getElementById('loading').style.display='block';
document.getElementById('result').style.display='none';
fetch('/api/web/query?platform='+encodeURIComponent(platform)+'&account='+encodeURIComponent(qq)).then(r=>r.json()).then(res=>{
document.getElementById('loading').style.display='none';
var html='';
if(res.data&&res.data.in_blacklist){
html='<div class="result in-blacklist"><h2>⚠️ 该QQ号在云黑名单中</h2>';
html+='<div class="info"><span class="info-label">QQ号</span><span>'+escHtml(res.data.qq)+'</span></div>';
if(res.data.nickname){html+='<div class="info"><span class="info-label">昵称</span><span>'+escHtml(res.data.nickname)+'</span></div>'}
if(res.data.reason){html+='<div class="info"><span class="info-label">原因</span><span>'+escHtml(res.data.reason)+'</span></div>'}
if(res.data.severity_text){html+='<div class="info"><span class="info-label">严重程度</span><span>'+escHtml(res.data.severity_text)+' · '+escHtml(res.data.severity_desc||'')+'</span></div>'}
if(res.data.tags&&res.data.tags.length){html+='<div class="info"><span class="info-label">标签</span><span class="tag-list">'+res.data.tags.map(t=>'<span class="tag">'+escHtml(t)+'</span>').join('')+'</span></div>'}
if(res.data.linked_accounts&&res.data.linked_accounts.length){html+='<div class="info"><span class="info-label">关联账号</span><span class="account-list">'+res.data.linked_accounts.map(a=>'<span class="account-item">'+escHtml(a.platform)+'：'+escHtml(a.account)+(a.nickname?'（'+escHtml(a.nickname)+'）':'')+'</span>').join('')+'</span></div>'}
html+='</div>';
}else{
html='<div class="result not-in-blacklist"><h2>平台暂未收录该账号</h2><div class="info"><span class="info-label">说明</span><span>暂未收录不代表绝对安全，仅表示当前平台暂无相关云黑记录。</span></div></div>';
}
document.getElementById('result').innerHTML=html;
document.getElementById('result').style.display='block';
}).catch(e=>{
document.getElementById('loading').style.display='none';
document.getElementById('result').innerHTML='<div class="error">查询失败，请稍后重试</div>';
document.getElementById('result').style.display='block';
});
}
document.getElementById('qqInput').addEventListener('keypress',function(e){if(e.key==='Enter')search()});
</script>
</body>
</html>`
	c.HTML(http.StatusOK, html)
	return nil
}

func webSubmit(c echo.Context) error {
	errorMsg := c.QueryParam("error")
	successMsg := c.QueryParam("success")

	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>提交云黑</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#f5f7fb;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;min-height:100vh;color:#111827;line-height:1.5;overflow-x:hidden}
body:before{content:"";position:fixed;inset:0;background:radial-gradient(circle at 50% -120px,rgba(37,99,235,.10),transparent 460px);pointer-events:none;z-index:-2}
body:after{content:"";position:fixed;inset:0;background-image:linear-gradient(rgba(15,23,42,.035) 1px,transparent 1px),linear-gradient(90deg,rgba(15,23,42,.035) 1px,transparent 1px);background-size:44px 44px;mask-image:linear-gradient(to bottom,rgba(0,0,0,.45),transparent 70%);pointer-events:none;z-index:-1}
.container{max-width:760px;margin:0 auto;padding:56px 22px}
.hero{background:rgba(255,255,255,.86);border:1px solid rgba(229,231,235,.9);border-radius:26px;padding:28px;box-shadow:0 24px 70px rgba(15,23,42,.10);backdrop-filter:blur(18px)}
.header{margin-bottom:24px;display:flex;justify-content:space-between;gap:18px;align-items:flex-end}
.header h1{font-size:38px;line-height:1.08;margin-bottom:10px;letter-spacing:-.045em;color:#111827;font-weight:900}
.header p{color:#6b7280;font-size:15px;max-width:420px}
.status-chip{display:inline-flex;align-items:center;gap:8px;padding:9px 12px;border:1px solid #e5e7eb;border-radius:999px;color:#374151;background:#f8fafc;font-size:13px;font-weight:800;white-space:nowrap}
.status-chip:before{content:"";width:8px;height:8px;border-radius:50%;background:#1f2937;box-shadow:0 0 0 5px rgba(31,41,55,.10)}
.tab-nav{display:flex;gap:6px;margin-bottom:20px;background:#f3f4f6;border:1px solid #e5e7eb;border-radius:14px;padding:5px}
.tab-nav a{flex:1;text-align:center;padding:12px 14px;text-decoration:none;color:#6b7280;border-radius:10px;font-weight:800;transition:background .2s,color .2s,transform .2s}
.tab-nav a.active{background:#111827;color:#fff}
.tab-nav a:not(.active):hover{background:#fff;color:#111827;transform:translateY(-1px)}
.card{background:#fff;border-radius:18px;padding:24px;border:1px solid #e5e7eb;box-shadow:0 16px 42px rgba(15,23,42,.08)}
.form-group{margin-bottom:18px}
.form-group label{display:block;margin-bottom:7px;color:#374151;font-size:14px;font-weight:700}
.form-group input,.form-group textarea,.form-group select{width:100%;padding:13px 14px;border:1px solid #d1d5db;border-radius:12px;background:#f9fafb;color:#111827;font-size:15px;transition:border .2s,box-shadow .2s}
.form-group input:focus,.form-group textarea:focus,.form-group select:focus{outline:none;border-color:#94a3b8;box-shadow:0 0 0 3px rgba(148,163,184,.22)}
.form-group input::placeholder,.form-group textarea::placeholder{color:#9ca3af}
.form-group textarea{height:132px;resize:vertical}
.form-group select{appearance:none;background-color:#f9fafb;cursor:pointer}
.form-group select option{background:#fff;color:#111827}
.btn{width:100%;padding:15px;background:#111827;color:#fff;border:none;border-radius:12px;font-size:16px;font-weight:800;cursor:pointer;transition:background .2s,transform .2s,opacity .2s}
.btn:hover{background:#1f2937;transform:translateY(-1px)}
.btn:disabled{cursor:not-allowed;opacity:.7;transform:none}
.error{background:#fff1f0;color:#b91c1c;padding:14px 15px;border-radius:12px;margin-bottom:18px;border:1px solid #fecaca}
.success{background:#ecfdf3;color:#166534;padding:14px 15px;border-radius:12px;margin-bottom:18px;border:1px solid #bbf7d0}
.submit-note{display:grid;grid-template-columns:repeat(3,1fr);gap:10px;margin-bottom:18px}.submit-note div{background:#fff;border:1px solid #e5e7eb;border-radius:14px;padding:12px;color:#6b7280;font-size:13px;box-shadow:0 10px 26px rgba(15,23,42,.05)}.submit-note strong{display:block;color:#111827;margin-bottom:3px}
.tag-grid{display:grid;grid-template-columns:repeat(4,1fr);gap:8px}.tag-grid label{display:flex;align-items:center;gap:6px;margin:0;padding:9px 10px;border:1px solid #e5e7eb;border-radius:10px;background:#f9fafb;color:#374151;font-size:13px}.tag-grid input{width:auto}.hint{color:#6b7280;font-size:13px;margin-top:6px}.severity-help{margin-top:8px;color:#6b7280;font-size:13px;line-height:1.6}
@media (max-width:720px){.container{padding:28px 14px}.hero{padding:18px;border-radius:20px}.header{display:block}.status-chip{margin-top:14px}.header h1{font-size:30px}.card{padding:18px;border-radius:16px}.tab-nav a{padding:11px 10px}.submit-note,.tag-grid{grid-template-columns:1fr}.form-group input,.form-group textarea,.form-group select{font-size:16px}}
</style>
</head>
<body>
<div class="container">
<div class="hero">
<div class="header">
<div>
<h1>云黑查询</h1>
<p>提交云黑记录</p>
</div>
<div class="status-chip">人工审核</div>
</div>
<div class="tab-nav">
<a href="/web/">查询</a>
<a href="/web/submit" class="active">提交</a>
<a href="/web/api">API</a>
</div>
<div class="submit-note"><div><strong>1. 填写QQ</strong>系统会先检查重复记录</div><div><strong>2. 描述原因</strong>请尽量提供清楚依据</div><div><strong>3. 等待审核</strong>管理员确认后生效</div></div>
<div class="card">`
	if errorMsg != "" {
		html += `<div class="error">` + esc(errorMsg) + `</div>`
	}
	if successMsg != "" {
		html += `<div class="success">` + esc(successMsg) + `</div>`
	}
	html += `
<form method="POST" action="/api/web/submit">
<div class="form-group"><label>关联主体名称</label><input type="text" name="subject_name" placeholder="例如：常用昵称/真实主体/团伙名称，可选"><p class="hint">多个平台账号会关联到同一个主体，便于后续查询时展示完整风险信息。</p></div>
<div class="form-group"><label>QQ号 *</label><input type="text" name="qq" required placeholder="请输入QQ号"></div>
<div class="form-group"><label>昵称</label><input type="text" name="nickname" placeholder="可选"></div>
<div class="form-group"><label>其他平台账号</label><textarea name="accounts" placeholder="每行一个，格式：平台:账号:昵称，例如\n微信:wx_123:张三\nTelegram:@test\n抖音:123456"></textarea><p class="hint">QQ 会自动作为主账号加入关联；其他账号按每行一个填写。</p></div>
<div class="form-group"><label>标签</label><div class="tag-grid">
<label><input type="checkbox" name="tags" value="诈骗">诈骗</label><label><input type="checkbox" name="tags" value="跑路">跑路</label><label><input type="checkbox" name="tags" value="恶意退款">恶意退款</label><label><input type="checkbox" name="tags" value="盗号">盗号</label><label><input type="checkbox" name="tags" value="广告骚扰">广告骚扰</label><label><input type="checkbox" name="tags" value="群内违规">群内违规</label><label><input type="checkbox" name="tags" value="虚假交易">虚假交易</label><label><input type="checkbox" name="tags" value="其他">其他</label>
</div></div>
<div class="form-group"><label>云黑原因 *</label><textarea name="reason" required placeholder="请详细描述云黑原因"></textarea></div>
<div class="form-group"><label>严重程度</label>
<select name="severity">
<option value="1">轻微</option>
<option value="2">一般</option>
<option value="3">较重</option>
<option value="4">严重</option>
<option value="5">极其严重</option>
</select>
<div class="severity-help">1 轻微纠纷 | 2 一般违规 | 3 明显恶意 | 4 严重欺诈 | 5 极高风险/惯犯</div>
</div>
<button type="submit" class="btn" id="submitBtn" onclick="this.disabled=true;this.textContent='提交中...';this.form.submit()">提交</button>
</form>
</div>
</div>
</div>
<script>
document.querySelector('form').addEventListener('submit',function(){document.getElementById('submitBtn').disabled=true;document.getElementById('submitBtn').textContent='提交中...'})
</script>
</body>
</html>`
	c.HTML(http.StatusOK, html)
	return nil
}

func webSubmitPost(c echo.Context) error {
	qq := c.FormValue("qq")
	nickname := c.FormValue("nickname")
	reason := c.FormValue("reason")
	severity := c.FormValue("severity")
	subjectName := c.FormValue("subject_name")
	tags := strings.Join(c.Request().Form["tags"], ",")

	if qq == "" || reason == "" {
		return c.Redirect(302, "/web/submit?error=请填写QQ号和云黑原因")
	}

	if !ValidateQQ(qq) {
		return c.Redirect(302, "/web/submit?error=QQ号格式不正确")
	}
	if len(nickname) > 50 {
		return c.Redirect(302, "/web/submit?error=昵称不能超过50个字符")
	}
	if len(reason) > 2000 {
		return c.Redirect(302, "/web/submit?error=云黑原因不能超过2000个字符")
	}
	if len(subjectName) > 100 {
		return c.Redirect(302, "/web/submit?error=主体名称不能超过100个字符")
	}

	minReason := GetSettingInt("submit_min_reason", 10)
	if !CheckReasonQuality(reason, minReason) {
		if minReason > 0 {
			return c.Redirect(302, fmt.Sprintf("/web/submit?error=云黑原因至少需要%d个有效字符，请详细描述", minReason))
		}
		return c.Redirect(302, "/web/submit?error=云黑原因无效，请详细描述")
	}

	clientIP := GetClientIP(c)
	account := qq
	cooldownMin := GetSettingInt("submit_cooldown", 30)
	if ok, remaining := CheckSubmitCooldown(clientIP, account, cooldownMin); !ok {
		return c.Redirect(302, fmt.Sprintf("/web/submit?error=提交过于频繁，请%d分钟后再试", remaining))
	}

	maxHour := GetSettingInt("submit_max_hour", 200)
	if !CheckGlobalSubmitLimit(maxHour) {
		return c.Redirect(302, "/web/submit?error=系统提交已达每小时上限，请稍后再试")
	}

	severityNum := parseSeverity(severity)
	qqNum64, _ := strconv.ParseInt(qq, 10, 64)
	qqNum := int(qqNum64)

	var cnt int
	err := DB.QueryRow("SELECT COUNT(*) FROM cloudblack_list WHERE qq = ?", qqNum).Scan(&cnt)
	if err == nil && cnt > 0 {
		return c.Redirect(302, "/web/submit?error=该QQ号已在云黑名单中")
	}

	err = DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE qq = ? AND status = 0", qqNum).Scan(&cnt)
	if err == nil && cnt > 0 {
		return c.Redirect(302, "/web/submit?error=该QQ号已提交，请等待审核")
	}

	accounts := buildAccounts(int64(qqNum), nickname, c.FormValue("accounts"))
	DB.Exec("INSERT INTO cloudblack_records (qq, nickname, reason, severity, status, subject_name, tags, accounts, created_at) VALUES (?, ?, ?, ?, 0, ?, ?, ?, datetime('now'))",
		qqNum, nickname, reason, severityNum, subjectName, tags, EncodeAccounts(accounts))
	LogAccess("submit", int64(qqNum), "web", "", 0, c)

	return c.Redirect(302, "/web/submit?success=提交成功，等待管理审核")
}

func webQuery(c echo.Context) error {
	platform := c.QueryParam("platform")
	account := c.QueryParam("account")
	if platform == "" {
		platform = "QQ"
	}
	if account == "" {
		account = c.QueryParam("qq")
	}
	if account == "" {
		return Success(c, "请输入账号", map[string]interface{}{"in_blacklist": false})
	}

	qqNum := int64(0)
	if platform == "QQ" {
		for _, ch := range account {
			if ch >= '0' && ch <= '9' {
				qqNum = qqNum*10 + int64(ch-'0')
			}
		}
	}
	if platform == "QQ" && qqNum < 10000 {
		return Success(c, "QQ号格式不正确", map[string]interface{}{"in_blacklist": false})
	}

	var record struct {
		QQ          int64
		Nickname    string
		Reason      string
		Severity    int
		Tags        string
		AccountsRaw string
	}

	var err error
	if platform == "QQ" {
		err = DB.QueryRow("SELECT qq, nickname, reason, severity, COALESCE(tags,''), COALESCE(accounts,'') FROM cloudblack_list WHERE qq = ? AND status = 1", qqNum).Scan(&record.QQ, &record.Nickname, &record.Reason, &record.Severity, &record.Tags, &record.AccountsRaw)
	} else {
		err = DB.QueryRow("SELECT l.qq, l.nickname, l.reason, l.severity, COALESCE(l.tags,''), COALESCE(l.accounts,'') FROM subject_accounts a JOIN cloudblack_list l ON l.subject_id = a.subject_id WHERE a.platform = ? AND a.account = ? AND l.status = 1 ORDER BY l.id DESC LIMIT 1", platform, account).Scan(&record.QQ, &record.Nickname, &record.Reason, &record.Severity, &record.Tags, &record.AccountsRaw)
	}
	LogAccess("query", qqNum, "web", "", 0, c)

	if err != nil {
		return Success(c, "平台暂未收录该账号，不代表绝对安全", map[string]interface{}{"in_blacklist": false, "platform": platform, "account": account, "note": "暂未收录仅表示当前平台暂无相关云黑记录"})
	}

	return Success(c, "success", map[string]interface{}{
		"in_blacklist":    true,
		"qq":              record.QQ,
		"nickname":        record.Nickname,
		"reason":          record.Reason,
		"severity":        record.Severity,
		"severity_text":   GetSeverityText(record.Severity),
		"severity_desc":   GetSeverityDesc(record.Severity),
		"tags":            splitTags(record.Tags),
		"linked_accounts": DecodeAccounts(record.AccountsRaw),
	})
}

func splitTags(raw string) []string {
	var tags []string
	for _, tag := range strings.Split(raw, ",") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func buildAccounts(qq int64, nickname, raw string) []LinkedAccount {
	accounts := []LinkedAccount{{Platform: "QQ", Account: strconv.FormatInt(qq, 10), Nickname: nickname}}
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}
		acc := LinkedAccount{Platform: strings.TrimSpace(parts[0]), Account: strings.TrimSpace(parts[1])}
		if len(parts) > 2 {
			acc.Nickname = strings.TrimSpace(parts[2])
		}
		if acc.Platform != "" && acc.Account != "" {
			accounts = append(accounts, acc)
		}
	}
	return accounts
}

func webAPIDoc(c echo.Context) error {
	html := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>API 文档</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{background:#f5f7fb;font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,"Helvetica Neue",Arial,sans-serif;min-height:100vh;color:#111827;line-height:1.6;overflow-x:hidden}
.container{max-width:860px;margin:0 auto;padding:56px 22px}
.hero{background:rgba(255,255,255,.86);border:1px solid rgba(229,231,235,.9);border-radius:26px;padding:28px;box-shadow:0 24px 70px rgba(15,23,42,.10);backdrop-filter:blur(18px)}
.header{margin-bottom:24px}
.header h1{font-size:32px;line-height:1.15;margin-bottom:8px;color:#111827;font-weight:900}
.header p{color:#6b7280;font-size:15px}
.tab-nav{display:flex;gap:6px;margin-bottom:24px;background:#f3f4f6;border:1px solid #e5e7eb;border-radius:14px;padding:5px}
.tab-nav a{flex:1;text-align:center;padding:12px 14px;text-decoration:none;color:#6b7280;border-radius:10px;font-weight:800;transition:background .2s,color .2s,transform .2s}
.tab-nav a.active{background:#111827;color:#fff}
.tab-nav a:not(.active):hover{background:#fff;color:#111827;transform:translateY(-1px)}
.endpoint{background:#fff;border:1px solid #e5e7eb;border-radius:18px;padding:22px;margin-bottom:18px;box-shadow:0 10px 26px rgba(15,23,42,.05)}
.endpoint h2{font-size:18px;margin-bottom:10px;color:#111827;display:flex;align-items:center;gap:10px}
.endpoint h2 .method{display:inline-flex;padding:4px 10px;border-radius:6px;font-size:12px;font-weight:800;color:#fff}
.method-get{background:#16a34a}
.method-post{background:#dc2626}
.endpoint .path{font-family:ui-monospace,SFMono-Regular,Consolas,monospace;font-size:15px;color:#374151;margin-bottom:14px;padding:8px 12px;background:#f3f4f6;border-radius:8px;word-break:break-all}
.endpoint p,.endpoint ul{color:#4b5563;font-size:14px;margin-bottom:8px}
.endpoint ul{padding-left:20px}
.endpoint code{background:#f1f5f9;border-radius:5px;padding:2px 6px;color:#dc2626;font-family:ui-monospace,SFMono-Regular,Consolas,monospace;font-size:13px}
.endpoint pre{background:#1f2937;color:#e5e7eb;padding:16px;border-radius:10px;overflow-x:auto;font-size:13px;line-height:1.5;margin-top:10px;font-family:ui-monospace,SFMono-Regular,Consolas,monospace}
.rate-info{background:#fff;border:1px solid #e5e7eb;border-radius:18px;padding:18px 22px;margin-bottom:18px;box-shadow:0 10px 26px rgba(15,23,42,.05);font-size:14px;color:#4b5563;line-height:1.7}
.rate-info strong{color:#111827}
@media(max-width:720px){.container{padding:28px 14px}.hero{padding:18px;border-radius:20px}.header h1{font-size:26px}.endpoint{padding:16px}.tab-nav a{padding:11px 10px}.endpoint pre{font-size:12px;padding:12px}}
</style>
</head>
<body>
<div class="container">
<div class="hero">
<div class="header"><h1>API 文档</h1><p>云黑系统提供 RESTful API 接口，支持查询、提交、审核等功能。</p></div>
<div class="tab-nav">
<a href="/web/">查询</a>
<a href="/web/submit">提交</a>
<a href="/web/api" class="active">API</a>
</div>

<div class="rate-info">
<strong>认证方式</strong><br>
以下接口支持不携带 API Key 直接调用（管理员可随时关闭或调整限额）：<br>
&bull; <code>GET /api/v1/query</code> &mdash; 单账号查询<br>
&bull; <code>GET /api/v1/check</code> &mdash; 快速检查<br>
&bull; <code>GET /api/v1/batch</code> &mdash; 批量查询<br>
&bull; <code>POST /api/v1/submit</code> &mdash; 提交云黑<br>
如需更高频率或更多权限，请联系管理员获取 API 密钥，支持三种方式传递：<br>
&bull; HTTP Header：<code>X-API-Key: your_key</code><br>
&bull; URL 参数：<code>?api_key=your_key</code><br>
&bull; 表单字段：<code>api_key=your_key</code><br><br>
<strong>速率限制</strong>：未提供密钥的请求受 RPM（每分钟次数）限制，具体数值由管理员在后台配置。携带有效密钥则使用更高限额。<br>
<strong>响应格式</strong>：所有响应均为 JSON，统一格式 <code>{"code": 0, "message": "...", "data": ...}</code>
</div>

<div class="endpoint">
<h2><span class="method method-get">GET</span> 单账号查询 <span style="font-size:12px;font-weight:600;color:#16a34a;background:#ecfdf3;padding:3px 8px;border-radius:4px;margin-left:8px">无需密钥</span></h2>
<div class="path">GET /api/v1/query?qq=123456789</div>
<p>查询指定 QQ 是否在云黑名单中，返回完整记录信息（含严重程度、标签、关联账号等）。</p>
<p><strong>参数</strong></p>
<ul>
<li><code>qq</code> <em>必填</em> — QQ 号（5-10位，首位非0）</li>
<li><code>api_key</code> <em>可选</em> — API 密钥</li>
</ul>
<pre>{
  "code": 0,
  "message": "success",
  "data": {
    "in_blacklist": true,
    "id": 1,
    "qq": 123456789,
    "nickname": "昵称",
    "reason": "详细原因",
    "severity": 4,
    "severity_text": "严重",
    "severity_desc": "严重影响，需立即处理",
    "status": 1,
    "status_text": "已通过",
    "tags": ["诈骗", "跑路"],
    "linked_accounts": [
      {"platform": "QQ", "account": "987654321", "nickname": "小号"}
    ],
    "evidence_img": [],
    "created_at": "2026-06-01 12:00:00",
    "reviewed_at": "2026-06-01 12:30:00"
  }
}</pre>
</div>

<div class="endpoint">
<h2><span class="method method-get">GET</span> 快速检查 <span style="font-size:12px;font-weight:600;color:#16a34a;background:#ecfdf3;padding:3px 8px;border-radius:4px;margin-left:8px">无需密钥</span></h2>
<div class="path">GET /api/v1/check?qq=123456789</div>
<p>轻量级检查，仅返回是否在黑名单中，不返回详细记录。</p>
<ul>
<li><code>qq</code> <em>必填</em> — QQ 号</li>
<li><code>api_key</code> <em>可选</em> — API 密钥</li>
</ul>
<pre>{
  "code": 0,
  "message": "success",
  "data": { "in_blacklist": true, "qq": 123456789 }
}</pre>
</div>

<div class="endpoint">
<h2><span class="method method-get">GET</span> 批量查询 <span style="font-size:12px;font-weight:600;color:#16a34a;background:#ecfdf3;padding:3px 8px;border-radius:4px;margin-left:8px">无需密钥</span></h2>
<div class="path">GET /api/v1/batch?qq_list=123456,789012,345678</div>
<p>一次查询多个 QQ，最多 100 个。支持逗号分隔列表或 JSON 数组。</p>
<p><strong>参数</strong>（二选一）</p>
<ul>
<li><code>qq_list</code> — 逗号分隔的 QQ 列表，如 <code>123456,789012</code></li>
<li><code>qq_array</code> — JSON 数组，如 <code>["123456","789012"]</code></li>
<li><code>api_key</code> <em>可选</em> — API 密钥</li>
</ul>
<pre>{
  "code": 0,
  "message": "success",
  "data": {
    "total": 3,
    "found": 1,
    "data": [
      { "qq": 123456, "in_blacklist": false },
      { "qq": 789012, "in_blacklist": true, "reason": "...", "severity": 3, ... },
      { "qq": 345678, "in_blacklist": false }
    ]
  }
}</pre>
</div>

<div class="endpoint">
<h2><span class="method method-post">POST</span> 提交云黑 <span style="font-size:12px;font-weight:600;color:#16a34a;background:#ecfdf3;padding:3px 8px;border-radius:4px;margin-left:8px">无需密钥</span></h2>
<div class="path">POST /api/v1/submit</div>
<p>提交新的云黑记录到审核队列。支持不携带密钥直接提交，受 RPM 和风控限制（同IP冷却期、原因字数、每小时上限）。</p>
<p><strong>参数</strong> — 支持 JSON body 或 form-data</p>
<ul>
<li><code>qq</code> <em>必填</em> — QQ 号</li>
<li><code>reason</code> <em>必填</em> — 云黑原因</li>
<li><code>nickname</code> <em>可选</em> — 昵称</li>
<li><code>severity</code> <em>可选</em> — 严重程度 1~5，默认 1</li>
<li><code>tags</code> <em>可选</em> — 标签（多选时传多个 <code>tags</code> 字段）</li>
<li><code>accounts</code> <em>可选</em> — 关联账号，每行一个，格式 <code>平台:账号:昵称(可选)</code></li>
<li><code>subject_name</code> <em>可选</em> — 主体名称</li>
<li><code>api_key</code> <em>可选</em> — API 密钥</li>
</ul>
<pre>{
  "code": 0,
  "message": "提交成功",
  "data": { "id": 42 }
}</pre>
</div>

<div class="endpoint">
<h2><span class="method method-get">GET</span> 审核列表 <span style="font-size:12px;font-weight:600;color:#b91c1c;background:#fff1f0;padding:3px 8px;border-radius:4px;margin-left:8px">需要 review 权限</span></h2>
<div class="path">GET /api/v1/review/list</div>
<p>获取待审核记录列表，最多返回 100 条。需要密钥且具备 <code>review</code> 权限。</p>
<ul>
<li><code>api_key</code> <em>必填</em> — 具备 review 权限的密钥</li>
</ul>
<pre>{
  "code": 0,
  "message": "success",
  "data": {
    "total": 2,
    "data": [
      { "id": 1, "qq": 123456789, "nickname": "...", "reason": "...", "severity": 3, "severity_text": "较重", "created_at": "2026-06-01 12:00:00" }
    ]
  }
}</pre>
</div>

<div class="endpoint">
<h2><span class="method method-post">POST</span> 审核操作 <span style="font-size:12px;font-weight:600;color:#b91c1c;background:#fff1f0;padding:3px 8px;border-radius:4px;margin-left:8px">需要 review 权限</span></h2>
<div class="path">POST /api/v1/review/action</div>
<p>审核通过或拒绝一条记录。需要密钥且具备 <code>review</code> 权限。</p>
<p><strong>参数</strong> — form-data</p>
<ul>
<li><code>id</code> <em>必填</em> — 记录 ID</li>
<li><code>action</code> <em>必填</em> — <code>approve</code> 或 <code>reject</code></li>
<li><code>note</code> <em>可选</em> — 审核备注</li>
<li><code>api_key</code> <em>必填</em> — 具备 review 权限的密钥</li>
</ul>
<pre>{
  "code": 0,
  "message": "操作成功",
  "data": { "id": "1", "action": "approve" }
}</pre>
</div>

</div>
</div>
</body>
</html>`
	return c.HTML(http.StatusOK, html)
}
