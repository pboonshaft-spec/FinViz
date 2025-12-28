import React, { useMemo } from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function TimeSeriesChart({ data }) {
  const { getTimeSeriesOptions } = useChartConfigs();

  const chartData = useMemo(() => {
    if (!data || !data.monthlyData || Object.keys(data.monthlyData).length === 0) {
      return null;
    }

    console.log('TimeSeriesChart: Processing data', data.monthlyData);

    // Sort months chronologically (keys are in YYYY-MM format)
    const months = Object.keys(data.monthlyData).sort();
    console.log('TimeSeriesChart: Months', months);

    // Format month labels for display (YYYY-MM -> "MMM YYYY")
    const monthLabels = months.map(m => {
      const [year, month] = m.split('-');
      const date = new Date(year, parseInt(month) - 1);
      return date.toLocaleDateString('en-US', { year: 'numeric', month: 'short' });
    });

    const income = months.map(m => data.monthlyData[m].income);
    const expenses = months.map(m => data.monthlyData[m].expenses);

    console.log('TimeSeriesChart: Income data', income);
    console.log('TimeSeriesChart: Expenses data', expenses);

    const series = [
      { name: 'Income', data: income },
      { name: 'Expenses', data: expenses }
    ];

    const options = getTimeSeriesOptions(monthLabels);

    return { series, options };
  }, [data, getTimeSeriesOptions]);

  if (!chartData) {
    return <p style={{ textAlign: 'center', padding: '40px', color: '#999' }}>No data available</p>;
  }

  return (
    <Chart
      options={chartData.options}
      series={chartData.series}
      type="area"
      height={350}
    />
  );
}

export default TimeSeriesChart;
