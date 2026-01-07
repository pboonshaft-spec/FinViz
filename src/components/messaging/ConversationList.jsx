import React from 'react';

function ConversationList({ conversations, selectedId, onSelect, isAdvisor }) {
  const formatTime = (dateString) => {
    if (!dateString) return '';
    const date = new Date(dateString);
    const now = new Date();
    const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else if (diffDays < 7) {
      return date.toLocaleDateString([], { weekday: 'short' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
  };

  const getUnreadCount = (conv) => {
    return isAdvisor ? conv.unreadCountAdvisor : conv.unreadCountClient;
  };

  const getDisplayName = (conv) => {
    return isAdvisor ? conv.clientName : conv.advisorName;
  };

  if (conversations.length === 0) {
    return (
      <div className="conversation-list-empty">
        <p>No conversations yet</p>
        {isAdvisor && <p className="hint">Select a client to start a conversation</p>}
      </div>
    );
  }

  return (
    <div className="conversation-list">
      {conversations.map((conv) => {
        const unread = getUnreadCount(conv);
        const isSelected = selectedId === conv.id;

        return (
          <div
            key={conv.id}
            className={`conversation-item ${isSelected ? 'selected' : ''} ${unread > 0 ? 'unread' : ''}`}
            onClick={() => onSelect(conv)}
          >
            <div className="conversation-avatar">
              {getDisplayName(conv)?.charAt(0)?.toUpperCase() || '?'}
            </div>
            <div className="conversation-info">
              <div className="conversation-header">
                <span className="conversation-name">{getDisplayName(conv)}</span>
                <span className="conversation-time">
                  {formatTime(conv.lastMessageAt)}
                </span>
              </div>
              {isAdvisor && conv.clientEmail && (
                <span className="conversation-email">{conv.clientEmail}</span>
              )}
            </div>
            {unread > 0 && (
              <span className="unread-badge">{unread > 99 ? '99+' : unread}</span>
            )}
          </div>
        );
      })}
    </div>
  );
}

export default ConversationList;
