// Color scheme constants
export const COLORS = {
  income: '#10b981',
  expenses: '#ef4444',
  primary: '#667eea',
  secondary: '#764ba2',
  categoryPalette: [
    '#667eea', '#764ba2', '#f093fb', '#4facfe', '#43e97b',
    '#fa709a', '#fee140', '#30cfd0', '#a8edea', '#fed6e3'
  ]
};

export function useChartConfigs() {
  function getTimeSeriesConfig(labels = []) {
    return {
      chart: {
        backgroundColor: '#FFFFFF',
        color: '#1A1A1A',
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        height: 400,
        padding: {
          top: 20,
          right: 20,
          bottom: 60,
          left: 80
        },
        grid: {
          show: true,
          stroke: '#e5e7eb',
          strokeWidth: 1,
          labels: {
            show: true,
            color: '#1A1A1A',
            fontSize: 14,
            xAxisLabels: {
              values: labels,
              show: true,
              fontSize: 14,
              color: '#1A1A1A'
            },
            yAxis: {
              show: true,
              useNiceScale: true
            }
          }
        },
        legend: {
          show: true,
          fontSize: 14,
          color: '#1A1A1A'
        },
        title: {
          show: false
        },
        tooltip: {
          show: true,
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          fontSize: 14,
          borderRadius: 4,
          borderColor: '#e5e7eb',
          borderWidth: 1
        }
      },
      line: {
        strokeWidth: 2,
        smooth: true,
        area: {
          useGradient: true,
          opacity: 20
        },
        labels: {
          show: false
        }
      },
      userOptions: {
        show: false
      }
    };
  }

  function getCategoryConfig() {
    return {
      style: {
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        chart: {
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          height: 400,
          layout: {
            labels: {
              dataLabels: {
                show: true,
                fontSize: 14
              },
              value: {
                show: true,
                fontSize: 14,
                prefix: '$'
              },
              percentage: {
                show: true,
                fontSize: 12
              }
            },
            donut: {
              strokeWidth: 100,
              borderWidth: 2,
              useLabelSlot: true
            }
          },
          legend: {
            show: true,
            fontSize: 12
          }
        }
      },
      userOptions: {
        show: false
      }
    };
  }

  function getCashFlowConfig(labels = []) {
    return {
      chart: {
        backgroundColor: '#FFFFFF',
        color: '#1A1A1A',
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        height: 400,
        padding: {
          top: 20,
          right: 20,
          bottom: 60,
          left: 80
        },
        grid: {
          show: true,
          stroke: '#e5e7eb'
        },
        legend: {
          show: true,
          fontSize: 14
        }
      },
      labels: labels,
      series: [
        { name: 'Income', color: COLORS.income },
        { name: 'Expenses', color: COLORS.expenses }
      ],
      userOptions: {
        show: false
      }
    };
  }

  function getTrendConfig(labels = []) {
    return {
      chart: {
        backgroundColor: '#FFFFFF',
        color: '#1A1A1A',
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        height: 400,
        padding: {
          top: 20,
          right: 20,
          bottom: 60,
          left: 80
        },
        grid: {
          show: true,
          stroke: '#e5e7eb',
          strokeWidth: 1,
          labels: {
            show: true,
            color: '#1A1A1A',
            fontSize: 14,
            xAxisLabels: {
              values: labels,
              show: true,
              fontSize: 12,
              color: '#1A1A1A',
              rotation: -45
            },
            yAxis: {
              show: true,
              useNiceScale: true
            }
          }
        },
        legend: {
          show: true,
          fontSize: 14,
          color: '#1A1A1A'
        },
        title: {
          show: false
        },
        tooltip: {
          show: true,
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          fontSize: 14,
          borderRadius: 4,
          borderColor: '#e5e7eb',
          borderWidth: 1
        }
      },
      line: {
        strokeWidth: 2,
        smooth: true,
        area: {
          useGradient: false,
          opacity: 0
        },
        labels: {
          show: false
        }
      },
      userOptions: {
        show: false
      }
    };
  }

  function getIncomeDistConfig() {
    return {
      style: {
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        backgroundColor: '#FFFFFF',
        color: '#1A1A1A',
        chart: {
          height: 400,
          padding: {
            top: 20,
            right: 20,
            bottom: 60,
            left: 80
          },
          grid: {
            show: true,
            stroke: '#e5e7eb'
          },
          legend: {
            show: true,
            fontSize: 14
          }
        }
      },
      type: 'bar',
      userOptions: {
        show: false
      }
    };
  }

  return {
    getTimeSeriesConfig,
    getCategoryConfig,
    getCashFlowConfig,
    getTrendConfig,
    getIncomeDistConfig
  };
}
