import React from 'react';
import Chart from 'react-apexcharts';
import { useChartConfigs } from '../../hooks/useChartConfigs';

function CategoryChart({ data }) {
  const { getCategoryOptions } = useChartConfigs();

  if (!data || !data.categories || Object.keys(data.categories).length === 0) {
    return <p style={{ textAlign: 'center', padding: '40px', color: '#999' }}>No expense categories found</p>;
  }

  const categories = Object.keys(data.categories).slice(0, 10);
  const amounts = categories.map(c => data.categories[c]);
  const total = amounts.reduce((a, b) => a + b, 0);

  const options = {
    ...getCategoryOptions(),
    labels: categories,
    plotOptions: {
      pie: {
        donut: {
          labels: {
            show: true,
            total: {
              show: true,
              label: 'Total Expenses',
              formatter: () => `$${total.toFixed(0)}`
            }
          }
        }
      }
    }
  };

  return (
    <Chart
      options={options}
      series={amounts}
      type="donut"
      height={380}
    />
  );
}

export default CategoryChart;
