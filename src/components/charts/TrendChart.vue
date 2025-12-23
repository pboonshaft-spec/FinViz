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

const { getTrendConfig } = useChartConfigs();

const chartData = computed(() => {
  if (!props.data?.dailyData) {
    return {
      categories: [],
      series: []
    };
  }

  const days = Object.keys(props.data.dailyData).sort();
  const amounts = days.map(d => props.data.dailyData[d]);

  return {
    categories: days,
    series: [
      {
        name: 'Daily Net',
        values: amounts,
        type: 'line',
        color: COLORS.secondary,
        smooth: true
      }
    ]
  };
});

const config = computed(() => getTrendConfig());
</script>
