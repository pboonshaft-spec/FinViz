import React from 'react';
import Chart from 'react-apexcharts';

function MonteCarloChart({ projection }) {
  if (!projection || !projection.projections) {
    return null;
  }

  const categories = projection.projections.map(p => `Year ${p.year}`);

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
          return `$${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}`;
        },
      },
    },
    dataLabels: {
      enabled: false,
    },
  };

  const series = [
    {
      name: '90th Percentile (Optimistic)',
      data: projection.projections.map(p => p.p90),
    },
    {
      name: '50th Percentile (Median)',
      data: projection.projections.map(p => p.p50),
    },
    {
      name: '10th Percentile (Conservative)',
      data: projection.projections.map(p => p.p10),
    },
  ];

  return (
    <div className="monte-carlo-chart">
      <Chart options={options} series={series} type="area" height={400} />

      <div className="projection-summary">
        <div className="summary-header">
          <h4>Projection Summary ({projection.summary.years} Years, {projection.summary.simulations.toLocaleString()} Simulations)</h4>
        </div>
        <div className="summary-stats">
          <div className="summary-stat">
            <span className="stat-label">Starting Net Worth</span>
            <span className="stat-value">
              ${projection.summary.startingNetWorth.toLocaleString('en-US', { minimumFractionDigits: 2 })}
            </span>
          </div>
          <div className="summary-stat optimistic">
            <span className="stat-label">Optimistic (90th)</span>
            <span className="stat-value">
              ${projection.summary.finalP90.toLocaleString('en-US', { minimumFractionDigits: 2 })}
            </span>
          </div>
          <div className="summary-stat median">
            <span className="stat-label">Median (50th)</span>
            <span className="stat-value">
              ${projection.summary.finalP50.toLocaleString('en-US', { minimumFractionDigits: 2 })}
            </span>
          </div>
          <div className="summary-stat conservative">
            <span className="stat-label">Conservative (10th)</span>
            <span className="stat-value">
              ${projection.summary.finalP10.toLocaleString('en-US', { minimumFractionDigits: 2 })}
            </span>
          </div>
        </div>
      </div>
    </div>
  );
}

export default MonteCarloChart;
