<template>
  <div class="chart-card" :class="{ 'full-width': fullWidth }">
    <div class="chart-title">{{ title }}</div>
    <slot></slot>
  </div>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    required: true
  },
  fullWidth: {
    type: Boolean,
    default: false
  }
});
</script>

<style scoped>
.chart-card {
  background: white;
  border-radius: 12px;
  padding: 25px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.chart-card.full-width {
  grid-column: 1 / -1;
}

.chart-title {
  font-size: 1.3rem;
  font-weight: 600;
  color: #333;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 2px solid #f0f0f0;
}
</style>
