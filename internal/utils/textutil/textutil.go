// Package textutil provides text processing utilities including HTML cleaning,
// markdown rendering, and text sanitization.
package textutil

import (
	"fmt"
	stdhtml "html"
	"regexp"
	"strings"

	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// selfClosingTags is the list of HTML self-closing tags to handle
const selfClosingTags = "img|br|hr|input|meta|link"

// Compile regex patterns once at package initialization for better performance
var (
	// Matches malformed opening tags like <p-->, <div-->
	malformedTagRegex = regexp.MustCompile(`<([a-zA-Z][a-zA-Z0-9]*)\s*--+>`)

	// Matches malformed self-closing tags with attributes like <img src="..." -->
	malformedSelfClosingWithAttrs = regexp.MustCompile(`<(` + selfClosingTags + `)\s+([^<>]+?)--+>`)

	// Matches malformed self-closing tags without attributes like <br-->
	malformedSelfClosingNoAttrs = regexp.MustCompile(`<(` + selfClosingTags + `)\s*--+>`)

	// Matches style attributes in HTML tags
	styleAttrRegex = regexp.MustCompile(`\s+style\s*=\s*"[^"]*"`)

	// Alternative style attribute with single quotes
	styleAttrSingleQuoteRegex = regexp.MustCompile(`\s+style\s*=\s*'[^']*'`)

	// Matches class attributes in HTML tags
	classAttrRegex = regexp.MustCompile(`\s+class\s*=\s*"[^"]*"`)

	// Alternative class attribute with single quotes
	classAttrSingleQuoteRegex = regexp.MustCompile(`\s+class\s*=\s*'[^']*'`)

	// Matches <style> tags and their content
	styleTagRegex = regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)

	// Matches <script> tags and their content
	scriptTagRegex = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
)

type protectedMathSegment struct {
	token string
	value string
}

type mathSegmentMatch struct {
	start int
	end   int
}

// CleanHTML sanitizes HTML content by fixing common malformed patterns
// and removing unwanted inline styles, classes, and scripts.
func CleanHTML(htmlContent string) string {
	if htmlContent == "" {
		return htmlContent
	}

	// Fix malformed opening tags like <p--> to <p>
	htmlContent = malformedTagRegex.ReplaceAllString(htmlContent, "<$1>")

	// Fix malformed self-closing tags
	htmlContent = malformedSelfClosingWithAttrs.ReplaceAllString(htmlContent, "<$1 $2>")
	htmlContent = malformedSelfClosingNoAttrs.ReplaceAllString(htmlContent, "<$1>")

	// Remove inline style attributes
	htmlContent = styleAttrRegex.ReplaceAllString(htmlContent, "")
	htmlContent = styleAttrSingleQuoteRegex.ReplaceAllString(htmlContent, "")

	// Remove class attributes
	htmlContent = classAttrRegex.ReplaceAllString(htmlContent, "")
	htmlContent = classAttrSingleQuoteRegex.ReplaceAllString(htmlContent, "")

	// Remove <style> tags and their content
	htmlContent = styleTagRegex.ReplaceAllString(htmlContent, "")

	// Remove <script> tags and their content
	htmlContent = scriptTagRegex.ReplaceAllString(htmlContent, "")

	return strings.TrimSpace(htmlContent)
}

// RenderMarkdown converts markdown text to safe HTML.
func RenderMarkdown(markdownText string) string {
	if markdownText == "" {
		return ""
	}

	protectedText, mathSegments := protectMathSegments(markdownText)

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{Flags: htmlFlags}
	renderer := mdhtml.NewRenderer(opts)

	htmlBytes := markdown.ToHTML([]byte(protectedText), p, renderer)
	return restoreMathSegments(string(htmlBytes), mathSegments)
}

// RenderMarkdownInline converts markdown to HTML without wrapping <p> tags.
func RenderMarkdownInline(markdownText string) string {
	if markdownText == "" {
		return ""
	}

	protectedText, mathSegments := protectMathSegments(markdownText)

	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)

	htmlFlags := mdhtml.CommonFlags | mdhtml.HrefTargetBlank
	opts := mdhtml.RendererOptions{Flags: htmlFlags}
	renderer := mdhtml.NewRenderer(opts)

	htmlBytes := markdown.ToHTML([]byte(protectedText), p, renderer)

	result := string(htmlBytes)
	result = strings.TrimPrefix(result, "<p>")
	result = strings.TrimSuffix(result, "</p>")
	result = strings.TrimSuffix(result, "<p />")

	return restoreMathSegments(result, mathSegments)
}

func protectMathSegments(text string) (string, []protectedMathSegment) {
	var segments []protectedMathSegment
	var result strings.Builder
	index := 0

	for index < len(text) {
		match := findNextMathSegment(text, index)
		if match == nil {
			result.WriteString(text[index:])
			break
		}

		result.WriteString(text[index:match.start])

		token := fmt.Sprintf("MRRSSMATHPLACEHOLDER%06d", len(segments))
		segments = append(segments, protectedMathSegment{
			token: token,
			value: text[match.start:match.end],
		})
		result.WriteString(token)
		index = match.end
	}

	return result.String(), segments
}

