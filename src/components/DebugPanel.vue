<template>
  <div class="debug-section">
    <h3>📋 Data Detection Log</h3>
    <div class="debug-content">
      <div v-for="(log, index) in logs" :key="index">{{ log }}</div>
    </div>
  </div>
</template>

<script setup>
defineProps({
  logs: {
    type: Array,
    required: true
  }
});
</script>

<style scoped>
.debug-section {
  background: white;
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
}

.debug-section h3 {
  color: #667eea;
  margin-bottom: 10px;
}

.debug-content {
  background: #f5f5f5;
  padding: 15px;
  border-radius: 8px;
  font-family: monospace;
  font-size: 0.9rem;
  max-height: 200px;
  overflow-y: auto;
}
</style>
