import React, { useCallback, useEffect, memo, useMemo } from 'react';
import ReactDOMServer from 'react-dom/server';
import cx from 'classnames';
import Highcharts from 'highcharts';
import PropTypes from 'prop-types';
import styles from './columnChart.module.scss';
import { Number as NumFormat, Text } from '../factorsComponents';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  FONT_FAMILY,
  METRIC_TYPES
} from '../../utils/constants';
import { CHART_COLOR_1 } from '../../constants/color.constants';
import { COLOR_CLASSNAMES } from '../../constants/charts.constants';
import { generateColors } from '../../utils/dataFormatter';
import TopLegends from 'Components/GroupedBarChart/TopLegends';

const defaultColors = generateColors(10);

function ColumnChart({
  series,
  categories,
  chartId,
  comparisonApplied,
  cardSize,
  multiColored,
  colors,
  valueMetricType,
  height,
  legendsProps
}) {
  useEffect(() => {
    if (comparisonApplied) {
      if (multiColored) {
        const stripes = colors.reduce((prev, curr, currIndex) => {
          return {
            ...prev,
            [`color_${currIndex}_stripes`]: {
              tagName: 'pattern',
              id: `stripes-${COLOR_CLASSNAMES[curr]}`,
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
          };
        }, {});
        Highcharts.setOptions({
          defs: stripes
        });
      } else {
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
    }
  }, [comparisonApplied, multiColored, colors]);

  const updatedSeries = useMemo(() => {
    if (!multiColored) {
      return series;
    }
    return series.map((s) => {
      return {
        ...s,
        data: s.data.map((d, index) => {
          return {
            y: d,
            className: COLOR_CLASSNAMES[colors[index]]
          };
        })
      };
    });
  }, [series, multiColored, colors]);

  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        type: 'column',
        animation: false,
        styledMode: comparisonApplied,
        style: {
          fontFamily: FONT_FAMILY
        },
        height
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
                <>
                  <NumFormat
                    number={self.point.y}
                    className='bar-chart-label'
                  />
                  {valueMetricType === METRIC_TYPES.percentType ? '%' : ''}
                </>
              );
            }
          },
          borderRadiusTopLeft: 5,
          borderRadiusTopRight: 5
        }
      },
      tooltip: {
        backgroundColor: 'white',
        borderWidth: 1,
        borderRadius: 12,
        shadow: false,
        useHTML: true,
        formatter() {
          const self = this;
          return ReactDOMServer.renderToString(
            <div className='flex flex-col row-gap-2'>
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
                    {valueMetricType === METRIC_TYPES.percentType ? '%' : ''}
                  </Text>
                </div>
              </div>
            </div>
          );
        }
      },
      series: updatedSeries
    });
  }, [
    cardSize,
    categories,
    chartId,
    comparisonApplied,
    updatedSeries,
    valueMetricType,
    height
  ]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return (
    <>
      {legendsProps != null && legendsProps.position === 'top' ? (
        <TopLegends
          cardSize={cardSize}
          colors={colors}
          showAllLegends={legendsProps.showAll}
          legends={categories}
        />
      ) : null}
      <div
        className={cx('w-full', styles.columnChart, {
          [styles.comparisonApplied]: comparisonApplied && !multiColored,
          [styles.multiColoredComparisonApplied]:
            comparisonApplied && multiColored
        })}
        id={chartId}
      />
      {legendsProps != null && legendsProps.position === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          colors={colors}
          showAllLegends={legendsProps.showAll}
          legends={categories}
        />
      ) : null}
    </>
  );
}

export default memo(ColumnChart);

ColumnChart.propTypes = {
  categories: PropTypes.arrayOf(PropTypes.string),
  series: PropTypes.arrayOf(
    PropTypes.shape({
      data: PropTypes.arrayOf(PropTypes.number)
    })
  ),
  chartId: PropTypes.string,
  comparisonApplied: PropTypes.bool,
  cardSize: PropTypes.number,
  multiColored: PropTypes.bool,
  colors: PropTypes.arrayOf(PropTypes.string),
  valueMetricType: PropTypes.string,
  height: PropTypes.number,
  legendsProps: PropTypes.shape({
    showAll: PropTypes.bool,
    position: PropTypes.oneOf(['top', 'bottom'])
  })
};

ColumnChart.defaultProps = {
  categories: [],
  series: [],
  chartId: 'columnChartContainer',
  comparisonApplied: false,
  cardSize: 1,
  multiColored: false,
  colors: defaultColors,
  valueMetricType: null,
  height: null,
  legendsProps: null
};
