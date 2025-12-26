import React, { useState } from 'react';

function TransactionList({ transactions }) {
  const [filter, setFilter] = useState('all'); // 'all', 'income', 'expenses'
  const [searchTerm, setSearchTerm] = useState('');

  const filteredTransactions = transactions.filter(t => {
    // Filter by type
    if (filter === 'income' && t.amount >= 0) return false;
    if (filter === 'expenses' && t.amount < 0) return false;

    // Filter by search term
    if (searchTerm) {
      const term = searchTerm.toLowerCase();
      return (
        t.name?.toLowerCase().includes(term) ||
        t.merchantName?.toLowerCase().includes(term) ||
        t.category?.toLowerCase().includes(term)
      );
    }

    return true;
  });

  const formatAmount = (amount) => {
    const absAmount = Math.abs(amount).toLocaleString('en-US', {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    });
    // In Plaid, positive = money out (expense), negative = money in (income)
    if (amount < 0) {
      return <span className="amount-positive">+${absAmount}</span>;
    }
    return <span className="amount-negative">-${absAmount}</span>;
  };

  const formatDate = (dateStr) => {
    const date = new Date(dateStr + 'T00:00:00');
    return date.toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric'
    });
  };

  return (
    <div className="transaction-list-container">
      <div className="transaction-filters">
        <div className="btn-group">
          <button
            className={`btn btn-sm ${filter === 'all' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setFilter('all')}
          >
            All
          </button>
          <button
            className={`btn btn-sm ${filter === 'income' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setFilter('income')}
          >
            Income
          </button>
          <button
            className={`btn btn-sm ${filter === 'expenses' ? 'btn-primary' : 'btn-secondary'}`}
            onClick={() => setFilter('expenses')}
          >
            Expenses
          </button>
        </div>
        <input
          type="text"
          className="search-input"
          placeholder="Search transactions..."
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
        />
      </div>

      <div className="transaction-list">
        {filteredTransactions.length === 0 ? (
          <div className="empty-state-small">
            <p>No transactions found</p>
          </div>
        ) : (
          filteredTransactions.slice(0, 50).map((t, idx) => (
            <div key={t.id || idx} className="transaction-item">
              <div className="transaction-info">
                <div className="transaction-name">
                  {t.merchantName || t.name}
                </div>
                <div className="transaction-meta">
                  <span className="transaction-date">{formatDate(t.date)}</span>
                  {t.category && (
                    <span className="transaction-category">{t.category}</span>
                  )}
                  {t.accountName && (
                    <span className="transaction-account">{t.accountName}</span>
                  )}
                </div>
              </div>
              <div className="transaction-amount">
                {formatAmount(t.amount)}
              </div>
            </div>
          ))
        )}
        {filteredTransactions.length > 50 && (
          <div className="transaction-more">
            Showing 50 of {filteredTransactions.length} transactions
          </div>
        )}
      </div>
    </div>
  );
}

export default TransactionList;
