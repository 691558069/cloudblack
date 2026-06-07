package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func RegisterAPIV1Routes(e *echo.Group, cfg *Config) {
	api := e.Group("/api/v1")
	rl := NewRateLimiter(cfg.RateLimit.API, time.Duration(cfg.RateLimit.Window)*time.Second)
	api.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			key := GetAPIKey(c)
			limit := cfg.RateLimit.API
			identifier := key
			if key != "" {
				identifier = "key:" + key
			} else {
				identifier = "public:" + GetClientIP(c) + ":" + c.Path()
				if c.Path() == "/api/v1/submit" {
					limit = GetSettingInt("public_submit_rpm", 5)
				} else {
					limit = GetSettingInt("public_query_rpm", 30)
				}
			}
			if !rl.AllowLimit(identifier, limit) {
				return Error(c, "请求过于频繁，请稍后重试", http.StatusTooManyRequests)
			}
			return next(c)
		}
	})

	api.POST("/submit", handleAPISubmit)
	api.GET("/query", handleAPIQuery)
	api.GET("/batch", handleAPIBatch)
	api.GET("/check", handleAPICheck)
	api.GET("/review/list", handleAPIReviewList)
	api.POST("/review/action", handleAPIReviewAction)
}

type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSONReturn(c echo.Context, code int, message string, data interface{}) error {
	c.JSON(code, APIResponse{Code: code, Message: message, Data: data})
	return nil
}

func Success(c echo.Context, message string, data interface{}) error {
	return JSONReturn(c, http.StatusOK, message, data)
}

func Error(c echo.Context, message string, code int) error {
	return JSONReturn(c, code, message, nil)
}

func GetAPIKey(c echo.Context) string {
	key := c.Request().Header.Get("X-API-Key")
	if key == "" {
		key = c.QueryParam("api_key")
	}
	if key == "" {
		key = c.FormValue("api_key")
	}
	return key
}

func GetClientIP(c echo.Context) string {
	if AppConfig != nil && AppConfig.Security.TrustCloudflare {
		ip := strings.TrimSpace(c.Request().Header.Get("CF-Connecting-IP"))
		if ip != "" {
			return ip
		}
		return c.RealIP()
	}

	ip := c.Request().Header.Get("X-Forwarded-For")
	if ip != "" {
		return strings.Split(ip, ",")[0]
	}
	ip = c.Request().Header.Get("X-Real-IP")
	if ip != "" {
		return ip
	}
	return c.RealIP()
}

func ValidateQQ(qq string) bool {
	matched, _ := regexp.MatchString(`^[1-9][0-9]{4,9}$`, qq)
	return matched
}

type APIKeyInfo struct {
	ID          int
	APIKey      string
	AdminID     int
	Permissions string
	Status      int
}

func GetAPIKeyInfo(apiKey string) (*APIKeyInfo, error) {
	if apiKey == "" {
		return nil, nil
	}

	var info APIKeyInfo
	err := DB.QueryRow("SELECT id, api_key, admin_id, permissions, status FROM api_keys WHERE api_key = ?", apiKey).Scan(
		&info.ID, &info.APIKey, &info.AdminID, &info.Permissions, &info.Status,
	)
	if err != nil {
		return nil, err
	}

	if info.Status != 1 {
		return nil, nil
	}

	return &info, nil
}

func HasPermission(info *APIKeyInfo, perm string) bool {
	if info == nil {
		return false
	}
	perms := strings.Split(info.Permissions, ",")
	for _, p := range perms {
		if strings.TrimSpace(p) == perm {
			return true
		}
	}
	return false
}

