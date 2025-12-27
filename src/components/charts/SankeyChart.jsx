import React, { useRef, useEffect } from 'react';
import ApexSankey from 'apexsankey';

function SankeyChart({ data }) {
  const containerRef = useRef(null);
  const chartRef = useRef(null);

  useEffect(() => {
    if (!containerRef.current || !data?.transactions?.length) return;

    const sankeyData = buildSankeyData(data.transactions);

    if (sankeyData.nodes.length === 0 || sankeyData.edges.length === 0) {
      return;
    }

    const options = {
      width: containerRef.current.offsetWidth || 800,
      height: 500,
      canvasStyle: 'background: transparent;',
      spacing: 150,
      nodeWidth: 20,
      fontFamily: "'Inter', sans-serif",
      fontColor: '#888',
      enableTooltip: true,
      enableExport: false
    };

    // Clear previous chart
    containerRef.current.innerHTML = '';

    try {
      chartRef.current = new ApexSankey(containerRef.current, options);
      chartRef.current.render(sankeyData);

      // Hide export button if it exists (fallback)
      const exportBtn = containerRef.current.querySelector('button');
      if (exportBtn) {
        exportBtn.style.display = 'none';
      }
    } catch (error) {
      console.error('Error rendering Sankey chart:', error);
    }

    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = '';
      }
    };
  }, [data]);

  const buildSankeyData = (transactions) => {
    const nodes = [];
    const edges = [];

    // Categories that are typically expenses - if these show as income, they're refunds
    const expenseCategories = new Set([
      'TRAVEL', 'FOOD_AND_DRINK', 'FOOD', 'SHOPPING', 'ENTERTAINMENT',
      'TRANSPORTATION', 'UTILITIES', 'RENT', 'SUBSCRIPTION', 'HEALTHCARE',
      'PERSONAL_CARE', 'GENERAL_MERCHANDISE', 'GROCERIES', 'GAS', 'AUTOMOTIVE',
      'TRANSFER_OUT' // Outbound transfers are expenses
    ]);

    // Categories that are income even if they appear as "transfers"
    const incomeCategories = new Set([
      'INCOME', 'INCOME_WAGES', 'INCOME_DIVIDENDS', 'INCOME_INTEREST',
      'TRANSFER_IN', 'TRANSFER_IN_ACCOUNT_TRANSFER', 'TRANSFER_IN_DEPOSIT'
    ]);

    // Aggregate income and expenses by category
    const incomeByCategory = new Map();
    const expensesByCategory = new Map();

    transactions.forEach(t => {
      const upperCategory = (t.category || '').toUpperCase();
      const upperSubcategory = (t.subcategory || '').toUpperCase();

      // Check if this is income based on category (regardless of amount sign)
      const isIncomeCategory = incomeCategories.has(upperCategory) ||
                               incomeCategories.has(upperSubcategory) ||
                               upperCategory.startsWith('INCOME') ||
                               upperCategory === 'TRANSFER_IN' ||
                               upperSubcategory.startsWith('INCOME');

      // Check if this is a transfer that should be excluded or categorized specially
      const isTransferOut = upperCategory === 'TRANSFER_OUT' ||
                            upperSubcategory.includes('TRANSFER_OUT');
      const isTransferIn = upperCategory === 'TRANSFER_IN' ||
                           upperSubcategory.includes('TRANSFER_IN');

      if (t.amount > 0 || isIncomeCategory) {
        // This is income
        if (t.amount <= 0 && !isIncomeCategory) {
          // Actually an expense, skip
          const category = t.category || 'Uncategorized';
          expensesByCategory.set(category, (expensesByCategory.get(category) || 0) + Math.abs(t.amount));
          return;
        }

        const isRefund = expenseCategories.has(upperCategory) && !isIncomeCategory;
        let category;

        if (isRefund) {
          category = 'Refunds';
        } else if (isIncomeCategory || upperCategory.startsWith('INCOME')) {
          category = 'Paycheck/Income';
        } else if (isTransferIn) {
          category = 'Transfers In';
        } else {
          category = t.category || t.description || 'Other Income';
        }

        const amount = Math.abs(t.amount);
        incomeByCategory.set(category, (incomeByCategory.get(category) || 0) + amount);
      } else {
        // This is an expense
        let category = t.category || 'Uncategorized';

        // Rename transfer out to be clearer
        if (isTransferOut || category.toUpperCase() === 'TRANSFER') {
          category = 'Transfers Out';
        }

        expensesByCategory.set(category, (expensesByCategory.get(category) || 0) + Math.abs(t.amount));
      }
    });

    // Color palette
    const incomeColors = ['#00d4aa', '#14b8a6', '#10b981', '#34d399'];
    const expenseColors = ['#ff6b6b', '#f87171', '#ef4444', '#dc2626', '#ec4899', '#f59e0b', '#8b5cf6', '#6366f1'];

    let incomeIdx = 0;
    let expenseIdx = 0;

    // Add income nodes
    incomeByCategory.forEach((amount, category) => {
      if (amount > 0) {
        nodes.push({
          id: `income_${category}`,
          title: category,
          color: incomeColors[incomeIdx % incomeColors.length]
        });
        incomeIdx++;
      }
    });

    // Add central "Cash Flow" node
    nodes.push({ id: 'cashflow', title: 'Cash Flow', color: '#6366f1' });

    // Add expense nodes
    expensesByCategory.forEach((amount, category) => {
      if (amount > 0) {
        nodes.push({
          id: `expense_${category}`,
          title: category,
          color: expenseColors[expenseIdx % expenseColors.length]
        });
        expenseIdx++;
      }
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

    return { nodes, edges };
  };

  if (!data?.transactions?.length) {
    return (
      <div style={{ textAlign: 'center', padding: '40px', color: '#666' }}>
        <p>No transaction flow data available</p>
      </div>
    );
  }

  return (
    <div
      ref={containerRef}
      style={{
        width: '100%',
        minHeight: '500px',
        background: 'transparent',
        borderRadius: '8px'
      }}
    />
  );
}

export default SankeyChart;
