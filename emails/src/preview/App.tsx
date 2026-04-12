import { createElement, useMemo, useState } from 'react';

import { toPlainText } from '@monetr/emails/toPlainText';

import { templateList } from './templates';

import { renderToStaticMarkup } from 'react-dom/server';

import styles from './App.module.scss';

type ViewMode = 'preview' | 'html' | 'text';

function cx(...classes: (string | false | undefined)[]): string {
  return classes.filter(Boolean).join(' ');
}

export function App() {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const [viewMode, setViewMode] = useState<ViewMode>('preview');
  const selected = templateList[selectedIndex];

  const renderedHtml = useMemo(() => {
    return renderToStaticMarkup(createElement(selected.component, selected.previewProps));
  }, [selected]);

  const plainText = useMemo(() => toPlainText(renderedHtml), [renderedHtml]);

  const viewModes: { key: ViewMode; label: string }[] = [
    { key: 'preview', label: 'Preview' },
    { key: 'text', label: 'Text' },
    { key: 'html', label: 'HTML' },
  ];

  return (
    <div className={styles.root}>
      <nav className={styles.sidebar}>
        <h2 className={styles.sidebarHeading}>Email Templates</h2>
        {templateList.map((template, i) => (
          <button
            key={template.name}
            onClick={() => setSelectedIndex(i)}
            className={cx(styles.templateButton, i === selectedIndex && styles.templateButtonActive)}
            type='button'
          >
            {template.name}
          </button>
        ))}
      </nav>

      <main className={styles.main}>
        <div className={styles.card}>
          <div className={styles.cardHeader}>
            <span>
              <strong className={styles.cardTitle}>{selected.name}</strong>
            </span>
            <div className={styles.viewModeButtons}>
              {viewModes.map(({ key, label }) => (
                <button
                  key={key}
                  onClick={() => setViewMode(key)}
                  className={cx(styles.viewModeButton, viewMode === key && styles.viewModeButtonActive)}
                  type='button'
                >
                  {label}
                </button>
              ))}
            </div>
          </div>

          {viewMode === 'preview' && (
            <div className={styles.previewContent}>{createElement(selected.component, selected.previewProps)}</div>
          )}
          {viewMode === 'text' && <pre className={cx(styles.pre, styles.preText)}>{plainText}</pre>}
          {viewMode === 'html' && <pre className={cx(styles.pre, styles.preHtml)}>{renderedHtml}</pre>}
        </div>
      </main>
    </div>
  );
}
