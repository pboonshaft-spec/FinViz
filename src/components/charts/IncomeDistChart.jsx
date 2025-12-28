import React, { useMemo } from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function IncomeDistChart({ data }) {
  const { getIncomeDistOptions } = useChartConfigs();

  const chartData = useMemo(() => {
    if (!data || !data.transactions) {
      return null;
    }

    const incomeTransactions = data.transactions.filter(t => t.amount > 0);

    const ranges = {
      '$0-100': 0,
      '$100-500': 0,
      '$500-1000': 0,
      '$1000-5000': 0,
      '$5000+': 0
    };

    incomeTransactions.forEach(t => {
      if (t.amount <= 100) ranges['$0-100']++;
      else if (t.amount <= 500) ranges['$100-500']++;
      else if (t.amount <= 1000) ranges['$500-1000']++;
      else if (t.amount <= 5000) ranges['$1000-5000']++;
      else ranges['$5000+']++;
    });

    const series = [
      { name: 'Count', data: Object.values(ranges) }
    ];

    const options = getIncomeDistOptions(ranges);

    return { series, options };
  }, [data, getIncomeDistOptions]);

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

export default IncomeDistChart;
