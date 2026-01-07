import React, { useEffect, useRef } from 'react';

function MessageThread({ messages, currentUserId, loading }) {
  const messagesEndRef = useRef(null);
  const containerRef = useRef(null);

  // Scroll to bottom on new messages
  useEffect(() => {
    if (messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [messages]);

  const formatTime = (dateString) => {
    const date = new Date(dateString);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return 'Today';
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else {
      return date.toLocaleDateString([], {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
      });
    }
  };

  // Group messages by date
  const groupedMessages = messages.reduce((groups, message) => {
    const date = new Date(message.createdAt).toDateString();
    if (!groups[date]) {
      groups[date] = [];
    }
    groups[date].push(message);
    return groups;
  }, {});

  // Reverse order for display (oldest first)
  const sortedDates = Object.keys(groupedMessages).sort(
    (a, b) => new Date(a) - new Date(b)
  );

  if (loading && messages.length === 0) {
    return (
      <div className="message-thread-loading">
        <div className="loading-spinner"></div>
        <p>Loading messages...</p>
      </div>
    );
  }

  if (messages.length === 0) {
    return (
      <div className="message-thread-empty">
        <div className="empty-icon">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
          </svg>
        </div>
        <p>No messages yet</p>
        <p className="hint">Start the conversation by sending a message</p>
      </div>
    );
  }

  return (
    <div className="message-thread" ref={containerRef}>
      {sortedDates.map((dateStr) => (
        <div key={dateStr} className="message-date-group">
          <div className="date-divider">
            <span>{formatDate(dateStr)}</span>
          </div>
          {groupedMessages[dateStr]
            .sort((a, b) => new Date(a.createdAt) - new Date(b.createdAt))
            .map((message) => {
              const isOwn = message.senderId === currentUserId || message.isOwn;
              const isDecrypted = message.decryptedContent !== undefined;
              const displayContent = isDecrypted
                ? message.decryptedContent || '[Unable to decrypt]'
                : message.encryptedContent.substring(0, 20) + '...';

              return (
                <div
                  key={message.id}
                  className={`message-bubble ${isOwn ? 'own' : 'other'}`}
                >
                  {!isOwn && message.senderName && (
                    <span className="message-sender">{message.senderName}</span>
                  )}
                  <div className="message-content">
                    {isDecrypted ? (
                      <p>{displayContent}</p>
                    ) : (
                      <p className="encrypted-placeholder">
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                          <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                        </svg>
                        Encrypted message
                      </p>
                    )}
                  </div>
                  <div className="message-meta">
                    <span className="message-time">{formatTime(message.createdAt)}</span>
                    {isOwn && message.readAt && (
                      <span className="message-read" title="Read">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                          <polyline points="20 6 9 17 4 12" />
                        </svg>
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
        </div>
      ))}
      <div ref={messagesEndRef} />
    </div>
  );
}

export default MessageThread;
