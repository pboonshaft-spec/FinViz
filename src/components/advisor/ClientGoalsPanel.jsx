import React, { useState, useEffect } from 'react';
import { useClientContext } from '../../contexts/ClientContext';
import { useAuth } from '../../contexts/AuthContext';
import './AdvisorStyles.css';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

const CATEGORIES = [
  { value: 'retirement', label: 'Retirement', color: '#8b5cf6', icon: 'M12 2a10 10 0 1 0 10 10A10 10 0 0 0 12 2zm0 18a8 8 0 1 1 8-8 8 8 0 0 1-8 8zm1-13h-2v6l4.25 2.52.77-1.28-3.02-1.79z' },
  { value: 'savings', label: 'Savings', color: '#10b981', icon: 'M19 5h-2V3H7v2H5c-1.1 0-2 .9-2 2v1c0 2.55 1.92 4.63 4.39 4.94A5.01 5.01 0 0 0 11 15.9V19H7v2h10v-2h-4v-3.1a5.01 5.01 0 0 0 3.61-2.96C19.08 12.63 21 10.55 21 8V7c0-1.1-.9-2-2-2z' },
  { value: 'debt', label: 'Debt Payoff', color: '#ef4444', icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z' },
  { value: 'investment', label: 'Investment', color: '#3b82f6', icon: 'M16 6l2.29 2.29-4.88 4.88-4-4L2 16.59 3.41 18l6-6 4 4 6.3-6.29L22 12V6z' },
  { value: 'education', label: 'Education', color: '#f59e0b', icon: 'M5 13.18v4L12 21l7-3.82v-4L12 17l-7-3.82zM12 3L1 9l11 6 9-4.91V17h2V9L12 3z' },
  { value: 'emergency', label: 'Emergency Fund', color: '#06b6d4', icon: 'M11.99 2C6.47 2 2 6.48 2 12s4.47 10 9.99 10C17.52 22 22 17.52 22 12S17.52 2 11.99 2zM12 20c-4.42 0-8-3.58-8-8s3.58-8 8-8 8 3.58 8 8-3.58 8-8 8zm.5-13H11v6l5.25 3.15.75-1.23-4.5-2.67z' },
  { value: 'major_purchase', label: 'Major Purchase', color: '#ec4899', icon: 'M19 3H5c-1.11 0-2 .9-2 2v14c0 1.1.89 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm-2 10h-4v4h-2v-4H7v-2h4V7h2v4h4v2z' },
  { value: 'other', label: 'Other', color: '#6b7280', icon: 'M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z' },
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

export default function ClientGoalsPanel() {
  const { activeClient } = useClientContext();
  const { token } = useAuth();
  const [goals, setGoals] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [filterStatus, setFilterStatus] = useState('');
  const [showAddForm, setShowAddForm] = useState(false);
  const [saving, setSaving] = useState(false);

  // New goal form
  const [newGoal, setNewGoal] = useState({
    title: '',
    description: '',
    category: 'other',
    priority: 'medium',
    targetAmount: '',
    targetDate: '',
  });

  // Edit state
  const [editingId, setEditingId] = useState(null);
  const [editGoal, setEditGoal] = useState({});

  const fetchGoals = async () => {
    if (!activeClient) return;

    try {
      setLoading(true);
      const params = new URLSearchParams();
      if (filterStatus) params.append('status', filterStatus);

      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/goals?${params.toString()}`,
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
  }, [activeClient?.id, filterStatus]);

  const handleAddGoal = async (e) => {
    e.preventDefault();
    if (!newGoal.title.trim() || !activeClient) return;

    try {
      setSaving(true);
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/goals`,
        {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            title: newGoal.title.trim(),
            description: newGoal.description.trim(),
            category: newGoal.category,
            priority: newGoal.priority,
            targetAmount: newGoal.targetAmount ? parseFloat(newGoal.targetAmount) : null,
            targetDate: newGoal.targetDate || null,
          }),
        }
      );

      if (!response.ok) throw new Error('Failed to create goal');
      const created = await response.json();
      setGoals(prev => [created, ...prev]);
      setNewGoal({
        title: '',
        description: '',
        category: 'other',
        priority: 'medium',
        targetAmount: '',
        targetDate: '',
      });
      setShowAddForm(false);
    } catch (err) {
      setError(err.message);
    } finally {
      setSaving(false);
    }
  };

  const handleUpdateGoal = async (goalId) => {
    try {
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/goals/${goalId}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            title: editGoal.title,
            description: editGoal.description,
            category: editGoal.category,
            status: editGoal.status,
            priority: editGoal.priority,
            targetAmount: editGoal.targetAmount ? parseFloat(editGoal.targetAmount) : null,
            currentAmount: editGoal.currentAmount ? parseFloat(editGoal.currentAmount) : null,
            targetDate: editGoal.targetDate || null,
          }),
        }
      );

      if (!response.ok) throw new Error('Failed to update goal');
      const updated = await response.json();
      setGoals(prev => prev.map(g => g.id === goalId ? updated : g));
      setEditingId(null);
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteGoal = async (goalId) => {
    if (!confirm('Delete this goal?')) return;

    try {
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/goals/${goalId}`,
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${token}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to delete goal');
      setGoals(prev => prev.filter(g => g.id !== goalId));
    } catch (err) {
      setError(err.message);
    }
  };

  const handleStatusChange = async (goal, newStatus) => {
    try {
      const response = await fetch(
        `${API_BASE_URL}/api/advisor/clients/${activeClient.id}/goals/${goal.id}`,
        {
          method: 'PUT',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({ status: newStatus }),
        }
      );

      if (!response.ok) throw new Error('Failed to update goal');
      const updated = await response.json();
      setGoals(prev => prev.map(g => g.id === goal.id ? updated : g));
    } catch (err) {
      setError(err.message);
    }
  };

  const startEdit = (goal) => {
    setEditingId(goal.id);
    setEditGoal({
      title: goal.title,
      description: goal.description || '',
      category: goal.category,
      status: goal.status,
      priority: goal.priority,
      targetAmount: goal.targetAmount || '',
      currentAmount: goal.currentAmount || '',
      targetDate: goal.targetDate || '',
    });
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
    if (!amount) return '';
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

  if (!activeClient) {
    return <div className="goals-panel empty">Select a client to view goals</div>;
  }

  return (
    <div className="goals-panel">
      <div className="goals-header">
        <h3>Financial Goals</h3>
        <div className="goals-actions">
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="goals-filter"
          >
            <option value="">All Statuses</option>
            {STATUSES.map(s => (
              <option key={s.value} value={s.value}>{s.label}</option>
            ))}
          </select>
          <button
            className="btn btn-primary"
            onClick={() => setShowAddForm(!showAddForm)}
          >
            {showAddForm ? 'Cancel' : '+ Add Goal'}
          </button>
        </div>
      </div>

      {error && (
        <div className="goals-error">
          {error}
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      {showAddForm && (
        <form className="add-goal-form" onSubmit={handleAddGoal}>
          <div className="form-row">
            <input
              type="text"
              value={newGoal.title}
              onChange={(e) => setNewGoal(prev => ({ ...prev, title: e.target.value }))}
              placeholder="Goal title"
              className="form-input"
              required
            />
          </div>
          <div className="form-row">
            <textarea
              value={newGoal.description}
              onChange={(e) => setNewGoal(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Description (optional)"
              rows={2}
              className="form-input"
            />
          </div>
          <div className="form-row form-row-grid">
            <div className="form-group">
              <label>Category</label>
              <select
                value={newGoal.category}
                onChange={(e) => setNewGoal(prev => ({ ...prev, category: e.target.value }))}
              >
                {CATEGORIES.map(cat => (
                  <option key={cat.value} value={cat.value}>{cat.label}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Priority</label>
              <select
                value={newGoal.priority}
                onChange={(e) => setNewGoal(prev => ({ ...prev, priority: e.target.value }))}
              >
                {PRIORITIES.map(p => (
                  <option key={p.value} value={p.value}>{p.label}</option>
                ))}
              </select>
            </div>
            <div className="form-group">
              <label>Target Amount</label>
              <input
                type="number"
                value={newGoal.targetAmount}
                onChange={(e) => setNewGoal(prev => ({ ...prev, targetAmount: e.target.value }))}
                placeholder="$0"
                min="0"
                step="100"
              />
            </div>
            <div className="form-group">
              <label>Target Date</label>
              <input
                type="date"
                value={newGoal.targetDate}
                onChange={(e) => setNewGoal(prev => ({ ...prev, targetDate: e.target.value }))}
              />
            </div>
          </div>
          <div className="form-actions">
            <button type="submit" className="btn btn-primary" disabled={saving}>
              {saving ? 'Creating...' : 'Create Goal'}
            </button>
            <button
              type="button"
              className="btn btn-secondary"
              onClick={() => setShowAddForm(false)}
            >
              Cancel
            </button>
          </div>
        </form>
      )}

      <div className="goals-list">
        {loading ? (
          <div className="goals-loading">Loading goals...</div>
        ) : goals.length === 0 ? (
          <div className="goals-empty">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 6v6l4 2" />
            </svg>
            <p>No goals set yet</p>
            <span>Create the first goal for {activeClient.name}</span>
          </div>
        ) : (
          goals.map(goal => {
            const catInfo = getCategoryInfo(goal.category);
            const statusInfo = getStatusInfo(goal.status);
            const priorityInfo = getPriorityInfo(goal.priority);
            const progress = getProgress(goal);
            const isEditing = editingId === goal.id;

            return (
              <div
                key={goal.id}
                className={`goal-card status-${goal.status}`}
              >
                {isEditing ? (
                  <div className="goal-edit-form">
                    <input
                      type="text"
                      value={editGoal.title}
                      onChange={(e) => setEditGoal(prev => ({ ...prev, title: e.target.value }))}
                      className="form-input"
                    />
                    <textarea
                      value={editGoal.description}
                      onChange={(e) => setEditGoal(prev => ({ ...prev, description: e.target.value }))}
                      rows={2}
                      className="form-input"
                    />
                    <div className="form-row form-row-grid">
                      <select
                        value={editGoal.category}
                        onChange={(e) => setEditGoal(prev => ({ ...prev, category: e.target.value }))}
                      >
                        {CATEGORIES.map(cat => (
                          <option key={cat.value} value={cat.value}>{cat.label}</option>
                        ))}
                      </select>
                      <select
                        value={editGoal.status}
                        onChange={(e) => setEditGoal(prev => ({ ...prev, status: e.target.value }))}
                      >
                        {STATUSES.map(s => (
                          <option key={s.value} value={s.value}>{s.label}</option>
                        ))}
                      </select>
                      <select
                        value={editGoal.priority}
                        onChange={(e) => setEditGoal(prev => ({ ...prev, priority: e.target.value }))}
                      >
                        {PRIORITIES.map(p => (
                          <option key={p.value} value={p.value}>{p.label}</option>
                        ))}
                      </select>
                    </div>
                    <div className="form-row form-row-grid">
                      <input
                        type="number"
                        value={editGoal.targetAmount}
                        onChange={(e) => setEditGoal(prev => ({ ...prev, targetAmount: e.target.value }))}
                        placeholder="Target $"
                      />
                      <input
                        type="number"
                        value={editGoal.currentAmount}
                        onChange={(e) => setEditGoal(prev => ({ ...prev, currentAmount: e.target.value }))}
                        placeholder="Current $"
                      />
                      <input
                        type="date"
                        value={editGoal.targetDate}
                        onChange={(e) => setEditGoal(prev => ({ ...prev, targetDate: e.target.value }))}
                      />
                    </div>
                    <div className="form-actions">
                      <button
                        className="btn btn-primary btn-sm"
                        onClick={() => handleUpdateGoal(goal.id)}
                      >
                        Save
                      </button>
                      <button
                        className="btn btn-secondary btn-sm"
                        onClick={() => setEditingId(null)}
                      >
                        Cancel
                      </button>
                    </div>
                  </div>
                ) : (
                  <>
                    <div className="goal-header">
                      <div className="goal-category" style={{ backgroundColor: catInfo.color }}>
                        {catInfo.label}
                      </div>
                      <div className="goal-priority" style={{ color: priorityInfo.color }}>
                        {priorityInfo.label}
                      </div>
                      <select
                        className="goal-status-select"
                        value={goal.status}
                        onChange={(e) => handleStatusChange(goal, e.target.value)}
                        style={{ color: statusInfo.color }}
                      >
                        {STATUSES.map(s => (
                          <option key={s.value} value={s.value}>{s.label}</option>
                        ))}
                      </select>
                    </div>

                    <h4 className="goal-title">{goal.title}</h4>
                    {goal.description && (
                      <p className="goal-description">{goal.description}</p>
                    )}

                    {(goal.targetAmount || goal.targetDate) && (
                      <div className="goal-details">
                        {goal.targetAmount && (
                          <div className="goal-amount">
                            <span className="label">Target:</span>
                            <span className="value">{formatCurrency(goal.targetAmount)}</span>
                            {goal.currentAmount > 0 && (
                              <span className="current">
                                (Current: {formatCurrency(goal.currentAmount)})
                              </span>
                            )}
                          </div>
                        )}
                        {goal.targetDate && (
                          <div className="goal-date">
                            <span className="label">Due:</span>
                            <span className="value">{formatDate(goal.targetDate)}</span>
                          </div>
                        )}
                      </div>
                    )}

                    {progress !== null && (
                      <div className="goal-progress">
                        <div className="progress-bar">
                          <div
                            className="progress-fill"
                            style={{
                              width: `${progress}%`,
                              backgroundColor: progress >= 100 ? '#10b981' : catInfo.color,
                            }}
                          />
                        </div>
                        <span className="progress-text">{progress}%</span>
                      </div>
                    )}

                    <div className="goal-actions">
                      <button
                        className="goal-action-btn"
                        onClick={() => startEdit(goal)}
                        title="Edit"
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7" />
                          <path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z" />
                        </svg>
                      </button>
                      <button
                        className="goal-action-btn delete"
                        onClick={() => handleDeleteGoal(goal.id)}
                        title="Delete"
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <polyline points="3 6 5 6 21 6" />
                          <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                        </svg>
                      </button>
                    </div>
                  </>
                )}
              </div>
            );
          })
        )}
      </div>
    </div>
  );
}
