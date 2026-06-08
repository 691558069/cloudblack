package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// AISettings 从 ai_settings 表读取的配置
type AISettings struct {
	EnableOfflineAI        int
	OfflineStart           string
	OfflineEnd             string
	AutoRejectConfidence   int
	AutoApproveConfidence  int
	MinIPCount             int
	MinQueryCount          int
	APIProvider            string
	APIKey                 string
	APIModel               string
	APIEndpoint            string
}

// AIReviewResult AI 分析结果
type AIReviewResult struct {
	Score    int
	Result   string
	Reason   string
	RawJSON  string
}

// GetAISettings 读取 AI 配置（id=1）
func GetAISettings() *AISettings {
	var s AISettings
	err := DB.QueryRow(`SELECT enable_offline_ai, offline_start, offline_end, auto_reject_confidence, auto_approve_confidence, min_ip_count, min_query_count, api_provider, api_key, api_model, api_endpoint FROM ai_settings WHERE id = 1`).Scan(
		&s.EnableOfflineAI, &s.OfflineStart, &s.OfflineEnd, &s.AutoRejectConfidence, &s.AutoApproveConfidence,
		&s.MinIPCount, &s.MinQueryCount, &s.APIProvider, &s.APIKey, &s.APIModel, &s.APIEndpoint,
	)
	if err != nil {
		return &AISettings{
			EnableOfflineAI: 0, OfflineStart: "23:00", OfflineEnd: "08:00",
			AutoRejectConfidence: 20, AutoApproveConfidence: 85,
			MinIPCount: 3, MinQueryCount: 50,
			APIProvider: "openai", APIModel: "gpt-4o-mini",
		}
	}
	return &s
}

// IsInOfflinePeriod 判断是否处于离线 AI 工作时段
func IsInOfflinePeriod(start, end string) bool {
	now := time.Now()
	nowStr := now.Format("15:04")
	if start <= end {
		return nowStr >= start && nowStr <= end
	}
	// 跨天（如 23:00 - 08:00）
	return nowStr >= start || nowStr <= end
}

// GetDistinctIPCount 统计 N 小时内某 QQ 被提交的独立 IP 数
func GetDistinctIPCount(qq int64, hours int) int {
	var count int
	DB.QueryRow(
		"SELECT COUNT(DISTINCT ip) FROM access_logs WHERE action = 'submit' AND qq = ? AND created_at > datetime('now', ?)",
		qq, "-"+strconv.Itoa(hours)+" hours",
	).Scan(&count)
	return count
}

// GetQueryCount 统计 N 天内某 QQ 被查询的次数
func GetQueryCount(qq int64, days int) int {
	var count int
	DB.QueryRow(
		"SELECT COUNT(*) FROM access_logs WHERE action = 'query' AND qq = ? AND created_at > datetime('now', ?)",
		qq, "-"+strconv.Itoa(days)+" days",
	).Scan(&count)
	return count
}

// GetModelsFromAPI 从外部 API 获取模型列表
func GetModelsFromAPI(endpoint, apiKey string) ([]string, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("endpoint is empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("api key is empty")
	}

	url := strings.TrimRight(endpoint, "/") + "/models"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 8 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response failed: %v", err)
	}

	var models []string
	for _, m := range result.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

// BuildPrompt 构建 AI 审核 Prompt
func BuildPrompt(qq int64, severity int, reason, tags, accounts, nickname string) string {
	return fmt.Sprintf(`你是一位云黑名单审核助手。你的任务是对用户提交的云黑记录进行可信度评分。

【重要原则】
- 文本可以被伪造，不要只看"写得好不好"
- 警惕模板化描述（多个提交文本高度相似）
- 警惕纯情绪发泄（"这个人是骗子"但没有具体信息）
- 警惕"听说"、"别人说的"等二手信息
- 警惕广告、网址、加群等无关内容

【提交信息】
QQ号: %d
严重程度: %d/5
昵称: %s
标签: %s
原因: %s
关联账号: %s

【评分规则】
0-20: 明显垃圾/广告/灌水/与云黑无关
21-40: 描述模糊、缺乏证据、纯情绪
41-60: 描述一般，有一定信息但不够详细
61-85: 描述详细，有具体金额/平台/时间/经过
86-100: 描述完整、结构化、有跨平台证据

【输出格式】（JSON）
{
  "score": 整数0-100,
  "result": "auto_reject" | "manual_review" | "auto_approve",
  "reason": "一句话中文解释"
}`, qq, severity, nickname, tags, reason, accounts)
}

