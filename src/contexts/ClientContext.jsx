import React, { createContext, useContext, useState, useCallback, useEffect, useRef } from 'react';
import { useAuth } from './AuthContext';
import { setClientContextRef } from '../hooks/useApi';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

const ClientContext = createContext(null);

export function ClientProvider({ children }) {
  const { token, user } = useAuth();
  const [activeClient, setActiveClient] = useState(null);
  const [clients, setClients] = useState([]);
  const [loadingClients, setLoadingClients] = useState(false);

  // Share context state with useApi hook
  const contextRef = useRef({ activeClient: null });

  // Initialize the ref on mount
  useEffect(() => {
    setClientContextRef(contextRef.current);
  }, []);

  // Fetch all clients for the advisor
  const fetchClients = useCallback(async () => {
    if (!token || user?.role !== 'advisor') return;

    setLoadingClients(true);
    try {
      const response = await fetch(`${API_BASE_URL}/api/advisor/clients`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setClients(data || []);
      }
    } catch (err) {
      console.error('Failed to fetch clients:', err);
    } finally {
      setLoadingClients(false);
    }
  }, [token, user?.role]);

  // Switch to a specific client's context
  const switchClient = useCallback((client) => {
    // Update ref synchronously BEFORE state update so useApi sees it immediately
    contextRef.current.activeClient = client;
    setActiveClient(client);
  }, []);

  // Clear client context (go back to advisor's own view)
  const clearClientContext = useCallback(() => {
    // Update ref synchronously BEFORE state update
    contextRef.current.activeClient = null;
    setActiveClient(null);
  }, []);

  // Get the API path prefix for the current context
  const getApiPath = useCallback((basePath) => {
    if (activeClient) {
      // Replace /api/ with /api/advisor/clients/{clientId}/
      return basePath.replace('/api/', `/api/advisor/clients/${activeClient.id}/`);
    }
    return basePath;
  }, [activeClient]);

  // Check if we're in client context
  const isInClientContext = !!activeClient;

  const value = {
    activeClient,
    clients,
    loadingClients,
    isInClientContext,
    fetchClients,
    switchClient,
    clearClientContext,
    getApiPath,
  };

  return (
    <ClientContext.Provider value={value}>
      {children}
    </ClientContext.Provider>
  );
}

export function useClientContext() {
  const context = useContext(ClientContext);
  if (!context) {
    throw new Error('useClientContext must be used within a ClientProvider');
  }
  return context;
}
