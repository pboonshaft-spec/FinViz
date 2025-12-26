import React from 'react';

const styles = {
  card: {
    background: '#1e1e1e',
    borderRadius: '16px',
    padding: '24px',
    border: '1px solid #2a2a2a'
  },
  title: {
    fontSize: '1rem',
    fontWeight: '600',
    color: '#fff',
    marginBottom: '20px',
    display: 'flex',
    alignItems: 'center',
    gap: '10px'
  },
  titleDot: {
    width: '8px',
    height: '8px',
    borderRadius: '50%',
    background: 'linear-gradient(135deg, #6366f1 0%, #8b5cf6 100%)'
  }
};

function ChartCard({ title, children, fullWidth }) {
  return (
    <div style={{
      ...styles.card,
      gridColumn: fullWidth ? '1 / -1' : 'auto'
    }}>
      <div style={styles.title}>
        <span style={styles.titleDot}></span>
        {title}
      </div>
      {children}
    </div>
  );
}

export default ChartCard;
