import React from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function TimeSeriesChart({ data }) {
  const { getTimeSeriesOptions } = useChartConfigs();

  if (!data || !data.monthlyData) {
    return <p style={{ textAlign: 'center', padding: '40px', color: '#999' }}>No data available</p>;
  }

  const months = Object.keys(data.monthlyData);
  const income = months.map(m => data.monthlyData[m].income);
  const expenses = months.map(m => data.monthlyData[m].expenses);

  const series = [
    { name: 'Income', data: income },
    { name: 'Expenses', data: expenses }
  ];

  const options = getTimeSeriesOptions(months);

  return (
    <Chart
      options={options}
      series={series}
      type="area"
      height={350}
    />
  );
}

export default TimeSeriesChart;
