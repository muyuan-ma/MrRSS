package core

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestFetchFullArticleContentRetriesForbiddenWithCookie(t *testing.T) {
	var hits int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)

		if !strings.Contains(r.Header.Get("User-Agent"), "Mozilla/5.0") {
			t.Errorf("User-Agent header = %q, want browser-like header", r.Header.Get("User-Agent"))
		}
		if !strings.Contains(r.Header.Get("Accept"), "text/html") {
			t.Errorf("Accept header = %q, want text/html", r.Header.Get("Accept"))
		}

		if _, err := r.Cookie("article_gate"); err != nil {
			http.SetCookie(w, &http.Cookie{
				Name:  "article_gate",
				Value: "open",
				Path:  "/",
			})
			http.Error(w, "cookie required", http.StatusForbidden)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(testFullArticleHTML()))
	}))
	defer server.Close()

	h := &Handler{}
	content, err := h.FetchFullArticleContent(server.URL + "/archives/11772")
	if err != nil {
		t.Fatalf("FetchFullArticleContent returned error: %v", err)
	}

	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Fatalf("request count = %d, want 2", got)
	}
	if !strings.Contains(content, "cookie protected article") {
		t.Fatalf("content does not include extracted article text: %s", content)
	}
}

func testFullArticleHTML() string {
	body := strings.Repeat(
		"This cookie protected article has enough readable body text for the parser to select the main content. ",
		20,
	)

	return `<!doctype html>
<html>
<head>
  <title>Cookie Protected Article</title>
</head>
<body>
  <nav>navigation clutter</nav>
  <article>
    <h1>Cookie Protected Article</h1>
    <p>` + body + `</p>
    <p>This is the unique cookie protected article phrase.</p>
  </article>
</body>
</html>`
}
