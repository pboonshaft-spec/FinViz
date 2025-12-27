import React, { useState, useMemo, useEffect } from 'react';
import { useApi } from '../../hooks/useApi';
import StatCard from '../StatCard';
import ChartCard from '../ChartCard';
import DebugPanel from '../DebugPanel';
import TimeSeriesChart from '../charts/TimeSeriesChart';
import CashFlowChart from '../charts/CashFlowChart';
import CategoryChart from '../charts/CategoryChart';
import TrendChart from '../charts/TrendChart';
import IncomeDistChart from '../charts/IncomeDistChart';
import SankeyChart from '../charts/SankeyChart';
import TransactionList from '../budget/TransactionList';
import AddDataDropdown from '../budget/AddDataDropdown';
import CSVImportModal from '../budget/CSVImportModal';

function BudgetTab() {
  const [processedData, setProcessedData] = useState(null);
  const [debugLogs, setDebugLogs] = useState([]);
  const [transactions, setTransactions] = useState([]);
  const [summary, setSummary] = useState(null);
  const [dateRange, setDateRange] = useState('30'); // days
  const [csvModalOpen, setCsvModalOpen] = useState(false);
  const [isImporting, setIsImporting] = useState(false);

  const { getTransactions, getTransactionSummary, syncTransactions, importCSV } = useApi();

  // Auto-sync and load Plaid transactions on mount
  useEffect(() => {
    autoSyncAndLoad();
  }, [dateRange]);

  const autoSyncAndLoad = async () => {
    const endDate = new Date().toISOString().split('T')[0];
    const startDate = new Date(Date.now() - parseInt(dateRange) * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

    try {
      // Silent sync from Plaid
      await syncTransactions(startDate, endDate);
    } catch (err) {
      // Sync may fail if no Plaid accounts linked - that's ok
      console.log('Auto-sync skipped:', err.message);
    }

    // Load whatever we have
    await loadPlaidData();
  };

  const loadPlaidData = async () => {
    const endDate = new Date().toISOString().split('T')[0];
    const startDate = new Date(Date.now() - parseInt(dateRange) * 24 * 60 * 60 * 1000).toISOString().split('T')[0];

    try {
      const [txns, sum] = await Promise.all([
        getTransactions(startDate, endDate),
        getTransactionSummary(startDate, endDate)
      ]);
      setTransactions(txns || []);
      setSummary(sum);

      // Convert to processedData format for charts
      if (txns && txns.length > 0) {
        const chartData = convertToChartFormat(txns, sum);
        setProcessedData(chartData);
      }
    } catch (err) {
      console.error('Failed to load transactions:', err);
    }
  };

  // Convert Plaid data to chart format (matching what charts expect)
  const convertToChartFormat = (txns, sum) => {
    // Group transactions by month for time series chart
    const monthlyData = {};
    const dailyData = {};

    txns.forEach(t => {
      // Parse date
      const date = new Date(t.date);
      const monthKey = date.toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
      const dayKey = t.date; // Already in YYYY-MM-DD format

      // Initialize monthly bucket
      if (!monthlyData[monthKey]) {
        monthlyData[monthKey] = { income: 0, expenses: 0, net: 0 };
      }

      // Initialize daily bucket
      if (!dailyData[dayKey]) {
        dailyData[dayKey] = 0;
      }

      // Plaid: negative = income, positive = expense
      if (t.amount < 0) {
        monthlyData[monthKey].income += Math.abs(t.amount);
        monthlyData[monthKey].net += Math.abs(t.amount);
      } else {
        monthlyData[monthKey].expenses += t.amount;
        monthlyData[monthKey].net -= t.amount;
        dailyData[dayKey] += t.amount; // Track daily spending
      }
    });

    // Categories for pie chart - must be object { "Food": 100, "Gas": 50 }
    const categories = {};
    (sum?.byCategory || []).forEach(c => {
      categories[c.category || 'Uncategorized'] = c.amount || 0;
    });

    // Transform transactions for SankeyChart (needs description field)
    // Also flip amount sign: Plaid uses positive=expense, but charts expect positive=income
    const chartTransactions = txns.map(t => ({
      ...t,
      amount: -t.amount, // Flip sign for chart convention
      description: t.name || t.merchantName || 'Transaction',
      category: t.category || 'Uncategorized',
      subcategory: t.subcategory || '', // Include subcategory for better classification
      isExpense: t.amount > 0
    }));

    return {
      totals: {
        income: sum?.totalIncome || 0,
        expenses: sum?.totalExpenses || 0,
        balance: sum?.netCashFlow || 0
      },
      monthlyData,
      dailyData,
      categories,
      transactions: chartTransactions
    };
  };

  const handleFilesSelected = async (files) => {
    setIsImporting(true);
    const importLogs = [];
    try {
      // Upload each file to the backend
      for (const file of files) {
        const result = await importCSV(file, 'transactions');
        importLogs.push(`✓ Imported ${result.imported} transactions from ${file.name}`);
        if (result.errors && result.errors.length > 0) {
          result.errors.forEach(e => importLogs.push(`⚠ ${e}`));
        }
      }

      // Reload data from backend after import
      setCsvModalOpen(false);
      setDebugLogs(importLogs);
      await loadPlaidData();
    } catch (err) {
      console.error('Failed to import CSV:', err);
      setDebugLogs([`✗ Import error: ${err.message}`]);
    } finally {
      setIsImporting(false);
    }
  };

  // Merge CSV data with Plaid data
  const mergeDataSources = (plaidData, csvData) => {
    return {
      totals: {
        income: (plaidData.totals?.income || 0) + (csvData.totals?.income || 0),
        expenses: (plaidData.totals?.expenses || 0) + (csvData.totals?.expenses || 0),
        balance: (plaidData.totals?.balance || 0) + (csvData.totals?.balance || 0)
      },
      monthlyData: mergeMonthlyData(plaidData.monthlyData || {}, csvData.monthlyData || {}),
      dailyData: mergeDailyData(plaidData.dailyData || {}, csvData.dailyData || {}),
      categories: mergeCategoriesData(plaidData.categories || [], csvData.categories || []),
      transactions: [...(plaidData.transactions || []), ...(csvData.transactions || [])]
    };
  };

  const mergeMonthlyData = (plaid, csv) => {
    const merged = { ...plaid };
    Object.entries(csv).forEach(([month, data]) => {
      if (merged[month]) {
        merged[month] = {
          income: merged[month].income + data.income,
          expenses: merged[month].expenses + data.expenses,
          net: merged[month].net + data.net
        };
      } else {
        merged[month] = { ...data };
      }
    });
    return merged;
  };

  const mergeDailyData = (plaid, csv) => {
    const merged = { ...plaid };
    Object.entries(csv).forEach(([day, amount]) => {
      merged[day] = (merged[day] || 0) + amount;
    });
    return merged;
  };

  const mergeCategoriesData = (plaidCategories, csvCategories) => {
    // Both should be objects { "Food": 100, "Gas": 50 }
    const merged = { ...plaidCategories };
    Object.entries(csvCategories).forEach(([category, amount]) => {
      merged[category] = (merged[category] || 0) + amount;
    });
    return merged;
  };

  const savingsRate = useMemo(() => {
    if (!processedData || processedData.totals.income === 0) return 0;
    return ((processedData.totals.balance / processedData.totals.income) * 100).toFixed(1);
  }, [processedData]);

  return (
    <div className="tab-content">
      <div className="tab-header">
        <div className="tab-header-text">
          <h2>Monthly Budget</h2>
          <p>Track income and expenses from your bank</p>
        </div>
        <div className="budget-controls">
          <div className="date-range-selector">
            <div className="btn-group">
              {[
                { value: '30', label: '30 Days' },
                { value: '90', label: '3 Months' },
                { value: '180', label: '6 Months' },
                { value: '365', label: '1 Year' },
              ].map(opt => (
                <button
                  key={opt.value}
                  className={`btn btn-sm ${dateRange === opt.value ? 'btn-primary' : 'btn-secondary'}`}
                  onClick={() => setDateRange(opt.value)}
                >
                  {opt.label}
                </button>
              ))}
            </div>
          </div>
          <AddDataDropdown onImportCSV={() => setCsvModalOpen(true)} />
        </div>
      </div>

      <DebugPanel logs={debugLogs} />

      {!processedData ? (
        <div className="empty-state">
          <div className="empty-state-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
              <line x1="3" y1="9" x2="21" y2="9" />
              <line x1="9" y1="21" x2="9" y2="9" />
            </svg>
          </div>
          <h2>No transaction data yet</h2>
          <p>Connect your bank account in Settings or import a CSV file to start tracking.</p>
          <button className="btn btn-primary" onClick={() => setCsvModalOpen(true)}>
            Import CSV
          </button>
        </div>
      ) : (
        <>

          <div className="stats-grid">
            <StatCard
              label="Total Income"
              value={`$${processedData.totals.income.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
              change="Revenue"
              changeType="positive"
              valueColor="#00d4aa"
            />
            <StatCard
              label="Total Expenses"
              value={`$${processedData.totals.expenses.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
              change="Spending"
              changeType="negative"
              valueColor="#ff6b6b"
            />
            <StatCard
              label="Net Balance"
              value={`$${processedData.totals.balance.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`}
              change={processedData.totals.balance >= 0 ? 'Surplus' : 'Deficit'}
              changeType={processedData.totals.balance >= 0 ? 'positive' : 'negative'}
              valueColor={processedData.totals.balance >= 0 ? '#00d4aa' : '#ff6b6b'}
            />
            <StatCard
              label="Savings Rate"
              value={`${savingsRate}%`}
              change={savingsRate >= 20 ? 'Great!' : 'Can improve'}
              changeType={savingsRate >= 20 ? 'positive' : 'negative'}
            />
          </div>

          <div className="charts-grid">
            <ChartCard title="Income vs Expenses Over Time" fullWidth>
              <TimeSeriesChart data={processedData} />
            </ChartCard>

            <ChartCard title="Spending by Category">
              <CategoryChart data={processedData} />
            </ChartCard>

            <ChartCard title="Monthly Cash Flow">
              <CashFlowChart data={processedData} />
            </ChartCard>

            <ChartCard title="Daily Spending Trend">
              <TrendChart data={processedData} />
            </ChartCard>

            <ChartCard title="Income Distribution">
              <IncomeDistChart data={processedData} />
            </ChartCard>

            <ChartCard title="Cash Flow Sankey" fullWidth>
              <SankeyChart data={processedData} />
            </ChartCard>

            {transactions.length > 0 && (
              <ChartCard title="Recent Transactions" fullWidth>
                <TransactionList transactions={transactions} />
              </ChartCard>
            )}
          </div>
        </>
      )}

      <CSVImportModal
        isOpen={csvModalOpen}
        onClose={() => setCsvModalOpen(false)}
        onFilesSelected={handleFilesSelected}
        isProcessing={isImporting}
      />
    </div>
  );
}

export default BudgetTab;