func restoreMathSegments(htmlContent string, segments []protectedMathSegment) string {
	for _, segment := range segments {
		htmlContent = strings.ReplaceAll(htmlContent, segment.token, stdhtml.EscapeString(segment.value))
	}
	return htmlContent
}

func findNextMathSegment(text string, from int) *mathSegmentMatch {
	var best *mathSegmentMatch

	candidates := []func(string, int) *mathSegmentMatch{
		func(source string, start int) *mathSegmentMatch {
			return findNextEnvironmentSegment(source, start, `\\begin{`, `\\end{`)
		},
		func(source string, start int) *mathSegmentMatch {
			return findNextEnvironmentSegment(source, start, `\begin{`, `\end{`)
		},
		func(source string, start int) *mathSegmentMatch {
			return findNextDelimitedSegment(source, start, `\\[`, `\\]`)
		},
		func(source string, start int) *mathSegmentMatch {
			return findNextDelimitedSegment(source, start, `\[`, `\]`)
		},
		func(source string, start int) *mathSegmentMatch {
			return findNextDelimitedSegment(source, start, `$$`, `$$`)
		},
		func(source string, start int) *mathSegmentMatch {
			return findNextDelimitedSegment(source, start, `\\(`, `\\)`)
		},
		func(source string, start int) *mathSegmentMatch {
			return findNextDelimitedSegment(source, start, `\(`, `\)`)
		},
		findNextDollarSegment,
	}

	for _, findCandidate := range candidates {
		candidate := findCandidate(text, from)
		if candidate == nil {
			continue
		}
		if best == nil || candidate.start < best.start || (candidate.start == best.start && candidate.end > best.end) {
			best = candidate
		}
	}

	return best
}

func findNextEnvironmentSegment(text string, from int, beginPrefix string, endPrefix string) *mathSegmentMatch {
	relativeStart := strings.Index(text[from:], beginPrefix)
	if relativeStart == -1 {
		return nil
	}

	start := from + relativeStart
	envNameStart := start + len(beginPrefix)
	envNameEndRelative := strings.IndexByte(text[envNameStart:], '}')
	if envNameEndRelative == -1 {
		return nil
	}

	envNameEnd := envNameStart + envNameEndRelative
	envName := text[envNameStart:envNameEnd]
	if envName == "" {
		return nil
	}

	endMarker := endPrefix + envName + "}"
	searchFrom := envNameEnd + 1
	relativeEnd := strings.Index(text[searchFrom:], endMarker)
	if relativeEnd == -1 {
		return nil
	}

	end := searchFrom + relativeEnd + len(endMarker)
	return &mathSegmentMatch{start: start, end: end}
}

func findNextDelimitedSegment(text string, from int, open string, close string) *mathSegmentMatch {
	relativeStart := strings.Index(text[from:], open)
	if relativeStart == -1 {
		return nil
	}

	start := from + relativeStart
	searchFrom := start + len(open)
	relativeEnd := strings.Index(text[searchFrom:], close)
	if relativeEnd == -1 {
		return nil
	}

	end := searchFrom + relativeEnd + len(close)
	return &mathSegmentMatch{start: start, end: end}
}

func findNextDollarSegment(text string, from int) *mathSegmentMatch {
	for {
		relativeStart := strings.IndexByte(text[from:], '$')
		if relativeStart == -1 {
			return nil
		}

		start := from + relativeStart
		if strings.HasPrefix(text[start:], "$$") {
			from = start + 2
			continue
		}

		searchFrom := start + 1
		relativeEnd := strings.IndexByte(text[searchFrom:], '$')
		if relativeEnd == -1 {
			return nil
		}

		end := searchFrom + relativeEnd + 1
		if end > start+2 && !strings.Contains(text[start+1:end-1], "\n") {
			return &mathSegmentMatch{start: start, end: end}
		}

		from = end
	}
}

// SanitizeHTML removes potentially dangerous HTML tags and attributes.
func SanitizeHTML(htmlContent string) string {
	if htmlContent == "" {
		return ""
	}

	// Remove script tags
	scriptRegex := regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	htmlContent = scriptRegex.ReplaceAllString(htmlContent, "")

	// Remove style tags
	styleRegex := regexp.MustCompile(`(?i)<style[^>]*>.*?</style>`)
	htmlContent = styleRegex.ReplaceAllString(htmlContent, "")

	// Remove iframe tags
	iframeRegex := regexp.MustCompile(`(?i)<iframe[^>]*>.*?</iframe>`)
	htmlContent = iframeRegex.ReplaceAllString(htmlContent, "")

	// Remove on* event handlers
	eventRegex := regexp.MustCompile(`(?i)\s+on\w+\s*=\s*["'][^"']*["']`)
	htmlContent = eventRegex.ReplaceAllString(htmlContent, "")

	// Remove javascript: protocol
	jsRegex := regexp.MustCompile(`(?i)javascript:`)
	htmlContent = jsRegex.ReplaceAllString(htmlContent, "")

	return htmlContent
}

// ConvertMarkdownToHTML converts markdown to safe HTML with sanitization.
func ConvertMarkdownToHTML(markdownText string) string {
	if markdownText == "" {
		return ""
	}

	htmlContent := RenderMarkdown(markdownText)
	return SanitizeHTML(htmlContent)
}
