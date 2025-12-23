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
  function getTimeSeriesConfig() {
    return {
      style: {
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        chart: {
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          height: 350,
          legend: {
            fontSize: 14
          },
          title: {
            fontSize: 16,
            color: '#1A1A1A',
            bold: true
          }
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
          height: 380,
          layout: {
            labels: {
              dataLabels: {
                show: true,
                fontSize: 14
              },
              value: {
                show: true,
                fontSize: 14
              }
            },
            donut: {
              strokeWidth: 80,
              useLabelSlot: true
            }
          },
          legend: {
            fontSize: 12
          }
        }
      },
      userOptions: {
        show: false
      }
    };
  }

  function getCashFlowConfig() {
    return {
      style: {
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        chart: {
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          height: 350
        }
      },
      userOptions: {
        show: false
      }
    };
  }

  function getTrendConfig() {
    return {
      style: {
        fontFamily: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif",
        chart: {
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          height: 350
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
        chart: {
          backgroundColor: '#FFFFFF',
          color: '#1A1A1A',
          height: 350
        }
      },
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
