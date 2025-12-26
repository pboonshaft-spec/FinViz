<template>
  <div v-if="hasData" ref="sankeyContainer" class="sankey-container"></div>
  <div v-else class="no-data">
    <p>No transaction flow data available</p>
  </div>
</template>

<script setup>
import { computed, ref, onMounted, watch } from 'vue';
import ApexSankey from 'apexsankey';

const props = defineProps({
  data: {
    type: Object,
    required: true
  }
});

const sankeyContainer = ref(null);
let sankeyChart = null;

const hasData = computed(() => {
  return props.data?.transactions && props.data.transactions.length > 0;
});

const buildSankeyData = () => {
  if (!hasData.value) {
    return { nodes: [], edges: [] };
  }

  try {
    const transactions = props.data.transactions;

    // Aggregate income and expenses by category
    const incomeByCategory = new Map();
    const expensesByCategory = new Map();

    transactions.forEach(t => {
      if (t.amount > 0) {
        const category = t.category || t.description || 'Other Income';
        incomeByCategory.set(category, (incomeByCategory.get(category) || 0) + t.amount);
      } else {
        const category = t.category || 'Uncategorized';
        expensesByCategory.set(category, (expensesByCategory.get(category) || 0) + Math.abs(t.amount));
      }
    });

    const nodes = [];
    const edges = [];

    // Add income nodes
    incomeByCategory.forEach((amount, category) => {
      if (amount > 0) {
        const nodeId = `income_${category}`;
        nodes.push({
          id: nodeId,
          title: category
        });
      }
    });

    // Add expense nodes
    expensesByCategory.forEach((amount, category) => {
      if (amount > 0) {
        const nodeId = `expense_${category}`;
        nodes.push({
          id: nodeId,
          title: category
        });
      }
    });

    // Add a central "Cash Flow" node
    nodes.push({
      id: 'cashflow',
      title: 'Cash Flow'
    });

    // Create edges from income to cash flow
    incomeByCategory.forEach((amount, category) => {
      if (amount > 0) {
        edges.push({
          source: `income_${category}`,
          target: 'cashflow',
          value: amount
        });
      }
    });

    // Create edges from cash flow to expenses
    expensesByCategory.forEach((amount, category) => {
      if (amount > 0) {
        edges.push({
          source: 'cashflow',
          target: `expense_${category}`,
          value: amount
        });
      }
    });

    console.log('Sankey data:', { nodes, edges });
    return { nodes, edges };
  } catch (error) {
    console.error('Error building sankey chart data:', error);
    return { nodes: [], edges: [] };
  }
};

const renderChart = () => {
  if (!sankeyContainer.value || !hasData.value) return;

  const data = buildSankeyData();

  if (data.nodes.length === 0) return;

  const options = {
    width: sankeyContainer.value.offsetWidth || 800,
    height: 600,
    canvasStyle: 'background: #FFFFFF;',
    spacing: 150,
    nodeWidth: 20,
  };

  try {
    if (sankeyChart) {
      sankeyContainer.value.innerHTML = '';
    }

    sankeyChart = new ApexSankey(sankeyContainer.value, options);
    sankeyChart.render(data);
  } catch (error) {
    console.error('Error rendering sankey chart:', error);
  }
};

onMounted(() => {
  renderChart();
});

watch(() => props.data, () => {
  renderChart();
}, { deep: true });
</script>

<style scoped>
.sankey-container {
  width: 100%;
  min-height: 600px;
  display: flex;
  justify-content: center;
  align-items: center;
}

.no-data {
  text-align: center;
  padding: 40px;
  color: #999;
}
</style>