func handleAPISubmit(c echo.Context) error {
	apiKey := GetAPIKey(c)
	var info *APIKeyInfo
	var err error
	submitterID := 0
	logKey := "public"
	if apiKey != "" {
		info, err = GetAPIKeyInfo(apiKey)
		if err != nil || info == nil {
			return Error(c, "API密钥无效", http.StatusUnauthorized)
		}
		if !HasPermission(info, "submit") {
			return Error(c, "无提交权限", http.StatusForbidden)
		}
		submitterID = info.AdminID
		logKey = apiKey
	}

	qq := c.FormValue("qq")
	nickname := c.FormValue("nickname")
	reason := c.FormValue("reason")
	severity := c.FormValue("severity")
	subjectName := c.FormValue("subject_name")
	c.Request().ParseForm()
	tags := strings.Join(c.Request().Form["tags"], ",")

	if qq == "" || !ValidateQQ(qq) {
		return Error(c, "QQ号格式不正确", http.StatusBadRequest)
	}

	if reason == "" {
		return Error(c, "请填写云黑原因", http.StatusBadRequest)
	}

	// Submit anti-abuse checks
	minReason := GetSettingInt("submit_min_reason", 10)
	if !CheckReasonQuality(reason, minReason) {
		if minReason > 0 {
			return Error(c, fmt.Sprintf("云黑原因至少需要%d个有效字符，请详细描述", minReason), http.StatusBadRequest)
		}
		return Error(c, "云黑原因无效，请详细描述", http.StatusBadRequest)
	}

	clientIP := GetClientIP(c)
	account := qq
	if platform := c.FormValue("platform"); platform != "" {
		account = platform + ":" + qq
	}
	cooldownMin := GetSettingInt("submit_cooldown", 30)
	if ok, remaining := CheckSubmitCooldown(clientIP, account, cooldownMin); !ok {
		return Error(c, fmt.Sprintf("提交过于频繁，请%d分钟后再试", remaining), http.StatusTooManyRequests)
	}

	maxHour := GetSettingInt("submit_max_hour", 200)
	if !CheckGlobalSubmitLimit(maxHour) {
		return Error(c, "系统提交已达每小时上限，请稍后再试", http.StatusTooManyRequests)
	}

	severityInt := 1
	if severity != "" {
		severityInt, _ = strconv.Atoi(severity)
		if severityInt < 1 || severityInt > 5 {
			severityInt = 1
		}
	}

	qqNum, _ := strconv.ParseInt(qq, 10, 64)

	var existing int
	err = DB.QueryRow("SELECT id FROM cloudblack_list WHERE qq = ?", qqNum).Scan(&existing)
	if err == nil {
		return Error(c, "该QQ号已在云黑名单中", http.StatusBadRequest)
	}

	err = DB.QueryRow("SELECT id FROM cloudblack_records WHERE qq = ? AND status = 0", qqNum).Scan(&existing)
	if err == nil {
		return Error(c, "该QQ号已提交待审核", http.StatusBadRequest)
	}

	tx, err := DB.Begin()
	if err != nil {
		return Error(c, "提交失败: 数据库错误", http.StatusInternalServerError)
	}
	accounts := buildAccounts(qqNum, nickname, c.FormValue("accounts"))
	res, err := tx.Exec(
		"INSERT INTO cloudblack_records (qq, nickname, reason, severity, submitter_id, status, subject_name, tags, accounts, created_at) VALUES (?, ?, ?, ?, ?, 0, ?, ?, ?, datetime('now'))",
		qqNum, nickname, reason, severityInt, submitterID, subjectName, tags, EncodeAccounts(accounts),
	)
	if err != nil {
		tx.Rollback()
		return Error(c, "提交失败: "+err.Error(), http.StatusInternalServerError)
	}

	recordID64, _ := res.LastInsertId()

	_, err = tx.Exec(
		"INSERT INTO stats_log (type, qq, api_key, ip, created_at) VALUES ('submit', ?, ?, ?, datetime('now'))",
		qqNum, logKey, GetClientIP(c),
	)
	if err != nil {
		tx.Rollback()
		return Error(c, "提交失败", http.StatusInternalServerError)
	}
	tx.Commit()

	recordID := int(recordID64)

	return Success(c, "提交成功", map[string]interface{}{"id": recordID})
}

