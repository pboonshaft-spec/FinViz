import React, { useMemo } from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function CashFlowChart({ data }) {
  const { getCashFlowOptions } = useChartConfigs();

  const chartData = useMemo(() => {
    if (!data || !data.monthlyData || Object.keys(data.monthlyData).length === 0) {
      return null;
    }

    // Sort months chronologically (keys are in YYYY-MM format)
    const months = Object.keys(data.monthlyData).sort();

    // Format month labels for display (YYYY-MM -> "MMM YYYY")
    const monthLabels = months.map(m => {
      const [year, month] = m.split('-');
      const date = new Date(year, parseInt(month) - 1);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
    });

    const netFlow = months.map(m => data.monthlyData[m].net);

    const series = [
      { name: 'Net Cash Flow', data: netFlow }
    ];

    const options = getCashFlowOptions(monthLabels);

    return { series, options };
  }, [data, getCashFlowOptions]);

  if (!chartData) {
    return <p style={{ textAlign: 'center', padding: '40px', color: '#999' }}>No data available</p>;
  }

  return (
    <Chart
      options={chartData.options}
      series={chartData.series}
      type="bar"
      height={350}
    />
  );
}

export default CashFlowChart;
