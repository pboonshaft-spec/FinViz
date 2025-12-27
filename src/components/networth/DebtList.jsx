import React, { useState, useMemo } from 'react';

// Icon components for debt types
const DebtIcons = {
  'Credit Card': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="1" y="4" width="22" height="16" rx="2" ry="2" />
      <line x1="1" y1="10" x2="23" y2="10" />
    </svg>
  ),
  'Mortgage': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  ),
  'Student Loan': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M22 10v6M2 10l10-5 10 5-10 5z" />
      <path d="M6 12v5c0 2 2 3 6 3s6-1 6-3v-5" />
    </svg>
  ),
  'Auto Loan': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <path d="M5 17H3a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2" />
      <path d="M19 17h2a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2h-2" />
      <path d="M5 17a2 2 0 1 0 4 0 2 2 0 0 0-4 0" />
      <path d="M15 17a2 2 0 1 0 4 0 2 2 0 0 0-4 0" />
      <path d="M9 17h6" />
      <path d="M3 9h18V7l-3-4H6L3 7v2z" />
    </svg>
  ),
  'Personal Loan': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <circle cx="12" cy="8" r="5" />
      <path d="M20 21a8 8 0 1 0-16 0" />
    </svg>
  ),
  'Other': (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <line x1="12" y1="1" x2="12" y2="23" />
      <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
    </svg>
  ),
};

// Infer debt type from name
function inferDebtType(name) {
  const lowerName = name.toLowerCase();

  if (lowerName.includes('credit') || lowerName.includes('card') || lowerName.includes('visa') ||
      lowerName.includes('mastercard') || lowerName.includes('amex') || lowerName.includes('discover')) {
    return 'Credit Card';
  }
  if (lowerName.includes('mortgage') || lowerName.includes('home loan') || lowerName.includes('house')) {
    return 'Mortgage';
  }
  if (lowerName.includes('student') || lowerName.includes('education') || lowerName.includes('college') ||
      lowerName.includes('university') || lowerName.includes('school')) {
    return 'Student Loan';
  }
  if (lowerName.includes('auto') || lowerName.includes('car') || lowerName.includes('vehicle') ||
      lowerName.includes('truck') || lowerName.includes('motorcycle')) {
    return 'Auto Loan';
  }
  if (lowerName.includes('personal') || lowerName.includes('line of credit') || lowerName.includes('loc')) {
    return 'Personal Loan';
  }
  return 'Other';
}

function DebtList({ debts, onEdit, onDelete }) {
  const [collapsedCategories, setCollapsedCategories] = useState({});

  // Group debts by inferred type
  const groupedDebts = useMemo(() => {
    const groups = {};
    debts.forEach(debt => {
      const type = inferDebtType(debt.name);
      if (!groups[type]) {
        groups[type] = { debts: [], total: 0 };
      }
      groups[type].debts.push(debt);
      groups[type].total += debt.currentBalance;
    });
    return groups;
  }, [debts]);

  const toggleCategory = (category) => {
    setCollapsedCategories(prev => ({
      ...prev,
      [category]: !prev[category]
    }));
  };

  if (debts.length === 0) {
    return (
      <div className="empty-list">
        <p>No debts added yet</p>
      </div>
    );
  }

  // Sort categories with a specific order
  const categoryOrder = ['Credit Card', 'Mortgage', 'Auto Loan', 'Student Loan', 'Personal Loan', 'Other'];
  const categories = Object.keys(groupedDebts).sort((a, b) => {
    return categoryOrder.indexOf(a) - categoryOrder.indexOf(b);
  });

  return (
    <div className="categorized-list">
      {categories.map(category => {
        const { debts: categoryDebts, total } = groupedDebts[category];
        const isCollapsed = collapsedCategories[category];
        const Icon = DebtIcons[category] || DebtIcons['Other'];

        return (
          <div key={category} className="category-group">
            <button
              className="category-header"
              onClick={() => toggleCategory(category)}
            >
              <div className="category-left">
                <span className={`collapse-icon ${isCollapsed ? '' : 'open'}`}>&#9656;</span>
                <span className="category-icon debt-icon">{Icon}</span>
                <span className="category-name">{category}</span>
                <span className="category-count">{categoryDebts.length}</span>
              </div>
              <span className="category-total negative">
                ${total.toLocaleString('en-US', { minimumFractionDigits: 2 })}
              </span>
            </button>

            {!isCollapsed && (
              <div className="category-items">
                {categoryDebts.map(debt => (
                  <div key={debt.id} className="item-card">
                    <div className="item-info">
                      <div className="item-name">{debt.name}</div>
                      <div className="item-type">
                        {debt.interestRate ? `${debt.interestRate}% APR` : 'No interest'}
                        {debt.minimumPayment && ` · $${debt.minimumPayment}/mo min`}
                      </div>
                    </div>
                    <div className="item-value negative">
                      ${debt.currentBalance.toLocaleString('en-US', { minimumFractionDigits: 2 })}
                    </div>
                    <div className="item-actions">
                      <button className="btn-icon" onClick={() => onEdit(debt)} title="Edit">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                        </svg>
                      </button>
                      <button className="btn-icon btn-danger" onClick={() => onDelete(debt.id)} title="Delete">
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

export default DebtList;
