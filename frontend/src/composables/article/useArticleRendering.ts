import { nextTick } from 'vue';
import katex from 'katex';
import hljs from 'highlight.js/lib/core';

// Import commonly used languages
import javascript from 'highlight.js/lib/languages/javascript';
import typescript from 'highlight.js/lib/languages/typescript';
import python from 'highlight.js/lib/languages/python';
import java from 'highlight.js/lib/languages/java';
import cpp from 'highlight.js/lib/languages/cpp';
import c from 'highlight.js/lib/languages/c';
import csharp from 'highlight.js/lib/languages/csharp';
import go from 'highlight.js/lib/languages/go';
import rust from 'highlight.js/lib/languages/rust';
import ruby from 'highlight.js/lib/languages/ruby';
import php from 'highlight.js/lib/languages/php';
import swift from 'highlight.js/lib/languages/swift';
import kotlin from 'highlight.js/lib/languages/kotlin';
import scala from 'highlight.js/lib/languages/scala';
import sql from 'highlight.js/lib/languages/sql';
import bash from 'highlight.js/lib/languages/bash';
import shell from 'highlight.js/lib/languages/shell';
import powershell from 'highlight.js/lib/languages/powershell';
import json from 'highlight.js/lib/languages/json';
import xml from 'highlight.js/lib/languages/xml';
import yaml from 'highlight.js/lib/languages/yaml';
import markdown from 'highlight.js/lib/languages/markdown';
import css from 'highlight.js/lib/languages/css';
import scss from 'highlight.js/lib/languages/scss';
import less from 'highlight.js/lib/languages/less';
import dockerfile from 'highlight.js/lib/languages/dockerfile';
import nginx from 'highlight.js/lib/languages/nginx';
import ini from 'highlight.js/lib/languages/ini';
import diff from 'highlight.js/lib/languages/diff';
import plaintext from 'highlight.js/lib/languages/plaintext';

// Register languages
hljs.registerLanguage('javascript', javascript);
hljs.registerLanguage('js', javascript);
hljs.registerLanguage('typescript', typescript);
hljs.registerLanguage('ts', typescript);
hljs.registerLanguage('python', python);
hljs.registerLanguage('py', python);
hljs.registerLanguage('java', java);
hljs.registerLanguage('cpp', cpp);
hljs.registerLanguage('c', c);
hljs.registerLanguage('csharp', csharp);
hljs.registerLanguage('cs', csharp);
hljs.registerLanguage('go', go);
hljs.registerLanguage('rust', rust);
hljs.registerLanguage('rs', rust);
hljs.registerLanguage('ruby', ruby);
hljs.registerLanguage('rb', ruby);
hljs.registerLanguage('php', php);
hljs.registerLanguage('swift', swift);
hljs.registerLanguage('kotlin', kotlin);
hljs.registerLanguage('kt', kotlin);
hljs.registerLanguage('scala', scala);
hljs.registerLanguage('sql', sql);
hljs.registerLanguage('bash', bash);
hljs.registerLanguage('sh', bash);
hljs.registerLanguage('shell', shell);
hljs.registerLanguage('powershell', powershell);
hljs.registerLanguage('ps1', powershell);
hljs.registerLanguage('json', json);
hljs.registerLanguage('xml', xml);
hljs.registerLanguage('html', xml);
hljs.registerLanguage('yaml', yaml);
hljs.registerLanguage('yml', yaml);
hljs.registerLanguage('markdown', markdown);
hljs.registerLanguage('md', markdown);
hljs.registerLanguage('css', css);
hljs.registerLanguage('scss', scss);
hljs.registerLanguage('less', less);
hljs.registerLanguage('dockerfile', dockerfile);
hljs.registerLanguage('docker', dockerfile);
hljs.registerLanguage('nginx', nginx);
hljs.registerLanguage('ini', ini);
hljs.registerLanguage('diff', diff);
hljs.registerLanguage('plaintext', plaintext);
hljs.registerLanguage('text', plaintext);

const katexMacros = {
  '\\msign': '\\operatorname{msign}',
  '\\sign': '\\operatorname{sign}',
  '\\diag': '\\operatorname{diag}',
  '\\rank': '\\operatorname{rank}',
  '\\tr': '\\operatorname{tr}',
  '\\trace': '\\operatorname{tr}',
  '\\argmax': '\\operatorname*{arg\\,max}',
  '\\argmin': '\\operatorname*{arg\\,min}',
};

