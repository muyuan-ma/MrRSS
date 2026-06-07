package core

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"MrRSS/internal/ai"
	"MrRSS/internal/models"
	"MrRSS/internal/summary"
	"MrRSS/internal/utils/textutil"
)

const (
	defaultDigestMaxArticles = 20
	maxDigestArticleChars    = 6000
)

// GenerateDailyDigest summarizes newly collected articles and stores a daily briefing.
func (h *Handler) GenerateDailyDigest(ctx context.Context, force bool) (*models.DailyDigest, error) {
	today := time.Now().Format("2006-01-02")
	if !force {
		existing, err := h.DB.GetDailyDigestByDate(today)
		if err == nil {
			existing.Articles, _ = h.DB.GetDailyDigestArticles(existing.ID)
			return existing, nil
		}
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
	}

	maxArticles := getIntSetting(h, "agent_digest_max_articles", defaultDigestMaxArticles)
	if maxArticles <= 0 {
		maxArticles = defaultDigestMaxArticles
	}

	since := time.Now().Add(-24 * time.Hour)
	if latest, err := h.DB.GetLatestDailyDigest(); err == nil && latest.GeneratedAt.After(since) && !force {
		since = latest.GeneratedAt
	}

	articles, err := h.DB.GetRecentArticlesForDigest(since, maxArticles)
	if err != nil {
		return nil, err
	}

	aiSummarizer, model, err := h.newDigestAISummarizer()
	if err != nil {
		return nil, err
	}

	language, _ := h.DB.GetSetting("language")
	memorySnapshot := h.buildAgentMemorySnapshot()
	perArticleItems := make([]models.DailyDigestArticle, 0, len(articles))

	for _, article := range articles {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if h.AITracker.IsLimitReached() {
			return nil, fmt.Errorf("AI usage limit reached")
		}

		content, _, err := h.GetArticleContent(article.ID)
		if err != nil {
			log.Printf("Daily digest: failed to get article content for %d: %v", article.ID, err)
		}
		if strings.TrimSpace(content) == "" {
			content = article.Title + "\n" + article.URL
		}

		h.AITracker.WaitForRateLimit()
		result, err := aiSummarizer.Complete(
			buildArticleSummarySystemPrompt(language),
			buildArticleSummaryUserPrompt(article, trimRunes(textutil.CleanHTML(content), maxDigestArticleChars), memorySnapshot, language),
		)
		if err != nil {
			return nil, fmt.Errorf("summarize article %d: %w", article.ID, err)
		}

		item := models.DailyDigestArticle{
			ArticleID:   article.ID,
			Title:       article.Title,
			URL:         article.URL,
			FeedTitle:   article.FeedTitle,
			PublishedAt: article.PublishedAt,
			Summary:     result.Summary,
		}
		perArticleItems = append(perArticleItems, item)

		_ = h.DB.UpdateArticleSummary(article.ID, result.Summary)
		h.AITracker.TrackSummary(content, result.Summary)
		_ = h.DB.IncrementStat("ai_summary")
	}

	digestContent := ""
	if len(perArticleItems) == 0 {
		digestContent = buildEmptyDigestContent(language)
	} else {
		if h.AITracker.IsLimitReached() {
			return nil, fmt.Errorf("AI usage limit reached")
		}
		h.AITracker.WaitForRateLimit()
		result, err := aiSummarizer.Complete(
			buildDailyDigestSystemPrompt(language),
			buildDailyDigestUserPrompt(perArticleItems, memorySnapshot, language),
		)
		if err != nil {
			return nil, fmt.Errorf("generate daily digest: %w", err)
		}
		digestContent = result.Summary
		h.AITracker.TrackSummary(buildDailyDigestUserPrompt(perArticleItems, memorySnapshot, language), digestContent)
		_ = h.DB.IncrementStat("ai_summary")
	}

	digest := &models.DailyDigest{
		DigestDate:     today,
		Title:          buildDigestTitle(today, language),
		Content:        digestContent,
		ArticleCount:   len(perArticleItems),
		Model:          model,
		MemorySnapshot: memorySnapshot,
	}

	digestID, err := h.DB.CreateDailyDigest(digest)
	if err != nil {
		return nil, err
	}
	digest.ID = digestID

	for i := range perArticleItems {
		perArticleItems[i].DigestID = digestID
		if err := h.DB.AddDailyDigestArticle(&perArticleItems[i]); err != nil {
			log.Printf("Daily digest: failed to store article summary %d: %v", perArticleItems[i].ArticleID, err)
		}
	}
	digest.Articles = perArticleItems

	return digest, nil
}

func (h *Handler) newDigestAISummarizer() (*summary.AISummarizer, string, error) {
	var apiKey, endpoint, model, customHeaders string
	if h.AIProfileProvider != nil {
		cfg, err := h.AIProfileProvider.GetConfigForFeature(ai.FeatureSummary)
		if err == nil && cfg != nil {
			apiKey = cfg.APIKey
			endpoint = cfg.Endpoint
			model = cfg.Model
			customHeaders = cfg.CustomHeaders
		}
	}

	if apiKey == "" && endpoint == "" {
		apiKey, _ = h.DB.GetEncryptedSetting("ai_api_key")
		endpoint, _ = h.DB.GetSetting("ai_endpoint")
		model, _ = h.DB.GetSetting("ai_model")
		customHeaders, _ = h.DB.GetSetting("ai_custom_headers")
	}

	language, _ := h.DB.GetSetting("language")
	aiSummarizer := summary.NewAISummarizerWithDB(apiKey, endpoint, model, h.DB)
	if customHeaders != "" {
		aiSummarizer.SetCustomHeaders(customHeaders)
	}
	aiSummarizer.SetLanguage(language)

	return aiSummarizer, model, nil
}

