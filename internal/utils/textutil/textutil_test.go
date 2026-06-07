package textutil

import (
	"strings"
	"testing"
)

func TestConvertMarkdownToHTMLPreservesLatexInlineMath(t *testing.T) {
	input := `公式为 \(\alpha = \left\{ d_{out} / d_{in} \right.\)，同时 **Markdown** 仍然工作。`

	html := ConvertMarkdownToHTML(input)

	if !strings.Contains(html, `\(\alpha = \left\{ d_{out} / d_{in} \right.\)`) {
		t.Fatalf("expected inline LaTeX to be preserved, got: %s", html)
	}
	if strings.Contains(html, "<em>") {
		t.Fatalf("expected LaTeX underscores not to become emphasis, got: %s", html)
	}
	if !strings.Contains(html, "<strong>Markdown</strong>") {
		t.Fatalf("expected ordinary markdown to render, got: %s", html)
	}
}

func TestConvertMarkdownToHTMLPreservesLatexEnvironment(t *testing.T) {
	input := `\begin{equation}\newcommand{\tr}{\mathop{\text{tr}}}\begin{aligned}\mathcal{F} &= \tr(\boldsymbol{\Sigma}_p) \label{eq:test}\end{aligned}\end{equation}`

	html := ConvertMarkdownToHTML(input)

	if !strings.Contains(html, `\begin{equation}`) || !strings.Contains(html, `\end{equation}`) {
		t.Fatalf("expected LaTeX environment to be preserved, got: %s", html)
	}
	if !strings.Contains(html, `\newcommand{\tr}`) {
		t.Fatalf("expected local macro definition to survive markdown conversion, got: %s", html)
	}
	if strings.Contains(html, "<em>") {
		t.Fatalf("expected LaTeX underscores not to become emphasis, got: %s", html)
	}
}

func TestConvertMarkdownToHTMLPreservesDoubleEscapedLatex(t *testing.T) {
	input := `摘要里可能出现 \\(\max(1,\cdot)\\) 这样的转义公式。`

	html := ConvertMarkdownToHTML(input)

	if !strings.Contains(html, `\\(\max(1,\cdot)\\)`) {
		t.Fatalf("expected double-escaped LaTeX delimiters to be preserved, got: %s", html)
	}
}
