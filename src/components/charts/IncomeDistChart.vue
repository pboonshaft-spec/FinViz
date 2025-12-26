<template>
  <div v-if="hasIncomeData">
    <VueUiXy :dataset="chartData" :config="config" />
  </div>
  <div v-else class="no-data">
    <p>No income transactions found</p>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { VueUiXy } from 'vue-data-ui';
import 'vue-data-ui/style.css';
import { COLORS } from '../../composables/useChartConfigs.js';

const props = defineProps({
  data: {
    type: Object,
    required: true
  }
});

const hasIncomeData = computed(() => {
  if (!props.data?.transactions) return false;
  const incomeTransactions = props.data.transactions.filter(t => t.amount > 0);
  if (incomeTransactions.length === 0) return false;

  // Also verify that at least one range will have data after filtering
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

  // Only render if we have at least one non-zero range
  const hasData = Object.values(ranges).some(count => count > 0);
  return hasData;
});

const chartData = computed(() => {
  if (!hasIncomeData.value) {
    return [];
  }

  const incomeTransactions = props.data.transactions.filter(t => t.amount > 0);
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

  const allRanges = Object.keys(ranges);
  const values = allRanges.map(range => ranges[range]);

  return [
    {
      name: 'Income Count',
      series: values,
      color: COLORS.income,
      type: 'bar',
      useArea: false,
      dataLabels: true
    }
  ];
});

const config = computed(() => {
  const incomeTransactions = props.data?.transactions?.filter(t => t.amount > 0) || [];
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

  const labels = Object.keys(ranges);

  return {
    chart: {
      fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
      backgroundColor: '#FFFFFF',
      color: '#1A1A1A',
      height: 400,
      padding: {
        top: 20,
        right: 20,
        bottom: 60,
        left: 80
      },
      grid: {
        show: true,
        stroke: '#e5e7eb',
        labels: {
          show: true,
          color: '#1A1A1A',
          fontSize: 14,
          xAxisLabels: {
            values: labels,
            show: true,
            fontSize: 14,
            color: '#1A1A1A'
          },
          yAxis: {
            show: true,
            useNiceScale: true
          }
        }
      },
      legend: {
        show: true,
        fontSize: 14,
        color: '#1A1A1A'
      },
      title: {
        show: false
      },
      tooltip: {
        show: true,
        backgroundColor: '#FFFFFF',
        color: '#1A1A1A',
        fontSize: 14,
        borderRadius: 4,
        borderColor: '#e5e7eb',
        borderWidth: 1
      }
    },
    bar: {
      borderRadius: 4,
      useGradient: true,
      labels: {
        show: true,
        offsetY: -8,
        color: '#1A1A1A'
      }
    },
    userOptions: {
      show: false
    }
  };
});
</script>

<style scoped>
.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