func (h *Handler) buildAgentMemorySnapshot() string {
	interests, _ := h.DB.GetSetting("agent_memory_interests")
	dislikes, _ := h.DB.GetSetting("agent_memory_dislikes")
	notes, _ := h.DB.GetSetting("agent_memory_notes")

	parts := []string{}
	if strings.TrimSpace(interests) != "" {
		parts = append(parts, "Interests:\n"+strings.TrimSpace(interests))
	}
	if strings.TrimSpace(dislikes) != "" {
		parts = append(parts, "Not interested in:\n"+strings.TrimSpace(dislikes))
	}
	if strings.TrimSpace(notes) != "" {
		parts = append(parts, "Other memory:\n"+strings.TrimSpace(notes))
	}
	if len(parts) == 0 {
		return "No explicit user memory yet."
	}
	return strings.Join(parts, "\n\n")
}

func getIntSetting(h *Handler, key string, fallback int) int {
	value, err := h.DB.GetSetting(key)
	if err != nil {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func buildArticleSummarySystemPrompt(language string) string {
	if strings.HasPrefix(language, "zh") {
		return "你是一个帮助用户筛选 RSS 信息的研究助理。请总结文章对用户可能有用的内容，而不是只判断是否相关。输出必须简洁、准确、保留重要术语。"
	}
	return "You are a research assistant for an RSS reader. Summarize what may be useful to the user, not merely what is topically related. Be concise, accurate, and preserve important terms."
}

func buildArticleSummaryUserPrompt(article models.Article, content, memory, language string) string {
	if strings.HasPrefix(language, "zh") {
		return fmt.Sprintf(`用户长期记忆：
%s

文章：
标题：%s
来源：%s
链接：%s

正文：
%s

请用中文输出：
1. 3-5 条要点摘要。
2. 这篇文章对用户可能有用的原因。
3. 如果文章价值不高，也请直说，并说明原因。`, memory, article.Title, article.FeedTitle, article.URL, content)
	}

	return fmt.Sprintf(`User memory:
%s

Article:
Title: %s
Source: %s
URL: %s

Content:
%s

Output in English:
1. 3-5 concise bullets.
2. Why this article may be useful to the user.
3. If it is low value, say so and explain why.`, memory, article.Title, article.FeedTitle, article.URL, content)
}

func buildDailyDigestSystemPrompt(language string) string {
	if strings.HasPrefix(language, "zh") {
		return "你是一个像 AI Agent 一样工作的 RSS 阅读助手。你的目标是从当天新增信息中提取对用户真正有用的内容，给出个性化优先级和建设性建议。输出 Markdown。"
	}
	return "You are an agentic RSS reading assistant. Extract what is genuinely useful from newly collected information, prioritize it for the user, and give constructive suggestions. Output Markdown."
}

func buildDailyDigestUserPrompt(items []models.DailyDigestArticle, memory, language string) string {
	var b strings.Builder
	if strings.HasPrefix(language, "zh") {
		b.WriteString("用户长期记忆：\n")
		b.WriteString(memory)
		b.WriteString("\n\n今天新增文章摘要：\n")
		for i, item := range items {
			fmt.Fprintf(&b, "\n%d. 标题：%s\n来源：%s\n链接：%s\n摘要：%s\n", i+1, item.Title, item.FeedTitle, item.URL, item.Summary)
		}
		b.WriteString("\n请生成中文每日简报，包含：\n- 今日最值得读的 3-7 篇及理由\n- 可以跳过或低优先级的内容\n- 对用户接下来学习、研究或行动的建议\n- 一句话总览\n")
		return b.String()
	}

	b.WriteString("User memory:\n")
	b.WriteString(memory)
	b.WriteString("\n\nNew article summaries:\n")
	for i, item := range items {
		fmt.Fprintf(&b, "\n%d. Title: %s\nSource: %s\nURL: %s\nSummary: %s\n", i+1, item.Title, item.FeedTitle, item.URL, item.Summary)
	}
	b.WriteString("\nCreate a daily briefing with:\n- The 3-7 most useful reads and why\n- Low-priority or skippable items\n- Constructive next actions for the user\n- A one-sentence overview\n")
	return b.String()
}

func buildEmptyDigestContent(language string) string {
	if strings.HasPrefix(language, "zh") {
		return "今天还没有发现新的可总结文章。"
	}
	return "No new articles were found for today's digest."
}

func buildDigestTitle(date, language string) string {
	if strings.HasPrefix(language, "zh") {
		return date + " 每日 AI 简报"
	}
	return date + " Daily AI Briefing"
}

func trimRunes(value string, max int) string {
	if max <= 0 || utf8.RuneCountInString(value) <= max {
		return value
	}
	runes := []rune(value)
	return string(runes[:max]) + "\n\n[Content truncated]"
}
