import React, { useState, useEffect, useRef } from 'react';
import { useChat } from '../../hooks/useChat';
import ChatMessage from './ChatMessage';
import ChatInput from './ChatInput';

export default function ChatPanel({ isOpen, onClose }) {
  const {
    messages,
    loading,
    error,
    sendMessage,
    clearMessages,
    initializeChat,
    checkChatStatus,
  } = useChat();

  const [isConfigured, setIsConfigured] = useState(true);
  const messagesEndRef = useRef(null);

  // Check if chat is configured and initialize
  useEffect(() => {
    if (isOpen) {
      checkChatStatus().then(configured => {
        setIsConfigured(configured);
        if (configured) {
          initializeChat();
        }
      });
    }
  }, [isOpen, checkChatStatus, initializeChat]);

  // Scroll to bottom when new messages arrive
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  const handleNewChat = () => {
    if (window.confirm('Start a new conversation? Current chat history will be cleared.')) {
      clearMessages();
      initializeChat();
    }
  };

  if (!isOpen) return null;

  return (
    <div className="chat-panel">
      {/* Header */}
      <div className="chat-header">
        <div className="chat-header-info">
          <div className="chat-avatar-header">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 16v-4M12 8h.01" />
            </svg>
          </div>
          <div>
            <h3 className="chat-title">Aurelia</h3>
            <p className="chat-subtitle">Financial Advisor</p>
          </div>
        </div>
        <div className="chat-header-actions">
          <button
            className="btn-icon"
            onClick={handleNewChat}
            title="New conversation"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M12 5v14M5 12h14" />
            </svg>
          </button>
          <button className="btn-icon" onClick={onClose} title="Close chat">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <line x1="18" y1="6" x2="6" y2="18" />
              <line x1="6" y1="6" x2="18" y2="18" />
            </svg>
          </button>
        </div>
      </div>

      {/* Messages */}
      <div className="chat-messages">
        {!isConfigured ? (
          <div className="chat-not-configured">
            <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <circle cx="12" cy="12" r="10" />
              <path d="M12 8v4M12 16h.01" />
            </svg>
            <h4>Chat Not Available</h4>
            <p>The AI assistant is not configured. Please set up the ANTHROPIC_API_KEY in your environment.</p>
          </div>
        ) : (
          <>
            {messages.map((message) => (
              <ChatMessage key={message.id} message={message} />
            ))}

            {loading && (
              <div className="chat-message chat-message-assistant">
                <div className="chat-avatar">
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M12 16v-4M12 8h.01" />
                  </svg>
                </div>
                <div className="chat-message-content">
                  <div className="chat-typing-indicator">
                    <span></span>
                    <span></span>
                    <span></span>
                  </div>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {/* Input */}
      {isConfigured && (
        <ChatInput onSend={sendMessage} disabled={loading} />
      )}
    </div>
  );
}
