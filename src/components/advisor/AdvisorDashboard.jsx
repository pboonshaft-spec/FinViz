import React, { useState } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useClientContext } from '../../contexts/ClientContext';
import ClientList from './ClientList';
import ClientDetailView from './ClientDetailView';
import ClientContextBanner from './ClientContextBanner';
import UserManagement from './UserManagement';
import MessagesTab from '../tabs/MessagesTab';
import './AdvisorStyles.css';

export default function AdvisorDashboard() {
  const { user, logout } = useAuth();
  const { activeClient, clearClientContext } = useClientContext();
  const [activeTab, setActiveTab] = useState('clients'); // 'clients' | 'admin'
  const [view, setView] = useState('list'); // 'list' | 'detail'

  const handleClientSelect = (client) => {
    setView('detail');
  };

  const handleBackToClients = () => {
    clearClientContext();
    setView('list');
  };

  const handleTabChange = (tab) => {
    if (activeClient) {
      clearClientContext();
    }
    setView('list');
    setActiveTab(tab);
  };

  return (
    <div className="advisor-dashboard">
      <header className="advisor-header">
        <div className="header-content">
          <h1>Advisor Portal</h1>
          <p>Manage your clients and their financial plans</p>
        </div>
        <div className="user-menu">
          <span className="user-name">{user?.name}</span>
          <span className="user-role">Advisor</span>
          <button className="btn btn-secondary btn-sm" onClick={logout}>
            Sign Out
          </button>
        </div>
      </header>

      {!activeClient && (
        <nav className="advisor-nav">
          <button
            className={`nav-tab ${activeTab === 'clients' ? 'active' : ''}`}
            onClick={() => handleTabChange('clients')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" />
              <circle cx="9" cy="7" r="4" />
              <path d="M23 21v-2a4 4 0 0 0-3-3.87" />
              <path d="M16 3.13a4 4 0 0 1 0 7.75" />
            </svg>
            Clients
          </button>
          <button
            className={`nav-tab ${activeTab === 'messages' ? 'active' : ''}`}
            onClick={() => handleTabChange('messages')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
            Messages
          </button>
          <button
            className={`nav-tab ${activeTab === 'admin' ? 'active' : ''}`}
            onClick={() => handleTabChange('admin')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="3" />
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
            </svg>
            Admin
          </button>
        </nav>
      )}

      {activeClient && (
        <ClientContextBanner
          client={activeClient}
          onBack={handleBackToClients}
        />
      )}

      <main className="advisor-main">
        {activeTab === 'clients' && !activeClient && view === 'list' && (
          <ClientList onClientSelect={handleClientSelect} />
        )}
        {activeTab === 'clients' && (view === 'detail' || activeClient) && (
          <ClientDetailView onBack={handleBackToClients} />
        )}
        {activeTab === 'messages' && !activeClient && (
          <MessagesTab />
        )}
        {activeTab === 'admin' && !activeClient && (
          <UserManagement />
        )}
      </main>
    </div>
  );
}
