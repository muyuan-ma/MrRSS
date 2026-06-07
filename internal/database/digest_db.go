package database

import (
	"database/sql"
	"time"

	"MrRSS/internal/models"
)

// GetRecentArticlesForDigest returns visible articles published after the given time.
func (db *DB) GetRecentArticlesForDigest(since time.Time, limit int) ([]models.Article, error) {
	db.WaitForReady()
	if limit <= 0 {
		limit = 20
	}

	rows, err := db.Query(`
		SELECT a.id, a.feed_id, a.title, a.url, a.image_url, a.audio_url, a.video_url,
			a.published_at, a.is_read, a.is_favorite, a.is_hidden, a.is_read_later,
			a.translated_title, a.summary, a.freshrss_item_id, f.title, a.author
		FROM articles a
		JOIN feeds f ON a.feed_id = f.id
		WHERE a.is_hidden = 0
			AND COALESCE(f.hide_from_timeline, 0) = 0
			AND a.published_at >= ?
		ORDER BY a.published_at ASC
		LIMIT ?
	`, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	articles := make([]models.Article, 0)
	for rows.Next() {
		var a models.Article
		var imageURL, audioURL, videoURL, translatedTitle, summary, freshrssItemID, author sql.NullString
		var publishedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.FeedID, &a.Title, &a.URL, &imageURL, &audioURL, &videoURL, &publishedAt, &a.IsRead, &a.IsFavorite, &a.IsHidden, &a.IsReadLater, &translatedTitle, &summary, &freshrssItemID, &a.FeedTitle, &author); err != nil {
			return nil, err
		}
		a.ImageURL = imageURL.String
		a.AudioURL = audioURL.String
		a.VideoURL = videoURL.String
		if publishedAt.Valid {
			a.PublishedAt = publishedAt.Time
		}
		a.TranslatedTitle = translatedTitle.String
		a.Summary = summary.String
		a.FreshRSSItemID = freshrssItemID.String
		a.Author = author.String
		articles = append(articles, a)
	}

	return articles, rows.Err()
}