const simpleInlineTags = new Set([
  'BR',
  'SPAN',
  'B',
  'I',
  'EM',
  'STRONG',
  'SUB',
  'SUP',
  'SMALL',
  'MARK',
  'U',
  'S',
]);

/**
 * Composable for enhanced article content rendering
 * Handles math formulas, code syntax highlighting, and other advanced rendering
 */
export function useArticleRendering() {
  interface MathMatch {
    start: number;
    end: number;
    math: string;
    isDisplay: boolean;
    environment?: string;
  }

  function normalizeEscapedCommandBackslashes(math: string): string {
    return math.replace(/\\\\(?=[A-Za-z])/g, '\\');
  }

  function stripOuterMathDelimiters(math: string): string {
    let result = math.trim();
    let changed = true;

    while (changed) {
      changed = false;
      const delimiterPairs = [
        { open: '\\\\(', close: '\\\\)' },
        { open: '\\\\[', close: '\\\\]' },
        { open: '\\(', close: '\\)' },
        { open: '\\[', close: '\\]' },
        { open: '$$', close: '$$' },
      ];

      for (const { open, close } of delimiterPairs) {
        if (
          result.startsWith(open) &&
          result.endsWith(close) &&
          result.length > open.length + close.length
        ) {
          result = result.slice(open.length, -close.length).trim();
          changed = true;
          break;
        }
      }
    }

    return result;
  }

  function normalizeTextCommands(math: string): string {
    return math.replace(/\\text\{([^{}]*)\}/g, (_match, content: string) => {
      const escapedContent = content.replace(/(^|[^\\])_/g, '$1\\_');
      return `\\text{${escapedContent}}`;
    });
  }

  function stripLocalMacroDefinitions(math: string): string {
    let result = '';
    let index = 0;

    while (index < math.length) {
      const command =
        math.startsWith('\\newcommand', index) || math.startsWith('\\renewcommand', index)
          ? math.startsWith('\\renewcommand', index)
            ? '\\renewcommand'
            : '\\newcommand'
          : '';

      if (!command) {
        result += math[index];
        index += 1;
        continue;
      }

      let nextIndex = index + command.length;
      const firstGroupEnd = consumeLatexGroup(math, nextIndex);
      if (firstGroupEnd === null) {
        result += math[index];
        index += 1;
        continue;
      }

      nextIndex = firstGroupEnd;
      const secondGroupEnd = consumeLatexGroup(math, nextIndex);
      if (secondGroupEnd === null) {
        result += math[index];
        index += 1;
        continue;
      }

      index = secondGroupEnd;
    }

    return result.trim();
  }

  function stripUnsupportedMathAnnotations(math: string): string {
    return math
      .replace(/\\label\{[^{}]*\}/g, '')
      .replace(/\\tag\{[^{}]*\}/g, '')
      .replace(/\\eqref\{[^{}]*\}/g, '')
      .replace(/\\ref\{[^{}]*\}/g, '')
      .replace(/\\labeleq\s*:\s*[-\w]+/g, '')
      .replace(/\\eqrefeq\s*:\s*[-\w]+/g, '')
      .replace(/\s+([&\\])/g, ' $1')
      .trim();
  }

  function cleanMathContent(math: string): string {
    return normalizeTextCommands(
      stripUnsupportedMathAnnotations(stripLocalMacroDefinitions(stripOuterMathDelimiters(math)))
    );
  }

  function consumeLatexGroup(source: string, startIndex: number): number | null {
    let index = startIndex;
    while (index < source.length && /\s/.test(source[index])) {
      index += 1;
    }

    if (source[index] !== '{') {
      return null;
    }

    let depth = 0;
    for (; index < source.length; index += 1) {
      const char = source[index];

      if (char === '\\') {
        index += 1;
        continue;
      }

      if (char === '{') {
        depth += 1;
      } else if (char === '}') {
        depth -= 1;
        if (depth === 0) {
          return index + 1;
        }
      }
    }

    return null;
  }

  function normalizeMathContent(math: string, environment?: string): string {
    const trimmedMath = stripOuterMathDelimiters(normalizeEscapedCommandBackslashes(math.trim()));
    if (!environment) {
      const wrappedEnvironment = trimmedMath.match(/^\\begin\{([a-zA-Z*]+)\}([\s\S]+)\\end\{\1\}$/);
      if (wrappedEnvironment) {
        return normalizeMathContent(wrappedEnvironment[2], wrappedEnvironment[1]);
      }

      return cleanMathContent(trimmedMath);
    }

    const normalizedEnvironment = environment.replace(/\*$/, '');

    if (normalizedEnvironment === 'equation') {
      return normalizeMathContent(cleanMathContent(trimmedMath));
    }
    if (normalizedEnvironment === 'align') {
      return `\\begin{aligned}${cleanMathContent(trimmedMath)}\\end{aligned}`;
    }
    if (normalizedEnvironment === 'gather') {
      return `\\begin{gathered}${cleanMathContent(trimmedMath)}\\end{gathered}`;
    }

    return `\\begin{${environment}}${cleanMathContent(trimmedMath)}\\end{${environment}}`;
  }

  function createMathElement(math: string, isDisplay: boolean, environment?: string): HTMLElement {
    const mathElement = document.createElement(isDisplay ? 'div' : 'span');
    mathElement.className = isDisplay ? 'katex-display' : 'katex-inline';
    katex.render(normalizeMathContent(math, environment), mathElement, {
      displayMode: isDisplay,
      throwOnError: false,
      strict: false,
      macros: katexMacros,
    });
    return mathElement;
  }

  function getMathMatches(text: string, includeBareMathBlock: boolean = false): MathMatch[] {
    // Order matters: wrappers should be matched before raw environments to avoid nested duplicates.
    const mathPatterns = [
      // Display wrappers should win over inline delimiters nested inside them.
      { regex: /\\\\\[([\s\S]+?)\\\\\]/g, isDisplay: true },
      // Display math: $$...$$
      { regex: /\$\$([\s\S]+?)\$\$/g, isDisplay: true },
      // Display math: \[...\]
      { regex: /\\\[([\s\S]+?)\\\]/g, isDisplay: true },
      { regex: /\\\\begin\{([a-zA-Z*]+)\}([\s\S]+?)\\\\end\{\1\}/g, isDisplay: true },
      // LaTeX environments: \begin{equation}...\end{equation}
      { regex: /\\begin\{([a-zA-Z*]+)\}([\s\S]+?)\\end\{\1\}/g, isDisplay: true },
      // Markdown/HTML converted summaries can preserve escaped delimiters as \\(...\\).
      { regex: /\\\\\(([\s\S]+?)\\\\\)/g, isDisplay: false },
      // Inline math: \(...\)
      { regex: /\\\(([\s\S]+?)\\\)/g, isDisplay: false },
      // Inline math: $...$ (single line, not empty, not starting/ending with space)
      { regex: /\$([^\s$][^$\n]*[^\s$]|\S)\$/g, isDisplay: false },
    ];

    const allMatches: MathMatch[] = [];

    for (const { regex, isDisplay } of mathPatterns) {
      let match;
      regex.lastIndex = 0;
      while ((match = regex.exec(text)) !== null) {
        const matchStart = match.index;
        const matchEnd = match.index + match[0].length;
        const isOverlapping = allMatches.some((m) => matchStart < m.end && matchEnd > m.start);

        if (!isOverlapping) {
          allMatches.push({
            start: matchStart,
            end: matchEnd,
            math: match[2] ?? match[1],
            isDisplay,
            environment: match[2] ? match[1] : undefined,
          });
        }
      }
    }

    if (includeBareMathBlock && allMatches.length === 0 && isBareMathBlock(text)) {
      allMatches.push({
        start: 0,
        end: text.length,
        math: text,
        isDisplay: true,
      });
    }

    if (allMatches.length === 0) {
      const danglingDollarMatch = getDanglingDollarMathMatch(text);
      if (danglingDollarMatch) {
        allMatches.push(danglingDollarMatch);
      }
    }

    return allMatches.sort((a, b) => a.start - b.start);
  }

  function getDanglingDollarMathMatch(text: string): MathMatch | null {
    const firstDollar = text.indexOf('$');
    if (firstDollar === -1 || text.indexOf('$', firstDollar + 1) !== -1) {
      return null;
    }

    const beforeDollar = text.slice(0, firstDollar);
    const afterDollar = text.slice(firstDollar + 1);
    const afterMath = afterDollar.trim();
    if (looksLikeMathExpression(afterMath) && !containsCJK(afterMath)) {
      const leadingWhitespace = afterDollar.length - afterDollar.trimStart().length;
      return {
        start: firstDollar,
        end: text.length,
        math: afterDollar.slice(leadingWhitespace),
        isDisplay: false,
      };
    }

    const beforeMath = beforeDollar.trim();
    if (looksLikeMathExpression(beforeMath) && !containsCJK(beforeMath)) {
      const mathStart = beforeDollar.length - beforeDollar.trimStart().length;
      return {
        start: mathStart,
        end: firstDollar + 1,
        math: beforeMath,
        isDisplay: false,
      };
    }

    return null;
  }

  function looksLikeMathExpression(text: string): boolean {
    const trimmedText = text.trim();
    if (trimmedText.length < 2) return false;

    return /\\[a-zA-Z]+|[_^{}]|[=<>]/.test(trimmedText);
  }

  function containsCJK(text: string): boolean {
    return /[\u4e00-\u9fff]/.test(text);
  }

  function isBareMathBlock(text: string): boolean {
    const trimmedText = text.trim();
    if (!trimmedText.startsWith('\\') || trimmedText.length < 8) {
      return false;
    }

    const hasMathCommand =
      /\\(?:frac|sqrt|sum|prod|int|boldsymbol|mathbf|mathop|operatorname|text|Vert|left|right|alpha|beta|gamma|eta|sigma|Phi|Delta)/.test(
        trimmedText
      );
    if (!hasMathCommand) {
      return false;
    }

    // Avoid treating ordinary escaped prose as a formula block.
    return !containsCJK(trimmedText);
  }

  function hasOnlyMatchedMath(text: string, matches: MathMatch[]): boolean {
    let remainingText = '';
    let lastIndex = 0;

    for (const match of matches) {
      remainingText += text.slice(lastIndex, match.start);
      lastIndex = match.end;
    }
    remainingText += text.slice(lastIndex);

    return remainingText.trim().length === 0;
  }

  function shouldSkipMathContainer(element: HTMLElement): boolean {
    return (
      !!element.closest('pre, code, script, style, .katex, .katex-display, .katex-inline') ||
      !!element.querySelector(
        'a, pre, code, script, style, table, img, video, audio, iframe, canvas, button, input, textarea'
      )
    );
  }

  function isSimpleMathContainer(element: HTMLElement): boolean {
    return Array.from(element.children).every((child) => {
      if (!simpleInlineTags.has(child.tagName)) return false;
      return !child.querySelector(
        'a, pre, code, script, style, table, img, video, audio, iframe, canvas, button, input, textarea'
      );
    });
  }

  function replaceTextWithMathFragments(
    target: Text | HTMLElement,
    text: string,
    matches: MathMatch[]
  ) {
    const fragments: (string | HTMLElement)[] = [];
    let lastIndex = 0;

    for (const { start, end, math, isDisplay, environment } of matches) {
      if (start > lastIndex) {
        fragments.push(text.substring(lastIndex, start));
      }

      try {
        fragments.push(createMathElement(math, isDisplay, environment));
      } catch (e) {
        console.error('Error rendering math:', e, 'Content:', math);
        fragments.push(text.substring(start, end));
      }

      lastIndex = end;
    }

    if (lastIndex < text.length) {
      fragments.push(text.substring(lastIndex));
    }

    if (target instanceof Text) {
      const parent = target.parentNode;
      if (!parent) return;

      for (const fragment of fragments) {
        parent.insertBefore(
          typeof fragment === 'string' ? document.createTextNode(fragment) : fragment,
          target
        );
      }
      parent.removeChild(target);
      return;
    }

    target.replaceChildren(
      ...fragments.map((fragment) =>
        typeof fragment === 'string' ? document.createTextNode(fragment) : fragment
      )
    );
  }

  function renderStandaloneMathBlocks(container: HTMLElement) {
    const candidates = Array.from(
      container.querySelectorAll('p, div, li, td, th, figcaption, blockquote')
    ) as HTMLElement[];

    for (const element of candidates) {
      if (shouldSkipMathContainer(element)) continue;

      const text = element.textContent || '';
      if (!text.includes('\\') && !text.includes('$')) continue;

      const matches = getMathMatches(text.trim(), true);
      const hasDisplayMath = matches.some((match) => match.isDisplay);
      const shouldRewriteMixedDisplayMath =
        hasDisplayMath && isSimpleMathContainer(element) && text.includes('\\');

      if (
        matches.length === 0 ||
        (!hasOnlyMatchedMath(text.trim(), matches) && !shouldRewriteMixedDisplayMath)
      ) {
        continue;
      }

      replaceTextWithMathFragments(element, text.trim(), matches);
    }
  }

  /**
   * Render math formulas in the content
   * Supports multiple formats:
   * - Display math: $$...$$ or \[...\]
   * - Inline math: $...$ or \(...\)
   * - LaTeX environments: \begin{equation}...\end{equation}, \begin{align}...\end{align}, etc.
   */
  function renderMathFormulas(container: HTMLElement) {
    if (!container) return;

    try {
      renderStandaloneMathBlocks(container);

      // First, handle pre-existing math elements with class 'math' or 'MathJax'
      const existingMathElements = container.querySelectorAll(
        '.math, .MathJax, [data-math], script[type*="math"]'
      );
      existingMathElements.forEach((el) => {
        if (el.classList.contains('katex') || el.querySelector('.katex')) return;

        const mathContent = el.getAttribute('data-math') || el.textContent || '';
        if (!mathContent.trim()) return;

        try {
          const isDisplay =
            el.classList.contains('math-display') ||
            el.classList.contains('display') ||
            el.tagName === 'DIV';
          const mathElement = createMathElement(mathContent, isDisplay);
          el.replaceWith(mathElement);
        } catch (e) {
          console.error('Error rendering existing math element:', e);
        }
      });

      // Find all text nodes that might contain math
      const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT, {
        acceptNode: (node) => {
          // Skip if already inside a math element
          const parent = node.parentElement;
          if (
            parent?.classList.contains('katex') ||
            parent?.classList.contains('katex-display') ||
            parent?.classList.contains('katex-inline') ||
            parent?.closest('.katex') ||
            parent?.closest('.katex-display') ||
            parent?.tagName === 'CODE' ||
            parent?.tagName === 'PRE' ||
            parent?.tagName === 'SCRIPT' ||
            parent?.tagName === 'STYLE'
          ) {
            return NodeFilter.FILTER_REJECT;
          }
          // Accept nodes that contain math delimiters
          const text = node.textContent || '';
          if (
            text.includes('$') ||
            text.includes('\\(') ||
            text.includes('\\[') ||
            text.includes('\\begin{')
          ) {
            return NodeFilter.FILTER_ACCEPT;
          }
          return NodeFilter.FILTER_REJECT;
        },
      });

      const nodesToProcess: Text[] = [];
      let currentNode: Node | null;
      while ((currentNode = walker.nextNode())) {
        nodesToProcess.push(currentNode as Text);
      }

      // Process each text node
      for (const node of nodesToProcess) {
        const text = node.textContent || '';
        if (!text) continue;

        const allMatches = getMathMatches(text);

        if (allMatches.length > 0) {
          replaceTextWithMathFragments(node, text, allMatches);
        }
      }
    } catch (e) {
      console.error('Error processing math formulas:', e);
    }
  }

  /**
   * Apply syntax highlighting to code blocks
   */
  function highlightCodeBlocks(container: HTMLElement) {
    if (!container) return;

    try {
      // Find all code blocks within pre elements
      const codeBlocks = container.querySelectorAll('pre code');

      codeBlocks.forEach((block) => {
        // Skip if already highlighted
        if (block.classList.contains('hljs')) return;

        const codeElement = block as HTMLElement;

        // Try to detect language from class
        let language = '';
        const classList = Array.from(codeElement.classList);
        for (const cls of classList) {
          if (cls.startsWith('language-') || cls.startsWith('lang-')) {
            language = cls.replace(/^(language-|lang-)/, '');
            break;
          }
        }

        // Highlight the code
        if (language && hljs.getLanguage(language)) {
          // Use specific language
          const result = hljs.highlight(codeElement.textContent || '', { language });
          codeElement.innerHTML = result.value;
          codeElement.classList.add('hljs');
        } else {
          // Auto-detect language
          const result = hljs.highlightAuto(codeElement.textContent || '');
          codeElement.innerHTML = result.value;
          codeElement.classList.add('hljs');
          if (result.language) {
            codeElement.dataset.detectedLanguage = result.language;
          }
        }
      });

      // Also handle standalone pre elements without code tags
      const preElements = container.querySelectorAll('pre:not(:has(code))');
      preElements.forEach((pre) => {
        if (pre.classList.contains('hljs')) return;

        const preElement = pre as HTMLElement;
        const result = hljs.highlightAuto(preElement.textContent || '');
        preElement.innerHTML = result.value;
        preElement.classList.add('hljs');
      });
    } catch (e) {
      console.error('Error highlighting code blocks:', e);
    }
  }

  /**
   * Apply all rendering enhancements to the container
   */
  async function enhanceRendering(containerSelector: string = '.prose-content') {
    await nextTick();
    const container = document.querySelector(containerSelector) as HTMLElement;
    if (!container) return;

    // Render math formulas
    renderMathFormulas(container);

    // Apply syntax highlighting to code blocks
    highlightCodeBlocks(container);
  }

  return {
    renderMathFormulas,
    highlightCodeBlocks,
    enhanceRendering,
  };
}
