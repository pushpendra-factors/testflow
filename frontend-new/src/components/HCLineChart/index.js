import React, { useCallback, useEffect, memo } from 'react';
import { Text, Number as NumFormat } from '../factorsComponents';
import styles from './styles.module.scss';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import Highcharts from 'highcharts';
import { high_charts_default_spacing } from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import TopLegends from '../GroupedBarChart/TopLegends';
import { generateColors } from '../../utils/dataFormatter';

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
}) {
  const colors = generateColors(data.length);
  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        height,
        spacing: cardSize !== 1 ? high_charts_default_spacing : spacing,
        style: {
          fontFamily: "'Work Sans', sans-serif",
        },
      },
      legend: {
        enabled: false,
      },
      title: {
        text: undefined,
      },
      xAxis: {
        categories,
        title: {
          enabled: false,
        },
        labels: {
          formatter: function () {
            if (frequency === 'hour') {
              return moment(this.value).format('MMM D, h A');
            } else if (frequency === 'date' || frequency === 'week') {
              return moment(this.value).format('MMM D');
            } else return moment(this.value).format('MMM YYYY');
          },
        },
      },
      yAxis: {
        min: 0,
        title: {
          enabled: false,
        },
      },
      credits: {
        enabled: false,
      },
      tooltip: {
        backgroundColor: 'white',
        borderWidth: 0,
        borderRadius: 12,
        useHTML: true,
        formatter: function () {
          const format = frequency === 'hour' ? 'MMM D, h A' : 'MMM D, YYYY';
          return ReactDOMServer.renderToString(
            <>
              <Text
                color='grey-8'
                weight='bold'
                type='title'
                extraClass='text-sm mb-0'
              >
                {moment(this.point.category).format(format)}
              </Text>
              <Text
                color='grey-2'
                type='title'
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >
                {this.point.series.name}
              </Text>
              <span className='flex items-center mt-1'>
                <LegendsCircle extraClass='mr-2' color={this.point.color} />
                <Text
                  color='grey-8'
                  type='title'
                  weight='bold'
                  extraClass='text-base mb-0'
                >
                  <NumFormat className='number' number={this.point.y} />
                </Text>
              </span>
            </>
          );
        },
      },
      plotOptions: {
        line: {
          marker: {
            symbol: 'circle',
          },
        },
      },
      series: data.map((d, index) => {
        return { ...d, color: colors[index] };
      }),
    });
  }, [cardSize, categories, data, frequency, height, spacing, chartId, colors]);

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