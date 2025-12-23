<template>
  <VueDataUi component="VueUiXy" :dataset="chartData" :config="config" />
</template>

<script setup>
import { computed } from 'vue';
import { VueDataUi } from 'vue-data-ui';
import { useChartConfigs, COLORS } from '../../composables/useChartConfigs.js';

const props = defineProps({
  data: {
    type: Object,
    required: true
  }
});

const { getTimeSeriesConfig } = useChartConfigs();

const chartData = computed(() => {
  if (!props.data?.monthlyData) {
    return {
      categories: [],
      series: []
    };
  }

  const months = Object.keys(props.data.monthlyData);
  const incomeData = months.map(m => props.data.monthlyData[m].income);
  const expensesData = months.map(m => props.data.monthlyData[m].expenses);

  return {
    categories: months,
    series: [
      {
        name: 'Income',
        values: incomeData,
        type: 'bar',
        color: COLORS.income
      },
      {
        name: 'Expenses',
        values: expensesData,
        type: 'bar',
        color: COLORS.expenses
      }
    ]
  };
});

const config = computed(() => getTimeSeriesConfig());
</script>
