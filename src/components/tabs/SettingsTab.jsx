import React, { useState } from 'react';
import ChartCard from '../ChartCard';
import PlaidLink from '../networth/PlaidLink';
import { useAuth } from '../../contexts/AuthContext';
import { useApi } from '../../hooks/useApi';

function SettingsTab() {
  const { user, logout } = useAuth();
  const { syncTransactions, syncPlaidAccounts } = useApi();
  const [syncing, setSyncing] = useState(false);

  const handleAccountsLinked = async () => {
    setSyncing(true);
    try {
      // Sync account balances to assets/debts
      await syncPlaidAccounts();

      // Sync transactions for Budget tab (last 90 days)
      const endDate = new Date().toISOString().split('T')[0];
      const startDate = new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      await syncTransactions(startDate, endDate);
    } catch (err) {
      console.error('Sync failed:', err);
    } finally {
      setSyncing(false);
    }
  };

  return (
    <div className="tab-content">
      <div className="tab-header">
        <div className="tab-header-text">
          <h2>Settings</h2>
          <p>Manage your account and bank connections</p>
        </div>
      </div>

      <div className="settings-section">
        <ChartCard title="Account">
          <div className="account-info">
            <div className="account-row">
              <span className="account-label">Email</span>
              <span className="account-value">{user?.email}</span>
            </div>
            <div className="account-row">
              <span className="account-label">Name</span>
              <span className="account-value">{user?.name}</span>
            </div>
            <button className="btn btn-secondary" onClick={logout}>
              Sign Out
            </button>
          </div>
        </ChartCard>
      </div>

      <div className="settings-section">
        <ChartCard title="Bank Connections">
          <p className="settings-description">
            Connect your bank accounts to automatically sync transactions and account balances.
            Your data is securely transferred using Plaid's bank-level encryption.
          </p>
          {syncing && (
            <div className="sync-status">
              <span className="spinner-sm"></span>
              Syncing accounts and transactions...
            </div>
          )}
          <PlaidLink onAccountsLinked={handleAccountsLinked} />
        </ChartCard>
      </div>

      <div className="settings-section">
        <ChartCard title="Data Management">
          <p className="settings-description">
            Your financial data is stored securely and is only accessible by you.
          </p>
          <div className="data-actions">
            <button className="btn btn-secondary" disabled>
              Export All Data (Coming Soon)
            </button>
          </div>
        </ChartCard>
      </div>
    </div>
  );
}

export default SettingsTab;
