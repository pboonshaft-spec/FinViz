import React, { useState, useMemo } from 'react';

// Icon components for asset types
const AssetIcons = {
  'Stocks (US)': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <polyline points="22 7 13.5 15.5 8.5 10.5 2 17" />
      <polyline points="16 7 22 7 22 13" />
    </svg>
  ),
  'Stocks (Intl)': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10" />
      <line x1="2" y1="12" x2="22" y2="12" />
      <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
    </svg>
  ),
  'Bonds': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="2" y="4" width="20" height="16" rx="2" />
      <line x1="2" y1="10" x2="22" y2="10" />
      <line x1="6" y1="4" x2="6" y2="20" />
    </svg>
  ),
  'Real Estate': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  ),
  'Cash/Savings': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="2" y="4" width="20" height="16" rx="2" />
      <circle cx="12" cy="12" r="4" />
      <line x1="2" y1="8" x2="4" y2="8" />
      <line x1="20" y1="8" x2="22" y2="8" />
      <line x1="2" y1="16" x2="4" y2="16" />
      <line x1="20" y1="16" x2="22" y2="16" />
    </svg>
  ),
  'Crypto': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="12" r="10" />
      <path d="M9.5 9.5c.5-1 1.5-1.5 2.5-1.5 1.5 0 2.5 1 2.5 2.5 0 1.5-1 2-2.5 2.5v1" />
      <circle cx="12" cy="16.5" r="0.5" fill="currentColor" />
    </svg>
  ),
  'default': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <line x1="12" y1="1" x2="12" y2="23" />
      <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
    </svg>
  ),
};

function AssetList({ assets, onEdit, onDelete }) {
  const [expandedCategories, setExpandedCategories] = useState({});

  // Group assets by type
  const groupedAssets = useMemo(() => {
    const groups = {};
    assets.forEach(asset => {
      const type = asset.assetType?.name || 'Other';
      if (!groups[type]) {
        groups[type] = { assets: [], total: 0 };
      }
      groups[type].assets.push(asset);
      groups[type].total += asset.currentValue;
    });
    return groups;
  }, [assets]);

  const toggleCategory = (category) => {
    setExpandedCategories(prev => ({
      ...prev,
      [category]: !prev[category]
    }));
  };

  if (assets.length === 0) {
    return (
      <div className="empty-list">
        <p>No assets added yet</p>
      </div>
    );
  }

  const categories = Object.keys(groupedAssets).sort();

  return (
    <div className="categorized-list">
      {categories.map(category => {
        const { assets: categoryAssets, total } = groupedAssets[category];
        const isExpanded = expandedCategories[category];
        const Icon = AssetIcons[category] || AssetIcons['default'];

        return (
          <div key={category} className="category-group">
            <button
              className="category-header"
              onClick={() => toggleCategory(category)}
            >
              <div className="category-left">
                <span className={`collapse-icon ${isExpanded ? 'open' : ''}`}>&#9656;</span>
                <span className="category-icon asset-icon">{Icon}</span>
                <span className="category-name">{category}</span>
                <span className="category-count">{categoryAssets.length}</span>
              </div>
              <span className="category-total positive">
                ${total.toLocaleString('en-US', { minimumFractionDigits: 2 })}
              </span>
            </button>

            {isExpanded && (
              <div className="category-items">
                {categoryAssets.map(asset => (
                  <div key={asset.id} className="item-card">
                    <div className="item-info">
                      <div className="item-name">{asset.name}</div>
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
            )}
          </div>
        );
      })}
    </div>
  );
}

export default AssetList;
