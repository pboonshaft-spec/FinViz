import React from 'react';
import Chart from 'react-apexcharts';

function MonteCarloChart({ projection }) {
  if (!projection || !projection.projections) {
    return null;
  }

  const { projections, summary, milestones, insights } = projection;

  // Find retirement year for annotation
  const retirementYearIndex = projections.findIndex(p => p.phase === 'distribution');

  const categories = projections.map(p =>
    p.age ? `Age ${p.age}` : `Year ${p.year}`
  );

  const options = {
    chart: {
      type: 'area',
      height: 400,
      background: 'transparent',
      toolbar: {
        show: true,
        tools: {
          download: true,
          selection: false,
          zoom: false,
          zoomin: false,
          zoomout: false,
          pan: false,
          reset: false,
        },
        export: {
          csv: {
            filename: 'monte-carlo-projection',
          },
          svg: {
            filename: 'monte-carlo-projection',
          },
          png: {
            filename: 'monte-carlo-projection',
          },
        },
        autoSelected: 'none',
      },
      zoom: {
        enabled: false,
      },
      animations: {
        enabled: true,
        speed: 800,
      },
    },
    colors: ['#00d4aa', '#6366f1', '#ff6b6b'],
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 0.3,
        opacityFrom: 0.5,
        opacityTo: 0.1,
        stops: [0, 100],
      },
    },
    stroke: {
      curve: 'smooth',
      width: 2,
    },
    xaxis: {
      categories,
      labels: {
        style: {
          colors: '#a0a0a0',
        },
        rotate: -45,
        rotateAlways: categories.length > 15,
      },
      axisBorder: {
        show: false,
      },
      axisTicks: {
        show: false,
      },
    },
    yaxis: {
      labels: {
        style: {
          colors: '#a0a0a0',
        },
        formatter: (value) => {
          if (value >= 1000000) {
            return `$${(value / 1000000).toFixed(1)}M`;
          }
          if (value >= 1000) {
            return `$${(value / 1000).toFixed(0)}K`;
          }
          if (value <= -1000000) {
            return `-$${(Math.abs(value) / 1000000).toFixed(1)}M`;
          }
          if (value <= -1000) {
            return `-$${(Math.abs(value) / 1000).toFixed(0)}K`;
          }
          return `$${value.toFixed(0)}`;
        },
      },
    },
    grid: {
      borderColor: '#2a2a2a',
      strokeDashArray: 3,
    },
    legend: {
      position: 'top',
      horizontalAlign: 'center',
      labels: {
        colors: '#e0e0e0',
      },
    },
    tooltip: {
      theme: 'dark',
      shared: true,
      y: {
        formatter: (value) => {
          return `$${value.toLocaleString('en-US', { minimumFractionDigits: 0, maximumFractionDigits: 0 })}`;
        },
      },
    },
    dataLabels: {
      enabled: false,
    },
    annotations: retirementYearIndex > 0 ? {
      xaxis: [{
        x: categories[retirementYearIndex],
        borderColor: '#f59e0b',
        strokeDashArray: 5,
        label: {
          borderColor: '#f59e0b',
          style: {
            color: '#1a1a2e',
            background: '#f59e0b',
          },
          text: 'Retirement',
        },
      }],
    } : {},
  };

  const series = [
    {
      name: '90th Percentile (Optimistic)',
      data: projections.map(p => p.p90),
    },
    {
      name: '50th Percentile (Median)',
      data: projections.map(p => p.p50),
    },
    {
      name: '10th Percentile (Conservative)',
      data: projections.map(p => p.p10),
    },
  ];

  // Determine success rate color and message
  const getSuccessRateStyle = (rate) => {
    if (rate >= 90) return { color: '#00d4aa', label: 'Excellent' };
    if (rate >= 75) return { color: '#6366f1', label: 'Good' };
    if (rate >= 50) return { color: '#f59e0b', label: 'Needs Work' };
    return { color: '#ff6b6b', label: 'At Risk' };
  };

  const successStyle = getSuccessRateStyle(summary.successRate);

  return (
    <div className="monte-carlo-chart">
      {/* Success Rate Banner */}
      {summary.successRate !== undefined && (
        <div className="success-rate-banner" style={{ borderColor: successStyle.color }}>
          <div className="success-rate-value" style={{ color: successStyle.color }}>
            {summary.successRate.toFixed(0)}%
          </div>
          <div className="success-rate-label">
            <strong>Success Rate</strong>
            <span>{successStyle.label}</span>
          </div>
          <div className="success-rate-info">
            Probability of not running out of money during retirement
          </div>
        </div>
      )}

      <Chart options={options} series={series} type="area" height={400} />

      {/* Summary Stats */}
      <div className="projection-summary">
        <div className="summary-header">
          <h4>
            Projection Summary
            ({summary.years} Years, {summary.simulations.toLocaleString()} Simulations)
          </h4>
        </div>
        <div className="summary-stats">
          <div className="summary-stat">
            <span className="stat-label">Starting Net Worth</span>
            <span className="stat-value">
              ${summary.startingNetWorth.toLocaleString('en-US', { minimumFractionDigits: 0 })}
            </span>
          </div>
          <div className="summary-stat optimistic">
            <span className="stat-label">Optimistic (90th)</span>
            <span className="stat-value">
              ${summary.finalP90.toLocaleString('en-US', { minimumFractionDigits: 0 })}
            </span>
          </div>
          <div className="summary-stat median">
            <span className="stat-label">Median (50th)</span>
            <span className="stat-value">
              ${summary.finalP50.toLocaleString('en-US', { minimumFractionDigits: 0 })}
            </span>
          </div>
          <div className="summary-stat conservative">
            <span className="stat-label">Conservative (10th)</span>
            <span className="stat-value">
              ${summary.finalP10.toLocaleString('en-US', { minimumFractionDigits: 0 })}
            </span>
          </div>
        </div>

        {/* Contribution/Withdrawal Summary */}
        {(summary.totalContributions > 0 || summary.totalWithdrawals > 0) && (
          <div className="summary-stats secondary">
            {summary.totalContributions > 0 && (
              <div className="summary-stat">
                <span className="stat-label">Avg Total Contributions</span>
                <span className="stat-value positive">
                  +${summary.totalContributions.toLocaleString('en-US', { minimumFractionDigits: 0 })}
                </span>
              </div>
            )}
            {summary.totalWithdrawals > 0 && (
              <div className="summary-stat">
                <span className="stat-label">Avg Total Withdrawals</span>
                <span className="stat-value negative">
                  -${summary.totalWithdrawals.toLocaleString('en-US', { minimumFractionDigits: 0 })}
                </span>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Milestones */}
      {milestones && milestones.length > 0 && (
        <div className="milestones-section">
          <h4>Key Milestones</h4>
          <div className="milestones-grid">
            {milestones.map((m, i) => (
              <div key={i} className="milestone-card">
                <div className="milestone-target">{m.description}</div>
                <div className="milestone-details">
                  <span className="milestone-year">~Year {m.medianYear}</span>
                  <span className="milestone-probability">
                    {m.probabilityPct.toFixed(0)}% likely
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Insights */}
      {insights && insights.length > 0 && (
        <div className="insights-section">
          <h4>Insights & Recommendations</h4>
          <div className="insights-list">
            {insights.map((insight, i) => (
              <div key={i} className={`insight-card ${insight.type}`}>
                <div className="insight-icon">
                  {insight.type === 'success' && '✓'}
                  {insight.type === 'info' && 'ℹ'}
                  {insight.type === 'warning' && '⚠'}
                  {insight.type === 'opportunity' && '💡'}
                </div>
                <div className="insight-content">
                  <strong>{insight.title}</strong>
                  <p>{insight.message}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

export default MonteCarloChart;
