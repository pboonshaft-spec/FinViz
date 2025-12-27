import React from 'react';
import ChatArtifact from './ChatArtifact';

// Simple markdown-like formatting
function formatContent(content) {
  if (!content) return '';

  // Split by code blocks first
  const parts = content.split(/(```[\s\S]*?```)/g);

  return parts.map((part, i) => {
    if (part.startsWith('```') && part.endsWith('```')) {
      // Code block
      const code = part.slice(3, -3).replace(/^[a-z]*\n?/, '');
      return (
        <pre key={i} className="chat-code-block">
          <code>{code}</code>
        </pre>
      );
    }

    // Process inline formatting
    return (
      <span key={i}>
        {part.split('\n').map((line, j) => {
          // Headers
          if (line.startsWith('### ')) {
            return <h4 key={j} className="chat-heading-3">{formatInline(line.slice(4))}</h4>;
          }
          if (line.startsWith('## ')) {
            return <h3 key={j} className="chat-heading-2">{formatInline(line.slice(3))}</h3>;
          }
          if (line.startsWith('# ')) {
            return <h2 key={j} className="chat-heading-1">{formatInline(line.slice(2))}</h2>;
          }

          // Bullet points
          if (line.startsWith('- ') || line.startsWith('* ')) {
            return <li key={j} className="chat-list-item">{formatInline(line.slice(2))}</li>;
          }

          // Numbered list
          const numberedMatch = line.match(/^(\d+)\.\s+(.+)/);
          if (numberedMatch) {
            return <li key={j} className="chat-list-item chat-numbered">{formatInline(numberedMatch[2])}</li>;
          }

          // Empty line = paragraph break
          if (line.trim() === '') {
            return <br key={j} />;
          }

          // Regular paragraph
          return <p key={j} className="chat-paragraph">{formatInline(line)}</p>;
        })}
      </span>
    );
  });
}

// Format inline elements (bold, italic, code)
function formatInline(text) {
  if (!text) return text;

  const parts = [];
  let remaining = text;
  let key = 0;

  while (remaining) {
    // Bold
    const boldMatch = remaining.match(/\*\*(.+?)\*\*/);
    // Italic
    const italicMatch = remaining.match(/\*(.+?)\*/);
    // Inline code
    const codeMatch = remaining.match(/`(.+?)`/);

    // Find earliest match
    const matches = [
      boldMatch && { type: 'bold', match: boldMatch, index: remaining.indexOf(boldMatch[0]) },
      italicMatch && !boldMatch?.index === italicMatch?.index && { type: 'italic', match: italicMatch, index: remaining.indexOf(italicMatch[0]) },
      codeMatch && { type: 'code', match: codeMatch, index: remaining.indexOf(codeMatch[0]) },
    ].filter(Boolean).sort((a, b) => a.index - b.index);

    if (matches.length === 0) {
      parts.push(remaining);
      break;
    }

    const first = matches[0];

    // Add text before match
    if (first.index > 0) {
      parts.push(remaining.slice(0, first.index));
    }

    // Add formatted element
    if (first.type === 'bold') {
      parts.push(<strong key={key++}>{first.match[1]}</strong>);
    } else if (first.type === 'italic') {
      parts.push(<em key={key++}>{first.match[1]}</em>);
    } else if (first.type === 'code') {
      parts.push(<code key={key++} className="chat-inline-code">{first.match[1]}</code>);
    }

    remaining = remaining.slice(first.index + first.match[0].length);
  }

  return parts;
}

export default function ChatMessage({ message }) {
  const isUser = message.role === 'user';
  const isError = message.isError;

  return (
    <div className={`chat-message ${isUser ? 'chat-message-user' : 'chat-message-assistant'} ${isError ? 'chat-message-error' : ''}`}>
      {!isUser && (
        <div className="chat-avatar">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="12" cy="12" r="10" />
            <path d="M12 16v-4M12 8h.01" />
          </svg>
        </div>
      )}

      <div className="chat-message-content">
        {/* Tool usage indicator */}
        {message.toolsUsed?.length > 0 && (
          <div className="chat-tools-used">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />
            </svg>
            <span>Used: {message.toolsUsed.join(', ')}</span>
          </div>
        )}

        {/* Message text */}
        <div className="chat-message-text">
          {formatContent(message.content)}
        </div>

        {/* A2UI Artifacts (charts, tables, metric cards) */}
        {message.artifacts?.map((artifact, index) => (
          <ChatArtifact key={index} artifact={artifact} />
        ))}

        {/* Timestamp */}
        <div className="chat-message-time">
          {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
        </div>
      </div>
    </div>
  );
}
