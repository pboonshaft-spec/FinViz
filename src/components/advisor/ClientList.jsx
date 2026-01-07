import React, { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useClientContext } from '../../contexts/ClientContext';
import './AdvisorStyles.css';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

export default function ClientList({ onClientSelect }) {
  const { token } = useAuth();
  const { clients, loadingClients, fetchClients, switchClient, activeClient } = useClientContext();
  const [showInviteModal, setShowInviteModal] = useState(false);
  const [inviteEmail, setInviteEmail] = useState('');
  const [inviting, setInviting] = useState(false);
  const [inviteError, setInviteError] = useState(null);
  const [inviteSuccess, setInviteSuccess] = useState(null);
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    fetchClients();
  }, [fetchClients]);

  const filteredClients = clients.filter(client =>
    client.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    client.email.toLowerCase().includes(searchQuery.toLowerCase())
  );

  const handleInviteClient = async (e) => {
    e.preventDefault();
    setInviting(true);
    setInviteError(null);
    setInviteSuccess(null);

    try {
      const response = await fetch(`${API_BASE_URL}/api/advisor/clients/invite`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email: inviteEmail }),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to send invitation');
      }

      setInviteSuccess(`Invitation sent to ${inviteEmail}`);
      setInviteEmail('');
      setTimeout(() => {
        setShowInviteModal(false);
        setInviteSuccess(null);
      }, 2000);
    } catch (err) {
      setInviteError(err.message);
    } finally {
      setInviting(false);
    }
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const handleClientClick = (client) => {
    switchClient(client);
    if (onClientSelect) {
      onClientSelect(client);
    }
  };

  if (loadingClients) {
    return <div className="client-list-loading">Loading clients...</div>;
  }

  return (
    <div className="client-list">
      <div className="client-list-header">
        <h2>Your Clients</h2>
        <button className="btn btn-primary btn-sm" onClick={() => setShowInviteModal(true)}>
          + Invite Client
        </button>
      </div>

      {clients.length > 0 && (
        <div className="client-search">
          <svg className="search-icon" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <path d="M21 21l-4.35-4.35" />
          </svg>
          <input
            type="text"
            placeholder="Search clients by name or email..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="client-search-input"
          />
          {searchQuery && (
            <button className="search-clear" onClick={() => setSearchQuery('')}>
              &times;
            </button>
          )}
        </div>
      )}

      {clients.length === 0 ? (
        <div className="client-list-empty">
          <p>No clients yet. Invite your first client to get started.</p>
        </div>
      ) : filteredClients.length === 0 ? (
        <div className="client-list-empty">
          <p>No clients match "{searchQuery}"</p>
          <button className="btn btn-secondary btn-sm" onClick={() => setSearchQuery('')}>
            Clear search
          </button>
        </div>
      ) : (
        <div className="client-cards">
          {filteredClients.map((client) => (
            <div
              key={client.id}
              className={`client-card ${activeClient?.id === client.id ? 'active' : ''}`}
              onClick={() => handleClientClick(client)}
            >
              <div className="client-card-header">
                <h3>{client.name}</h3>
                <span className={`client-status ${client.status}`}>{client.status}</span>
              </div>
              <div className="client-card-email">{client.email}</div>
              <div className="client-card-metrics">
                <div className="client-metric">
                  <span className="metric-label">Net Worth</span>
                  <span className="metric-value">{formatCurrency(client.netWorth || 0)}</span>
                </div>
                <div className="client-metric">
                  <span className="metric-label">Assets</span>
                  <span className="metric-value">{formatCurrency(client.totalAssets || 0)}</span>
                </div>
                <div className="client-metric">
                  <span className="metric-label">Debts</span>
                  <span className="metric-value negative">{formatCurrency(client.totalDebts || 0)}</span>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showInviteModal && (
        <div className="modal-overlay" onClick={() => setShowInviteModal(false)}>
          <div className="modal-content" onClick={(e) => e.stopPropagation()}>
            <h3>Invite Client</h3>
            <form onSubmit={handleInviteClient}>
              <div className="form-group">
                <label htmlFor="inviteEmail">Client Email</label>
                <input
                  type="email"
                  id="inviteEmail"
                  value={inviteEmail}
                  onChange={(e) => setInviteEmail(e.target.value)}
                  placeholder="client@example.com"
                  required
                />
              </div>
              {inviteError && <div className="error-message">{inviteError}</div>}
              {inviteSuccess && <div className="success-message">{inviteSuccess}</div>}
              <div className="modal-actions">
                <button type="button" className="btn btn-secondary" onClick={() => setShowInviteModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary" disabled={inviting}>
                  {inviting ? 'Sending...' : 'Send Invitation'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
