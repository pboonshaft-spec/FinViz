import React, { useState, useEffect } from 'react';
import { useApi } from '../../hooks/useApi';
import { useAuth } from '../../contexts/AuthContext';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

const CATEGORIES = [
  { value: 'retirement', label: 'Retirement', color: '#8b5cf6' },
  { value: 'savings', label: 'Savings', color: '#10b981' },
  { value: 'debt', label: 'Debt Payoff', color: '#ef4444' },
  { value: 'investment', label: 'Investment', color: '#3b82f6' },
  { value: 'education', label: 'Education', color: '#f59e0b' },
  { value: 'emergency', label: 'Emergency Fund', color: '#06b6d4' },
  { value: 'major_purchase', label: 'Major Purchase', color: '#ec4899' },
  { value: 'other', label: 'Other', color: '#6b7280' },
];

const STATUSES = [
  { value: 'pending', label: 'Pending', color: '#6b7280' },
  { value: 'in_progress', label: 'In Progress', color: '#3b82f6' },
  { value: 'completed', label: 'Completed', color: '#10b981' },
  { value: 'on_hold', label: 'On Hold', color: '#f59e0b' },
];

const PRIORITIES = [
  { value: 'high', label: 'High', color: '#ef4444' },
  { value: 'medium', label: 'Medium', color: '#f59e0b' },
  { value: 'low', label: 'Low', color: '#6b7280' },
];

