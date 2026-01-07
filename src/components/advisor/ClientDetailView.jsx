import React, { useState } from 'react';
import { useClientContext } from '../../contexts/ClientContext';
import NetWorthTab from '../tabs/NetWorthTab';
import BudgetTab from '../tabs/BudgetTab';
import ClientNotesPanel from './ClientNotesPanel';
import ClientMessagesPanel from './ClientMessagesPanel';
import ClientGoalsPanel from './ClientGoalsPanel';
import ChatPanel from '../chat/ChatPanel';
import './AdvisorStyles.css';

export default function ClientDetailView({ onBack }) {
  const { activeClient } = useClientContext();
  const [activeTab, setActiveTab] = useState('networth');
  const [isChatOpen, setIsChatOpen] = useState(false);

  if (!activeClient) {
    return (
      <div className="client-detail-empty">
        <p>Select a client to view their details.</p>
        <button className="btn btn-secondary" onClick={onBack}>
          Back to Client List
        </button>
      </div>
    );
  }

  return (
    <div className={`client-detail-view ${isChatOpen ? 'chat-open' : ''}`}>
      <div className="client-detail-content">
        <div className="client-detail-header">
          <div className="client-info">
            <h2>{activeClient.name}</h2>
            <span className="client-email">{activeClient.email}</span>
          </div>
          <button
            className={`btn-chat-toggle ${isChatOpen ? 'active' : ''}`}
            onClick={() => setIsChatOpen(!isChatOpen)}
            title="Chat with Aurelia"
          >
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
            <span>Ask Aurelia</span>
          </button>
        </div>

        <nav className="client-tab-nav">
          <button
            className={`tab-btn ${activeTab === 'networth' ? 'active' : ''}`}
            onClick={() => setActiveTab('networth')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="12" y1="1" x2="12" y2="23" />
              <path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
            </svg>
            Net Worth
          </button>
          <button
            className={`tab-btn ${activeTab === 'budget' ? 'active' : ''}`}
            onClick={() => setActiveTab('budget')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <line x1="3" y1="9" x2="21" y2="9" />
              <line x1="9" y1="21" x2="9" y2="9" />
            </svg>
            Budget
          </button>
          <button
            className={`tab-btn ${activeTab === 'goals' ? 'active' : ''}`}
            onClick={() => setActiveTab('goals')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 6v6l4 2" />
            </svg>
            Goals
          </button>
          <button
            className={`tab-btn ${activeTab === 'notes' ? 'active' : ''}`}
            onClick={() => setActiveTab('notes')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
            </svg>
            Notes
          </button>
          <button
            className={`tab-btn ${activeTab === 'messages' ? 'active' : ''}`}
            onClick={() => setActiveTab('messages')}
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
            </svg>
            Messages
          </button>
        </nav>

        <main className="client-tab-container">
          {activeTab === 'networth' && <NetWorthTab />}
          {activeTab === 'budget' && <BudgetTab />}
          {activeTab === 'goals' && <ClientGoalsPanel />}
          {activeTab === 'notes' && <ClientNotesPanel />}
          {activeTab === 'messages' && <ClientMessagesPanel />}
        </main>
      </div>

      <ChatPanel isOpen={isChatOpen} onClose={() => setIsChatOpen(false)} />
    </div>
  );
}
