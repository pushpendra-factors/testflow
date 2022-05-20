import React, { useEffect, memo, useCallback } from 'react';
import Highcharts from 'highcharts';
import styles from './styles.module.scss';
import {
  high_charts_barLine_default_spacing,
  BAR_CHART_XAXIS_TICK_LENGTH
} from '../../utils/constants';
import { renderBigLengthTicks } from '../../utils/dataFormatter';
import TopLegends from '../GroupedBarChart/TopLegends';

function HCBarLineChart({
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = high_charts_barLine_default_spacing,
  chartId = 'lineChartContainer',
  categories,
  series,
  legends,
  generateTooltip
}) {
  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        height,
        spacing: cardSize !== 1 ? high_charts_barLine_default_spacing : spacing,
        style: {
          fontFamily: "'Work Sans', sans-serif"
        }
      },
      title: {
        text: undefined
      },
      credits: {
        enabled: false
      },
      plotOptions: {
        series: {
          pointPadding: 0
        }
      },
      xAxis: [
        {
          categories,
          labels: {
            formatter() {
              return renderBigLengthTicks(
                this.value,
                BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
              );
            }
          }
        }
      ],
      yAxis: [
        {
          // left yAxis
          labels: {
            style: {
              color: '#3e516c'
            }
          },
          title: {
            margin: 20,
            text: 'Unique Users',
            style: {
              color: '#8692A3'
            }
          }
        },
        {
          // right yAxis
          title: {
            margin: 20,
            text: 'Cost per conversion',
            style: {
              color: '#8692A3'
            }
          },
          labels: {
            style: {
              color: '#d4787d'
            }
          },
          opposite: true
        }
      ],
      tooltip: {
        outside: true,
        shared: false,
        backgroundColor: 'white',
        borderWidth: 0,
        borderRadius: 12,
        padding: 16,
        useHTML: true,
        formatter() {
          return generateTooltip(this.point.category);
        }
      },
      legend: {
        enabled: false
      },
      series
    });
  }, [cardSize, categories, chartId, height, series, spacing, generateTooltip]);

  useEffect(() => {
    drawChart();
  }, [cardSize, drawChart]);
  return (
    <>
      {legendsPosition === 'top' ? (
        <TopLegends
          cardSize={cardSize}
          legends={legends}
          colors={['#4d7db4', '#d4787d']}
          showFullLengthLegends={cardSize === 1}
        />
      ) : null}
      <div className={styles.barLineChart} id={chartId}></div>
      <svg width="0" height="0">
        <defs>
          <pattern
            id="hatch-left"
            patternUnits="userSpaceOnUse"
            width="4"
            height="4"
          >
            <path d="M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2"></path>
          </pattern>
        </defs>
      </svg>
      {legendsPosition === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          legends={legends}
          colors={['#4d7db4', '#d4787d']}
          showFullLengthLegends={cardSize === 1}
        />
      ) : null}
    </>
  );
}

export default memo(HCBarLineChart);
