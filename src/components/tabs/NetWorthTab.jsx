import React, { useState, useEffect, useMemo } from 'react';
import { useApi } from '../../hooks/useApi';
import StatCard from '../StatCard';
import ChartCard from '../ChartCard';
import AssetList from '../networth/AssetList';
import DebtList from '../networth/DebtList';
import AssetForm from '../networth/AssetForm';
import DebtForm from '../networth/DebtForm';
import MonteCarloChart from '../networth/MonteCarloChart';
import SimulationInputs from '../networth/SimulationInputs';

function NetWorthTab() {
  const [assets, setAssets] = useState([]);
  const [debts, setDebts] = useState([]);
  const [assetTypes, setAssetTypes] = useState([]);
  const [projection, setProjection] = useState(null);
  const [simulationParams, setSimulationParams] = useState({});
  const [simulationLoading, setSimulationLoading] = useState(false);
  const [showAssetForm, setShowAssetForm] = useState(false);
  const [showDebtForm, setShowDebtForm] = useState(false);
  const [editingAsset, setEditingAsset] = useState(null);
  const [editingDebt, setEditingDebt] = useState(null);

  const {
    loading,
    error,
    getAssets,
    getDebts,
    getAssetTypes,
    createAsset,
    updateAsset,
    deleteAsset,
    createDebt,
    updateDebt,
    deleteDebt,
    runMonteCarlo
  } = useApi();

  // Load initial data
  useEffect(() => {
    loadData();
  }, []);

  const loadData = async () => {
    try {
      const [assetsData, debtsData, typesData] = await Promise.all([
        getAssets(),
        getDebts(),
        getAssetTypes(),
      ]);
      setAssets(assetsData || []);
      setDebts(debtsData || []);
      setAssetTypes(typesData || []);
    } catch (err) {
      console.error('Failed to load data:', err);
    }
  };

  // Calculate totals
  const totals = useMemo(() => {
    const totalAssets = assets.reduce((sum, a) => sum + a.currentValue, 0);
    const totalDebts = debts.reduce((sum, d) => sum + d.currentBalance, 0);
    return {
      assets: totalAssets,
      debts: totalDebts,
      netWorth: totalAssets - totalDebts,
    };
  }, [assets, debts]);

  // Run Monte Carlo simulation
  const handleRunSimulation = async () => {
    setSimulationLoading(true);
    try {
      const result = await runMonteCarlo(simulationParams);
      setProjection(result);
    } catch (err) {
      console.error('Failed to run simulation:', err);
    } finally {
      setSimulationLoading(false);
    }
  };

  // Asset handlers
  const handleAddAsset = async (asset) => {
    try {
      await createAsset(asset);
      await loadData();
      setShowAssetForm(false);
      setProjection(null); // Clear projection to recalculate
    } catch (err) {
      console.error('Failed to create asset:', err);
    }
  };

  const handleUpdateAsset = async (id, updates) => {
    try {
      await updateAsset(id, updates);
      await loadData();
      setEditingAsset(null);
      setProjection(null);
    } catch (err) {
      console.error('Failed to update asset:', err);
    }
  };

  const handleDeleteAsset = async (id) => {
    if (!confirm('Are you sure you want to delete this asset?')) return;
    try {
      await deleteAsset(id);
      await loadData();
      setProjection(null);
    } catch (err) {
      console.error('Failed to delete asset:', err);
    }
  };

  // Debt handlers
  const handleAddDebt = async (debt) => {
    try {
      await createDebt(debt);
      await loadData();
      setShowDebtForm(false);
      setProjection(null);
    } catch (err) {
      console.error('Failed to create debt:', err);
    }
  };

  const handleUpdateDebt = async (id, updates) => {
    try {
      await updateDebt(id, updates);
      await loadData();
      setEditingDebt(null);
      setProjection(null);
    } catch (err) {
      console.error('Failed to update debt:', err);
    }
  };

  const handleDeleteDebt = async (id) => {
    if (!confirm('Are you sure you want to delete this debt?')) return;
    try {
      await deleteDebt(id);
      await loadData();
      setProjection(null);
    } catch (err) {
      console.error('Failed to delete debt:', err);
    }
  };

  return (
    <div className="tab-content">
      <div className="tab-header">
        <div className="tab-header-text">
          <h2>Net Worth</h2>
          <p>Track assets, debts, and project future wealth</p>
        </div>
      </div>

      {error && (
        <div className="error-banner">
          <span>Error: {error}</span>
        </div>
      )}

      <div className="stats-grid">
        <StatCard
          label="Total Assets"
          value={`$${totals.assets.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
          change={`${assets.length} accounts`}
          changeType="positive"
          valueColor="#00d4aa"
        />
        <StatCard
          label="Total Debts"
          value={`$${totals.debts.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
          change={`${debts.length} accounts`}
          changeType="negative"
          valueColor="#ff6b6b"
        />
        <StatCard
          label="Net Worth"
          value={`$${totals.netWorth.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
          change={totals.netWorth >= 0 ? 'Positive' : 'Negative'}
          changeType={totals.netWorth >= 0 ? 'positive' : 'negative'}
          valueColor={totals.netWorth >= 0 ? '#00d4aa' : '#ff6b6b'}
        />
        <StatCard
          label="Debt Ratio"
          value={totals.assets > 0 ? `${((totals.debts / totals.assets) * 100).toFixed(1)}%` : '0%'}
          change={totals.debts / totals.assets < 0.3 ? 'Healthy' : 'High'}
          changeType={totals.debts / totals.assets < 0.3 ? 'positive' : 'negative'}
        />
      </div>

      <div className="networth-grid">
        <div className="networth-column">
          <div className="list-header">
            <h3>Assets</h3>
            <button className="btn btn-primary btn-sm" onClick={() => setShowAssetForm(true)}>
              + Add Asset
            </button>
          </div>

          {showAssetForm && (
            <AssetForm
              assetTypes={assetTypes}
              onSubmit={handleAddAsset}
              onCancel={() => setShowAssetForm(false)}
            />
          )}

          {editingAsset && (
            <AssetForm
              assetTypes={assetTypes}
              asset={editingAsset}
              onSubmit={(updates) => handleUpdateAsset(editingAsset.id, updates)}
              onCancel={() => setEditingAsset(null)}
            />
          )}

          <AssetList
            assets={assets}
            onEdit={setEditingAsset}
            onDelete={handleDeleteAsset}
          />
        </div>

        <div className="networth-column">
          <div className="list-header">
            <h3>Debts</h3>
            <button className="btn btn-primary btn-sm" onClick={() => setShowDebtForm(true)}>
              + Add Debt
            </button>
          </div>

          {showDebtForm && (
            <DebtForm
              onSubmit={handleAddDebt}
              onCancel={() => setShowDebtForm(false)}
            />
          )}

          {editingDebt && (
            <DebtForm
              debt={editingDebt}
              onSubmit={(updates) => handleUpdateDebt(editingDebt.id, updates)}
              onCancel={() => setEditingDebt(null)}
            />
          )}

          <DebtList
            debts={debts}
            onEdit={setEditingDebt}
            onDelete={handleDeleteDebt}
          />
        </div>
      </div>

      <ChartCard title="Monte Carlo Projection" fullWidth>
        {assets.length === 0 ? (
          <div className="empty-state-small">
            <p>Add assets to run Monte Carlo projections</p>
          </div>
        ) : (
          <>
            <SimulationInputs
              params={simulationParams}
              onChange={setSimulationParams}
              onRun={handleRunSimulation}
              loading={simulationLoading}
              disabled={assets.length === 0}
            />
            {projection && <MonteCarloChart projection={projection} />}
          </>
        )}
      </ChartCard>
    </div>
  );
}

export default NetWorthTab;
