import React, { useCallback, useEffect, memo } from 'react';
import ReactDOMServer from 'react-dom/server';
import cx from 'classnames';
import Highcharts from 'highcharts';
import PropTypes from 'prop-types';
import styles from './columnChart.module.scss';
import { Number as NumFormat, Text } from '../factorsComponents';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  FONT_FAMILY
} from '../../utils/constants';
import { CHART_COLOR_1 } from '../../constants/color.constants';

function ColumnChart({
  series,
  categories,
  chartId,
  comparisonApplied,
  cardSize
}) {
  if (comparisonApplied) {
    Highcharts.setOptions({
      defs: {
        stripes: {
          tagName: 'pattern',
          id: 'columnChartStripes',
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
              fill: CHART_COLOR_1
            },
            {
              tagName: 'path',
              d: 'M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2',
              stroke: '#fff',
              strokeWidth: '1px'
            }
          ]
        }
      }
    });
  }

  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        type: 'column',
        animation: false,
        styledMode: comparisonApplied,
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
      yAxis: {
        title: {
          text: null
        }
      },
      credits: {
        enabled: false
      },
      xAxis: {
        categories,
        labels: {
          formatter() {
            const self = this;
            const label = self.value;
            const tickLength = BAR_CHART_XAXIS_TICK_LENGTH[cardSize];
            if (label.length > tickLength) {
              return `${label.substr(0, tickLength)}...`;
            }
            return label;
          }
        }
      },
      plotOptions: {
        column: {
          pointPadding: 0
        },
        series: {
          dataLabels: {
            align: 'center',
            enabled: true,
            useHTML: true,
            formatter() {
              const self = this;
              return ReactDOMServer.renderToString(
                <NumFormat number={self.point.y} />
              );
            }
          },
          borderRadiusTopLeft: 5,
          borderRadiusTopRight: 5
        }
      },
      tooltip: {
        backgroundColor: 'red',
        borderWidth: 0,
        borderRadius: 12,
        borderColor: 'black',
        useHTML: true,
        formatter() {
          const self = this;
          return ReactDOMServer.renderToString(
            <div className='flex flex-col row-gap-2 bannat'>
              <Text
                extraClass={styles.infoText}
                type='title'
                level={7}
                color='grey-2'
              >
                {self.point.category}
              </Text>
              <div className={cx('flex flex-col')}>
                <div className='flex items-center col-gap-1'>
                  <Text weight='bold' type='title' color='grey-6' level={5}>
                    <NumFormat number={self.point.y} />
                  </Text>
                </div>
              </div>
            </div>
          );
        }
      },
      series
    });
  }, [categories, series]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return (
    <div
      className={cx(styles.columnChart, {
        [styles.comparisonApplied]: comparisonApplied
      })}
      id={chartId}
    />
  );
}

export default memo(ColumnChart);

ColumnChart.propTypes = {
  categories: PropTypes.arrayOf(PropTypes.string),
  series: PropTypes.arrayOf(
    PropTypes.shape({
      data: PropTypes.arrayOf(PropTypes.number),
      color: PropTypes.string
    })
  ),
  chartId: PropTypes.string,
  comparisonApplied: PropTypes.bool,
  cardSize: PropTypes.number
};

ColumnChart.defaultProps = {
  categories: [],
  series: [],
  chartId: 'columnChartContainer',
  comparisonApplied: false,
  cardSize: 1
};
