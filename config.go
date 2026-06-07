package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	memStat "github.com/shirou/gopsutil/v3/mem"
	_ "modernc.org/sqlite"
)

var StartTime = time.Now()

type Config struct {
	Debug     bool            `json:"debug"`
	Timezone  string          `json:"timezone"`
	Port      string          `json:"port"`
	DataDir   string          `json:"-"`
	DB        DBConfig        `json:"db"`
	Admin     AdminConfig     `json:"admin"`
	Upload    UploadConfig    `json:"upload"`
	System    SystemConfig    `json:"system"`
	RateLimit RateLimitConfig `json:"rate_limit"`
	Security  SecurityConfig  `json:"security"`
}

type RateLimitConfig struct {
	API    int `json:"api"`
	Web    int `json:"web"`
	Admin  int `json:"admin"`
	Window int `json:"window"`
}

type SecurityConfig struct {
	TrustCloudflare bool `json:"trust_cloudflare"`
	SecureCookie    bool `json:"secure_cookie"`
}

type DBConfig struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Host     string `json:"host"`
	Port     string `json:"port"`
	DBName   string `json:"dbname"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type AdminConfig struct {
	Salt   string `json:"salt"`
	Expire int    `json:"expire"`
}

type UploadConfig struct {
	Path    string   `json:"path"`
	MaxSize int      `json:"max_size"`
	Allowed []string `json:"allowed"`
}

type SystemConfig struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type LinkedAccount struct {
	Platform string `json:"platform"`
	Account  string `json:"account"`
	Nickname string `json:"nickname,omitempty"`
}

var DB *sql.DB
var AppConfig *Config

func LoadConfig(path string) *Config {
	cfg := &Config{
		Debug:    false,
		Timezone: "Asia/Shanghai",
		Port:     "8080",
		DB: DBConfig{
			Type: "sqlite",
			Path: "data/cloudblack.db",
		},
		Admin: AdminConfig{
			Salt:   "cloudblack_2024_secure_salt",
			Expire: 7200,
		},
		Upload: UploadConfig{
			Path:    "web/assets/uploads",
			MaxSize: 5 * 1024 * 1024,
			Allowed: []string{"jpg", "jpeg", "png", "gif", "webp"},
		},
		System: SystemConfig{
			Name:    "云黑系统",
			Version: "1.0.0",
		},
		RateLimit: RateLimitConfig{
			API:    30,
			Web:    5,
			Admin:  10,
			Window: 60,
		},
		Security: SecurityConfig{
			TrustCloudflare: true,
			SecureCookie:    false,
		},
	}

	if _, err := os.Stat(path); err == nil {
		data, _ := os.ReadFile(path)
		json.Unmarshal(data, cfg)
	} else if os.IsNotExist(err) {
		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			_ = os.MkdirAll(dir, 0755)
		}
		_ = SaveConfig(cfg, path)
	}

	return cfg
}

func SaveConfig(cfg *Config, path string) error {
	data, _ := json.MarshalIndent(cfg, "", "  ")
	return os.WriteFile(path, data, 0644)
}

func InitDB(cfg *Config) error {
	var err error
	dbPath := cfg.DataDir + "/" + "cloudblack.db"

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("database open failed: %v", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	if err := createTables(); err != nil {
		return err
	}

	var count int
	err = DB.QueryRow("SELECT COUNT(*) FROM admins").Scan(&count)
	if err == nil && count == 0 {
		hash, _ := HashPassword("123456")
		DB.Exec("INSERT INTO admins (username, password, nickname, role, must_change_password, created_at) VALUES (?, ?, ?, 1, 1, datetime('now'))",
			"admin", hash, "超级管理员")
		log.Println("========================================")
		log.Println("  首次启动 - 初始管理员账号")
		log.Println("  用户名: admin")
		log.Println("  密码:   123456")
		log.Println("  请登录后立即修改默认密码！")
		log.Println("========================================")
	}

	return nil
}

func createTables() error {
	tables := []string{
		`CREATE TABLE IF NOT EXISTS cloudblack_list (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			qq INTEGER UNIQUE NOT NULL,
			nickname TEXT,
			reason TEXT NOT NULL,
			severity INTEGER DEFAULT 1,
			evidence_images TEXT,
			submitter_id INTEGER,
			status INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now')),
			reviewed_at TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS cloudblack_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			qq INTEGER NOT NULL,
			nickname TEXT,
			reason TEXT NOT NULL,
			severity INTEGER DEFAULT 1,
			evidence_images TEXT,
			submitter_id INTEGER,
			status INTEGER DEFAULT 0,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS admins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			nickname TEXT,
			role INTEGER DEFAULT 1,
			last_login TEXT,
			created_at TEXT DEFAULT (datetime('now')),
			must_change_password INTEGER DEFAULT 1
		)`,
		`CREATE TABLE IF NOT EXISTS api_keys (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			api_key TEXT UNIQUE NOT NULL,
			admin_id INTEGER NOT NULL,
			permissions TEXT NOT NULL,
			note TEXT,
			status INTEGER DEFAULT 1,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS admin_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			admin_id INTEGER NOT NULL,
			action TEXT NOT NULL,
			detail TEXT,
			ip TEXT,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS admin_sessions (
			token TEXT PRIMARY KEY,
			admin_id INTEGER NOT NULL,
			expires_at TEXT NOT NULL,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS stats_log (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			qq INTEGER,
			api_key TEXT,
			ip TEXT,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS system_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS black_subjects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			display_name TEXT,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE TABLE IF NOT EXISTS subject_accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			subject_id INTEGER NOT NULL,
			platform TEXT NOT NULL,
			account TEXT NOT NULL,
			nickname TEXT,
			created_at TEXT DEFAULT (datetime('now'))
		)`,
		`CREATE INDEX IF NOT EXISTS idx_cloudblack_qq ON cloudblack_list(qq)`,
		`CREATE INDEX IF NOT EXISTS idx_cloudblack_status ON cloudblack_list(status)`,
		`CREATE INDEX IF NOT EXISTS idx_records_qq ON cloudblack_records(qq)`,
		`CREATE INDEX IF NOT EXISTS idx_subject_accounts ON subject_accounts(platform, account)`,
	}

	for _, sql := range tables {
		if _, err := DB.Exec(sql); err != nil {
			return fmt.Errorf("create table failed: %v", err)
		}
	}

	ensureColumn("cloudblack_list", "subject_id", "INTEGER")
	ensureColumn("cloudblack_list", "subject_name", "TEXT")
	ensureColumn("cloudblack_list", "tags", "TEXT")
	ensureColumn("cloudblack_list", "accounts", "TEXT")
	ensureColumn("cloudblack_list", "reviewed_by", "INTEGER")
	ensureColumn("cloudblack_list", "review_note", "TEXT")
	ensureColumn("cloudblack_records", "subject_id", "INTEGER")
	ensureColumn("cloudblack_records", "subject_name", "TEXT")
	ensureColumn("cloudblack_records", "tags", "TEXT")
	ensureColumn("cloudblack_records", "accounts", "TEXT")
	ensureColumn("cloudblack_records", "reviewed_by", "INTEGER")
	ensureColumn("cloudblack_records", "reviewed_at", "TEXT")
	ensureColumn("cloudblack_records", "review_note", "TEXT")
	ensureColumn("cloudblack_records", "reject_reason", "TEXT")
	ensureColumn("admin_sessions", "csrf_token", "TEXT")

	defaults := map[string]string{
		"public_query_rpm":  "30",
		"public_submit_rpm": "5",
	}
	for key, value := range defaults {
		DB.Exec("INSERT OR IGNORE INTO system_settings (key, value, updated_at) VALUES (?, ?, datetime('now'))", key, value)
	}

	return nil
}

func ensureColumn(table, column, typ string) {
	rows, err := DB.Query("PRAGMA table_info(" + table + ")")
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt interface{}
		rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk)
		if name == column {
			return
		}
	}
	DB.Exec("ALTER TABLE " + table + " ADD COLUMN " + column + " " + typ)
}

func GetSetting(key, defaultValue string) string {
	var value string
	err := DB.QueryRow("SELECT value FROM system_settings WHERE key = ?", key).Scan(&value)
	if err != nil || value == "" {
		return defaultValue
	}
	return value
}

func GetSettingInt(key string, defaultValue int) int {
	value := GetSetting(key, strconv.Itoa(defaultValue))
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return defaultValue
	}
	return n
}

func SetSetting(key string, value string) error {
	_, err := DB.Exec("INSERT INTO system_settings (key, value, updated_at) VALUES (?, ?, datetime('now')) ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = datetime('now')", key, value)
	return err
}

// Submit cooldown tracker: ip+account -> last submit time
var (
	submitCooldown   = make(map[string]time.Time)
	submitCooldownMu sync.Mutex
)

func init() {
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			submitCooldownMu.Lock()
			cutoff := time.Now().Add(-2 * time.Hour)
			for k, t := range submitCooldown {
				if t.Before(cutoff) {
					delete(submitCooldown, k)
				}
			}
			submitCooldownMu.Unlock()
		}
	}()
}

func CheckSubmitCooldown(ip, account string, minutes int) (bool, int) {
	if minutes <= 0 {
		return true, 0
	}
	key := ip + ":" + account
	submitCooldownMu.Lock()
	defer submitCooldownMu.Unlock()
	lastTime, exists := submitCooldown[key]
	now := time.Now()
	if exists && now.Sub(lastTime) < time.Duration(minutes)*time.Minute {
		remaining := int(time.Duration(minutes)*time.Minute-now.Sub(lastTime)) / 60
		if remaining < 1 {
			remaining = 1
		}
		return false, remaining
	}
	// prevent unbounded growth under attack: cap at 50000 entries
	const maxEntries = 50000
	if !exists && len(submitCooldown) >= maxEntries {
		return false, minutes
	}
	submitCooldown[key] = now
	return true, 0
}

func CheckGlobalSubmitLimit(maxPerHour int) bool {
	if maxPerHour <= 0 {
		return true
	}
	var cnt int
	DB.QueryRow("SELECT COUNT(*) FROM cloudblack_records WHERE created_at > datetime('now','-1 hour')").Scan(&cnt)
	return cnt < maxPerHour
}

func CheckReasonQuality(reason string, minLen int) bool {
	if minLen <= 0 {
		return true
	}
	s := strings.TrimSpace(reason)
	if len([]rune(s)) < minLen {
		return false
	}
	// reject purely repetitive or gibberish
	if isGarbageText(s) {
		return false
	}
	return true
}

func isGarbageText(s string) bool {
	if len(s) == 0 {
		return true
	}
	r := []rune(s)
	// all same character repeated
	same := true
	for i := 1; i < len(r); i++ {
		if r[i] != r[0] {
			same = false
			break
		}
	}
	if same {
		return true
	}
	// pure digits
	allDigit := true
	for _, ch := range r {
		if ch < '0' || ch > '9' {
			allDigit = false
			break
		}
	}
	if allDigit {
		return true
	}
	// pure ASCII punctuation/gibberish (no CJK, no letter)
	hasLetterOrCJK := false
	for _, ch := range r {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch >= 0x4e00 {
			hasLetterOrCJK = true
			break
		}
	}
	if !hasLetterOrCJK {
		return true
	}
	return false
}

func Now() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func GetSeverityText(severity int) string {
	texts := []string{"", "轻微", "一般", "较重", "严重", "极其严重"}
	if severity < 1 || severity > 5 {
		return "未知"
	}
	return texts[severity]
}

func GetSeverityDesc(severity int) string {
	descs := []string{"", "轻微纠纷或低风险行为", "一般违规或存在争议风险", "明显恶意行为或较高风险", "严重欺诈、长期恶意或造成较大损失", "极高风险、惯犯或团伙相关行为"}
	if severity < 1 || severity > 5 {
		return "未知严重程度"
	}
	return descs[severity]
}

func GetDefaultTags() []string {
	return []string{"诈骗", "跑路", "恶意退款", "盗号", "广告骚扰", "群内违规", "虚假交易", "其他"}
}

func EncodeAccounts(accounts []LinkedAccount) string {
	data, _ := json.Marshal(accounts)
	return string(data)
}

func DecodeAccounts(raw string) []LinkedAccount {
	if raw == "" {
		return nil
	}
	var accounts []LinkedAccount
	json.Unmarshal([]byte(raw), &accounts)
	return accounts
}

func GetStatusText(status int) string {
	texts := []string{"待审核", "已通过", "已拒绝"}
	if status < 0 || status > 2 {
		return "未知"
	}
	return texts[status]
}

// Server status monitoring
type ServerStatus struct {
	CPUPercent  float64
	MemAlloc    uint64
	MemSys      uint64
	SysMemUsed  uint64
	SysMemTotal uint64
}

var LiveStatus ServerStatus
var statusMu sync.RWMutex

func GetLiveStatus() ServerStatus {
	statusMu.RLock()
	defer statusMu.RUnlock()
	return LiveStatus
}

func InitStatusSampler() {
	go func() {
		cpu.Percent(0, false)
		for {
			time.Sleep(2 * time.Second)

			cpuPercents, err := cpu.Percent(0, false)
			var cpuVal float64
			if err == nil && len(cpuPercents) > 0 {
				cpuVal = math.Round(cpuPercents[0]*10) / 10
			}

			var mem runtime.MemStats
			runtime.ReadMemStats(&mem)

			var sysUsed, sysTotal uint64
			if vm, err := memStat.VirtualMemory(); err == nil {
				sysUsed = vm.Used
				sysTotal = vm.Total
			}

			statusMu.Lock()
			LiveStatus.CPUPercent = cpuVal
			LiveStatus.MemAlloc = mem.Alloc
			LiveStatus.MemSys = mem.Sys
			LiveStatus.SysMemUsed = sysUsed
			LiveStatus.SysMemTotal = sysTotal
			statusMu.Unlock()
		}
	}()

	time.Sleep(100 * time.Millisecond)
	cpuPercents, _ := cpu.Percent(500*time.Millisecond, false)
	var cpuVal float64
	if len(cpuPercents) > 0 {
		cpuVal = math.Round(cpuPercents[0]*10) / 10
	}
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	var sysUsed, sysTotal uint64
	if vm, err := memStat.VirtualMemory(); err == nil {
		sysUsed = vm.Used
		sysTotal = vm.Total
	}

	statusMu.Lock()
	LiveStatus.CPUPercent = cpuVal
	LiveStatus.MemAlloc = mem.Alloc
	LiveStatus.MemSys = mem.Sys
	LiveStatus.SysMemUsed = sysUsed
	LiveStatus.SysMemTotal = sysTotal
	statusMu.Unlock()
}
