import React from 'react';
import './AdvisorStyles.css';

export default function ClientContextBanner({ client, onBack }) {
  return (
    <div className="client-context-banner">
      <div className="banner-content">
        <span className="banner-icon">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2" />
            <circle cx="12" cy="7" r="4" />
          </svg>
        </span>
        <span className="banner-text">
          Viewing: <strong>{client.name}</strong> ({client.email})
        </span>
      </div>
      <button className="btn btn-outline btn-sm" onClick={onBack}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
          <line x1="19" y1="12" x2="5" y2="12" />
          <polyline points="12 19 5 12 12 5" />
        </svg>
        Back to Clients
      </button>
    </div>
  );
}
