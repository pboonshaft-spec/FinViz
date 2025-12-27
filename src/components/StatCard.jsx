import React from 'react';

const styles = {
  card: {
    background: '#1e1e1e',
    borderRadius: '14px',
    padding: '20px 24px',
    border: '1px solid #2a2a2a',
    transition: 'all 0.2s ease'
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'flex-start',
    marginBottom: '12px'
  },
  label: {
    color: '#888',
    fontSize: '0.8rem',
    fontWeight: '500',
    textTransform: 'uppercase',
    letterSpacing: '0.5px'
  },
  badge: {
    display: 'inline-flex',
    alignItems: 'center',
    gap: '4px',
    padding: '4px 8px',
    borderRadius: '6px',
    fontSize: '0.75rem',
    fontWeight: '500'
  },
  value: {
    fontSize: '1.75rem',
    fontWeight: '600',
    letterSpacing: '-0.02em'
  }
};

function StatCard({ label, value, change, changeType, valueColor }) {
  const badgeStyle = {
    ...styles.badge,
    background: changeType === 'positive' ? 'rgba(0, 212, 170, 0.15)' : 'rgba(255, 107, 107, 0.15)',
    color: changeType === 'positive' ? '#00d4aa' : '#ff6b6b'
  };

  return (
    <div style={styles.card}>
      <div style={styles.header}>
        <span style={styles.label}>{label}</span>
        <span style={badgeStyle}>
          {changeType === 'positive' ? '↑' : '↓'} {change}
        </span>
      </div>
      <div style={{ ...styles.value, color: valueColor || '#fff' }}>
        {value}
      </div>
    </div>
  );
}

export default StatCard;
