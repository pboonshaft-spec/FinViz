import { useState, useCallback, useContext } from 'react';
import { useAuth } from '../contexts/AuthContext';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

// We need to import ClientContext directly to avoid circular deps
// This is a late import - the context may not exist yet
let clientContextRef = null;
export function setClientContextRef(ctx) {
  clientContextRef = ctx;
}

export function useApi() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { token, logout, isAdvisor } = useAuth();

  // Try to get active client from context ref
  const getActiveClientId = () => {
    if (clientContextRef && isAdvisor) {
      try {
        return clientContextRef.activeClient?.id;
      } catch {
        return null;
      }
    }
    return null;
  };

  // Transform endpoint for client context
  const getContextualEndpoint = (endpoint) => {
    const clientId = getActiveClientId();
    if (!clientId) return endpoint;

    // Don't transform these endpoints - they are global or user-specific
    const excludeFromTransform = [
      '/api/advisor/',      // Already advisor-specific
      '/api/auth/',         // Auth endpoints
      '/api/asset-types',   // Global - same for all users
      '/api/plaid/',        // Plaid is user-specific, not client-contextual
      '/api/import/',       // Import is user-specific
      '/api/health',        // Health check
      '/api/chat/status',   // Chat status
      '/api/invitation',    // Invitations
      '/api/invitations',   // Invitations
      '/api/transactions/sync', // Sync is user-specific (uses their Plaid)
      '/api/messages/',     // Messaging is user-specific, not client-contextual
    ];

    if (excludeFromTransform.some(prefix => endpoint.startsWith(prefix))) {
      return endpoint;
    }

    if (endpoint.startsWith('/api/')) {
      // Transform /api/assets to /api/advisor/clients/{clientId}/assets
      return endpoint.replace('/api/', `/api/advisor/clients/${clientId}/`);
    }
    return endpoint;
  };

  const request = useCallback(async (endpoint, options = {}) => {
    // Apply client context transformation
    const contextualEndpoint = getContextualEndpoint(endpoint);
    setLoading(true);
    setError(null);

    try {
      const headers = {
        'Content-Type': 'application/json',
        ...options.headers,
      };

      // Add auth header if token exists
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`${API_BASE_URL}${contextualEndpoint}`, {
        ...options,
        headers,
      });

      if (response.status === 401) {
        // Token is invalid or expired
        logout();
        throw new Error('Session expired. Please log in again.');
      }

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error ${response.status}`);
      }

      const data = await response.json();
      return data;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [token, logout]);

  // Assets API
  const getAssets = useCallback(() => request('/api/assets'), [request]);

  const createAsset = useCallback((asset) => request('/api/assets', {
    method: 'POST',
    body: JSON.stringify(asset),
  }), [request]);

  const updateAsset = useCallback((id, updates) => request(`/api/assets/${id}`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  }), [request]);

  const deleteAsset = useCallback((id) => request(`/api/assets/${id}`, {
    method: 'DELETE',
  }), [request]);

  // Asset Types API
  const getAssetTypes = useCallback(() => request('/api/asset-types'), [request]);

  // Debts API
  const getDebts = useCallback(() => request('/api/debts'), [request]);

  const createDebt = useCallback((debt) => request('/api/debts', {
    method: 'POST',
    body: JSON.stringify(debt),
  }), [request]);

  const updateDebt = useCallback((id, updates) => request(`/api/debts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  }), [request]);

  const deleteDebt = useCallback((id) => request(`/api/debts/${id}`, {
    method: 'DELETE',
  }), [request]);

  // Monte Carlo API
  const runMonteCarlo = useCallback((params = {}) => {
    return request('/api/monte-carlo', {
      method: 'POST',
      body: JSON.stringify({ params }),
    });
  }, [request]);

  // Report Generation API
  const generateReport = useCallback(async (options = {}) => {
    const contextualEndpoint = getContextualEndpoint('/api/reports/generate');

    const headers = {
      'Content-Type': 'application/json',
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const response = await fetch(`${API_BASE_URL}${contextualEndpoint}`, {
      method: 'POST',
      headers,
      body: JSON.stringify(options),
    });

    if (response.status === 401) {
      logout();
      throw new Error('Session expired. Please log in again.');
    }

    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new Error(errorData.error || `HTTP error ${response.status}`);
    }

    // Return blob for PDF download
    const blob = await response.blob();
    const contentDisposition = response.headers.get('Content-Disposition');
    let filename = 'financial_plan.pdf';
    if (contentDisposition) {
      const match = contentDisposition.match(/filename="(.+)"/);
      if (match) filename = match[1];
    }

    return { blob, filename };
  }, [token, logout]);

  // CSV Import API
  const importCSV = useCallback(async (file, type) => {
    setLoading(true);
    setError(null);

    try {
      const formData = new FormData();
      formData.append('file', file);
      formData.append('type', type);

      const headers = {};
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }

      const response = await fetch(`${API_BASE_URL}/api/import/csv`, {
        method: 'POST',
        headers,
        body: formData,
      });

      if (response.status === 401) {
        logout();
        throw new Error('Session expired. Please log in again.');
      }

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP error ${response.status}`);
      }

      return await response.json();
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [token, logout]);

  // Plaid API
  const getPlaidStatus = useCallback(() => {
    // This is a public endpoint, doesn't need auth
    return fetch(`${API_BASE_URL}/api/plaid/status`)
      .then(res => res.json());
  }, []);

  const createLinkToken = useCallback(() => request('/api/plaid/link-token', {
    method: 'POST',
  }), [request]);

  const exchangeToken = useCallback((publicToken) => request('/api/plaid/exchange-token', {
    method: 'POST',
    body: JSON.stringify({ publicToken }),
  }), [request]);

  const getPlaidItems = useCallback(() => request('/api/plaid/items'), [request]);

  const deletePlaidItem = useCallback((id, deleteData = false) => request(`/api/plaid/items/${id}?delete_data=${deleteData}`, {
    method: 'DELETE',
  }), [request]);

  const getPlaidAccounts = useCallback(() => request('/api/plaid/accounts'), [request]);

  const syncPlaidAccounts = useCallback(() => request('/api/plaid/sync', {
    method: 'POST',
  }), [request]);

  // Transactions API
  const getTransactions = useCallback((startDate, endDate, category) => {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    if (category) params.append('category', category);
    return request(`/api/transactions?${params.toString()}`);
  }, [request]);

  const getTransactionSummary = useCallback((startDate, endDate) => {
    const params = new URLSearchParams();
    if (startDate) params.append('start_date', startDate);
    if (endDate) params.append('end_date', endDate);
    return request(`/api/transactions/summary?${params.toString()}`);
  }, [request]);

  const getCategories = useCallback(() => request('/api/transactions/categories'), [request]);

  const syncTransactions = useCallback((startDate, endDate) => request('/api/transactions/sync', {
    method: 'POST',
    body: JSON.stringify({ startDate, endDate }),
  }), [request]);

  // Admin API (Advisor only)
  const getAdvisors = useCallback(() => request('/api/advisor/admin/advisors'), [request]);

  const getAdvisor = useCallback((id) => request(`/api/advisor/admin/advisors/${id}`), [request]);

  const createAdvisor = useCallback((advisor) => request('/api/advisor/admin/advisors', {
    method: 'POST',
    body: JSON.stringify(advisor),
  }), [request]);

  const updateAdvisor = useCallback((id, updates) => request(`/api/advisor/admin/advisors/${id}`, {
    method: 'PUT',
    body: JSON.stringify(updates),
  }), [request]);

  const deleteAdvisor = useCallback((id) => request(`/api/advisor/admin/advisors/${id}`, {
    method: 'DELETE',
  }), [request]);

  const getAllUsers = useCallback(() => request('/api/advisor/admin/users'), [request]);

  const claimClient = useCallback((clientId) => request('/api/advisor/admin/claim-client', {
    method: 'POST',
    body: JSON.stringify({ clientId }),
  }), [request]);

  const assignClient = useCallback((clientId, advisorId) => request('/api/advisor/admin/assign-client', {
    method: 'POST',
    body: JSON.stringify({ clientId, advisorId }),
  }), [request]);

  return {
    loading,
    error,
    getAssets,
    createAsset,
    updateAsset,
    deleteAsset,
    getAssetTypes,
    getDebts,
    createDebt,
    updateDebt,
    deleteDebt,
    runMonteCarlo,
    generateReport,
    importCSV,
    // Plaid
    getPlaidStatus,
    createLinkToken,
    exchangeToken,
    getPlaidItems,
    deletePlaidItem,
    getPlaidAccounts,
    syncPlaidAccounts,
    // Transactions
    getTransactions,
    getTransactionSummary,
    getCategories,
    syncTransactions,
    // Admin (Advisor only)
    getAdvisors,
    getAdvisor,
    createAdvisor,
    updateAdvisor,
    deleteAdvisor,
    getAllUsers,
    claimClient,
    assignClient,
    // Messaging
    getConversations: useCallback(() => request('/api/messages/conversations'), [request]),
    startConversation: useCallback((clientId) => request('/api/messages/conversations', {
      method: 'POST',
      body: JSON.stringify({ clientId }),
    }), [request]),
    getConversation: useCallback((id) => request(`/api/messages/conversations/${id}`), [request]),
    getMessages: useCallback((conversationId, before) => {
      const params = new URLSearchParams();
      if (before) params.append('before', before);
      const query = params.toString() ? `?${params.toString()}` : '';
      return request(`/api/messages/conversations/${conversationId}/messages${query}`);
    }, [request]),
    sendMessage: useCallback((conversationId, encryptedContent, nonce) => request(`/api/messages/conversations/${conversationId}/messages`, {
      method: 'POST',
      body: JSON.stringify({ encryptedContent, nonce }),
    }), [request]),
    markAsRead: useCallback((conversationId) => request(`/api/messages/conversations/${conversationId}/read`, {
      method: 'POST',
    }), [request]),
    getUnreadCounts: useCallback(() => request('/api/messages/unread'), [request]),
    registerPublicKey: useCallback((data) => request('/api/messages/keys', {
      method: 'POST',
      body: JSON.stringify(data),
    }), [request]),
    getPublicKey: useCallback((userId) => request(`/api/messages/keys/${userId}`), [request]),
    // Document Vault
    getDocuments: useCallback((category) => {
      const params = new URLSearchParams();
      if (category) params.append('category', category);
      return request(`/api/documents?${params.toString()}`);
    }, [request]),
    uploadDocument: useCallback(async (file, metadata) => {
      setLoading(true);
      setError(null);
      try {
        const formData = new FormData();
        formData.append('file', file);
        if (metadata.name) formData.append('name', metadata.name);
        if (metadata.category) formData.append('category', metadata.category);
        if (metadata.description) formData.append('description', metadata.description);
        if (metadata.year) formData.append('year', metadata.year.toString());
        if (metadata.clientId) formData.append('client_id', metadata.clientId.toString());

        const headers = {};
        if (token) headers['Authorization'] = `Bearer ${token}`;

        const response = await fetch(`${API_BASE_URL}/api/documents/upload`, {
          method: 'POST',
          headers,
          body: formData,
        });

        if (response.status === 401) {
          logout();
          throw new Error('Session expired');
        }

        if (!response.ok) {
          const errorData = await response.json().catch(() => ({}));
          throw new Error(errorData.error || `Upload failed`);
        }

        return await response.json();
      } catch (err) {
        setError(err.message);
        throw err;
      } finally {
        setLoading(false);
      }
    }, [token, logout]),
    downloadDocument: useCallback(async (docId, filename) => {
      const headers = {};
      if (token) headers['Authorization'] = `Bearer ${token}`;

      const response = await fetch(`${API_BASE_URL}/api/documents/${docId}/download`, { headers });

      if (response.status === 401) {
        logout();
        throw new Error('Session expired');
      }

      if (!response.ok) {
        throw new Error('Download failed');
      }

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      a.remove();
    }, [token, logout]),
    deleteDocument: useCallback((docId) => request(`/api/documents/${docId}`, {
      method: 'DELETE',
    }), [request]),
    shareDocument: useCallback((docId, shareWithId, permission, expiresIn) => request(`/api/documents/${docId}/share`, {
      method: 'POST',
      body: JSON.stringify({ share_with_id: shareWithId, permission, expires_in: expiresIn }),
    }), [request]),
  };
}
