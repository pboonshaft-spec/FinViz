export const COLORS = {
  income: '#00d4aa',
  expenses: '#ff6b6b',
  primary: '#6366f1',
  secondary: '#8b5cf6',
  categoryPalette: [
    '#6366f1', '#8b5cf6', '#ec4899', '#3b82f6', '#00d4aa',
    '#f59e0b', '#ef4444', '#14b8a6', '#a855f7', '#f97316'
  ]
};

const baseChartOptions = {
  chart: {
    fontFamily: "'Inter', -apple-system, BlinkMacSystemFont, sans-serif",
    toolbar: {
      show: true,
      tools: {
        download: true,
        selection: false,
        zoom: true,
        zoomin: true,
        zoomout: true,
        pan: false,
        reset: true
      }
    },
    background: 'transparent'
  },
  theme: {
    mode: 'dark'
  },
  grid: {
    borderColor: '#2a2a2a',
    strokeDashArray: 4
  },
  tooltip: {
    theme: 'dark',
    style: {
      fontSize: '12px'
    }
  },
  legend: {
    labels: {
      colors: '#888'
    }
  },
  xaxis: {
    labels: {
      style: {
        colors: '#666'
      }
    },
    axisBorder: {
      color: '#2a2a2a'
    },
    axisTicks: {
      color: '#2a2a2a'
    }
  },
  yaxis: {
    labels: {
      style: {
        colors: '#666'
      }
    }
  }
};

export function useChartConfigs() {
  const getTimeSeriesOptions = (months) => ({
    ...baseChartOptions,
    chart: {
      ...baseChartOptions.chart,
      type: 'area',
      height: 350
    },
    colors: [COLORS.income, COLORS.expenses],
    stroke: {
      curve: 'smooth',
      width: 2
    },
    fill: {
      type: 'gradient',
      gradient: {
        opacityFrom: 0.4,
        opacityTo: 0.05
      }
    },
    xaxis: {
      ...baseChartOptions.xaxis,
      categories: months,
      labels: {
        ...baseChartOptions.xaxis.labels,
        rotate: -45
      }
    },
    yaxis: {
      ...baseChartOptions.yaxis,
      labels: {
        ...baseChartOptions.yaxis.labels,
        formatter: (val) => `$${val.toFixed(0)}`
      }
    },
    tooltip: {
      ...baseChartOptions.tooltip,
      y: {
        formatter: (val) => `$${val.toFixed(2)}`
      }
    },
    legend: {
      ...baseChartOptions.legend,
      position: 'top'
    },
    dataLabels: { enabled: false }
  });

  const getCashFlowOptions = (months) => ({
    ...baseChartOptions,
    chart: {
      ...baseChartOptions.chart,
      type: 'bar',
      height: 350
    },
    colors: [COLORS.primary],
    plotOptions: {
      bar: {
        distributed: true,
        borderRadius: 4,
        colors: {
          ranges: [{
            from: -100000,
            to: 0,
            color: COLORS.expenses
          }, {
            from: 0.01,
            to: 100000,
            color: COLORS.income
          }]
        }
      }
    },
    dataLabels: { enabled: false },
    xaxis: {
      ...baseChartOptions.xaxis,
      categories: months,
      labels: {
        ...baseChartOptions.xaxis.labels,
        rotate: -45
      }
    },
    yaxis: {
      ...baseChartOptions.yaxis,
      labels: {
        ...baseChartOptions.yaxis.labels,
        formatter: (val) => `$${val.toFixed(0)}`
      }
    },
    tooltip: {
      ...baseChartOptions.tooltip,
      y: {
        formatter: (val) => `$${val.toFixed(2)}`
      }
    },
    legend: {
      show: false
    }
  });

  const getCategoryOptions = () => ({
    ...baseChartOptions,
    chart: {
      ...baseChartOptions.chart,
      type: 'donut',
      height: 380
    },
    colors: COLORS.categoryPalette,
    stroke: {
      width: 0
    },
    legend: {
      ...baseChartOptions.legend,
      position: 'bottom',
      fontSize: '12px'
    },
    dataLabels: {
      enabled: true,
      formatter: (val) => `${val.toFixed(1)}%`,
      style: {
        fontSize: '11px',
        fontWeight: '500'
      },
      dropShadow: {
        enabled: false
      }
    },
    plotOptions: {
      pie: {
        donut: {
          size: '75%',
          labels: {
            show: true,
            name: {
              color: '#fff'
            },
            value: {
              color: '#888',
              formatter: (val) => `$${parseFloat(val).toFixed(0)}`
            },
            total: {
              show: true,
              label: 'Total',
              color: '#888',
              formatter: (w) => {
                const total = w.globals.seriesTotals.reduce((a, b) => a + b, 0);
                return `$${total.toFixed(0)}`;
              }
            }
          }
        }
      }
    },
    tooltip: {
      ...baseChartOptions.tooltip,
      y: {
        formatter: (val) => `$${val.toFixed(2)}`
      }
    }
  });

  const getTrendOptions = (days) => ({
    ...baseChartOptions,
    chart: {
      ...baseChartOptions.chart,
      type: 'line',
      height: 350,
      zoom: { enabled: true }
    },
    colors: [COLORS.secondary],
    stroke: {
      curve: 'smooth',
      width: 3
    },
    xaxis: {
      ...baseChartOptions.xaxis,
      type: 'datetime',
      categories: days
    },
    yaxis: {
      ...baseChartOptions.yaxis,
      labels: {
        ...baseChartOptions.yaxis.labels,
        formatter: (val) => `$${val.toFixed(0)}`
      }
    },
    tooltip: {
      ...baseChartOptions.tooltip,
      x: {
        format: 'MMM dd, yyyy'
      },
      y: {
        formatter: (val) => `$${val.toFixed(2)}`
      }
    }
  });

  const getIncomeDistOptions = (ranges) => ({
    ...baseChartOptions,
    chart: {
      ...baseChartOptions.chart,
      type: 'bar',
      height: 350
    },
    colors: [COLORS.income],
    plotOptions: {
      bar: {
        horizontal: true,
        distributed: false,
        borderRadius: 4
      }
    },
    dataLabels: {
      enabled: true,
      style: {
        colors: ['#fff']
      }
    },
    xaxis: {
      ...baseChartOptions.xaxis,
      categories: Object.keys(ranges)
    },
    yaxis: {
      ...baseChartOptions.yaxis,
      title: {
        text: 'Transaction Range',
        style: { color: '#666' }
      }
    }
  });

  return {
    getTimeSeriesOptions,
    getCashFlowOptions,
    getCategoryOptions,
    getTrendOptions,
    getIncomeDistOptions,
    COLORS
  };
}
