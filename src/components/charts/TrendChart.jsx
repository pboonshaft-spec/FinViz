import React from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function TrendChart({ data }) {
  const { getTrendOptions } = useChartConfigs();

  if (!data || !data.dailyData || Object.keys(data.dailyData).length === 0) {
    return <p style={{ textAlign: 'center', padding: '40px', color: '#999' }}>No daily data available</p>;
  }

  const days = Object.keys(data.dailyData).sort();
  const amounts = days.map(d => data.dailyData[d]);

  const series = [
    { name: 'Daily Net', data: amounts }
  ];

  const options = getTrendOptions(days);

  return (
    <Chart
      options={options}
      series={series}
      type="line"
      height={350}
    />
  );
}

export default TrendChart;
