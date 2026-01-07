import React, { useState } from 'react';

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
  employerMatch: 0,
  employerMatchLimit: 0,
  volatility: 0.15,
  pensionIncome: 0,
  withdrawalStrategy: 'fixed',
  retirementTaxRate: 0.22,
  excludeCreditCardDebt: true,
  enableGlidePath: false,
};

function SimulationInputs({ params, onChange, onRun, loading, disabled, onGenerateReport, reportLoading, hasResults }) {
  const [showGrowthSettings, setShowGrowthSettings] = useState(false);
  const [showRetirementIncome, setShowRetirementIncome] = useState(false);
  const [showAdvanced, setShowAdvanced] = useState(false);

  const currentParams = { ...DEFAULT_PARAMS, ...params };

  const handleChange = (field, value) => {
    onChange({ ...currentParams, [field]: value });
  };

  const formatPercent = (value) => (value * 100).toFixed(1);
  const parsePercent = (str) => parseFloat(str) / 100;

  return (
    <div className="simulation-inputs">
      {/* Tier 1: Essential (Always Visible) */}
      <div className="input-section essential">
        <div className="input-row">
          <div className="input-group">
            <label>Current Age</label>
            <input
              type="number"
              value={currentParams.currentAge}
              onChange={(e) => handleChange('currentAge', parseInt(e.target.value) || 0)}
              min={18}
              max={100}
            />
          </div>
          <div className="input-group">
            <label>Retirement Age</label>
            <input
              type="number"
              value={currentParams.retirementAge}
              onChange={(e) => handleChange('retirementAge', parseInt(e.target.value) || 0)}
              min={currentParams.currentAge}
              max={100}
            />
          </div>
          <div className="input-group">
            <label>Time Horizon (Years)</label>
            <input
              type="number"
              value={currentParams.timeHorizonYears}
              onChange={(e) => handleChange('timeHorizonYears', parseInt(e.target.value) || 0)}
              min={1}
              max={80}
            />
          </div>
        </div>

        <div className="input-row">
          <div className="input-group wide">
            <label>Monthly Contribution</label>
            <div className="input-with-prefix">
              <span className="prefix">$</span>
              <input
                type="text"
                inputMode="numeric"
                value={currentParams.monthlyContribution}
                onChange={(e) => handleChange('monthlyContribution', parseInt(e.target.value.replace(/[^0-9]/g, '')) || 0)}
              />
            </div>
          </div>
          <div className="input-group wide">
            <label>Retirement Monthly Spending</label>
            <div className="input-with-prefix">
              <span className="prefix">$</span>
              <input
                type="text"
                inputMode="numeric"
                value={currentParams.retirementSpending}
                onChange={(e) => handleChange('retirementSpending', parseInt(e.target.value.replace(/[^0-9]/g, '')) || 0)}
              />
            </div>
          </div>
        </div>

        <div className="input-row">
          <label className="checkbox-label">
            <input
              type="checkbox"
              checked={currentParams.excludeCreditCardDebt}
              onChange={(e) => handleChange('excludeCreditCardDebt', e.target.checked)}
            />
            <span>Exclude credit card debt from projection</span>
            <span className="checkbox-hint">Credit cards are typically paid monthly, not amortized</span>
          </label>
        </div>
      </div>

      {/* Tier 2: Growth & Inflation Settings */}
      <div className="input-section collapsible">
        <button
          className="section-toggle"
          onClick={() => setShowGrowthSettings(!showGrowthSettings)}
        >
          <span className={`toggle-icon ${showGrowthSettings ? 'open' : ''}`}>&#9656;</span>
          Growth & Inflation Settings
        </button>
        {showGrowthSettings && (
          <div className="section-content">
            <div className="input-row">
              <div className="input-group">
                <label>Expected Return</label>
                <div className="input-with-suffix">
                  <input
                    type="number"
                    value={formatPercent(currentParams.expectedReturn)}
                    onChange={(e) => handleChange('expectedReturn', parsePercent(e.target.value))}
                    step={0.5}
                  />
                  <span className="suffix">%</span>
                </div>
                <span className="input-hint">Average annual investment return</span>
              </div>
              <div className="input-group">
                <label>Volatility</label>
                <div className="input-with-suffix">
                  <input
                    type="number"
                    value={formatPercent(currentParams.volatility)}
                    onChange={(e) => handleChange('volatility', parsePercent(e.target.value))}
                    step={0.5}
                  />
                  <span className="suffix">%</span>
                </div>
                <span className="input-hint">Standard deviation of returns</span>
              </div>
            </div>
            <div className="input-row">
              <div className="input-group">
                <label>Inflation Rate</label>
                <div className="input-with-suffix">
                  <input
                    type="number"
                    value={formatPercent(currentParams.inflationRate)}
                    onChange={(e) => handleChange('inflationRate', parsePercent(e.target.value))}
                    step={0.1}
                  />
                  <span className="suffix">%</span>
                </div>
              </div>
              <div className="input-group">
                <label>Contribution Growth</label>
                <div className="input-with-suffix">
                  <input
                    type="number"
                    value={formatPercent(currentParams.contributionGrowth)}
                    onChange={(e) => handleChange('contributionGrowth', parsePercent(e.target.value))}
                    step={0.5}
                  />
                  <span className="suffix">%</span>
                </div>
                <span className="input-hint">Annual salary increase</span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Tier 2: Retirement Income */}
      <div className="input-section collapsible">
        <button
          className="section-toggle"
          onClick={() => setShowRetirementIncome(!showRetirementIncome)}
        >
          <span className={`toggle-icon ${showRetirementIncome ? 'open' : ''}`}>&#9656;</span>
          Retirement Income Sources
        </button>
        {showRetirementIncome && (
          <div className="section-content">
            <div className="input-row">
              <div className="input-group">
                <label>Social Security (Monthly)</label>
                <div className="input-with-prefix">
                  <span className="prefix">$</span>
                  <input
                    type="number"
                    value={currentParams.socialSecurityAmount}
                    onChange={(e) => handleChange('socialSecurityAmount', parseFloat(e.target.value) || 0)}
                    min={0}
                    step={100}
                  />
                </div>
              </div>
              <div className="input-group">
                <label>SS Start Age</label>
                <input
                  type="number"
                  value={currentParams.socialSecurityAge}
                  onChange={(e) => handleChange('socialSecurityAge', parseInt(e.target.value) || 67)}
                  min={62}
                  max={70}
                />
              </div>
            </div>
            <div className="input-row">
              <div className="input-group">
                <label>Pension Income (Monthly)</label>
                <div className="input-with-prefix">
                  <span className="prefix">$</span>
                  <input
                    type="number"
                    value={currentParams.pensionIncome}
                    onChange={(e) => handleChange('pensionIncome', parseFloat(e.target.value) || 0)}
                    min={0}
                    step={100}
                  />
                </div>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Tier 3: Advanced Settings */}
      <div className="input-section collapsible">
        <button
          className="section-toggle"
          onClick={() => setShowAdvanced(!showAdvanced)}
        >
          <span className={`toggle-icon ${showAdvanced ? 'open' : ''}`}>&#9656;</span>
          Advanced Settings
        </button>
        {showAdvanced && (
          <div className="section-content">
            <div className="input-row">
              <div className="input-group">
                <label>Employer 401k Match</label>
                <div className="input-with-suffix">
                  <input
                    type="number"
                    value={formatPercent(currentParams.employerMatch)}
                    onChange={(e) => handleChange('employerMatch', parsePercent(e.target.value))}
                    step={1}
                    min={0}
                    max={100}
                  />
                  <span className="suffix">%</span>
                </div>
                <span className="input-hint">% of your contribution matched</span>
              </div>
              <div className="input-group">
                <label>Match Limit (Annual)</label>
                <div className="input-with-prefix">
                  <span className="prefix">$</span>
                  <input
                    type="number"
                    value={currentParams.employerMatchLimit}
                    onChange={(e) => handleChange('employerMatchLimit', parseFloat(e.target.value) || 0)}
                    min={0}
                    step={500}
                  />
                </div>
              </div>
            </div>
            <div className="input-row">
              <div className="input-group">
                <label>Withdrawal Strategy</label>
                <select
                  value={currentParams.withdrawalStrategy}
                  onChange={(e) => handleChange('withdrawalStrategy', e.target.value)}
                >
                  <option value="fixed">Fixed 4% Rule</option>
                  <option value="dynamic">Dynamic (4% of current)</option>
                  <option value="guardrails">Guardrails (3-5%)</option>
                </select>
                <span className="input-hint">How withdrawals adjust over time</span>
              </div>
              <div className="input-group">
                <label>Retirement Tax Rate</label>
                <div className="input-with-suffix">
                  <input
                    type="number"
                    value={formatPercent(currentParams.retirementTaxRate)}
                    onChange={(e) => handleChange('retirementTaxRate', parsePercent(e.target.value))}
                    step={1}
                    min={0}
                    max={50}
                  />
                  <span className="suffix">%</span>
                </div>
              </div>
            </div>
            <div className="input-row">
              <label className="checkbox-label">
                <input
                  type="checkbox"
                  checked={currentParams.enableGlidePath}
                  onChange={(e) => handleChange('enableGlidePath', e.target.checked)}
                />
                <span>Enable Glide Path (Target Date)</span>
                <span className="checkbox-hint">Auto-adjusts stock/bond allocation by age (90% stocks young → 40% at retirement)</span>
              </label>
            </div>
          </div>
        )}
      </div>

      <div className="simulation-actions">
        <button
          className="btn btn-primary btn-lg"
          onClick={onRun}
          disabled={loading || disabled}
        >
          {loading ? 'Running Simulation...' : 'Run Simulation'}
        </button>
        {onGenerateReport && (
          <button
            className="btn btn-secondary btn-lg"
            onClick={onGenerateReport}
            disabled={reportLoading || disabled}
            title={!hasResults ? 'Run a simulation first to include projections' : 'Generate PDF report'}
          >
            {reportLoading ? 'Generating...' : 'Download Report'}
          </button>
        )}
        <span className="simulation-info">5,000 Monte Carlo simulations</span>
      </div>
    </div>
  );
}

export default SimulationInputs;
