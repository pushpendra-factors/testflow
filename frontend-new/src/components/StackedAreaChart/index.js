import React, { useCallback, useEffect, memo } from 'react';
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
  formatCount,
  generateColors
} from '../../utils/dataFormatter';
import TopLegends from '../GroupedBarChart/TopLegends';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';

function StackedAreaChart({
  categories,
  data,
  frequency,
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = highChartsDefaultSpacing,
  chartId = 'areaChartContainer',
  showAllLegends = false
}) {
  const colors = generateColors(data.length);
  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        type: 'area',
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
        title: {
          enabled: false
        }
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
        formatter() {
          const self = this;
          const format = getDateFormatForTimeSeriesChart({ frequency });
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
                  <NumFormat
                    className='number'
                    number={self.point.y}
                    shortHand={self.point.y >= 1000}
                  />
                </Text>
              </span>
              <Text
                type='title'
                color='grey-2'
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >{`${formatCount(self.point.percentage, 1)}% (${
                self.point.y
              } of ${self.point.stackTotal})`}</Text>
            </>
          );
        }
      },
      plotOptions: {
        area: {
          stacking: 'normal',
          lineWidth: 2,
          marker: {
            symbol: 'circle'
          }
        }
      },
      series: data.map((d, index) => ({ ...d, color: colors[index] }))
    });
  }, [cardSize, categories, chartId, data, frequency, height, spacing, colors]);

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
      <div id={chartId} className={styles.areaChart} />
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

export default memo(StackedAreaChart);
