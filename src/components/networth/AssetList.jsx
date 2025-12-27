import React from 'react';

function AssetList({ assets, onEdit, onDelete }) {
  if (assets.length === 0) {
    return (
      <div className="empty-list">
        <p>No assets added yet</p>
      </div>
    );
  }

  return (
    <div className="item-list">
      {assets.map(asset => (
        <div key={asset.id} className="item-card">
          <div className="item-info">
            <div className="item-name">{asset.name}</div>
            <div className="item-type">{asset.assetType?.name || 'Unknown Type'}</div>
          </div>
          <div className="item-value positive">
            ${asset.currentValue.toLocaleString('en-US', { minimumFractionDigits: 2 })}
          </div>
          <div className="item-actions">
            <button className="btn-icon" onClick={() => onEdit(asset)} title="Edit">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
              </svg>
            </button>
            <button className="btn-icon btn-danger" onClick={() => onDelete(asset.id)} title="Delete">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                <polyline points="3 6 5 6 21 6" />
                <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
              </svg>
            </button>
          </div>
        </div>
      ))}
    </div>
  );
}

export default AssetList;
