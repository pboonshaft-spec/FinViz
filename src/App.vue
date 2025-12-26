<template>
  <div class="container">
    <div class="header">
      <h1>💰 Financial Analytics Dashboard</h1>
      <p>Upload your financial CSV files to visualize your data</p>
    </div>

    <FileUpload @files-parsed="handleFilesParsed" />

    <div v-if="processedData" class="stats-grid">
      <StatCard
        label="Total Income"
        :value="processedData.totals.income"
        change="Revenue"
        changeType="positive"
      />
      <StatCard
        label="Total Expenses"
        :value="processedData.totals.expenses"
        change="Spending"
        changeType="negative"
      />
      <StatCard
        label="Net Balance"
        :value="processedData.totals.balance"
        :change="processedData.totals.balance >= 0 ? 'Surplus' : 'Deficit'"
        :changeType="processedData.totals.balance >= 0 ? 'positive' : 'negative'"
      />
      <StatCard
        label="Savings Rate"
        :value="savingsRate"
        :change="savingsRate >= 20 ? 'Great!' : 'Can improve'"
        :changeType="savingsRate >= 20 ? 'positive' : 'negative'"
        suffix="%"
      />
    </div>

    <div v-if="processedData" class="charts-grid">
      <ChartCard title="Income vs Expenses Over Time" :fullWidth="true">
        <TimeSeriesChart :data="processedData" />
      </ChartCard>

      <ChartCard title="Spending by Category">
        <CategoryChart :data="processedData" />
      </ChartCard>

      <ChartCard title="Monthly Cash Flow">
        <CashFlowChart :data="processedData" />
      </ChartCard>

      <ChartCard title="Daily Spending Trend">
        <TrendChart :data="processedData" />
      </ChartCard>

      <ChartCard title="Income Distribution">
        <IncomeDistChart :data="processedData" />
      </ChartCard>

      <ChartCard title="Cash Flow Sankey" :fullWidth="true">
        <SankeyChart :data="processedData" />
      </ChartCard>
    </div>

    <div v-if="isLoading" class="loading">Processing your data...</div>
  </div>
</template>

<script setup>
import { ref, computed } from 'vue';
import FileUpload from './components/FileUpload.vue';
import StatCard from './components/StatCard.vue';
import ChartCard from './components/ChartCard.vue';
import TimeSeriesChart from './components/charts/TimeSeriesChart.vue';
import CategoryChart from './components/charts/CategoryChart.vue';
import CashFlowChart from './components/charts/CashFlowChart.vue';
import TrendChart from './components/charts/TrendChart.vue';
import IncomeDistChart from './components/charts/IncomeDistChart.vue';
import SankeyChart from './components/charts/SankeyChart.vue';
import { useDataProcessor } from './composables/useDataProcessor.js';

const processedData = ref(null);
const isLoading = ref(false);

const savingsRate = computed(() => {
  if (!processedData.value || processedData.value.totals.income === 0) return 0;
  return ((processedData.value.totals.balance / processedData.value.totals.income) * 100).toFixed(1);
});

const handleFilesParsed = async (filesData) => {
  isLoading.value = true;

  const processor = useDataProcessor();
  const result = processor.processData(filesData);

  processedData.value = result.data;

  isLoading.value = false;
};
</script>
