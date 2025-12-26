<template>
  <div v-if="hasData">
    <VueDataUi component="VueUiXy" :dataset="chartData" :config="config" />
  </div>
  <div v-else class="no-data">
    <p>No daily data available</p>
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

const { getTrendConfig } = useChartConfigs();

const hasData = computed(() => {
  return props.data?.dailyData && Object.keys(props.data.dailyData).length > 0;
});

const chartData = computed(() => {
  if (!props.data?.dailyData) {
    return [];
  }

  const days = Object.keys(props.data.dailyData).sort();

  return [
    {
      name: 'Daily Net',
      series: days.map(day => props.data.dailyData[day] || 0),
      type: 'line',
      color: COLORS.secondary,
      useArea: false,
      dataLabels: false,
      smooth: true
    }
  ];
});

const config = computed(() => {
  const days = Object.keys(props.data?.dailyData || {}).sort();
  return getTrendConfig(days);
});
</script>

<style scoped>
.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
