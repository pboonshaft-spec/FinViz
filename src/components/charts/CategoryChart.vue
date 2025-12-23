<template>
  <div v-if="hasCategories">
    <VueDataUi component="VueUiDonut" :dataset="chartData" :config="config" />
  </div>
  <div v-else class="no-data">
    <p>No expense categories found</p>
  </div>
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

const { getCategoryConfig } = useChartConfigs();

const hasCategories = computed(() => {
  return props.data?.categories && Object.keys(props.data.categories).length > 0;
});

const chartData = computed(() => {
  if (!props.data?.categories) return [];

  const categories = Object.keys(props.data.categories).slice(0, 10);

  return categories.map((cat, idx) => ({
    name: cat,
    value: props.data.categories[cat],
    color: COLORS.categoryPalette[idx % COLORS.categoryPalette.length]
  }));
});

const config = computed(() => getCategoryConfig());
</script>

<style scoped>
.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
