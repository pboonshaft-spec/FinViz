import React, { useRef, useEffect, useState } from 'react';
import ApexSankey from 'apexsankey';

function SankeyChart({ data }) {
  const containerRef = useRef(null);
  const chartRef = useRef(null);
  const wrapperRef = useRef(null);
  const [showMenu, setShowMenu] = useState(false);

  // Close menu when clicking outside
  useEffect(() => {
    const handleClickOutside = (event) => {
      if (wrapperRef.current && !wrapperRef.current.contains(event.target)) {
        setShowMenu(false);
      }
    };

    if (showMenu) {
      document.addEventListener('mousedown', handleClickOutside);
      return () => document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [showMenu]);

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
      spacing: 50,
      nodeWidth: 50,
      fontFamily: "'Inter', sans-serif",
      fontColor: '#000000ff',
      enableTooltip: true,
      enableExport: false,
      tooltipId: 'sankey-tooltip-container',
      tooltipBorderColor: '#BCBCBC',
      tooltipBGColor: '#FFFFFF',
      tooltipTemplate: ({ source, target, value }) => {
        return `
          <div style='display:flex;align-items:center;gap:5px;padding:8px;'>
            <div style='width:15px;height:15px;background-color:${source.color};border-radius:2px;'></div>
            <div style='font-weight:500;'>${source.title}</div>
            <div style='color:#666;'>→</div>
            <div style='width:15px;height:15px;background-color:${target.color};border-radius:2px;'></div>
            <div style='font-weight:500;'>${target.title}</div>
            <div style='font-weight:700;'>: $${value.toLocaleString('en-US', { minimumFractionDigits: 2, maximumFractionDigits: 2 })}</div>
          </div>
        `;
      }
    };

    // Clear previous chart
    containerRef.current.innerHTML = '';

    try {
      chartRef.current = new ApexSankey(containerRef.current, options);
      chartRef.current.render(sankeyData);

      // Remove all default export buttons completely (not just hide)
      setTimeout(() => {
        const buttons = containerRef.current.querySelectorAll('button');
        buttons.forEach(btn => {
          btn.remove();
        });

        // Also remove any export-related elements
        const exportElements = containerRef.current.querySelectorAll('[class*="export"], [class*="menu"], [id*="export"]');
        exportElements.forEach(el => {
          el.remove();
        });
      }, 100);
    } catch (error) {
      console.error('Error rendering Sankey chart:', error);
    }

    return () => {
      if (containerRef.current) {
        containerRef.current.innerHTML = '';
      }
    };
  }, [data]);

  const handleDownload = (format) => {
    if (!containerRef.current) return;

    const canvas = containerRef.current.querySelector('canvas');
    if (!canvas) return;

    const link = document.createElement('a');
    link.download = `sankey-chart.${format}`;

    if (format === 'png') {
      link.href = canvas.toDataURL('image/png');
    } else if (format === 'svg') {
      // For SVG, we'll use PNG as fallback since canvas doesn't natively support SVG export
      link.href = canvas.toDataURL('image/png');
    }

    link.click();
    setShowMenu(false);
  };

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
    <div ref={wrapperRef} style={{ position: 'relative', width: '100%', minHeight: '500px' }}>
      {/* Custom Toolbar matching ApexCharts style */}
      <div style={{
        position: 'absolute',
        top: '0',
        right: '0',
        zIndex: 11,
        display: 'flex',
        alignItems: 'center',
        gap: '4px'
      }}>
        <div style={{ position: 'relative' }}>
          <button
            onClick={() => setShowMenu(!showMenu)}
            style={{
              background: 'transparent',
              border: 'none',
              cursor: 'pointer',
              padding: '8px',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              color: '#6E7079',
              fontSize: '14px',
              transition: 'all 0.15s ease'
            }}
            onMouseEnter={(e) => e.currentTarget.style.color = '#fff'}
            onMouseLeave={(e) => e.currentTarget.style.color = '#6E7079'}
            title="Menu"
          >
            {/* Hamburger icon (three lines) */}
            <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
              <path d="M3 18h18v-2H3v2zm0-5h18v-2H3v2zm0-7v2h18V6H3z"/>
            </svg>
          </button>

          {/* Dropdown Menu */}
          {showMenu && (
            <div style={{
              position: 'absolute',
              top: '100%',
              right: '0',
              marginTop: '4px',
              background: '#2b2b2f',
              border: '1px solid #3f3f46',
              borderRadius: '4px',
              boxShadow: '0 4px 6px rgba(0, 0, 0, 0.3)',
              minWidth: '140px',
              zIndex: 1000
            }}>
              <button
                onClick={() => handleDownload('png')}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  background: 'transparent',
                  border: 'none',
                  color: '#e5e7eb',
                  textAlign: 'left',
                  cursor: 'pointer',
                  fontSize: '13px',
                  transition: 'background 0.15s ease'
                }}
                onMouseEnter={(e) => e.currentTarget.style.background = '#3f3f46'}
                onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
              >
                Download PNG
              </button>
              <button
                onClick={() => handleDownload('svg')}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  background: 'transparent',
                  border: 'none',
                  color: '#e5e7eb',
                  textAlign: 'left',
                  cursor: 'pointer',
                  fontSize: '13px',
                  transition: 'background 0.15s ease',
                  borderTop: '1px solid #3f3f46'
                }}
                onMouseEnter={(e) => e.currentTarget.style.background = '#3f3f46'}
                onMouseLeave={(e) => e.currentTarget.style.background = 'transparent'}
              >
                Download SVG
              </button>
            </div>
          )}
        </div>
      </div>

      {/* Chart Container */}
      <div
        ref={containerRef}
        style={{
          width: '100%',
          minHeight: '500px',
          background: 'transparent',
          borderRadius: '8px'
        }}
      />
    </div>
  );
}

export default SankeyChart;
