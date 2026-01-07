import React, { useState, useEffect } from 'react';
import { useAuth } from '../../contexts/AuthContext';
import { useClientContext } from '../../contexts/ClientContext';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8085';

export default function SimulationHistory({ onLoadSimulation }) {
  const { token } = useAuth();
  const { getApiPath, isInClientContext } = useClientContext();
  const [simulations, setSimulations] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [selectedSimulation, setSelectedSimulation] = useState(null);
  const [showDetails, setShowDetails] = useState(false);

  useEffect(() => {
    fetchSimulations();
  }, [isInClientContext]);

  const fetchSimulations = async () => {
    setLoading(true);
    setError(null);
    try {
      const apiPath = getApiPath('/api/simulations');
      const response = await fetch(`${API_BASE_URL}${apiPath}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch simulations');
      }

      const data = await response.json();
      setSimulations(data || []);
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  const fetchSimulationDetails = async (id) => {
    try {
      const apiPath = getApiPath(`/api/simulations/${id}`);
      const response = await fetch(`${API_BASE_URL}${apiPath}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        throw new Error('Failed to fetch simulation details');
      }

      const data = await response.json();
      setSelectedSimulation(data);
      setShowDetails(true);
    } catch (err) {
      console.error('Failed to fetch simulation details:', err);
    }
  };

  const deleteSimulation = async (id) => {
    if (!confirm('Are you sure you want to delete this simulation?')) return;

    try {
      const apiPath = getApiPath(`/api/simulations/${id}`);
      const response = await fetch(`${API_BASE_URL}${apiPath}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        setSimulations(simulations.filter(s => s.id !== id));
        if (selectedSimulation?.id === id) {
          setSelectedSimulation(null);
          setShowDetails(false);
        }
      }
    } catch (err) {
      console.error('Failed to delete simulation:', err);
    }
  };

  const toggleFavorite = async (id, currentValue) => {
    try {
      const apiPath = getApiPath(`/api/simulations/${id}`);
      const response = await fetch(`${API_BASE_URL}${apiPath}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ isFavorite: !currentValue }),
      });

      if (response.ok) {
        setSimulations(simulations.map(s =>
          s.id === id ? { ...s, isFavorite: !currentValue } : s
        ));
      }
    } catch (err) {
      console.error('Failed to toggle favorite:', err);
    }
  };

  const formatCurrency = (value) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const formatDate = (dateString) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getSuccessRateColor = (rate) => {
    if (rate >= 85) return 'success';
    if (rate >= 70) return 'warning';
    return 'danger';
  };

  if (loading) {
    return <div className="simulation-history-loading">Loading simulation history...</div>;
  }

  if (error) {
    return <div className="simulation-history-error">{error}</div>;
  }

  return (
    <div className="simulation-history">
      <div className="simulation-history-header">
        <h3>Saved Simulations</h3>
        <button className="btn btn-secondary btn-sm" onClick={fetchSimulations}>
          Refresh
        </button>
      </div>

      {simulations.length === 0 ? (
        <div className="simulation-history-empty">
          <p>No saved simulations yet. Run a Monte Carlo simulation to see your projection history.</p>
        </div>
      ) : (
        <div className="simulation-list">
          {simulations.map((sim) => (
            <div key={sim.id} className="simulation-card">
              <div className="simulation-card-header">
                <div className="simulation-title">
                  <button
                    className={`btn-favorite ${sim.isFavorite ? 'active' : ''}`}
                    onClick={() => toggleFavorite(sim.id, sim.isFavorite)}
                    title={sim.isFavorite ? 'Remove from favorites' : 'Add to favorites'}
                  >
                    <svg width="16" height="16" viewBox="0 0 24 24" fill={sim.isFavorite ? 'currentColor' : 'none'} stroke="currentColor" strokeWidth="2">
                      <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                    </svg>
                  </button>
                  <span className="sim-name">{sim.name || `Simulation #${sim.id}`}</span>
                </div>
                <span className={`success-rate ${getSuccessRateColor(sim.successRate)}`}>
                  {sim.successRate.toFixed(1)}%
                </span>
              </div>

              <div className="simulation-card-details">
                <div className="sim-stat">
                  <span className="stat-label">Time Horizon</span>
                  <span className="stat-value">{sim.timeHorizonYears} years</span>
                </div>
                <div className="sim-stat">
                  <span className="stat-label">Final Value (P50)</span>
                  <span className="stat-value">{formatCurrency(sim.finalP50)}</span>
                </div>
                <div className="sim-stat">
                  <span className="stat-label">Starting Net Worth</span>
                  <span className="stat-value">{formatCurrency(sim.startingNetWorth)}</span>
                </div>
              </div>

              {sim.notes && (
                <div className="simulation-notes">
                  <p>{sim.notes}</p>
                </div>
              )}

              <div className="simulation-card-footer">
                <span className="sim-date">{formatDate(sim.createdAt)}</span>
                <div className="sim-actions">
                  <button
                    className="btn btn-sm btn-secondary"
                    onClick={() => fetchSimulationDetails(sim.id)}
                  >
                    View Details
                  </button>
                  {onLoadSimulation && (
                    <button
                      className="btn btn-sm btn-primary"
                      onClick={() => onLoadSimulation(sim)}
                    >
                      Load
                    </button>
                  )}
                  <button
                    className="btn btn-sm btn-danger"
                    onClick={() => deleteSimulation(sim.id)}
                  >
                    Delete
                  </button>
                </div>
              </div>
            </div>
          ))}
        </div>
      )}

      {showDetails && selectedSimulation && (
        <div className="modal-overlay" onClick={() => setShowDetails(false)}>
          <div className="simulation-details-modal" onClick={(e) => e.stopPropagation()}>
            <div className="modal-header">
              <h3>{selectedSimulation.name || `Simulation #${selectedSimulation.id}`}</h3>
              <button className="btn-close" onClick={() => setShowDetails(false)}>×</button>
            </div>
            <div className="modal-body">
              <div className="detail-section">
                <h4>Parameters</h4>
                <div className="detail-grid">
                  <div className="detail-item">
                    <span className="label">Time Horizon</span>
                    <span className="value">{selectedSimulation.params?.timeHorizonYears || selectedSimulation.timeHorizonYears} years</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Current Age</span>
                    <span className="value">{selectedSimulation.params?.currentAge}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Retirement Age</span>
                    <span className="value">{selectedSimulation.params?.retirementAge || 65}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Monthly Contribution</span>
                    <span className="value">{formatCurrency(selectedSimulation.params?.monthlyContribution || 0)}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Expected Return</span>
                    <span className="value">{((selectedSimulation.params?.expectedReturn || 0.07) * 100).toFixed(1)}%</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Volatility</span>
                    <span className="value">{((selectedSimulation.params?.volatility || 0.15) * 100).toFixed(1)}%</span>
                  </div>
                </div>
              </div>

              <div className="detail-section">
                <h4>Results</h4>
                <div className="detail-grid">
                  <div className="detail-item">
                    <span className="label">Success Rate</span>
                    <span className={`value ${getSuccessRateColor(selectedSimulation.successRate)}`}>
                      {selectedSimulation.successRate.toFixed(1)}%
                    </span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Starting Net Worth</span>
                    <span className="value">{formatCurrency(selectedSimulation.startingNetWorth)}</span>
                  </div>
                  <div className="detail-item">
                    <span className="label">Final P50</span>
                    <span className="value">{formatCurrency(selectedSimulation.finalP50)}</span>
                  </div>
                </div>
              </div>

              {selectedSimulation.notes && (
                <div className="detail-section">
                  <h4>Notes</h4>
                  <p>{selectedSimulation.notes}</p>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
