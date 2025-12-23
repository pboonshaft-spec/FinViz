<template>
  <div class="stat-card">
    <div class="stat-label">{{ label }}</div>
    <div class="stat-value" :style="{ color: valueColor }">
      {{ prefix }}{{ formattedValue }}{{ suffix }}
    </div>
    <div class="stat-change" :class="changeType">
      {{ changeIcon }} {{ change }}
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue';

const props = defineProps({
  label: {
    type: String,
    required: true
  },
  value: {
    type: Number,
    required: true
  },
  change: {
    type: String,
    default: ''
  },
  changeType: {
    type: String,
    default: 'positive',
    validator: (value) => ['positive', 'negative'].includes(value)
  },
  prefix: {
    type: String,
    default: '$'
  },
  suffix: {
    type: String,
    default: ''
  }
});

const formattedValue = computed(() => {
  if (props.suffix === '%') {
    return props.value;
  }
  return props.value.toLocaleString('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  });
});

const valueColor = computed(() => {
  if (props.label === 'Total Expenses') return '#ef4444';
  if (props.label === 'Net Balance') {
    return props.value >= 0 ? '#10b981' : '#ef4444';
  }
  return '#667eea';
});

const changeIcon = computed(() => {
  return props.changeType === 'positive' ? '↑' : '↓';
});
</script>

<style scoped>
.stat-card {
  background: white;
  border-radius: 12px;
  padding: 25px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.2);
  transition: transform 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-5px);
}

.stat-label {
  color: #666;
  font-size: 0.9rem;
  text-transform: uppercase;
  letter-spacing: 1px;
  margin-bottom: 10px;
}

.stat-value {
  font-size: 2rem;
  font-weight: bold;
}

.stat-change {
  font-size: 0.9rem;
  margin-top: 8px;
}

.stat-change.positive {
  color: #10b981;
}

.stat-change.negative {
  color: #ef4444;
}
</style>
