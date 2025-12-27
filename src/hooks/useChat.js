import { useState, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

export function useChat() {
  const [messages, setMessages] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const [artifacts, setArtifacts] = useState([]);
  const { token, logout } = useAuth();

  // Check if chat is configured
  const checkChatStatus = useCallback(async () => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/chat/status`);
      const data = await response.json();
      return data.configured;
    } catch (err) {
      console.error('Failed to check chat status:', err);
      return false;
    }
  }, []);

  // Send a message to Aurelia
  const sendMessage = useCallback(async (content) => {
    if (!content.trim()) return;

    // Add user message to state
    const userMessage = {
      id: Date.now(),
      role: 'user',
      content: content.trim(),
      timestamp: new Date().toISOString(),
    };

    setMessages(prev => [...prev, userMessage]);
    setLoading(true);
    setError(null);

    try {
      // Build message history for context
      const messageHistory = [...messages, userMessage].map(m => ({
        role: m.role,
        content: m.content,
      }));

      const response = await fetch(`${API_BASE_URL}/api/chat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({
          messages: messageHistory,
        }),
      });

      if (response.status === 401) {
        logout();
        throw new Error('Session expired. Please log in again.');
      }

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error ${response.status}`);
      }

      const data = await response.json();

      // Add assistant message to state
      const assistantMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: data.response,
        timestamp: new Date().toISOString(),
        toolsUsed: data.toolsUsed || [],
        artifacts: data.artifacts || [],
        tokenUsage: data.tokenUsage,
      };

      setMessages(prev => [...prev, assistantMessage]);

      // Track all artifacts
      if (data.artifacts?.length > 0) {
        setArtifacts(prev => [...prev, ...data.artifacts]);
      }

      return assistantMessage;
    } catch (err) {
      setError(err.message);

      // Add error message to chat
      const errorMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: `I apologize, but I encountered an error: ${err.message}. Please try again.`,
        timestamp: new Date().toISOString(),
        isError: true,
      };

      setMessages(prev => [...prev, errorMessage]);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [messages, token, logout]);

  // Clear conversation
  const clearMessages = useCallback(() => {
    setMessages([]);
    setArtifacts([]);
    setError(null);
  }, []);

  // Get initial greeting message
  const getGreeting = useCallback(() => {
    return {
      id: 0,
      role: 'assistant',
      content: `Hello! I'm Aurelia, your financial advisor. I have access to your financial data and can help you with:

- **Portfolio Analysis**: Review your assets and investment allocation
- **Debt Strategy**: Analyze your debts and suggest payoff strategies
- **Net Worth Tracking**: Understand your overall financial picture
- **Tax Planning**: Research tax implications and strategies
- **Financial Research**: Look up current rates, regulations, and market information

What would you like to explore today?`,
      timestamp: new Date().toISOString(),
      isGreeting: true,
    };
  }, []);

  // Initialize with greeting
  const initializeChat = useCallback(() => {
    if (messages.length === 0) {
      setMessages([getGreeting()]);
    }
  }, [messages.length, getGreeting]);

  return {
    messages,
    loading,
    error,
    artifacts,
    sendMessage,
    clearMessages,
    initializeChat,
    checkChatStatus,
  };
}
