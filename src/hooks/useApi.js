import { useState, useCallback } from 'react';
import { useAuth } from '../contexts/AuthContext';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

export function useApi() {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);
  const { token, logout } = useAuth();

  const request = useCallback(async (endpoint, options = {}) => {
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

      const response = await fetch(`${API_BASE_URL}${endpoint}`, {
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
  };
}
