import { useState, createElement } from 'react';
import { templates } from './templates';

export function App() {
  const [selectedIndex, setSelectedIndex] = useState(0);
  const selected = templates[selectedIndex];

  return (
    <div style={{ display: 'flex', height: '100vh', fontFamily: 'system-ui, sans-serif' }}>
      {/* Sidebar */}
      <nav style={{
        width: '240px',
        borderRight: '1px solid #e5e7eb',
        padding: '16px',
        backgroundColor: '#fafafa',
        flexShrink: 0,
      }}>
        <h2 style={{ margin: '0 0 16px', fontSize: '14px', fontWeight: 600, color: '#6b7280', textTransform: 'uppercase', letterSpacing: '0.05em' }}>
          Email Templates
        </h2>
        {templates.map((template, i) => (
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
        <div style={{ maxWidth: '700px', margin: '0 auto', backgroundColor: '#fff', borderRadius: '8px', boxShadow: '0 1px 3px rgba(0,0,0,0.1)' }}>
          <div style={{ padding: '12px 16px', borderBottom: '1px solid #e5e7eb', fontSize: '13px', color: '#6b7280' }}>
            Preview: <strong style={{ color: '#111827' }}>{selected.name}</strong>
          </div>
          <div style={{ padding: '16px' }}>
            {createElement(selected.component, selected.previewProps)}
          </div>
        </div>
      </main>
    </div>
  );
}
