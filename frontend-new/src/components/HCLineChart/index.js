/* eslint-disable react/no-this-in-sfc */
import React, { useCallback, useEffect, memo } from 'react';
import cx from 'classnames';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import Highcharts from 'highcharts';

import { get, has } from 'lodash';
import { Number as NumFormat, Text } from 'factorsComponents';
import {
  high_charts_default_spacing as highChartsDefaultSpacing,
  FONT_FAMILY
} from '../../utils/constants';
import TopLegends from '../GroupedBarChart/TopLegends';
import { addQforQuarter, generateColors } from '../../utils/dataFormatter';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';
import styles from './styles.module.scss';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';

function LineChart({
  categories,
  data,
  frequency,
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = highChartsDefaultSpacing,
  chartId = 'lineChartContainer',
  showAllLegends = false,
  comparisonApplied = false,
  compareCategories,
  secondaryYAxisIndices = []
}) {
  const dateFormat = getDateFormatForTimeSeriesChart({ frequency });
  const metricTypes = data.reduce(
    (result, d) => ({
      ...result,
      [d.name]: get(d, 'metricType', null)
    }),
    {}
  );

  const colors = generateColors(
    data.filter((d) => !has(d, 'compareIndex')).length
  );

  const drawChart = useCallback(() => {
    const chartConfig = {
      chart: {
        height,
        spacing: cardSize !== 1 ? highChartsDefaultSpacing : spacing,
        style: {
          fontFamily: FONT_FAMILY
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
      yAxis: [
        {
          min: 0,
          title: {
            enabled: false
          }
        },
        {
          min: 0,
          title: {
            enabled: false
          },
          opposite: true
        }
      ],
      credits: {
        enabled: false
      },
      tooltip: {
        backgroundColor: 'white',
        borderWidth: 1,
        borderRadius: 12,
        shadow: false,
        useHTML: true,
        formatter() {
          const metricType = get(metricTypes, this.point.series.name, null);
          const value = this.point.y;
          let timestamp = this.point.category;
          if (
            comparisonApplied &&
            has(this.point.series.userOptions, 'compareIndex')
          ) {
            timestamp = compareCategories[this.point.index];
          }
          return ReactDOMServer.renderToString(
            <div className='flex flex-col row-gap-2'>
              <Text
                extraClass={styles.infoText}
                type='title'
                level={7}
                color='grey-2'
              >
                {this.point.series.name}
              </Text>
              <div className={cx('flex flex-col')}>
                <Text type='title' color='grey' level={7}>
                  {addQforQuarter(frequency) +
                    moment(timestamp).format(dateFormat)}
                </Text>
                <div className='flex items-center col-gap-1'>
                  <Text weight='bold' type='title' color='grey-6' level={5}>
                    {metricType != null && metricType !== '' ? (
                      getFormattedKpiValue({
                        value,
                        metricType
                      })
                    ) : (
                      <NumFormat number={value} />
                    )}
                  </Text>
                </div>
              </div>
            </div>
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
      series: data.map((d, index, dataSet) => {
        const isCompareLine = has(d, 'compareIndex');
        const compareIndex = isCompareLine
          ? dataSet.findIndex((s) => s.index === d.compareIndex)
          : null;
        return {
          ...d,
          color: !isCompareLine ? colors[index] : colors[compareIndex],
          yAxis: isCompareLine
            ? secondaryYAxisIndices.includes(compareIndex)
              ? 1
              : 0
            : secondaryYAxisIndices.includes(index)
            ? 1
            : 0
        };
      })
    };

    Highcharts.chart(chartId, chartConfig);
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
    dateFormat,
    secondaryYAxisIndices,
    compareCategories,
    comparisonApplied
  ]);

  useEffect(() => {
    drawChart();
  }, [cardSize, drawChart]);

  return (
    <>
      {legendsPosition === 'top' ? (
        <TopLegends
          cardSize={cardSize}
          colors={colors}
          showAllLegends={showAllLegends}
          legends={data
            .filter((d) => !has(d, 'compareIndex'))
            .map((d) => d.name)}
        />
      ) : null}
      <div className={styles.areaChart} id={chartId} />
      {legendsPosition === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data
            .filter((d) => !has(d, 'compareIndex'))
            .map((d) => d.name)}
          colors={colors}
          showAllLegends={showAllLegends}
        />
      ) : null}
    </>
  );
}

export default memo(LineChart);
