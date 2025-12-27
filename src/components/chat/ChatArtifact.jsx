import React from 'react';
import Chart from 'react-apexcharts';

// Render A2UI artifacts (charts, tables, metric cards)
export default function ChatArtifact({ artifact }) {
  if (!artifact || !artifact.type) return null;

  switch (artifact.type) {
    case 'chart':
      return <ChartArtifact artifact={artifact} />;
    case 'table':
      return <TableArtifact artifact={artifact} />;
    case 'metric_card':
      return <MetricCardArtifact artifact={artifact} />;
    default:
      return null;
  }
}

// Chart artifact using ApexCharts
function ChartArtifact({ artifact }) {
  const { chart_type, title, data, colors } = artifact;

  // Convert data to ApexCharts format based on chart type
  let options = {
    chart: {
      type: chart_type === 'donut' ? 'donut' : chart_type,
      background: 'transparent',
      toolbar: { show: false },
    },
    title: {
      text: title,
      style: {
        color: 'var(--text-primary)',
        fontSize: '14px',
        fontWeight: 600,
      },
    },
    theme: {
      mode: 'dark',
    },
    colors: colors || ['#6366f1', '#00d4aa', '#ff6b6b', '#feca57', '#54a0ff'],
  };

  let series = [];

  if (chart_type === 'pie' || chart_type === 'donut') {
    // Pie/donut expects { labels: [], series: [] }
    options.labels = data.map(d => d.label || d.name);
    series = data.map(d => d.value);
    options.legend = {
      position: 'bottom',
      labels: { colors: 'var(--text-secondary)' },
    };
  } else if (chart_type === 'bar') {
    // Bar chart
    options.xaxis = {
      categories: data.map(d => d.label || d.name),
      labels: { style: { colors: 'var(--text-secondary)' } },
    };
    options.yaxis = {
      labels: {
        style: { colors: 'var(--text-secondary)' },
        formatter: (val) => formatCurrency(val),
      },
    };
    options.plotOptions = {
      bar: { borderRadius: 4, horizontal: false },
    };
    series = [{ name: 'Value', data: data.map(d => d.value) }];
  } else if (chart_type === 'line' || chart_type === 'area') {
    // Line/area chart
    options.xaxis = {
      categories: data.map(d => d.label || d.date || d.name),
      labels: { style: { colors: 'var(--text-secondary)' } },
    };
    options.yaxis = {
      labels: {
        style: { colors: 'var(--text-secondary)' },
        formatter: (val) => formatCurrency(val),
      },
    };
    options.stroke = { curve: 'smooth', width: 2 };
    if (chart_type === 'area') {
      options.fill = {
        type: 'gradient',
        gradient: { opacityFrom: 0.4, opacityTo: 0.1 },
      };
    }
    series = [{ name: 'Value', data: data.map(d => d.value) }];
  }

  return (
    <div className="chat-artifact chat-artifact-chart">
      <Chart
        options={options}
        series={series}
        type={chart_type === 'donut' ? 'donut' : chart_type}
        height={250}
      />
    </div>
  );
}

// Table artifact
function TableArtifact({ artifact }) {
  const { title, headers, rows } = artifact;

  return (
    <div className="chat-artifact chat-artifact-table">
      {title && <h4 className="chat-artifact-title">{title}</h4>}
      <div className="chat-table-wrapper">
        <table className="chat-table">
          <thead>
            <tr>
              {headers.map((header, i) => (
                <th key={i}>{header}</th>
              ))}
            </tr>
          </thead>
          <tbody>
            {rows.map((row, i) => (
              <tr key={i}>
                {(Array.isArray(row) ? row : Object.values(row)).map((cell, j) => (
                  <td key={j}>{formatCell(cell)}</td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

// Metric card artifact
function MetricCardArtifact({ artifact }) {
  const { label, value, change, trend } = artifact;

  const trendClass = trend === 'up' ? 'trend-up' : trend === 'down' ? 'trend-down' : '';

  return (
    <div className="chat-artifact chat-artifact-metric">
      <div className="chat-metric-label">{label}</div>
      <div className="chat-metric-value">{value}</div>
      {change && (
        <div className={`chat-metric-change ${trendClass}`}>
          {trend === 'up' && (
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polyline points="23 6 13.5 15.5 8.5 10.5 1 18" />
              <polyline points="17 6 23 6 23 12" />
            </svg>
          )}
          {trend === 'down' && (
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polyline points="23 18 13.5 8.5 8.5 13.5 1 6" />
              <polyline points="17 18 23 18 23 12" />
            </svg>
          )}
          {change}
        </div>
      )}
    </div>
  );
}

// Helper to format currency values
function formatCurrency(value) {
  if (typeof value !== 'number') return value;
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(value);
}

// Helper to format table cells
function formatCell(value) {
  if (value === null || value === undefined) return '-';
  if (typeof value === 'number') {
    // Check if it looks like currency (large numbers)
    if (Math.abs(value) >= 100) {
      return formatCurrency(value);
    }
    // Check if it looks like a percentage
    if (value >= -1 && value <= 1 && value !== 0) {
      return `${(value * 100).toFixed(1)}%`;
    }
    return value.toLocaleString();
  }
  return String(value);
}
