import React, { useCallback, useEffect, memo } from 'react';
import styles from './styles.module.scss';
import Highcharts from 'highcharts';
import { HIGH_CHARTS_SCATTER_PLOT_DEFAULT_SPACING, FONT_FAMILY } from '../../utils/constants';

function ScatterPlotChart({
  series,
  yAxisTitle = 'Unique Users',
  xAxisTitle = 'Cost Per Conversion',
  generateTooltip,
  spacing = HIGH_CHARTS_SCATTER_PLOT_DEFAULT_SPACING,
  chartId = 'areaChartContainer',
  cardSize = 1,
  height = null
}) {
  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        type: 'scatter',
        height,
        spacing:
          cardSize !== 1 ? HIGH_CHARTS_SCATTER_PLOT_DEFAULT_SPACING : spacing,
        style: {
          fontFamily: FONT_FAMILY
        }
      },
      title: {
        text: undefined
      },
      credits: {
        enabled: false
      },
      tooltip: {
        shared: false,
        backgroundColor: 'white',
        borderWidth: 0,
        borderRadius: 12,
        padding: 16,
        useHTML: true,
        formatter() {
          return generateTooltip(this.point.index);
        }
      },
      xAxis: {
        gridLineDashStyle: 'Dash',
        gridLineWidth: 1,
        gridLineColor: '#D9D9D9',
        labels: {
          style: {
            color: '#8692A3',
            fontSize: '12px'
          }
        },
        title: {
          margin: 10,
          enabled: true,
          text: xAxisTitle,
          style: {
            color: '#3E516C',
            fontSize: '14px'
          }
        }
      },
      yAxis: [
        {
          gridLineDashStyle: 'Dash',
          gridLineColor: '#D9D9D9',
          gridLineWidth: 1,
          min: 0,
          labels: {
            style: {
              color: '#8692A3',
              fontSize: '12px'
            }
          },
          title: {
            margin: 20,
            text: yAxisTitle,
            style: {
              color: '#3E516C',
              fontSize: '14px'
            }
          }
        }
      ],
      legend: {
        enabled: false
      },
      plotOptions: {
        scatter: {
          marker: {
            symbol: 'circle',
            radius: 5,
            states: {
              hover: {
                enabled: true
              }
            }
          },
          states: {
            hover: {
              marker: {
                enabled: false
              }
            }
          }
        }
      },
      series
    });
  }, [series, xAxisTitle, yAxisTitle, chartId, height, generateTooltip]);

  useEffect(() => {
    drawChart();
  }, [cardSize, drawChart]);

  return <div id={chartId} className={styles.scatterPlotChart}></div>;
}

export default memo(ScatterPlotChart);
