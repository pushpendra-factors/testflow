import React, { useCallback, useEffect, memo, useMemo } from 'react';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import Highcharts from 'highcharts';
import styles from './styles.module.scss';
import { Text, Number as NumFormat } from '../factorsComponents';
import {
  high_charts_default_spacing as highChartsDefaultSpacing,
  FONT_FAMILY
} from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import {
  addQforQuarter,
  calculatePercentage,
  generateColors
} from '../../utils/dataFormatter';
import TopLegends from '../GroupedBarChart/TopLegends';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';

function StackedBarChart({
  categories,
  data,
  frequency,
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = highChartsDefaultSpacing,
  chartId = 'barChartContainer',
  showAllLegends = false,
  dateWiseTotals = []
}) {
  const computePointValues = useCallback(
    (point) => {
      const categoryIndex = categories.indexOf(point.category);
      const value = point.y;
      const total =
        categoryIndex > -1 && dateWiseTotals.length > 0
          ? dateWiseTotals[categoryIndex]
          : point.stackTotal;

      return {
        percent: calculatePercentage(value, total, 1),
        total
      };
    },
    [categories, dateWiseTotals]
  );

  const colors = useMemo(() => {
    return generateColors(data.length);
  }, [data.length]);

  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        type: 'column',
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
      credits: {
        enabled: false
      },
      xAxis: {
        categories,
        title: {
          enabled: false
        },
        labels: {
          formatter() {
            const self = this;
            if (frequency === 'hour') {
              return moment(self.value).format('MMM D, h A');
            }
            if (frequency === 'date' || frequency === 'week') {
              return moment(self.value).format('MMM D');
            }
            if (frequency === 'month') {
              return moment(self.value).format('MMM YYYY');
            }
            return `${`Q${moment(self.value).format('Q, YYYY')}`}`;
          }
        }
      },
      yAxis: {
        min: 0,
        title: {
          enabled: false
        },
        stackLabels: {
          enabled: false
          // formatter() {
          //   const self = this;
          //   return ReactDOMServer.renderToString(
          //     <NumFormat shortHand={self.total >= 1000} number={self.total} />
          //   );
          // }
        }
      },
      tooltip: {
        backgroundColor: 'white',
        borderWidth: 1,
        shadow: false,
        borderRadius: 12,
        useHTML: true,
        formatter() {
          const self = this;
          const format = getDateFormatForTimeSeriesChart({ frequency });
          const values = computePointValues(self.point);
          return ReactDOMServer.renderToString(
            <>
              <Text
                color='grey-8'
                weight='bold'
                type='title'
                extraClass='text-sm mb-0'
              >
                {addQforQuarter(frequency) +
                  moment(self.point.category).format(format)}
              </Text>
              <Text
                color='grey-2'
                type='title'
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >
                {self.point.series.name}
              </Text>
              <span className='flex items-center mt-1'>
                <LegendsCircle extraClass='mr-2' color={self.point.color} />
                <Text
                  color='grey-8'
                  type='title'
                  weight='bold'
                  extraClass='text-base mb-0'
                >
                  <NumFormat className='number' number={self.point.y} />
                </Text>
              </span>
              <Text
                type='title'
                color='grey-2'
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >
                {`${values.percent}% (${self.point.y} of ${
                  values.total
                })`}
              </Text>
            </>
          );
        }
      },
      plotOptions: {
        column: {
          stacking: 'normal'
        }
      },
      series: data.map((d, index) => ({ ...d, color: colors[index] }))
    });
  }, [
    chartId,
    height,
    cardSize,
    spacing,
    categories,
    data,
    frequency,
    computePointValues,
    colors
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
      <div id={chartId} className={styles.columnChart} />
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

export default memo(StackedBarChart);
