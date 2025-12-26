import React from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function CashFlowChart({ data }) {
  const { getCashFlowOptions } = useChartConfigs();

  if (!data || !data.monthlyData) {
    return <p style={{ textAlign: 'center', padding: '40px', color: '#999' }}>No data available</p>;
  }

  const months = Object.keys(data.monthlyData);
  const netFlow = months.map(m => data.monthlyData[m].net);

  const series = [
    { name: 'Net Cash Flow', data: netFlow }
  ];

  const options = getCashFlowOptions(months);

  return (
    <Chart
      options={options}
      series={series}
      type="bar"
      height={350}
    />
  );
}

export default CashFlowChart;
