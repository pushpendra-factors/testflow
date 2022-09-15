import React, { useCallback, useEffect, memo } from 'react';
import cx from 'classnames';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import Highcharts from 'highcharts';

import { Text, Number as NumFormat } from '../factorsComponents';
import { high_charts_default_spacing } from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import TopLegends from '../GroupedBarChart/TopLegends';
import { addQforQuarter, generateColors } from '../../utils/dataFormatter';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';
import styles from './styles.module.scss';
import { get, has } from 'lodash';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';

function LineChart({
  categories,
  data,
  frequency,
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = high_charts_default_spacing,
  chartId = 'lineChartContainer',
  showAllLegends = false,
  comparisonApplied = false,
  compareCategories
}) {
  const dateFormat = getDateFormatForTimeSeriesChart({ frequency });
  const metricTypes = data.reduce((result, d) => {
    return {
      ...result,
      [d.name]: get(d, 'metricType', null)
    };
  }, {});

  const colors = generateColors(
    data.filter((d) => !has(d, 'compareIndex')).length
  );

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
          const value = this.point.y;
          let timestamp = this.point.category;
          if (
            comparisonApplied &&
            has(this.point.series.userOptions, 'compareIndex')
          ) {
            timestamp = compareCategories[this.point.index];
          }
          return ReactDOMServer.renderToString(
            <>
              <div className="flex flex-col row-gap-2">
                <Text
                  extraClass={styles.infoText}
                  type="title"
                  level={7}
                  color="grey-2"
                >
                  {this.point.series.name}
                </Text>
                <div className={cx('flex flex-col')}>
                  <Text type="title" color="grey" level={7}>
                    {addQforQuarter(frequency) +
                      moment(timestamp).format(dateFormat)}
                  </Text>
                  <div className="flex items-center col-gap-1">
                    <Text weight="bold" type="title" color="grey-6" level={5}>
                      {metricType != null && metricType !== '' ? (
                        getFormattedKpiValue({
                          value,
                          metricType
                        })
                      ) : (
                        <NumFormat number={value} />
                      )}
                    </Text>
                    {/* {comparisonApplied && (
                      <>
                        {changeIcon}
                        <Text level={7} type="title" color="grey">
                          <NumFormat number={Math.abs(10)} />%
                        </Text>
                      </>
                    )} */}
                  </div>
                </div>
              </div>
              {/* <Text
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
                  {metricType ? (
                    <div className="number">
                      {getFormattedKpiValue({
                        value: this.point.y,
                        metricType
                      })}
                    </div>
                  ) : (
                    <NumFormat className="number" number={this.point.y} />
                  )}
                </Text>
              </span> */}
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
      series: data.map((d, index, dataSet) => {
        const isCompareLine = has(d, 'compareIndex');
        const compareIndex = isCompareLine
          ? dataSet.findIndex((s) => s.index === d.compareIndex)
          : null;
        return {
          ...d,
          color: !isCompareLine ? colors[index] : colors[compareIndex]
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
          colors={colors}
          showAllLegends={showAllLegends}
          legends={data
            .filter((d) => !has(d, 'compareIndex'))
            .map((d) => d.name)}
        />
      ) : null}
      <div className={styles.areaChart} id={chartId}></div>
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
