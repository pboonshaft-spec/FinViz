import React, { useState, useEffect, useCallback } from 'react';
import { useClientContext } from '../../contexts/ClientContext';
import { useAuth } from '../../contexts/AuthContext';
import { useApi } from '../../hooks/useApi';
import MessageThread from '../messaging/MessageThread';
import MessageInput from '../messaging/MessageInput';
import {
  initializeEncryption,
  encryptMessage,
  decryptMessage,
  getStoredKeys,
  getUserPublicKey,
} from '../../utils/encryption';
import './AdvisorStyles.css';

export default function ClientMessagesPanel() {
  const { activeClient } = useClientContext();
  const { user } = useAuth();
  const {
    getConversations,
    startConversation,
    getMessages,
    sendMessage,
    markAsRead,
    registerPublicKey,
    getPublicKey,
  } = useApi();

  const [conversation, setConversation] = useState(null);
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [encryptionReady, setEncryptionReady] = useState(false);
  const [userKeys, setUserKeys] = useState(null);
  const [clientPublicKey, setClientPublicKey] = useState(null);
  const [error, setError] = useState(null);

  // Initialize encryption on mount
  useEffect(() => {
    const init = async () => {
      try {
        const keys = await initializeEncryption(registerPublicKey);
        setUserKeys(keys);
        setEncryptionReady(true);
      } catch (error) {
        console.error('Failed to initialize encryption:', error);
        setError('Failed to initialize secure messaging');
      }
    };
    init();
  }, [registerPublicKey]);

  // Find or create conversation with active client
  useEffect(() => {
    if (!activeClient) return;

    const findOrCreateConversation = async () => {
      try {
        setLoading(true);
        setError(null);

        // Get all conversations
        const convs = await getConversations();

        // Find existing conversation with this client
        const existingConv = convs?.find(c => c.clientId === activeClient.id);

        if (existingConv) {
          setConversation(existingConv);
        } else {
          // Start new conversation with client
          const newConv = await startConversation(activeClient.id);
          setConversation(newConv);
        }
      } catch (err) {
        console.error('Failed to find/create conversation:', err);
        setError('Failed to load conversation');
      } finally {
        setLoading(false);
      }
    };

    findOrCreateConversation();
  }, [activeClient?.id]);

  // Load client's public key
  useEffect(() => {
    if (!activeClient) return;

    const loadClientKey = async () => {
      try {
        const pubKey = await getUserPublicKey(activeClient.id, getPublicKey);
        setClientPublicKey(pubKey);
      } catch (err) {
        console.error('Failed to load client public key:', err);
      }
    };

    loadClientKey();
  }, [activeClient?.id, getPublicKey]);

  // Load messages when conversation is set
  useEffect(() => {
    if (!conversation) return;

    const loadMessages = async () => {
      try {
        setMessagesLoading(true);
        const msgs = await getMessages(conversation.id);

        // Decrypt messages
        const keys = getStoredKeys();
        if (keys && msgs && clientPublicKey) {
          const decryptedMsgs = await Promise.all(
            msgs.map(async (msg) => {
              const isOwn = msg.senderId === user?.id;
              try {
                const decryptedContent = decryptMessage(
                  msg.encryptedContent,
                  msg.nonce,
                  isOwn ? keys.publicKey : clientPublicKey,
                  keys.secretKey
                );
                return { ...msg, decryptedContent: decryptedContent || '[Decryption failed]' };
              } catch (e) {
                return { ...msg, decryptedContent: '[Decryption failed]' };
              }
            })
          );
          setMessages(decryptedMsgs);
        } else {
          setMessages(msgs || []);
        }

        // Mark as read
        await markAsRead(conversation.id);
      } catch (err) {
        console.error('Failed to load messages:', err);
      } finally {
        setMessagesLoading(false);
      }
    };

    loadMessages();

    // Poll for new messages every 5 seconds
    const pollInterval = setInterval(loadMessages, 5000);
    return () => clearInterval(pollInterval);
  }, [conversation?.id, clientPublicKey]);

  const handleSendMessage = async (messageText) => {
    if (!conversation || !encryptionReady || !userKeys) {
      throw new Error('Not ready to send messages');
    }

    if (!clientPublicKey) {
      throw new Error('Client encryption key not available');
    }

    // Encrypt the message
    const { encryptedContent, nonce } = encryptMessage(
      messageText,
      clientPublicKey,
      userKeys.secretKey
    );

    // Send encrypted message
    const sentMsg = await sendMessage(conversation.id, encryptedContent, nonce);

    // Add to local messages with decrypted content
    setMessages((prev) => [
      ...prev,
      { ...sentMsg, decryptedContent: messageText, isOwn: true },
    ]);
  };

  if (!activeClient) {
    return (
      <div className="client-messages-panel empty">
        <p>Select a client to view messages</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="client-messages-panel loading">
        <div className="loading-spinner"></div>
        <p>Loading conversation...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="client-messages-panel error">
        <p>{error}</p>
        <button className="btn btn-secondary" onClick={() => window.location.reload()}>
          Retry
        </button>
      </div>
    );
  }

  return (
    <div className="client-messages-panel">
      <div className="messages-panel-header">
        <div className="conversation-info">
          <h3>Messages with {activeClient.name}</h3>
          <span className="encryption-badge">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
              <path d="M7 11V7a5 5 0 0 1 10 0v4" />
            </svg>
            End-to-end encrypted
          </span>
        </div>
        {!encryptionReady && (
          <div className="encryption-status">
            <span className="status-indicator initializing"></span>
            Setting up encryption...
          </div>
        )}
      </div>

      <div className="messages-content">
        <MessageThread
          messages={messages}
          currentUserId={user?.id}
          loading={messagesLoading}
        />
      </div>

      <div className="messages-input-wrapper">
        <MessageInput
          onSend={handleSendMessage}
          disabled={!encryptionReady || !clientPublicKey}
          placeholder={
            !encryptionReady
              ? 'Setting up encryption...'
              : !clientPublicKey
              ? 'Waiting for client encryption key...'
              : `Message ${activeClient.name}...`
          }
        />
      </div>
    </div>
  );
}
