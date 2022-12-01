import React, { useEffect, memo, useCallback } from 'react';
import ReactDOMServer from 'react-dom/server';
import Highcharts from 'highcharts';
import styles from './styles.module.scss';
import { Number as NumFormat } from '../factorsComponents';

import {
  HIGH_CHARTS_BARLINE_DEFAULT_SPACING,
  BAR_CHART_XAXIS_TICK_LENGTH,
  FONT_FAMILY
} from '../../utils/constants';
import { renderBigLengthTicks } from '../../utils/dataFormatter';
import { CHART_COLOR_1, CHART_COLOR_8 } from '../../constants/color.constants';
import TopLegends from '../GroupedBarChart/TopLegends';

function HCBarLineChart({
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = HIGH_CHARTS_BARLINE_DEFAULT_SPACING,
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
        spacing: cardSize !== 1 ? HIGH_CHARTS_BARLINE_DEFAULT_SPACING : spacing,
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
      plotOptions: {
        series: {
          pointPadding: 0,
          dataLabels: {
            align: 'center',
            enabled: true,
            useHTML: true,
            formatter() {
              return ReactDOMServer.renderToString(
                <NumFormat number={this.point.y} className='bar-chart-label' />
              );
            }
          },
          borderRadiusTopLeft: 5,
          borderRadiusTopRight: 5
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
              color: CHART_COLOR_1
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
              color: CHART_COLOR_8
            }
          },
          opposite: true
        }
      ],
      tooltip: {
        outside: true,
        shared: false,
        backgroundColor: 'white',
        borderWidth: 1,
        borderRadius: 12,
        shadow: false,
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
          colors={[CHART_COLOR_1, CHART_COLOR_8]}
          showFullLengthLegends={cardSize === 1}
        />
      ) : null}
      <div className={styles.barLineChart} id={chartId}></div>
      <svg width='0' height='0'>
        <defs>
          <pattern
            id='hatch-left'
            patternUnits='userSpaceOnUse'
            width='4'
            height='4'
          >
            <path d='M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2'></path>
          </pattern>
        </defs>
      </svg>
      {legendsPosition === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          legends={legends}
          colors={[CHART_COLOR_1, CHART_COLOR_8]}
          showFullLengthLegends={cardSize === 1}
        />
      ) : null}
    </>
  );
}

export default memo(HCBarLineChart);
