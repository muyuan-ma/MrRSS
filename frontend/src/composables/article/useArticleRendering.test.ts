import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import katex from 'katex';
import { useArticleRendering } from './useArticleRendering';

// Mock KaTeX
vi.mock('katex', () => ({
  default: {
    render: vi.fn((formula, element) => {
      element.innerHTML = `<span class="katex">${formula}</span>`;
    }),
  },
}));

describe('useArticleRendering', () => {
  let container: HTMLElement;

  beforeEach(() => {
    // Create a fresh container for each test
    container = document.createElement('div');
    container.className = 'prose-content';
    document.body.appendChild(container);
  });

  afterEach(() => {
    // Clean up
    document.body.removeChild(container);
  });

  describe('renderMathFormulas', () => {
    it('should render inline math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>The formula is $E = mc^2$ here.</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('E = mc^2');
    });

    it('should render parenthesized inline math with ordinary parentheses inside', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>截断操作是 \\(\\max(1,\\cdot)\\) 的形式。</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('\\max(1,\\cdot)');
      expect(container.innerHTML).not.toContain('\\(\\max');
    });

    it('should render display math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>$$\\int_{0}^{\\infty} e^{-x} dx$$</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
    });

    it('should not render math inside code blocks', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<pre><code>$E = mc^2$</code></pre>';
      const originalHTML = container.innerHTML;
      renderMathFormulas(container);

      // HTML should remain unchanged
      expect(container.innerHTML).toBe(originalHTML);
    });

    it('should handle text without math formulas', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>This is just plain text.</p>';
      const originalHTML = container.innerHTML;
      renderMathFormulas(container);

      // HTML should remain unchanged
      expect(container.innerHTML).toBe(originalHTML);
    });

    it('should handle mixed content with both inline and display math', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>Inline $a = b$ and display $$c = d$$</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('katex-display');
    });

    it('should render LaTeX equation environments as display math', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>\\begin{equation}p_i = \\frac{\\sigma_i}{\\sum_j \\sigma_j}\\end{equation}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('p_i = \\frac');
      expect(container.innerHTML).not.toContain('\\begin{equation}');
    });

    it('should render LaTeX align environments with an aligned wrapper', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>\\begin{align}a &= b \\\\ c &= d\\end{align}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('\\begin{aligned}');
    });

    it('should strip local macro definitions before rendering aligned environments', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>\\begin{aligned}\\newcommand{\\msign}{\\mathop{\\text{msign}}}\\boldsymbol{M}_t &= \\msign(\\boldsymbol{M}_{t-1})\\end{aligned}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).not.toContain('\\newcommand');
      expect(container.innerHTML).toContain('\\msign');
    });

    it('should render wrapped LaTeX environments without duplicate delimiters', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>\\[\\begin{equation}x = y\\end{equation}\\]</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('x = y');
      expect(container.innerHTML).not.toContain('\\[');
      expect(container.innerHTML).not.toContain('\\begin{equation}');
    });

    it('should strip nested escaped math delimiters before rendering', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>$$\\\\(\\alpha = \\sqrt{x}\\\\)$$</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('\\alpha = \\sqrt{x}');
      expect(container.innerHTML).not.toContain('\\\\(');
      expect(container.innerHTML).not.toContain('\\\\)');
    });

    it('should strip LaTeX labels and reference annotations', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>\\begin{equation}\\begin{aligned}a &= b \\label{eq:test} \\labeleq : E-Hq \\eqrefeq : w-p-q\\end{aligned}\\end{equation}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('a &amp;= b');
      expect(container.innerHTML).not.toContain('\\label');
      expect(container.innerHTML).not.toContain('\\labeleq');
      expect(container.innerHTML).not.toContain('\\eqrefeq');
    });

    it('should render double-escaped inline math delimiters from summaries', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>截断操作是 \\\\(\\max(1,\\cdot)\\\\) 的形式。</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('\\max(1,\\cdot)');
      expect(container.innerHTML).not.toContain('\\\\(');
      expect(container.innerHTML).not.toContain('\\\\)');
    });

    it('should render double-escaped LaTeX environments from summaries', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>公式为 \\\\begin{equation}\\\\begin{aligned}a &= b \\\\[5pt] c &= d\\\\end{aligned}\\\\end{equation}。</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('\\begin{aligned}');
      expect(container.innerHTML).not.toContain('\\\\begin{equation}');
    });

    it('should render LaTeX environments split by HTML line breaks', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>\\begin{equation}<br>\\begin{aligned}<br>a &= b \\\\[4pt]<br>c &= d<br>\\end{aligned}<br>\\end{equation}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('a &amp;= b');
      expect(container.innerHTML).not.toContain('\\begin{equation}');
    });

    it('should render display environments inside mixed prose paragraphs', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>那么\\begin{equation}\\begin{aligned}a &= b \\\\[4pt] c &= d\\end{aligned}\\end{equation}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('那么');
      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).not.toContain('\\begin{equation}');
    });

    it('should render dangling dollar math fragments split across blocks', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>此时的SVD要写成 $\\boldsymbol{U}</p><p>\\boldsymbol{V}^{\\top}$，于是</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      expect(container.innerHTML).toContain('此时的SVD要写成');
      expect(container.innerHTML).toContain('，于是');
      expect(container.innerHTML).not.toContain('$\\boldsymbol{U}');
      expect(container.innerHTML).not.toContain('\\boldsymbol{V}^{\\top}$');
    });

    it('should render bare LaTeX display blocks', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML =
        '<p>\\boldsymbol{\\beta} = \\mathop{\\text{desc_sort}}(\\boldsymbol{s}, \\text{axis=0})_{[lk/n:lk/n+1]}</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-display');
      expect(container.innerHTML).toContain('desc\\_sort');
    });

    it('should render custom Muon macros', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<p>The update uses $\\Phi = \\msign(M) + \\tr(W)$ here.</p>';
      renderMathFormulas(container);

      expect(container.innerHTML).toContain('katex-inline');
      const lastRenderCall = vi.mocked(katex.render).mock.calls.at(-1);
      expect(lastRenderCall?.[2]).toMatchObject({
        macros: {
          '\\msign': '\\operatorname{msign}',
          '\\tr': '\\operatorname{tr}',
        },
      });
    });

    it('should not crash on empty container', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '';
      expect(() => renderMathFormulas(container)).not.toThrow();
    });

    it('should skip math in script tags', () => {
      const { renderMathFormulas } = useArticleRendering();

      container.innerHTML = '<script>var x = "$E = mc^2$";</script><p>Normal text</p>';
      const scriptContent = container.querySelector('script')?.textContent;

      renderMathFormulas(container);

      // Script content should remain unchanged
      expect(container.querySelector('script')?.textContent).toBe(scriptContent);
    });
  });

  describe('enhanceRendering', () => {
    it('should enhance rendering for the specified container', async () => {
      const { enhanceRendering } = useArticleRendering();

      container.innerHTML = '<p>The equation is $E = mc^2$.</p>';
      await enhanceRendering('.prose-content');

      expect(container.innerHTML).toContain('katex');
    });

    it('should handle non-existent container gracefully', async () => {
      const { enhanceRendering } = useArticleRendering();

      // Should not throw error
      await expect(enhanceRendering('.non-existent')).resolves.not.toThrow();
    });
  });
});
