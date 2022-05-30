import React, {
  memo, useCallback, useEffect, useRef
} from 'react';
import ReactDOMServer from 'react-dom/server';
import Highcharts from 'highcharts';
import get from 'lodash/get';
import { Text, Number as NumFormat } from '../factorsComponents';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  METRIC_TYPES
} from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import { formatDuration } from '../../utils/dataFormatter';
import styles from './index.module.scss';

function HorizontalBarChart({
  series, categories, height, width, cardSize
}) {
  const chartRef = useRef(null);
  const drawChart = useCallback(() => {
    Highcharts.chart(chartRef.current, {
      chart: {
        type: 'bar',
        animation: false,
        height,
        width,
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
      tooltip: {
        backgroundColor: 'white',
        borderWidth: 0,
        borderRadius: 12,
        useHTML: true,
        formatter() {
          const metricType = get(this.point, 'metricType', null);
          return ReactDOMServer.renderToString(
            <>
              <Text
                color="grey-2"
                type="title"
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >
                {this.point.category}
              </Text>
              <span className="flex items-center mt-1">
                <LegendsCircle extraClass="mr-2" color={this.point.color} />
                <Text
                  color="grey-8"
                  type="title"
                  weight="bold"
                  extraClass="text-base mb-0"
                >
                  {metricType === METRIC_TYPES.dateType ? (
                    <div className="number">{formatDuration(this.point.y)}</div>
                  ) : (
                    <NumFormat className="number" number={this.point.y} />
                  )}
                </Text>
              </span>
            </>
          );
        }
      },
      xAxis: {
        categories,
        grid: {
          enabled: false
        },
        labels: {
          useHTML: true,
          formatter() {
            return ReactDOMServer.renderToString(
              <>
                <Text
                  color="grey-2"
                  type="title"
                  extraClass={`${styles.xAxisLabels} mb-0`}
                >
                  {this.value.length > BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
                    ? this.value.substr(
                      0,
                      BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
                    ) + '...'
                    : this.value}
                </Text>
              </>
            );
          }
        }
      },
      yAxis: {
        labels: {
          enabled: false
        },
        grid: {
          borderWidth: 0
        },
        title: {
          text: undefined
        }
      },
      legend: {
        enabled: false
      },
      plotOptions: {
        series: {
          animation: false,
          dataLabels: {
            enabled: true,
            formatter() {
              const metricType = get(this.point, 'metricType', null);
              return ReactDOMServer.renderToString(
                metricType === METRIC_TYPES.dateType ? (
                  formatDuration(this.y)
                ) : (
                  <NumFormat number={this.y} shortHand={true} />
                )
              );
            }
          }
        }
      },
      series
    });
  }, [series, categories, height, width, cardSize]);

  useEffect(() => {
    drawChart();
  }, [drawChart, cardSize]);

  return <div ref={chartRef} className={styles.horizontalBarChart}></div>;
}

export default memo(HorizontalBarChart);
