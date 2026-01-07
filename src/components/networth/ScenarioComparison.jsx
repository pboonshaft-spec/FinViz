import React, { useState } from 'react';
import Chart from 'react-apexcharts';
import { useApi } from '../../hooks/useApi';

const DEFAULT_PARAMS = {
  timeHorizonYears: 30,
  monthlyContribution: 0,
  retirementAge: 65,
  currentAge: 35,
  expectedReturn: 0.07,
  inflationRate: 0.03,
  contributionGrowth: 0.02,
  retirementSpending: 0,
  socialSecurityAmount: 0,
  socialSecurityAge: 67,
  volatility: 0.15,
  withdrawalStrategy: 'fixed',
  retirementTaxRate: 0.22,
  excludeCreditCardDebt: true,
  enableGlidePath: false,
};

const SCENARIO_COLORS = ['#00d4aa', '#6366f1', '#f59e0b', '#ef4444', '#06b6d4'];

function ScenarioComparison({ baseParams, onClose }) {
  const { post, loading } = useApi();
  const [scenarios, setScenarios] = useState([
    { name: 'Current Plan', params: { ...DEFAULT_PARAMS, ...baseParams } },
    { name: 'Save More', params: { ...DEFAULT_PARAMS, ...baseParams, monthlyContribution: (baseParams?.monthlyContribution || 0) + 500 } },
  ]);
  const [results, setResults] = useState(null);
  const [editingScenario, setEditingScenario] = useState(null);

  const addScenario = () => {
    if (scenarios.length >= 5) return;
    setScenarios([...scenarios, {
      name: `Scenario ${scenarios.length + 1}`,
      params: { ...DEFAULT_PARAMS, ...baseParams }
    }]);
  };

  const removeScenario = (index) => {
    if (scenarios.length <= 2) return;
    setScenarios(scenarios.filter((_, i) => i !== index));
  };

  const updateScenario = (index, field, value) => {
    const updated = [...scenarios];
    if (field === 'name') {
      updated[index].name = value;
    } else {
      updated[index].params[field] = value;
    }
    setScenarios(updated);
  };

  const runComparison = async () => {
    try {
      const response = await post('/api/monte-carlo/scenarios', { scenarios });
      setResults(response);
    } catch (err) {
      console.error('Failed to run comparison:', err);
    }
  };

  // Quick scenario presets
  const applyPreset = (index, preset) => {
    const updated = [...scenarios];
    const base = scenarios[0].params;

    switch (preset) {
      case 'save_more':
        updated[index].name = 'Save $500 More/Month';
        updated[index].params = { ...base, monthlyContribution: base.monthlyContribution + 500 };
        break;
      case 'retire_early':
        updated[index].name = 'Retire 3 Years Earlier';
        updated[index].params = { ...base, retirementAge: base.retirementAge - 3 };
        break;
      case 'retire_later':
        updated[index].name = 'Retire 3 Years Later';
        updated[index].params = { ...base, retirementAge: base.retirementAge + 3 };
        break;
      case 'spend_less':
        updated[index].name = 'Spend 20% Less in Retirement';
        updated[index].params = { ...base, retirementSpending: base.retirementSpending * 0.8 };
        break;
      case 'aggressive':
        updated[index].name = 'More Aggressive (Higher Risk)';
        updated[index].params = { ...base, expectedReturn: 0.09, volatility: 0.20 };
        break;
      case 'conservative':
        updated[index].name = 'More Conservative (Lower Risk)';
        updated[index].params = { ...base, expectedReturn: 0.05, volatility: 0.10 };
        break;
      default:
        break;
    }
    setScenarios(updated);
  };

  const formatCurrency = (val) => {
    if (val >= 1000000) return `$${(val / 1000000).toFixed(1)}M`;
    if (val >= 1000) return `$${(val / 1000).toFixed(0)}K`;
    return `$${val.toFixed(0)}`;
  };

  const chartOptions = results ? {
    chart: {
      type: 'line',
      height: 350,
      background: 'transparent',
      toolbar: { show: false },
      animations: { enabled: true, speed: 800 },
    },
    colors: SCENARIO_COLORS.slice(0, results.scenarios.length),
    stroke: { curve: 'smooth', width: 2 },
    xaxis: {
      categories: results.scenarios[0]?.projections.map(p => p.age ? `Age ${p.age}` : `Year ${p.year}`) || [],
      labels: { style: { colors: '#a0a0a0' }, rotate: -45 },
    },
    yaxis: {
      labels: {
        style: { colors: '#a0a0a0' },
        formatter: (val) => formatCurrency(val),
      },
    },
    grid: { borderColor: '#2a2a2a', strokeDashArray: 3 },
    legend: { position: 'top', labels: { colors: '#e0e0e0' } },
    tooltip: {
      theme: 'dark',
      y: { formatter: (val) => formatCurrency(val) },
    },
  } : {};

  const chartSeries = results ? results.scenarios.map(s => ({
    name: s.name,
    data: s.projections.map(p => p.p50),
  })) : [];

  return (
    <div className="scenario-comparison">
      <div className="scenario-comparison-header">
        <h3>Compare Scenarios</h3>
        <button className="btn btn-secondary btn-sm" onClick={onClose}>Close</button>
      </div>

      {/* Scenario Builder */}
      <div className="scenario-builder">
        <div className="scenarios-list">
          {scenarios.map((scenario, index) => (
            <div key={index} className="scenario-card" style={{ borderLeftColor: SCENARIO_COLORS[index] }}>
              <div className="scenario-card-header">
                <input
                  className="scenario-name-input"
                  value={scenario.name}
                  onChange={(e) => updateScenario(index, 'name', e.target.value)}
                />
                {index > 0 && scenarios.length > 2 && (
                  <button className="btn-remove" onClick={() => removeScenario(index)}>×</button>
                )}
              </div>

              {/* Quick presets for non-baseline scenarios */}
              {index > 0 && (
                <div className="scenario-presets">
                  <select onChange={(e) => { applyPreset(index, e.target.value); e.target.value = ''; }}>
                    <option value="">Apply preset...</option>
                    <option value="save_more">Save $500 More/Month</option>
                    <option value="retire_early">Retire 3 Years Earlier</option>
                    <option value="retire_later">Retire 3 Years Later</option>
                    <option value="spend_less">Spend 20% Less in Retirement</option>
                    <option value="aggressive">More Aggressive</option>
                    <option value="conservative">More Conservative</option>
                  </select>
                </div>
              )}

              <div className="scenario-params">
                <div className="param-row">
                  <label>Monthly Savings</label>
                  <div className="input-with-prefix">
                    <span className="prefix">$</span>
                    <input
                      type="number"
                      value={scenario.params.monthlyContribution}
                      onChange={(e) => updateScenario(index, 'monthlyContribution', parseInt(e.target.value) || 0)}
                    />
                  </div>
                </div>
                <div className="param-row">
                  <label>Retirement Age</label>
                  <input
                    type="number"
                    value={scenario.params.retirementAge}
                    onChange={(e) => updateScenario(index, 'retirementAge', parseInt(e.target.value) || 0)}
                  />
                </div>
                <div className="param-row">
                  <label>Monthly Spending</label>
                  <div className="input-with-prefix">
                    <span className="prefix">$</span>
                    <input
                      type="number"
                      value={scenario.params.retirementSpending}
                      onChange={(e) => updateScenario(index, 'retirementSpending', parseInt(e.target.value) || 0)}
                    />
                  </div>
                </div>

                <button
                  className="btn-expand"
                  onClick={() => setEditingScenario(editingScenario === index ? null : index)}
                >
                  {editingScenario === index ? 'Less Options' : 'More Options'}
                </button>

                {editingScenario === index && (
                  <div className="expanded-params">
                    <div className="param-row">
                      <label>Expected Return</label>
                      <div className="input-with-suffix">
                        <input
                          type="number"
                          step="0.1"
                          value={(scenario.params.expectedReturn * 100).toFixed(1)}
                          onChange={(e) => updateScenario(index, 'expectedReturn', parseFloat(e.target.value) / 100 || 0)}
                        />
                        <span className="suffix">%</span>
                      </div>
                    </div>
                    <div className="param-row">
                      <label>Volatility</label>
                      <div className="input-with-suffix">
                        <input
                          type="number"
                          step="0.1"
                          value={(scenario.params.volatility * 100).toFixed(1)}
                          onChange={(e) => updateScenario(index, 'volatility', parseFloat(e.target.value) / 100 || 0)}
                        />
                        <span className="suffix">%</span>
                      </div>
                    </div>
                    <div className="param-row">
                      <label>Social Security</label>
                      <div className="input-with-prefix">
                        <span className="prefix">$</span>
                        <input
                          type="number"
                          value={scenario.params.socialSecurityAmount}
                          onChange={(e) => updateScenario(index, 'socialSecurityAmount', parseInt(e.target.value) || 0)}
                        />
                      </div>
                    </div>
                    <div className="param-row">
                      <label>SS Start Age</label>
                      <input
                        type="number"
                        value={scenario.params.socialSecurityAge}
                        onChange={(e) => updateScenario(index, 'socialSecurityAge', parseInt(e.target.value) || 0)}
                      />
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}

          {scenarios.length < 5 && (
            <button className="btn-add-scenario" onClick={addScenario}>
              + Add Scenario
            </button>
          )}
        </div>

        <button
          className="btn btn-primary btn-run-comparison"
          onClick={runComparison}
          disabled={loading}
        >
          {loading ? 'Running...' : 'Compare Scenarios'}
        </button>
      </div>

      {/* Results */}
      {results && (
        <div className="comparison-results">
          <h4>Comparison Results</h4>

          {/* Best Scenario Banner */}
          <div className="best-scenario-banner">
            <span className="label">Best Scenario:</span>
            <span className="value">{results.bestScenario}</span>
          </div>

          {/* Chart */}
          <div className="comparison-chart">
            <h5>Median Wealth Over Time (P50)</h5>
            <Chart options={chartOptions} series={chartSeries} type="line" height={350} />
          </div>

          {/* Summary Table */}
          <div className="comparison-table">
            <table>
              <thead>
                <tr>
                  <th>Scenario</th>
                  <th>Success Rate</th>
                  <th>Final P50</th>
                  <th>Total Contributions</th>
                </tr>
              </thead>
              <tbody>
                {results.scenarios.map((s, i) => (
                  <tr key={i} className={s.name === results.bestScenario ? 'best' : ''}>
                    <td>
                      <span className="color-dot" style={{ backgroundColor: SCENARIO_COLORS[i] }}></span>
                      {s.name}
                    </td>
                    <td className={s.summary.successRate >= 80 ? 'good' : s.summary.successRate >= 50 ? 'ok' : 'bad'}>
                      {s.summary.successRate.toFixed(1)}%
                    </td>
                    <td>{formatCurrency(s.summary.finalP50)}</td>
                    <td>{formatCurrency(s.summary.totalContributions)}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Recommendations */}
          {results.comparisons?.length > 0 && (
            <div className="comparison-recommendations">
              <h5>Analysis</h5>
              {results.comparisons.map((c, i) => (
                <div key={i} className="recommendation-card">
                  <div className="comparison-header">
                    <span>{c.scenarioA}</span>
                    <span className="vs">vs</span>
                    <span>{c.scenarioB}</span>
                  </div>
                  <div className="comparison-diff">
                    <span className={c.successRateDiff >= 0 ? 'positive' : 'negative'}>
                      {c.successRateDiff >= 0 ? '+' : ''}{c.successRateDiff.toFixed(1)}% success rate
                    </span>
                    <span className={c.finalP50Diff >= 0 ? 'positive' : 'negative'}>
                      {c.finalP50Diff >= 0 ? '+' : ''}{formatCurrency(c.finalP50Diff)} final wealth
                    </span>
                  </div>
                  <p className="recommendation-text">{c.recommendation}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default ScenarioComparison;
