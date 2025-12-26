<template>
  <div v-if="hasData">
    <VueDataUi component="VueUiXy" :dataset="chartData" :config="config" />
  </div>
  <div v-else class="no-data">
    <p>No monthly data available</p>
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

const { getTimeSeriesConfig } = useChartConfigs();

const hasData = computed(() => {
  return props.data?.monthlyData && Object.keys(props.data.monthlyData).length > 0;
});

const chartData = computed(() => {
  if (!props.data?.monthlyData) {
    return [];
  }

  const months = Object.keys(props.data.monthlyData);

  return [
    {
      name: 'Income',
      series: months.map(month => props.data.monthlyData[month].income || 0),
      color: COLORS.income,
      type: 'line',
      useArea: true,
      dataLabels: false,
      smooth: true
    },
    {
      name: 'Expenses',
      series: months.map(month => props.data.monthlyData[month].expenses || 0),
      color: COLORS.expenses,
      type: 'line',
      useArea: true,
      dataLabels: false,
      smooth: true
    }
  ];
});

const config = computed(() => {
  const months = Object.keys(props.data?.monthlyData || {});
  return getTimeSeriesConfig(months);
});
</script>

<style scoped>
.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
