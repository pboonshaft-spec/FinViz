import React, { useState, useCallback, useEffect } from 'react';
import { usePlaidLink } from 'react-plaid-link';
import { useApi } from '../../hooks/useApi';
import ConfirmModal from '../common/ConfirmModal';

function PlaidLinkButton({ onSuccess, onExit }) {
  const [linkToken, setLinkToken] = useState(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const { createLinkToken, exchangeToken } = useApi();

  // Generate link token on mount
  useEffect(() => {
    const generateToken = async () => {
      setLoading(true);
      setError(null);
      try {
        const response = await createLinkToken();
        setLinkToken(response.linkToken);
      } catch (err) {
        setError(err.message);
      } finally {
        setLoading(false);
      }
    };

    generateToken();
  }, []);

  const handleSuccess = useCallback(async (publicToken, metadata) => {
    setLoading(true);
    setError(null);
    try {
      const response = await exchangeToken(publicToken);
      if (onSuccess) {
        onSuccess(response);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }, [exchangeToken, onSuccess]);

  const handleExit = useCallback((err, metadata) => {
    if (err) {
      setError(err.message || 'Link was closed');
    }
    if (onExit) {
      onExit(err, metadata);
    }
  }, [onExit]);

  const config = {
    token: linkToken,
    onSuccess: handleSuccess,
    onExit: handleExit,
  };

  const { open, ready } = usePlaidLink(config);

  if (error) {
    return (
      <div className="plaid-error">
        <span>Error: {error}</span>
        <button className="btn btn-sm btn-secondary" onClick={() => setError(null)}>
          Dismiss
        </button>
      </div>
    );
  }

  return (
    <button
      className="btn btn-primary"
      onClick={() => open()}
      disabled={!ready || loading}
    >
      {loading ? (
        <>
          <span className="spinner-sm"></span>
          Connecting...
        </>
      ) : (
        <>
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M12 2v4m0 12v4M4.93 4.93l2.83 2.83m8.48 8.48l2.83 2.83M2 12h4m12 0h4M4.93 19.07l2.83-2.83m8.48-8.48l2.83-2.83"/>
          </svg>
          Connect Bank Account
        </>
      )}
    </button>
  );
}

function PlaidLink({ onAccountsLinked }) {
  const [plaidStatus, setPlaidStatus] = useState({ configured: false, checked: false });
  const [linkedItems, setLinkedItems] = useState([]);
  const [syncing, setSyncing] = useState(false);
  const [deleteModal, setDeleteModal] = useState({ isOpen: false, itemId: null });

  const { getPlaidStatus, getPlaidItems, syncPlaidAccounts, deletePlaidItem } = useApi();

  // Check if Plaid is configured
  useEffect(() => {
    const checkStatus = async () => {
      try {
        const status = await getPlaidStatus();
        setPlaidStatus({ configured: status.configured, checked: true });
      } catch (err) {
        setPlaidStatus({ configured: false, checked: true });
      }
    };
    checkStatus();
  }, []);

  // Load linked items
  useEffect(() => {
    if (plaidStatus.configured) {
      loadLinkedItems();
    }
  }, [plaidStatus.configured]);

  const loadLinkedItems = async () => {
    try {
      const items = await getPlaidItems();
      setLinkedItems(items);
    } catch (err) {
      console.error('Failed to load Plaid items:', err);
    }
  };

  const handleSuccess = (response) => {
    loadLinkedItems();
    if (onAccountsLinked) {
      onAccountsLinked(response);
    }
  };

  const handleSync = async () => {
    setSyncing(true);
    try {
      const result = await syncPlaidAccounts();
      if (onAccountsLinked) {
        onAccountsLinked(result);
      }
    } catch (err) {
      console.error('Sync failed:', err);
    } finally {
      setSyncing(false);
    }
  };

  const handleDeleteClick = (itemId) => {
    setDeleteModal({ isOpen: true, itemId });
  };

  const handleDeleteConfirm = async (choice) => {
    const { itemId } = deleteModal;
    setDeleteModal({ isOpen: false, itemId: null });

    if (choice === null || choice === 'cancel') return;

    const deleteData = choice === 'yes';

    try {
      await deletePlaidItem(itemId, deleteData);
      loadLinkedItems();
      if (onAccountsLinked) {
        onAccountsLinked(); // Refresh parent data
      }
    } catch (err) {
      console.error('Failed to delete item:', err);
    }
  };

  if (!plaidStatus.checked) {
    return <div className="plaid-loading">Checking Plaid status...</div>;
  }

  if (!plaidStatus.configured) {
    return (
      <div className="plaid-not-configured">
        <p>Bank account linking is not configured.</p>
        <p className="hint">Add PLAID_CLIENT_ID and PLAID_SECRET to your .env file.</p>
      </div>
    );
  }

  return (
    <div className="plaid-container">
      <div className="plaid-header">
        <h4>Linked Accounts</h4>
        <div className="plaid-actions">
          {linkedItems.length > 0 && (
            <button
              className="btn btn-secondary btn-sm"
              onClick={handleSync}
              disabled={syncing}
            >
              {syncing ? 'Syncing...' : 'Sync Balances'}
            </button>
          )}
          <PlaidLinkButton onSuccess={handleSuccess} />
        </div>
      </div>

      {linkedItems.length === 0 ? (
        <div className="plaid-empty">
          <p>No bank accounts linked yet.</p>
          <p className="hint">Click "Connect Bank Account" to automatically import your accounts.</p>
        </div>
      ) : (
        <div className="plaid-items">
          {linkedItems.map(item => (
            <div key={item.id} className="plaid-item">
              <div className="plaid-item-info">
                <span className="institution-name">{item.institutionName || 'Unknown Bank'}</span>
                <span className="item-status">{item.status}</span>
              </div>
              <button
                className="btn-icon btn-danger"
                onClick={() => handleDeleteClick(item.id)}
                title="Remove connection"
              >
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                  <polyline points="3 6 5 6 21 6" />
                  <path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
                </svg>
              </button>
            </div>
          ))}
        </div>
      )}

      <ConfirmModal
        isOpen={deleteModal.isOpen}
        title="Remove Bank Connection"
        message="Do you also want to delete the assets, debts, and transactions synced from this bank?"
        options={[
          { label: 'Yes, delete all data', value: 'yes', variant: 'danger' },
          { label: 'No, keep data', value: 'no', variant: 'primary' },
          { label: 'Cancel', value: 'cancel', variant: 'secondary' },
        ]}
        onSelect={handleDeleteConfirm}
      />
    </div>
  );
}

export default PlaidLink;
