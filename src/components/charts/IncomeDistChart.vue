<template>
  <VueDataUi component="VueUiHorizontalBar" :dataset="chartData" :config="config" />
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

const { getIncomeDistConfig } = useChartConfigs();

const chartData = computed(() => {
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

  return Object.keys(ranges).map(range => ({
    name: range,
    value: ranges[range],
    color: COLORS.income
  }));
});

const config = computed(() => getIncomeDistConfig());
</script>