export default function GoalsTab() {
  const { token } = useAuth();
  const [goals, setGoals] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterStatus, setFilterStatus] = useState('');
  const [updatingId, setUpdatingId] = useState(null);

  const fetchGoals = async () => {
    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (filterStatus) params.append('status', filterStatus);

      const response = await fetch(
        `${API_BASE_URL}/api/goals?${params.toString()}`,
        {
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to fetch goals');
      const data = await response.json();
      setGoals(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchGoals();
  }, [filterStatus]);

  const handleUpdateProgress = async (goalId, newAmount) => {
    try {
      setUpdatingId(goalId);
      const response = await fetch(
        `${API_BASE_URL}/api/goals/${goalId}/progress`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ currentAmount: parseFloat(newAmount) }),
        }
      );

      if (!response.ok) throw new Error('Failed to update progress');
      const updated = await response.json();
      setGoals(prev => prev.map(g => g.id === goalId ? updated : g));
    } catch (err) {
      setError(err.message);
    } finally {
      setUpdatingId(null);
    }
  };

  const getCategoryInfo = (categoryValue) => {
    return CATEGORIES.find(c => c.value === categoryValue) || CATEGORIES[CATEGORIES.length - 1];
  };

  const getStatusInfo = (statusValue) => {
    return STATUSES.find(s => s.value === statusValue) || STATUSES[0];
  };

  const getPriorityInfo = (priorityValue) => {
    return PRIORITIES.find(p => p.value === priorityValue) || PRIORITIES[1];
  };

  const formatCurrency = (amount) => {
    if (!amount) return '$0';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(amount);
  };

  const formatDate = (dateStr) => {
    if (!dateStr) return '';
    return new Date(dateStr).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  };

  const getProgress = (goal) => {
    if (!goal.targetAmount || !goal.currentAmount) return null;
    return Math.min(100, Math.round((goal.currentAmount / goal.targetAmount) * 100));
  };

  const getDaysRemaining = (targetDate) => {
    if (!targetDate) return null;
    const today = new Date();
    const target = new Date(targetDate);
    const diffTime = target - today;
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    return diffDays;
  };

  // Separate active and completed goals
  const activeGoals = goals.filter(g => g.status !== 'completed');
  const completedGoals = goals.filter(g => g.status === 'completed');

  return (
    <div className="goals-tab">
      <div className="goals-tab-header">
        <div className="header-content">
          <h2>My Financial Goals</h2>
          <p>Track your progress towards financial milestones set by your advisor</p>
        </div>
        <div className="goals-tab-filter">
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
          >
            <option value="">All Goals</option>
            {STATUSES.map(s => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
        </div>
      </div>

      {error && (
        <div className="goals-tab-error">
          {error}
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      {loading ? (
        <div className="goals-tab-loading">
          <div className="loading-spinner"></div>
          <p>Loading your goals...</p>
        </div>
      ) : goals.length === 0 ? (
        <div className="goals-tab-empty">
          <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 6v6l4 2" />
          </svg>
          <h3>No Goals Yet</h3>
          <p>Your financial advisor hasn't set any goals for you yet. Check back later or reach out to discuss your financial objectives.</p>
        </div>
      ) : (
        <>
          {/* Summary Stats */}
          <div className="goals-summary">
            <div className="summary-stat">
              <span className="stat-value">{activeGoals.length}</span>
              <span className="stat-label">Active Goals</span>
            </div>
            <div className="summary-stat">
              <span className="stat-value">{completedGoals.length}</span>
              <span className="stat-label">Completed</span>
            </div>
            <div className="summary-stat">
              <span className="stat-value">
                {activeGoals.filter(g => g.priority === 'high').length}
              </span>
              <span className="stat-label">High Priority</span>
            </div>
          </div>

          {/* Active Goals */}
          {activeGoals.length > 0 && (
            <div className="goals-section">
              <h3>Active Goals</h3>
              <div className="goals-grid">
                {activeGoals.map(goal => {
                  const catInfo = getCategoryInfo(goal.category);
                  const statusInfo = getStatusInfo(goal.status);
                  const priorityInfo = getPriorityInfo(goal.priority);
                  const progress = getProgress(goal);
                  const daysRemaining = getDaysRemaining(goal.targetDate);

                  return (
                    <div
                      key={goal.id}
                      className={`goal-card priority-${goal.priority}`}
                    >
                      <div className="goal-card-header">
                        <span
                          className="goal-category-badge"
                          style={{ backgroundColor: catInfo.color }}
                        >
                          {catInfo.label}
                        </span>
                        <span
                          className="goal-status-badge"
                          style={{ color: statusInfo.color }}
                        >
                          {statusInfo.label}
                        </span>
                      </div>

                      <h4 className="goal-card-title">{goal.title}</h4>

                      {goal.description && (
                        <p className="goal-card-description">{goal.description}</p>
                      )}

                      {goal.targetAmount && (
                        <div className="goal-card-amount">
                          <div className="amount-display">
                            <span className="current">{formatCurrency(goal.currentAmount || 0)}</span>
                            <span className="separator">/</span>
                            <span className="target">{formatCurrency(goal.targetAmount)}</span>
                          </div>
                          {progress !== null && (
                            <div className="goal-progress-container">
                              <div className="progress-bar">
                                <div
                                  className="progress-fill"
                                  style={{
                                    width: `${progress}%`,
                                    backgroundColor: catInfo.color,
                                  }}
                                />
                              </div>
                              <span className="progress-percent">{progress}%</span>
                            </div>
                          )}
                          <div className="progress-update">
                            <input
                              type="number"
                              placeholder="Update progress..."
                              min="0"
                              step="100"
                              onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                  handleUpdateProgress(goal.id, e.target.value);
                                  e.target.value = '';
                                }
                              }}
                              disabled={updatingId === goal.id}
                            />
                            <span className="hint">Press Enter to update</span>
                          </div>
                        </div>
                      )}

                      <div className="goal-card-footer">
                        {goal.targetDate && (
                          <div className={`due-date ${daysRemaining && daysRemaining < 30 ? 'urgent' : ''}`}>
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                              <rect x="3" y="4" width="18" height="18" rx="2" ry="2" />
                              <line x1="16" y1="2" x2="16" y2="6" />
                              <line x1="8" y1="2" x2="8" y2="6" />
                              <line x1="3" y1="10" x2="21" y2="10" />
                            </svg>
                            {formatDate(goal.targetDate)}
                            {daysRemaining !== null && daysRemaining > 0 && (
                              <span className="days-left">({daysRemaining} days left)</span>
                            )}
                          </div>
                        )}
                        <span
                          className="priority-badge"
                          style={{ color: priorityInfo.color }}
                        >
                          {priorityInfo.label} Priority
                        </span>
                      </div>
                    </div>
                  );
                })}
              </div>
            </div>
          )}

          {/* Completed Goals */}
          {completedGoals.length > 0 && (
            <div className="goals-section completed-section">
              <h3>Completed Goals</h3>
              <div className="goals-grid">
                {completedGoals.map(goal => {
                  const catInfo = getCategoryInfo(goal.category);

                  return (
                    <div key={goal.id} className="goal-card completed">
                      <div className="goal-card-header">
                        <span
                          className="goal-category-badge"
                          style={{ backgroundColor: catInfo.color }}
                        >
                          {catInfo.label}
                        </span>
                        <span className="completed-badge">
                          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                            <polyline points="20 6 9 17 4 12" />
                          </svg>
                          Completed
                        </span>
                      </div>
                      <h4 className="goal-card-title">{goal.title}</h4>
                      {goal.targetAmount && (
                        <p className="goal-card-amount">
                          {formatCurrency(goal.targetAmount)}
                        </p>
                      )}
                      {goal.completedAt && (
                        <p className="completed-date">
                          Completed on {formatDate(goal.completedAt)}
                        </p>
                      )}
                    </div>
                  );
                })}
              </div>
            </div>
          )}
        </>
      )}
    </div>
  );
}
