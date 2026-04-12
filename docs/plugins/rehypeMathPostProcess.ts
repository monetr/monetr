// Rehype plugin that post-processes Shiki-rendered math code blocks into KaTeX.
//
// Problem: rspress runs Shiki BEFORE user rehype plugins. So remark-math creates
// math MDAST nodes → they become <pre><code class="language-math"> in HAST →
// Shiki converts them to styled code blocks → rehype-katex never sees them.
//
// Solution: This plugin runs AFTER Shiki and finds <pre> elements containing
// math code blocks, extracts the LaTeX source, renders it with KaTeX, and
// parses the KaTeX HTML into proper HAST nodes to replace the code block.
//
// Inline math ($...$) is unaffected — Shiki only processes <pre><code> blocks,
// not standalone <code> elements, so rehype-katex handles inline math fine.

import type { Element, ElementContent, Nodes, Root } from 'hast';
import { fromHtml } from 'hast-util-from-html';
import * as katex from 'katex';
import { SKIP, visit } from 'unist-util-visit';

function isMathCodeBlock(node: Element): boolean {
  const props = node.properties;
  // rspress/Shiki uses raw HTML attribute names (lang) not HAST convention (dataLang)
  if (props.dataLang === 'math' || props['data-lang'] === 'math' || props.lang === 'math') {
    return true;
  }
  const code = node.children.find((c): c is Element => c.type === 'element' && c.tagName === 'code');
  if (!code) {
    return false;
  }
  const codeProps = code.properties;
  if (codeProps.dataLang === 'math' || codeProps['data-lang'] === 'math' || codeProps.lang === 'math') {
    return true;
  }
  const classes = codeProps.className || codeProps.class;
  if (Array.isArray(classes)) {
    return classes.some(c => typeof c === 'string' && c.includes('language-math'));
  }
  if (typeof classes === 'string') {
    return classes.includes('language-math');
  }
  return false;
}

function extractText(node: Nodes): string {
  if (node.type === 'text') {
    return node.value;
  }
  if ('children' in node) {
    return node.children.map(extractText).join('');
  }
  return '';
}

export default function rehypeMathPostProcess() {
  return (tree: Root) => {
    visit(tree, 'element', (node, index, parent) => {
      if (node.tagName !== 'pre' || index === undefined || !parent) {
        return;
      }
      if (!isMathCodeBlock(node)) {
        return;
      }

      const latex = extractText(node).trim();
      if (!latex) {
        return;
      }

      try {
        const html = katex.renderToString(latex, {
          displayMode: true,
          throwOnError: false,
        });

        // Parse KaTeX HTML into proper HAST nodes
        const fragment = fromHtml(html, { fragment: true });

        // Wrap in a display-math container div
        const wrapper: Element = {
          type: 'element',
          tagName: 'div',
          properties: { className: ['math', 'math-display'] },
          children: fragment.children as ElementContent[],
        };

        parent.children[index] = wrapper;
      } catch {
        // If KaTeX fails, leave the code block as-is
      }

      return SKIP;
    });
  };
}