func handleAPIQuery(c echo.Context) error {
	apiKey := GetAPIKey(c)
	logKey := "public"
	if apiKey != "" {
		info, err := GetAPIKeyInfo(apiKey)
		if err != nil || info == nil {
			return Error(c, "API密钥无效", http.StatusUnauthorized)
		}
		if !HasPermission(info, "query") {
			return Error(c, "无查询权限", http.StatusForbidden)
		}
		logKey = apiKey
	}

	qq := c.QueryParam("qq")
	if qq == "" || !ValidateQQ(qq) {
		return Error(c, "QQ号格式不正确", http.StatusBadRequest)
	}

	qqNum, _ := strconv.ParseInt(qq, 10, 64)

	type Record struct {
		ID         int
		QQ         int64
		Nickname   sql.NullString
		Reason     string
		Severity   int
		Status     int
		Evidence   sql.NullString
		CreatedAt  string
		ReviewedAt sql.NullString
		Tags       string
		Accounts   string
	}

	var record Record
	err := DB.QueryRow(
		"SELECT id, qq, nickname, reason, severity, status, evidence_images, created_at, reviewed_at, COALESCE(tags,''), COALESCE(accounts,'') FROM cloudblack_list WHERE qq = ?",
		qqNum,
	).Scan(&record.ID, &record.QQ, &record.Nickname, &record.Reason, &record.Severity, &record.Status, &record.Evidence, &record.CreatedAt, &record.ReviewedAt, &record.Tags, &record.Accounts)

	DB.Exec(
		"INSERT INTO stats_log (type, qq, api_key, ip, created_at) VALUES ('query', ?, ?, ?, datetime('now'))",
		qqNum, logKey, GetClientIP(c),
	)

	if err != nil {
		return Success(c, "平台暂未收录该QQ号，不代表绝对安全", map[string]interface{}{"in_blacklist": false, "note": "暂未收录仅表示当前平台暂无相关云黑记录"})
	}

	var images []string
	if record.Evidence.String != "" {
		json.Unmarshal([]byte(record.Evidence.String), &images)
	}

	return Success(c, "success", map[string]interface{}{
		"id":              record.ID,
		"qq":              record.QQ,
		"nickname":        record.Nickname.String,
		"reason":          record.Reason,
		"severity":        record.Severity,
		"severity_text":   GetSeverityText(record.Severity),
		"severity_desc":   GetSeverityDesc(record.Severity),
		"status":          record.Status,
		"status_text":     GetStatusText(record.Status),
		"evidence_img":    images,
		"tags":            splitTags(record.Tags),
		"linked_accounts": DecodeAccounts(record.Accounts),
		"created_at":      record.CreatedAt,
		"reviewed_at":     record.ReviewedAt.String,
	})
}

func handleAPIBatch(c echo.Context) error {
	apiKey := GetAPIKey(c)
	logKey := "public"
	if apiKey != "" {
		info, err := GetAPIKeyInfo(apiKey)
		if err != nil || info == nil {
			return Error(c, "API密钥无效", http.StatusUnauthorized)
		}
		if !HasPermission(info, "query") {
			return Error(c, "无查询权限", http.StatusForbidden)
		}
		logKey = apiKey
	}

	qqList := c.QueryParam("qq_list")
	qqArray := c.QueryParam("qq_array")

	var qqStrs []string
	if qqList != "" {
		qqStrs = strings.Split(qqList, ",")
	} else if qqArray != "" {
		json.Unmarshal([]byte(qqArray), &qqStrs)
	} else {
		return Error(c, "缺少QQ号列表", http.StatusBadRequest)
	}

	var qqNums []int64
	for _, q := range qqStrs {
		q = strings.TrimSpace(q)
		if ValidateQQ(q) {
			n, _ := strconv.ParseInt(q, 10, 64)
			qqNums = append(qqNums, n)
		}
	}

	if len(qqNums) > 100 {
		return Error(c, "单次最多查询100个QQ号", http.StatusBadRequest)
	}

	type Record struct {
		ID         int
		QQ         int64
		Nickname   sql.NullString
		Reason     string
		Severity   int
		Status     int
		Evidence   sql.NullString
		CreatedAt  string
		ReviewedAt sql.NullString
		Tags       string
		Accounts   string
	}

	records := make(map[int64]Record)
	for _, qq := range qqNums {
		var r Record
		err := DB.QueryRow(
			"SELECT id, qq, nickname, reason, severity, status, evidence_images, created_at, reviewed_at, COALESCE(tags,''), COALESCE(accounts,'') FROM cloudblack_list WHERE qq = ?",
			qq,
		).Scan(&r.ID, &r.QQ, &r.Nickname, &r.Reason, &r.Severity, &r.Status, &r.Evidence, &r.CreatedAt, &r.ReviewedAt, &r.Tags, &r.Accounts)
		if err == nil {
			records[qq] = r
		}
	}

	DB.Exec(
		"INSERT INTO stats_log (type, api_key, ip, created_at) VALUES ('batch_query', ?, ?, datetime('now'))",
		logKey, GetClientIP(c),
	)

	var result []map[string]interface{}
	for _, qq := range qqNums {
		if r, ok := records[qq]; ok {
			var images []string
			if r.Evidence.String != "" {
				json.Unmarshal([]byte(r.Evidence.String), &images)
			}
			result = append(result, map[string]interface{}{
				"id":              r.ID,
				"qq":              r.QQ,
				"nickname":        r.Nickname.String,
				"reason":          r.Reason,
				"severity":        r.Severity,
				"severity_text":   GetSeverityText(r.Severity),
				"status":          r.Status,
				"status_text":     GetStatusText(r.Status),
				"evidence_img":    images,
				"tags":            splitTags(r.Tags),
				"linked_accounts": DecodeAccounts(r.Accounts),
				"created_at":      r.CreatedAt,
				"reviewed_at":     r.ReviewedAt.String,
				"in_blacklist":    true,
			})
		} else {
			result = append(result, map[string]interface{}{
				"qq":           qq,
				"in_blacklist": false,
			})
		}
	}

	return Success(c, "success", map[string]interface{}{
		"total": len(qqNums),
		"found": len(records),
		"data":  result,
	})
}

