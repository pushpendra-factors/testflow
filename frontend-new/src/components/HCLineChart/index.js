import React, { useCallback, useEffect, memo } from 'react';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import Highcharts from 'highcharts';

import { Text, Number as NumFormat } from '../factorsComponents';
import {
  high_charts_default_spacing,
  METRIC_TYPES
} from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import TopLegends from '../GroupedBarChart/TopLegends';
import {
  addQforQuarter,
  generateColors,
  formatDuration
} from '../../utils/dataFormatter';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';
import styles from './styles.module.scss';
import { get } from 'lodash';

function LineChart({
  categories,
  data,
  frequency,
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = high_charts_default_spacing,
  chartId = 'lineChartContainer',
  showAllLegends = false
}) {
  const dateFormat = getDateFormatForTimeSeriesChart({ frequency });
  const metricTypes = data.reduce((result, d) => {
    return {
      ...result,
      [d.name]: get(d, 'metricType', null)
    };
  }, {});
  const colors = generateColors(data.length);

  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        height,
        spacing: cardSize !== 1 ? high_charts_default_spacing : spacing,
        style: {
          fontFamily: "'Work Sans', sans-serif"
        }
      },
      legend: {
        enabled: false
      },
      title: {
        text: undefined
      },
      xAxis: {
        categories,
        title: {
          enabled: false
        },
        labels: {
          formatter() {
            return (
              addQforQuarter(frequency) + moment(this.value).format(dateFormat)
            );
          }
        }
      },
      yAxis: {
        min: 0,
        title: {
          enabled: false
        }
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
          const metricType = get(metricTypes, this.point.series.name, null);
          return ReactDOMServer.renderToString(
            <>
              <Text
                color="grey-8"
                weight="bold"
                type="title"
                extraClass="text-sm mb-0"
              >
                {addQforQuarter(frequency) +
                  moment(this.point.category).format(dateFormat)}
              </Text>
              <Text
                color="grey-2"
                type="title"
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >
                {this.point.series.name}
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
      plotOptions: {
        line: {
          marker: {
            symbol: 'circle'
          }
        }
      },
      series: data.map((d, index) => {
        return {
          ...d,
          color: colors[index]
        };
      })
    });
  }, [
    cardSize,
    categories,
    data,
    frequency,
    height,
    spacing,
    chartId,
    colors,
    metricTypes,
    dateFormat
  ]);

  useEffect(() => {
    drawChart();
  }, [cardSize, drawChart]);

  return (
    <>
      {legendsPosition === 'top' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data.map((d) => d.name)}
          colors={colors}
          showAllLegends={showAllLegends}
        />
      ) : null}
      <div className={styles.areaChart} id={chartId}></div>
      {legendsPosition === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data.map((d) => d.name)}
          colors={colors}
          showAllLegends={showAllLegends}
        />
      ) : null}
    </>
  );
}

export default memo(LineChart);
