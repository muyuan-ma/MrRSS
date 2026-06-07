package core

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"
)

// startDailyDigestScheduler checks once per minute whether today's AI digest should run.
func (h *Handler) startDailyDigestScheduler(ctx context.Context) {
	log.Println("Starting daily AI digest scheduler")

	runIfDue := func() {
		enabled, _ := h.DB.GetSetting("agent_digest_enabled")
		if enabled != "true" {
			return
		}

		now := time.Now()
		if !isDigestTimeDue(now, getDigestTime(h)) {
			return
		}

		today := now.Format("2006-01-02")
		if _, err := h.DB.GetDailyDigestByDate(today); err == nil {
			return
		} else if err != nil && err != sql.ErrNoRows {
			log.Printf("Daily digest: failed to check existing digest: %v", err)
			return
		}

		go func() {
			digest, err := h.GenerateDailyDigest(ctx, false)
			if err != nil {
				log.Printf("Daily digest generation failed: %v", err)
				return
			}
			log.Printf("Daily digest generated for %s with %d articles", digest.DigestDate, digest.ArticleCount)
		}()
	}

	runIfDue()

	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping daily AI digest scheduler")
			return
		case <-ticker.C:
			runIfDue()
		}
	}
}

func getDigestTime(h *Handler) string {
	value, _ := h.DB.GetSetting("agent_digest_time")
	value = strings.TrimSpace(value)
	if value == "" {
		return "08:30"
	}
	return value
}

func isDigestTimeDue(now time.Time, configured string) bool {
	target, err := time.Parse("15:04", configured)
	if err != nil {
		target, _ = time.Parse("15:04", "08:30")
	}
	due := time.Date(now.Year(), now.Month(), now.Day(), target.Hour(), target.Minute(), 0, 0, now.Location())
	return !now.Before(due)
}