func handleAPICheck(c echo.Context) error {
	apiKey := GetAPIKey(c)
	if apiKey != "" {
		info, err := GetAPIKeyInfo(apiKey)
		if err != nil || info == nil {
			return Error(c, "API密钥无效", http.StatusUnauthorized)
		}
		if !HasPermission(info, "query") {
			return Error(c, "无查询权限", http.StatusForbidden)
		}
	}

	qq := c.QueryParam("qq")
	if qq == "" || !ValidateQQ(qq) {
		return Error(c, "QQ号格式不正确", http.StatusBadRequest)
	}

	qqNum, _ := strconv.ParseInt(qq, 10, 64)

	type Record struct {
		QQ int64
	}

	var record Record
	err := DB.QueryRow("SELECT qq FROM cloudblack_list WHERE qq = ?", qqNum).Scan(&record.QQ)

	if err != nil {
		return Success(c, "平台暂未收录该QQ号，不代表绝对安全", map[string]interface{}{"in_blacklist": false, "qq": qqNum, "note": "暂未收录仅表示当前平台暂无相关云黑记录"})
	}

	return Success(c, "success", map[string]interface{}{"in_blacklist": true, "qq": qqNum})
}

func requireAPIReviewPermission(c echo.Context) (*APIKeyInfo, error) {
	apiKey := GetAPIKey(c)
	if apiKey == "" {
		return nil, fmt.Errorf("缺少API密钥")
	}
	info, err := GetAPIKeyInfo(apiKey)
	if err != nil || info == nil {
		return nil, fmt.Errorf("API密钥无效")
	}
	if !HasPermission(info, "review") {
		return nil, fmt.Errorf("无审核权限")
	}
	return info, nil
}

func handleAPIReviewList(c echo.Context) error {
	_, err := requireAPIReviewPermission(c)
	if err != nil {
		return Error(c, err.Error(), http.StatusUnauthorized)
	}

	rows, err := DB.Query("SELECT id, qq, nickname, reason, severity, created_at FROM cloudblack_records WHERE status = 0 ORDER BY id DESC LIMIT 100")
	if err != nil {
		return Error(c, "获取审核列表失败", http.StatusInternalServerError)
	}
	defer rows.Close()

	var records []map[string]interface{}
	for rows.Next() {
		var id, severity int
		var qq int64
		var nickname, reason, createdAt string
		rows.Scan(&id, &qq, &nickname, &reason, &severity, &createdAt)
		records = append(records, map[string]interface{}{
			"id":            id,
			"qq":            qq,
			"nickname":      nickname,
			"reason":        reason,
			"severity":      severity,
			"severity_text": GetSeverityText(severity),
			"created_at":    createdAt,
		})
	}

	return Success(c, "success", map[string]interface{}{"total": len(records), "data": records})
}

func handleAPIReviewAction(c echo.Context) error {
	info, err := requireAPIReviewPermission(c)
	if err != nil {
		return Error(c, err.Error(), http.StatusUnauthorized)
	}

	id := c.FormValue("id")
	action := c.FormValue("action")
	note := c.FormValue("note")
	if id == "" || (action != "approve" && action != "reject") {
		return Error(c, "参数错误", http.StatusBadRequest)
	}

	err = performReviewAction(id, action, info.AdminID, note)
	if err != nil {
		return Error(c, err.Error(), http.StatusBadRequest)
	}

	DB.Exec("INSERT INTO admin_logs (admin_id, action, detail, ip, created_at) VALUES (?, ?, ?, ?, datetime('now'))", info.AdminID, "api_review_"+action, "record_id="+id, GetClientIP(c))
	return Success(c, "操作成功", map[string]interface{}{"id": id, "action": action})
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func RandomString(length int) string {
	byteLen := length
	if byteLen < 16 {
		byteLen = 16
	}
	b := make([]byte, byteLen)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	encoded := hex.EncodeToString(b)
	if len(encoded) > length {
		return encoded[:length]
	}
	return encoded
}
