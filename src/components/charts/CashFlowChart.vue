<template>
  <div v-if="hasData">
    <VueDataUi component="VueUiStackbar" :dataset="chartData" :config="config" />
  </div>
  <div v-else class="no-data">
    <p>No cash flow data available</p>
  </div>
</template>

<script setup>
import { computed } from 'vue';
import { VueDataUi } from 'vue-data-ui';
import 'vue-data-ui/style.css';
import { useChartConfigs, COLORS } from '../../composables/useChartConfigs.js';

const props = defineProps({
  data: {
    type: Object,
    required: true
  }
});

const { getCashFlowConfig } = useChartConfigs();

const hasData = computed(() => {
  return props.data?.monthlyData && Object.keys(props.data.monthlyData).length > 0;
});

const chartData = computed(() => {
  if (!props.data?.monthlyData) {
    return [];
  }

  const months = Object.keys(props.data.monthlyData);

  return months.map(month => ({
    name: month,
    series: [
      props.data.monthlyData[month].income || 0,
      props.data.monthlyData[month].expenses || 0
    ]
  }));
});

const config = computed(() => {
  const months = Object.keys(props.data?.monthlyData || {});
  return getCashFlowConfig(months);
});
</script>

<style scoped>
.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