// CallAIReview 调用外部 AI API 进行审核
func CallAIReview(qq int64, severity int, reason, tags, accounts, nickname string) (*AIReviewResult, error) {
	settings := GetAISettings()
	if settings.APIKey == "" {
		return nil, fmt.Errorf("api key not configured")
	}

	endpoint := settings.APIEndpoint
	if endpoint == "" {
		if settings.APIProvider == "openai" {
			endpoint = "https://api.openai.com/v1"
		} else if settings.APIProvider == "deepseek" {
			endpoint = "https://api.deepseek.com/v1"
		}
	}

	prompt := BuildPrompt(qq, severity, reason, tags, accounts, nickname)

	reqBody := map[string]interface{}{
		"model":    settings.APIModel,
		"messages": []map[string]string{{"role": "user", "content": prompt}},
		"temperature": 0.3,
		"max_tokens": 500,
	}
	jsonBody, _ := json.Marshal(reqBody)

	url := strings.TrimRight(endpoint, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+settings.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response: %v", err)
	}
	if len(apiResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := apiResp.Choices[0].Message.Content
	return parseAIResponse(content, string(body))
}

// parseAIResponse 解析 AI 返回内容
func parseAIResponse(content, rawJSON string) (*AIReviewResult, error) {
	content = strings.TrimSpace(content)

	// 尝试从 JSON 中提取
	var result struct {
		Score  int    `json:"score"`
		Result string `json:"result"`
		Reason string `json:"reason"`
	}

	// 如果 content 包含 ```json 代码块
	if idx := strings.Index(content, "{"); idx != -1 {
		jsonStr := content[idx:]
		if endIdx := strings.LastIndex(jsonStr, "}"); endIdx != -1 {
			jsonStr = jsonStr[:endIdx+1]
			if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
				return &AIReviewResult{
					Score:   result.Score,
					Result:  result.Result,
					Reason:  result.Reason,
					RawJSON: rawJSON,
				}, nil
			}
		}
	}

	// 如果解析失败，返回 manual_review
	return &AIReviewResult{
		Score:   50,
		Result:  "manual_review",
		Reason:  "AI 返回格式异常",
		RawJSON: rawJSON,
	}, nil
}

// PerformOfflineReview 核心审核逻辑
func PerformOfflineReview(recordID int, qq int64, reason, tags, accounts, nickname string, severity int) {
	settings := GetAISettings()
	if settings.EnableOfflineAI == 0 {
		return
	}
	if !IsInOfflinePeriod(settings.OfflineStart, settings.OfflineEnd) {
		return
	}

	// 1. 行为数据
	ipCount := GetDistinctIPCount(qq, 72)
	queryCount := GetQueryCount(qq, 30)

	// 2. 本地规则：先检查是否是垃圾
	if isGarbageText(reason) {
		SaveAIReviewLog(recordID, "submit", &AIReviewResult{Score: 10, Result: "auto_reject", Reason: "本地规则：垃圾文本"}, ipCount, queryCount, "auto_reject")
		AutoRejectRecord(recordID)
		return
	}

	// 3. 调用 AI
	aiResult, err := CallAIReview(qq, severity, reason, tags, accounts, nickname)
	if err != nil {
		SaveAIReviewLog(recordID, "submit", &AIReviewResult{Score: 0, Result: "manual_review", Reason: "AI 调用失败: " + err.Error()}, ipCount, queryCount, "manual_review")
		return
	}

	// 4. 决策
	finalDecision := "manual_review"

	// 自动拒绝：AI 评分极低
	if aiResult.Score <= settings.AutoRejectConfidence {
		finalDecision = "auto_reject"
	}

	// 自动通过：AI 评分高 + 行为证据充分
	if aiResult.Score >= settings.AutoApproveConfidence {
		if ipCount >= settings.MinIPCount || queryCount >= settings.MinQueryCount {
			finalDecision = "auto_approve"
		}
	}

	// 5. 执行
	SaveAIReviewLog(recordID, "submit", aiResult, ipCount, queryCount, finalDecision)

	switch finalDecision {
	case "auto_reject":
		AutoRejectRecord(recordID)
	case "auto_approve":
		AutoApproveRecord(recordID, aiResult.Reason)
	}
}

// AutoApproveRecord AI 自动通过记录
func AutoApproveRecord(recordID int, aiReason string) {
	var qq int64
	var nickname, reason string
	var severity int
	var tags, accounts string

	err := DB.QueryRow("SELECT qq, nickname, reason, severity, COALESCE(tags,''), COALESCE(accounts,'') FROM cloudblack_records WHERE id = ?", recordID).Scan(
		&qq, &nickname, &reason, &severity, &tags, &accounts,
	)
	if err != nil {
		return
	}

	// 使用 performReviewAction 相同的逻辑
	performReviewAction(strconv.Itoa(recordID), "approve", 0, "AI 自动通过: "+aiReason)
}

// AutoRejectRecord AI 自动拒绝记录
func AutoRejectRecord(recordID int) {
	DB.Exec("UPDATE cloudblack_records SET status = 2, reviewed_at = datetime('now') WHERE id = ?", recordID)
}

// SaveAIReviewLog 保存 AI 审核日志
func SaveAIReviewLog(recordID int, action string, aiResult *AIReviewResult, ipCount, queryCount int, finalDecision string) {
	DB.Exec(`INSERT INTO ai_review_logs (record_id, action, ai_result, ai_score, ai_reason, raw_response, behavior_ip_count, behavior_query_count, final_status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))`,
		recordID, action, finalDecision, aiResult.Score, aiResult.Reason, aiResult.RawJSON, ipCount, queryCount, finalDecision,
	)
}
