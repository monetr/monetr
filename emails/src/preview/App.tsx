import { createElement, useMemo, useState } from 'react';

import { templateList } from './templates';

import { renderToStaticMarkup } from 'react-dom/server';

type ViewMode = 'preview' | 'html' | 'text';

function htmlToPlainText(html: string): string {
  // Remove hidden preview div
  let text = html.replace(/<div[^>]*data-skip-in-text[^>]*>[\s\S]*?<\/div>\s*<\/div>/gi, '');
  // Remove style and head tags entirely
  text = text.replace(/<(style|head)[^>]*>[\s\S]*?<\/\1>/gi, '');
  // Replace <hr> with a line of dashes
  text = text.replace(/<hr[^>]*>/gi, '\n' + '-'.repeat(80) + '\n');
  // Replace <br> with newline
  text = text.replace(/<br\s*\/?>/gi, '\n');
  // Replace block-level closing tags with newlines
  text = text.replace(/<\/(p|h[1-6]|div|tr|table|tbody)>/gi, '\n');
  // Extract link text with href
  text = text.replace(/<a[^>]*href="([^"]*)"[^>]*>([\s\S]*?)<\/a>/gi, (_, href, content) => {
    const linkText = content.replace(/<[^>]+>/g, '').trim();
    if (href === `mailto:${linkText}` || href === linkText) {
      return linkText;
    }
    return `${linkText} [${href}]`;
  });
  // Remove remaining tags
  text = text.replace(/<[^>]+>/g, '');
  // Decode common HTML entities
  text = text
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;/g, "'")
    .replace(/&nbsp;/g, ' ');
  // Remove zero-width characters used for preview padding
  text = text.replace(/[\u200C\u200B\u00A0]+/g, ' ');
  // Collapse multiple blank lines
  text = text.replace(/\n{3,}/g, '\n\n');
  // Trim lines and overall
  text = text
    .split('\n')
    .map(l => l.trim())
    .join('\n')
    .trim();
  return text;
}

export function App() {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [viewMode, setViewMode] = useState<ViewMode>('preview');
  const selected = templateList[selectedIndex];

  const renderedHtml = useMemo(() => {
    return renderToStaticMarkup(createElement(selected.component, selected.previewProps));
  }, [selected]);

  const plainText = useMemo(() => htmlToPlainText(renderedHtml), [renderedHtml]);

  const viewModes: { key: ViewMode; label: string }[] = [
    { key: 'preview', label: 'Preview' },
    { key: 'text', label: 'Text' },
    { key: 'html', label: 'HTML' },
  ];

  return (
    <div style={{ display: 'flex', height: '100vh', fontFamily: 'system-ui, sans-serif' }}>
      {/* Sidebar */}
      <nav
        style={{
          width: '240px',
          borderRight: '1px solid #e5e7eb',
          padding: '16px',
          backgroundColor: '#fafafa',
          flexShrink: 0,
        }}
      >
        <h2
          style={{
            margin: '0 0 16px',
            fontSize: '14px',
            fontWeight: 600,
            color: '#6b7280',
            textTransform: 'uppercase',
            letterSpacing: '0.05em',
          }}
        >
          Email Templates
        </h2>
        {templateList.map((template, i) => (
          <button
            key={template.name}
            onClick={() => setSelectedIndex(i)}
            style={{
              display: 'block',
              width: '100%',
              padding: '8px 12px',
              marginBottom: '4px',
              border: 'none',
              borderRadius: '6px',
              textAlign: 'left',
              cursor: 'pointer',
              fontSize: '14px',
              backgroundColor: i === selectedIndex ? '#4E1AA0' : 'transparent',
              color: i === selectedIndex ? '#fff' : '#374151',
            }}
          >
            {template.name}
          </button>
        ))}
      </nav>

      {/* Preview pane */}
      <main style={{ flex: 1, overflow: 'auto', backgroundColor: '#f3f4f6', padding: '24px' }}>
        <div
          style={{
            maxWidth: '700px',
            margin: '0 auto',
            backgroundColor: '#fff',
            borderRadius: '8px',
            boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
          }}
        >
          {/* Header with view mode toggle */}
          <div
            style={{
              padding: '12px 16px',
              borderBottom: '1px solid #e5e7eb',
              fontSize: '13px',
              color: '#6b7280',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'space-between',
            }}
          >
            <span>
              <strong style={{ color: '#111827' }}>{selected.name}</strong>
            </span>
            <div style={{ display: 'flex', gap: '4px' }}>
              {viewModes.map(({ key, label }) => (
                <button
                  key={key}
                  onClick={() => setViewMode(key)}
                  style={{
                    padding: '4px 10px',
                    border: '1px solid',
                    borderColor: viewMode === key ? '#4E1AA0' : '#d1d5db',
                    borderRadius: '4px',
                    fontSize: '12px',
                    cursor: 'pointer',
                    backgroundColor: viewMode === key ? '#4E1AA0' : '#fff',
                    color: viewMode === key ? '#fff' : '#374151',
                  }}
                >
                  {label}
                </button>
              ))}
            </div>
          </div>

          {/* Content */}
          {viewMode === 'preview' && (
            <div style={{ padding: '16px' }}>{createElement(selected.component, selected.previewProps)}</div>
          )}
          {viewMode === 'text' && (
            <pre
              style={{
                padding: '16px',
                margin: 0,
                fontSize: '13px',
                lineHeight: '1.6',
                fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
                whiteSpace: 'pre-wrap',
                wordWrap: 'break-word',
                color: '#374151',
              }}
            >
              {plainText}
            </pre>
          )}
          {viewMode === 'html' && (
            <pre
              style={{
                padding: '16px',
                margin: 0,
                fontSize: '12px',
                lineHeight: '1.5',
                fontFamily: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
                whiteSpace: 'pre-wrap',
                wordWrap: 'break-word',
                color: '#374151',
                maxHeight: '80vh',
                overflow: 'auto',
              }}
            >
              {renderedHtml}
            </pre>
          )}
        </div>
      </main>
    </div>
  );
}