// CreateDailyDigest stores a generated daily digest.
func (db *DB) CreateDailyDigest(digest *models.DailyDigest) (int64, error) {
	db.WaitForReady()
	var existingID int64
	if err := db.QueryRow("SELECT id FROM daily_digests WHERE digest_date = ?", digest.DigestDate).Scan(&existingID); err == nil {
		_, err = db.Exec(`
			UPDATE daily_digests
			SET title = ?, content = ?, article_count = ?, model = ?, memory_snapshot = ?,
				generated_at = CURRENT_TIMESTAMP, notified_at = NULL
			WHERE id = ?
		`, digest.Title, digest.Content, digest.ArticleCount, digest.Model, digest.MemorySnapshot, existingID)
		if err != nil {
			return 0, err
		}
		_, _ = db.Exec("DELETE FROM daily_digest_articles WHERE digest_id = ?", existingID)
		return existingID, nil
	}

	result, err := db.Exec(`
		INSERT INTO daily_digests
			(digest_date, title, content, article_count, model, memory_snapshot, generated_at, notified_at)
		VALUES (?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, NULL)
	`, digest.DigestDate, digest.Title, digest.Content, digest.ArticleCount, digest.Model, digest.MemorySnapshot)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// AddDailyDigestArticle stores a per-article summary for a digest.
func (db *DB) AddDailyDigestArticle(item *models.DailyDigestArticle) error {
	db.WaitForReady()
	_, err := db.Exec(`
		INSERT OR REPLACE INTO daily_digest_articles
			(digest_id, article_id, summary, recommendation, relevance_score, created_at)
		VALUES (?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
	`, item.DigestID, item.ArticleID, item.Summary, item.Recommendation, item.RelevanceScore)
	return err
}

// GetDailyDigestByDate returns the digest for a date if it exists.
func (db *DB) GetDailyDigestByDate(date string) (*models.DailyDigest, error) {
	db.WaitForReady()
	row := db.QueryRow(`
		SELECT id, digest_date, title, content, article_count, model, memory_snapshot, generated_at, notified_at
		FROM daily_digests
		WHERE digest_date = ?
	`, date)
	return scanDailyDigest(row)
}

// GetLatestDailyDigest returns the most recently generated digest.
func (db *DB) GetLatestDailyDigest() (*models.DailyDigest, error) {
	db.WaitForReady()
	row := db.QueryRow(`
		SELECT id, digest_date, title, content, article_count, model, memory_snapshot, generated_at, notified_at
		FROM daily_digests
		ORDER BY digest_date DESC, generated_at DESC
		LIMIT 1
	`)
	return scanDailyDigest(row)
}

// GetDailyDigests lists recent digests.
func (db *DB) GetDailyDigests(limit int) ([]models.DailyDigest, error) {
	db.WaitForReady()
	if limit <= 0 {
		limit = 20
	}
	rows, err := db.Query(`
		SELECT id, digest_date, title, content, article_count, model, memory_snapshot, generated_at, notified_at
		FROM daily_digests
		ORDER BY digest_date DESC, generated_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	digests := make([]models.DailyDigest, 0)
	for rows.Next() {
		digest, err := scanDailyDigestRows(rows)
		if err != nil {
			return nil, err
		}
		digests = append(digests, *digest)
	}
	return digests, rows.Err()
}

// GetDailyDigestArticles lists the per-article summaries for a digest.
func (db *DB) GetDailyDigestArticles(digestID int64) ([]models.DailyDigestArticle, error) {
	db.WaitForReady()
	rows, err := db.Query(`
		SELECT dda.id, dda.digest_id, dda.article_id, a.title, a.url, f.title, a.published_at,
			dda.summary, dda.recommendation, dda.relevance_score, dda.created_at
		FROM daily_digest_articles dda
		JOIN articles a ON dda.article_id = a.id
		JOIN feeds f ON a.feed_id = f.id
		WHERE dda.digest_id = ?
		ORDER BY a.published_at ASC
	`, digestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]models.DailyDigestArticle, 0)
	for rows.Next() {
		var item models.DailyDigestArticle
		var publishedAt, createdAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.DigestID, &item.ArticleID, &item.Title, &item.URL, &item.FeedTitle, &publishedAt, &item.Summary, &item.Recommendation, &item.RelevanceScore, &createdAt); err != nil {
			return nil, err
		}
		if publishedAt.Valid {
			item.PublishedAt = publishedAt.Time
		}
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// MarkDailyDigestNotified records that the frontend has notified the user.
func (db *DB) MarkDailyDigestNotified(id int64) error {
	db.WaitForReady()
	_, err := db.Exec("UPDATE daily_digests SET notified_at = CURRENT_TIMESTAMP WHERE id = ?", id)
	return err
}

func scanDailyDigest(row *sql.Row) (*models.DailyDigest, error) {
	var digest models.DailyDigest
	var generatedAt sql.NullTime
	var notifiedAt sql.NullTime
	if err := row.Scan(&digest.ID, &digest.DigestDate, &digest.Title, &digest.Content, &digest.ArticleCount, &digest.Model, &digest.MemorySnapshot, &generatedAt, &notifiedAt); err != nil {
		return nil, err
	}
	if generatedAt.Valid {
		digest.GeneratedAt = generatedAt.Time
	}
	if notifiedAt.Valid {
		digest.NotifiedAt = &notifiedAt.Time
	}
	return &digest, nil
}

func scanDailyDigestRows(rows *sql.Rows) (*models.DailyDigest, error) {
	var digest models.DailyDigest
	var generatedAt sql.NullTime
	var notifiedAt sql.NullTime
	if err := rows.Scan(&digest.ID, &digest.DigestDate, &digest.Title, &digest.Content, &digest.ArticleCount, &digest.Model, &digest.MemorySnapshot, &generatedAt, &notifiedAt); err != nil {
		return nil, err
	}
	if generatedAt.Valid {
		digest.GeneratedAt = generatedAt.Time
	}
	if notifiedAt.Valid {
		digest.NotifiedAt = &notifiedAt.Time
	}
	return &digest, nil
}
