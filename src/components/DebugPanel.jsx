import React, { useState } from 'react';

const styles = {
  wrapper: {
    marginBottom: '24px'
  },
  toggle: {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '8px',
    padding: '8px 16px',
    background: '#1e1e1e',
    border: '1px solid #2a2a2a',
    borderRadius: '8px',
    color: '#888',
    fontSize: '0.85rem',
    cursor: 'pointer',
    transition: 'all 0.2s ease'
  },
  panel: {
    marginTop: '12px',
    background: '#1e1e1e',
    borderRadius: '12px',
    padding: '16px',
    border: '1px solid #2a2a2a'
  },
  content: {
    background: '#151515',
    padding: '16px',
    borderRadius: '8px',
    fontFamily: "'SF Mono', 'Fira Code', monospace",
    fontSize: '0.8rem',
    color: '#888',
    maxHeight: '200px',
    overflowY: 'auto',
    lineHeight: '1.6'
  },
  logLine: {
    borderBottom: '1px solid #222',
    paddingBottom: '4px',
    marginBottom: '4px'
  }
};

function DebugPanel({ logs }) {
  const [isOpen, setIsOpen] = useState(false);

  if (!logs || logs.length === 0) return null;

  return (
    <div style={styles.wrapper}>
      <button
        style={styles.toggle}
        onClick={() => setIsOpen(!isOpen)}
      >
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
          <polyline points="14 2 14 8 20 8" />
          <line x1="16" y1="13" x2="8" y2="13" />
          <line x1="16" y1="17" x2="8" y2="17" />
        </svg>
        {isOpen ? 'Hide' : 'Show'} Processing Log ({logs.length})
      </button>

      {isOpen && (
        <div style={styles.panel}>
          <div style={styles.content}>
            {logs.map((log, index) => (
              <div key={index} style={styles.logLine}>{log}</div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default DebugPanel;
