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

const { getCashFlowConfig } = useChartConfigs();

const chartData = computed(() => {
  if (!props.data?.monthlyData) {
    return {
      categories: [],
      series: []
    };
  }

  const months = Object.keys(props.data.monthlyData);
  const netFlowData = months.map(m => ({
    value: props.data.monthlyData[m].net,
    color: props.data.monthlyData[m].net >= 0 ? COLORS.income : COLORS.expenses
  }));

  return {
    categories: months,
    series: [
      {
        name: 'Net Cash Flow',
        values: netFlowData.map(d => d.value),
        type: 'bar',
        datapoints: netFlowData
      }
    ]
  };
});

const config = computed(() => getCashFlowConfig());
</script>
