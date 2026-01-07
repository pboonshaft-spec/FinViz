import React, { useState, useEffect, useCallback } from 'react';
import { useApi } from '../../hooks/useApi';
import { useAuth } from '../../contexts/AuthContext';
import ConversationList from '../messaging/ConversationList';
import MessageThread from '../messaging/MessageThread';
import MessageInput from '../messaging/MessageInput';
import {
  initializeEncryption,
  encryptMessage,
  decryptMessage,
  getStoredKeys,
  getUserPublicKey,
} from '../../utils/encryption';

function MessagesTab() {
  const { user } = useAuth();
  const {
    getConversations,
    getMessages,
    sendMessage,
    markAsRead,
    registerPublicKey,
    getPublicKey,
    startConversation,
  } = useApi();

  const [conversations, setConversations] = useState([]);
  const [selectedConversation, setSelectedConversation] = useState(null);
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(true);
  const [messagesLoading, setMessagesLoading] = useState(false);
  const [encryptionReady, setEncryptionReady] = useState(false);
  const [userKeys, setUserKeys] = useState(null);
  const [otherUserPublicKey, setOtherUserPublicKey] = useState(null);

  // Initialize encryption on mount
  useEffect(() => {
    const init = async () => {
      try {
        const keys = await initializeEncryption(registerPublicKey);
        setUserKeys(keys);
        setEncryptionReady(true);
      } catch (error) {
        console.error('Failed to initialize encryption:', error);
      }
    };
    init();
  }, [registerPublicKey]);

  // Load conversations
  useEffect(() => {
    loadConversations();
  }, []);

  const loadConversations = async () => {
    try {
      setLoading(true);
      const convs = await getConversations();
      setConversations(convs || []);
    } catch (error) {
      console.error('Failed to load conversations:', error);
    } finally {
      setLoading(false);
    }
  };

  // Load messages when conversation changes
  useEffect(() => {
    if (selectedConversation) {
      loadMessages(selectedConversation.id);
      loadOtherUserKey(selectedConversation);
    }
  }, [selectedConversation?.id]);

  const loadOtherUserKey = async (conv) => {
    const otherUserId = user?.role === 'advisor' ? conv.clientId : conv.advisorId;
    try {
      const pubKey = await getUserPublicKey(otherUserId, getPublicKey);
      setOtherUserPublicKey(pubKey);
    } catch (error) {
      console.error('Failed to get other user public key:', error);
    }
  };

  const loadMessages = async (conversationId) => {
    try {
      setMessagesLoading(true);
      const msgs = await getMessages(conversationId);

      // Decrypt messages
      const keys = getStoredKeys();
      if (keys && msgs) {
        const decryptedMsgs = await Promise.all(
          msgs.map(async (msg) => {
            // For own messages, we need the recipient's public key to decrypt
            // For received messages, we need the sender's public key
            const senderId = msg.senderId;
            const isOwn = senderId === user?.id;

            let senderPubKey = otherUserPublicKey;
            if (!senderPubKey) {
              const otherUserId = user?.role === 'advisor'
                ? selectedConversation.clientId
                : selectedConversation.advisorId;
              senderPubKey = await getUserPublicKey(otherUserId, getPublicKey);
            }

            if (senderPubKey) {
              try {
                // For own messages, decrypt with our key and recipient's public key
                // For received messages, decrypt with sender's public key and our secret key
                const decryptedContent = decryptMessage(
                  msg.encryptedContent,
                  msg.nonce,
                  isOwn ? keys.publicKey : senderPubKey,
                  keys.secretKey
                );
                return { ...msg, decryptedContent: decryptedContent || '[Decryption failed]' };
              } catch (e) {
                return { ...msg, decryptedContent: '[Decryption failed]' };
              }
            }
            return { ...msg, decryptedContent: '[Key not available]' };
          })
        );
        setMessages(decryptedMsgs);
      } else {
        setMessages(msgs || []);
      }

      // Mark as read
      await markAsRead(conversationId);

      // Update unread count in conversation list
      setConversations((prev) =>
        prev.map((c) =>
          c.id === conversationId
            ? {
                ...c,
                unreadCountAdvisor: user?.role === 'advisor' ? 0 : c.unreadCountAdvisor,
                unreadCountClient: user?.role !== 'advisor' ? 0 : c.unreadCountClient,
              }
            : c
        )
      );
    } catch (error) {
      console.error('Failed to load messages:', error);
    } finally {
      setMessagesLoading(false);
    }
  };

  const handleSelectConversation = (conv) => {
    setSelectedConversation(conv);
    setMessages([]);
    setOtherUserPublicKey(null);
  };

  const handleSendMessage = async (messageText) => {
    if (!selectedConversation || !encryptionReady || !userKeys) {
      throw new Error('Not ready to send messages');
    }

    // Get recipient's public key
    let recipientPubKey = otherUserPublicKey;
    if (!recipientPubKey) {
      const otherUserId = user?.role === 'advisor'
        ? selectedConversation.clientId
        : selectedConversation.advisorId;
      recipientPubKey = await getUserPublicKey(otherUserId, getPublicKey);
      setOtherUserPublicKey(recipientPubKey);
    }

    if (!recipientPubKey) {
      throw new Error('Cannot encrypt: recipient public key not available');
    }

    // Encrypt the message
    const { encryptedContent, nonce } = encryptMessage(
      messageText,
      recipientPubKey,
      userKeys.secretKey
    );

    // Send encrypted message
    const sentMsg = await sendMessage(selectedConversation.id, encryptedContent, nonce);

    // Add to local messages with decrypted content
    setMessages((prev) => [
      ...prev,
      { ...sentMsg, decryptedContent: messageText, isOwn: true },
    ]);

    // Update conversation list
    setConversations((prev) =>
      prev.map((c) =>
        c.id === selectedConversation.id
          ? { ...c, lastMessageAt: new Date().toISOString() }
          : c
      )
    );
  };

  const isAdvisor = user?.role === 'advisor';

  return (
    <div className="tab-content messages-tab">
      <div className="tab-header">
        <div className="tab-header-text">
          <h2>Messages</h2>
          <p>Secure, end-to-end encrypted messaging with your {isAdvisor ? 'clients' : 'advisor'}</p>
        </div>
      </div>

      {!encryptionReady && (
        <div className="encryption-status">
          <span className="status-indicator initializing"></span>
          Initializing secure connection...
        </div>
      )}

      <div className="messages-container">
        <div className="conversations-panel">
          <div className="panel-header">
            <h3>Conversations</h3>
          </div>
          {loading ? (
            <div className="loading-state">Loading...</div>
          ) : (
            <ConversationList
              conversations={conversations}
              selectedId={selectedConversation?.id}
              onSelect={handleSelectConversation}
              isAdvisor={isAdvisor}
            />
          )}
        </div>

        <div className="messages-panel">
          {selectedConversation ? (
            <>
              <div className="panel-header">
                <div className="conversation-title">
                  <h3>
                    {isAdvisor
                      ? selectedConversation.clientName
                      : selectedConversation.advisorName}
                  </h3>
                  <span className="encryption-badge">
                    <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                      <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                      <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                    </svg>
                    Encrypted
                  </span>
                </div>
              </div>
              <MessageThread
                messages={messages}
                currentUserId={user?.id}
                loading={messagesLoading}
              />
              <MessageInput
                onSend={handleSendMessage}
                disabled={!encryptionReady || !otherUserPublicKey}
                placeholder={
                  !encryptionReady
                    ? 'Setting up encryption...'
                    : !otherUserPublicKey
                    ? 'Waiting for recipient key...'
                    : 'Type a secure message...'
                }
              />
            </>
          ) : (
            <div className="no-conversation-selected">
              <div className="empty-icon">
                <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                </svg>
              </div>
              <h3>Select a conversation</h3>
              <p>Choose from your existing conversations or start a new one</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default MessagesTab;
