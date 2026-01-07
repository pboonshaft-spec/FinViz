import React, { useState, useEffect } from 'react';
import { useApi } from '../../hooks/useApi';
import { useAuth } from '../../contexts/AuthContext';

export default function UserManagement() {
  const { user: currentUser } = useAuth();
  const [advisors, setAdvisors] = useState([]);
  const [allUsers, setAllUsers] = useState([]);
  const [activeTab, setActiveTab] = useState('advisors');
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [selectedAdvisor, setSelectedAdvisor] = useState(null);
  const [selectedClient, setSelectedClient] = useState(null);
  const [assignAdvisorId, setAssignAdvisorId] = useState('');
  const [formData, setFormData] = useState({ name: '', email: '', password: '' });
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState(null);
  const [successMessage, setSuccessMessage] = useState(null);

  const {
    getAdvisors,
    getAllUsers,
    createAdvisor,
    updateAdvisor,
    deleteAdvisor,
    claimClient,
    assignClient
  } = useApi();

  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    setIsLoading(true);
    setError(null);
    try {
      const [advisorsData, usersData] = await Promise.all([
        getAdvisors(),
        getAllUsers()
      ]);
      setAdvisors(advisorsData || []);
      setAllUsers(usersData || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateAdvisor = async (e) => {
    e.preventDefault();
    setError(null);
    try {
      const result = await createAdvisor(formData);
      setSuccessMessage(
        result.temporaryPassword
          ? `Advisor created! Temporary password: ${result.temporaryPassword}`
          : 'Advisor created successfully!'
      );
      setShowCreateModal(false);
      setFormData({ name: '', email: '', password: '' });
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleUpdateAdvisor = async (e) => {
    e.preventDefault();
    setError(null);
    try {
      const updates = {};
      if (formData.name && formData.name !== selectedAdvisor.name) updates.name = formData.name;
      if (formData.email && formData.email !== selectedAdvisor.email) updates.email = formData.email;
      if (formData.password) updates.password = formData.password;

      if (Object.keys(updates).length === 0) {
        setShowEditModal(false);
        return;
      }

      await updateAdvisor(selectedAdvisor.id, updates);
      setSuccessMessage('Advisor updated successfully!');
      setShowEditModal(false);
      setSelectedAdvisor(null);
      setFormData({ name: '', email: '', password: '' });
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleDeleteAdvisor = async (advisor) => {
    if (!window.confirm(`Are you sure you want to delete ${advisor.name}? This action cannot be undone.`)) {
      return;
    }

    setError(null);
    try {
      await deleteAdvisor(advisor.id);
      setSuccessMessage('Advisor deleted successfully!');
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleClaimClient = async (client) => {
    setError(null);
    try {
      await claimClient(client.id);
      setSuccessMessage(`Successfully claimed ${client.name} as your client!`);
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const handleAssignClient = async (e) => {
    e.preventDefault();
    if (!assignAdvisorId || !selectedClient) return;

    setError(null);
    try {
      await assignClient(selectedClient.id, parseInt(assignAdvisorId));
      const advisor = advisors.find(a => a.id === parseInt(assignAdvisorId));
      setSuccessMessage(`Successfully assigned ${selectedClient.name} to ${advisor?.name || 'advisor'}!`);
      setShowAssignModal(false);
      setSelectedClient(null);
      setAssignAdvisorId('');
      loadData();
    } catch (err) {
      setError(err.message);
    }
  };

  const openEditModal = (advisor) => {
    setSelectedAdvisor(advisor);
    setFormData({ name: advisor.name, email: advisor.email, password: '' });
    setShowEditModal(true);
  };

  const openAssignModal = (client) => {
    setSelectedClient(client);
    setAssignAdvisorId('');
    setShowAssignModal(true);
  };

  const formatDate = (dateStr) => {
    return new Date(dateStr).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  };

  const isAlreadyMyClient = (user) => {
    return user.advisors?.some(a => a.id === currentUser?.id);
  };

  const clients = allUsers.filter(u => u.role === 'client');

  if (isLoading) {
    return (
      <div className="admin-panel">
        <div className="loading-spinner"></div>
        <p>Loading users...</p>
      </div>
    );
  }

  return (
    <div className="admin-panel">
      <div className="admin-header">
        <h2>User Management</h2>
        <p>Manage advisors, claim clients, and assign clients to advisors</p>
      </div>

      {error && (
        <div className="alert alert-error">
          {error}
          <button onClick={() => setError(null)}>Dismiss</button>
        </div>
      )}

      {successMessage && (
        <div className="alert alert-success">
          {successMessage}
          <button onClick={() => setSuccessMessage(null)}>Dismiss</button>
        </div>
      )}

      <div className="admin-tabs">
        <button
          className={`admin-tab ${activeTab === 'advisors' ? 'active' : ''}`}
          onClick={() => setActiveTab('advisors')}
        >
          Advisors ({advisors.length})
        </button>
        <button
          className={`admin-tab ${activeTab === 'clients' ? 'active' : ''}`}
          onClick={() => setActiveTab('clients')}
        >
          Clients ({clients.length})
        </button>
        <button
          className={`admin-tab ${activeTab === 'all-users' ? 'active' : ''}`}
          onClick={() => setActiveTab('all-users')}
        >
          All Users ({allUsers.length})
        </button>
      </div>

      {activeTab === 'advisors' && (
        <div className="admin-content">
          <div className="admin-toolbar">
            <button
              className="btn btn-primary"
              onClick={() => {
                setFormData({ name: '', email: '', password: '' });
                setShowCreateModal(true);
              }}
            >
              + Add Advisor
            </button>
          </div>

          <div className="admin-table-container">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Clients</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {advisors.length === 0 ? (
                  <tr>
                    <td colSpan="5" className="empty-state">
                      No advisors found. Create the first advisor to get started.
                    </td>
                  </tr>
                ) : (
                  advisors.map((advisor) => (
                    <tr key={advisor.id}>
                      <td>
                        {advisor.name}
                        {advisor.id === currentUser?.id && (
                          <span className="you-badge">You</span>
                        )}
                      </td>
                      <td>{advisor.email}</td>
                      <td>{advisor.clientCount}</td>
                      <td>{formatDate(advisor.createdAt)}</td>
                      <td>
                        <div className="action-buttons">
                          <button
                            className="btn btn-sm btn-secondary"
                            onClick={() => openEditModal(advisor)}
                          >
                            Edit
                          </button>
                          {advisor.id !== currentUser?.id && (
                            <button
                              className="btn btn-sm btn-danger"
                              onClick={() => handleDeleteAdvisor(advisor)}
                            >
                              Delete
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'clients' && (
        <div className="admin-content">
          <div className="admin-table-container">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Assigned To</th>
                  <th>Created</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {clients.length === 0 ? (
                  <tr>
                    <td colSpan="5" className="empty-state">
                      No clients found.
                    </td>
                  </tr>
                ) : (
                  clients.map((client) => (
                    <tr key={client.id}>
                      <td>{client.name}</td>
                      <td>{client.email}</td>
                      <td>
                        {client.advisors && client.advisors.length > 0 ? (
                          <div className="advisor-tags">
                            {client.advisors.map((advisor) => (
                              <span
                                key={advisor.id}
                                className={`advisor-tag ${advisor.id === currentUser?.id ? 'is-you' : ''}`}
                              >
                                {advisor.name}
                                {advisor.id === currentUser?.id && ' (You)'}
                              </span>
                            ))}
                          </div>
                        ) : (
                          <span className="unassigned">Unassigned</span>
                        )}
                      </td>
                      <td>{formatDate(client.createdAt)}</td>
                      <td>
                        <div className="action-buttons">
                          {!isAlreadyMyClient(client) && (
                            <button
                              className="btn btn-sm btn-primary"
                              onClick={() => handleClaimClient(client)}
                            >
                              Claim
                            </button>
                          )}
                          <button
                            className="btn btn-sm btn-secondary"
                            onClick={() => openAssignModal(client)}
                          >
                            Assign
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {activeTab === 'all-users' && (
        <div className="admin-content">
          <div className="admin-table-container">
            <table className="admin-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Email</th>
                  <th>Role</th>
                  <th>Assigned To</th>
                  <th>Created</th>
                </tr>
              </thead>
              <tbody>
                {allUsers.length === 0 ? (
                  <tr>
                    <td colSpan="5" className="empty-state">
                      No users found.
                    </td>
                  </tr>
                ) : (
                  allUsers.map((user) => (
                    <tr key={user.id}>
                      <td>
                        {user.name}
                        {user.id === currentUser?.id && (
                          <span className="you-badge">You</span>
                        )}
                      </td>
                      <td>{user.email}</td>
                      <td>
                        <span className={`role-badge role-${user.role}`}>
                          {user.role}
                        </span>
                      </td>
                      <td>
                        {user.role === 'client' && user.advisors && user.advisors.length > 0 ? (
                          <div className="advisor-tags">
                            {user.advisors.map((advisor) => (
                              <span
                                key={advisor.id}
                                className={`advisor-tag ${advisor.id === currentUser?.id ? 'is-you' : ''}`}
                              >
                                {advisor.name}
                              </span>
                            ))}
                          </div>
                        ) : user.role === 'client' ? (
                          <span className="unassigned">Unassigned</span>
                        ) : (
                          <span className="not-applicable">-</span>
                        )}
                      </td>
                      <td>{formatDate(user.createdAt)}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Create Advisor Modal */}
      {showCreateModal && (
        <div className="modal-overlay" onClick={() => setShowCreateModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Create New Advisor</h3>
              <button className="modal-close" onClick={() => setShowCreateModal(false)}>
                &times;
              </button>
            </div>
            <form onSubmit={handleCreateAdvisor}>
              <div className="modal-body">
                <div className="form-group">
                  <label htmlFor="name">Name</label>
                  <input
                    type="text"
                    id="name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="email">Email</label>
                  <input
                    type="email"
                    id="email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="password">Password (optional)</label>
                  <input
                    type="password"
                    id="password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    placeholder="Leave blank to generate"
                  />
                  <small>If left blank, a temporary password will be generated</small>
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowCreateModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Create Advisor
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Advisor Modal */}
      {showEditModal && selectedAdvisor && (
        <div className="modal-overlay" onClick={() => setShowEditModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Edit Advisor</h3>
              <button className="modal-close" onClick={() => setShowEditModal(false)}>
                &times;
              </button>
            </div>
            <form onSubmit={handleUpdateAdvisor}>
              <div className="modal-body">
                <div className="form-group">
                  <label htmlFor="edit-name">Name</label>
                  <input
                    type="text"
                    id="edit-name"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="edit-email">Email</label>
                  <input
                    type="email"
                    id="edit-email"
                    value={formData.email}
                    onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                    required
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="edit-password">New Password (optional)</label>
                  <input
                    type="password"
                    id="edit-password"
                    value={formData.password}
                    onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                    placeholder="Leave blank to keep current"
                  />
                </div>
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowEditModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary">
                  Save Changes
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Assign Client Modal */}
      {showAssignModal && selectedClient && (
        <div className="modal-overlay" onClick={() => setShowAssignModal(false)}>
          <div className="modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>Assign Client</h3>
              <button className="modal-close" onClick={() => setShowAssignModal(false)}>
                &times;
              </button>
            </div>
            <form onSubmit={handleAssignClient}>
              <div className="modal-body">
                <p className="assign-client-info">
                  Assign <strong>{selectedClient.name}</strong> to an advisor:
                </p>
                <div className="form-group">
                  <label htmlFor="assign-advisor">Select Advisor</label>
                  <select
                    id="assign-advisor"
                    value={assignAdvisorId}
                    onChange={(e) => setAssignAdvisorId(e.target.value)}
                    required
                  >
                    <option value="">Choose an advisor...</option>
                    {advisors.map((advisor) => {
                      const alreadyAssigned = selectedClient.advisors?.some(a => a.id === advisor.id);
                      return (
                        <option
                          key={advisor.id}
                          value={advisor.id}
                          disabled={alreadyAssigned}
                        >
                          {advisor.name} {advisor.id === currentUser?.id ? '(You)' : ''}
                          {alreadyAssigned ? ' - Already assigned' : ''}
                        </option>
                      );
                    })}
                  </select>
                </div>
                {selectedClient.advisors && selectedClient.advisors.length > 0 && (
                  <div className="current-advisors">
                    <small>Currently assigned to: {selectedClient.advisors.map(a => a.name).join(', ')}</small>
                  </div>
                )}
              </div>
              <div className="modal-footer">
                <button type="button" className="btn btn-secondary" onClick={() => setShowAssignModal(false)}>
                  Cancel
                </button>
                <button type="submit" className="btn btn-primary" disabled={!assignAdvisorId}>
                  Assign Client
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
