import React, { memo, useCallback, useEffect, useRef } from 'react';
import PropTypes from 'prop-types';
import cx from 'classnames';
import ReactDOMServer from 'react-dom/server';
import Highcharts from 'highcharts';
import get from 'lodash/get';
import { Text, Number as NumFormat } from '../factorsComponents';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  FONT_FAMILY
} from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import styles from './index.module.scss';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import { COLOR_CLASSNAMES } from '../../constants/charts.constants';
import { visualizationColors } from '../../utils/dataFormatter';

function HorizontalBarChart({
  series,
  categories,
  height,
  width,
  cardSize,
  comparisonApplied,
  hideXAxis
}) {
  const chartRef = useRef(null);

  useEffect(() => {
    if (comparisonApplied) {
      const stripes = visualizationColors.reduce(
        (prev, curr, currIndex) => ({
          ...prev,
          [`color_${currIndex}_stripes`]: {
            tagName: 'pattern',
            id: `barChartStripes${currIndex}`,
            patternUnits: 'userSpaceOnUse',
            width: 4,
            height: 4,
            children: [
              {
                tagName: 'rect', // Solid background
                x: 0,
                y: 0,
                width: 4,
                height: 4,
                fill: curr
              },
              {
                tagName: 'path',
                d: 'M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2',
                stroke: '#fff',
                strokeWidth: '1px'
              }
            ]
          }
        }),
        {}
      );
      Highcharts.setOptions({
        defs: stripes
      });
    }
  }, [comparisonApplied]);

  const drawChart = useCallback(() => {
    Highcharts.chart(chartRef.current, {
      chart: {
        type: 'bar',
        animation: false,
        height,
        width,
        styledMode: comparisonApplied,
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
        backgroundColor: 'white',
        borderWidth: 1,
        shadow: false,
        borderRadius: 12,
        useHTML: true,
        outside: true,
        formatter() {
          const self = this;
          const metricType = get(self.point, 'metricType', null);
          return ReactDOMServer.renderToString(
            <>
              <Text
                color='grey-2'
                type='title'
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >
                {self.point.category}
              </Text>
              <span className='flex items-center mt-1'>
                <LegendsCircle extraClass='mr-2' color={self.point.color} />
                <Text
                  color='grey-8'
                  level={6}
                  type='title'
                  weight='bold'
                  extraClass='text-base mb-0'
                >
                  {metricType ? (
                    <div className='number'>
                      {getFormattedKpiValue({
                        value: self.point.y,
                        metricType
                      })}
                    </div>
                  ) : (
                    <NumFormat
                      className='number'
                      number={self.point.y}
                      shortHand={self.point.y >= 1000}
                    />
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
            const self = this;
            return ReactDOMServer.renderToString(
              <Text
                color='grey-2'
                type='title'
                extraClass={`${styles.xAxisLabels} mb-0`}
              >
                {self.value.length > BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
                  ? `${self.value.substr(
                      0,
                      BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
                    )}...`
                  : self.value}
              </Text>
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
              const self = this;
              const metricType = get(self.point, 'metricType', null);
              return ReactDOMServer.renderToString(
                metricType ? (
                  getFormattedKpiValue({ value: self.y, metricType })
                ) : (
                  <NumFormat number={self.y} shortHand={self.y >= 1000} />
                )
              );
            }
          }
        }
      },
      series: series.map((s) => ({
        ...s,
        data: s.data.map((d) => ({
          ...d,
          className: COLOR_CLASSNAMES[d.color]
        }))
      }))
    });
  }, [series, categories, height, width, cardSize, comparisonApplied]);

  useEffect(() => {
    drawChart();
  }, [drawChart, cardSize]);

  return (
    <div
      ref={chartRef}
      className={cx('w-full', styles.horizontalBarChart, {
        [styles.comparisonApplied]: comparisonApplied,
        [styles['no-x-axis']]: hideXAxis
      })}
    />
  );
}

export default memo(HorizontalBarChart);

HorizontalBarChart.propTypes = {
  categories: PropTypes.arrayOf(PropTypes.string),
  series: PropTypes.arrayOf(
    PropTypes.shape({
      name: PropTypes.string,
      data: PropTypes.arrayOf(
        PropTypes.shape({
          y: PropTypes.number,
          color: PropTypes.string
        })
      )
    })
  ),
  comparisonApplied: PropTypes.bool,
  cardSize: PropTypes.number,
  height: PropTypes.number,
  width: PropTypes.number,
  hideXAxis: PropTypes.bool
};

HorizontalBarChart.defaultProps = {
  categories: [],
  series: [],
  comparisonApplied: false,
  cardSize: 1,
  height: null,
  width: null,
  hideXAxis: false
};
