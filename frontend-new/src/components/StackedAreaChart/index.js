import React, { useState, useCallback, useEffect } from 'react';
import { Text, Number as NumFormat } from '../factorsComponents';
import styles from './styles.module.scss';
import ReactDOMServer from 'react-dom/server';
import moment from 'moment';
import Highcharts from 'highcharts';
import HighchartsReact from 'highcharts-react-official';
import { high_charts_default_spacing } from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import { formatCount } from '../../utils/dataFormatter';
import TopLegends from '../GroupedBarChart/TopLegends';

function StackedAreaChart({
  categories,
  data,
  frequency,
  height = null,
  legendsPosition = 'bottom',
  cardSize = 1,
  spacing = high_charts_default_spacing,
}) {
  const getChartOptions = useCallback(() => {
    return {
      chart: {
        type: 'area',
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
            return frequency === 'hour'
              ? moment(this.value).format('MMM D, h A')
              : moment(this.value).format('MMM D');
          },
        },
      },
      yAxis: {
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
              <Text
                type='title'
                color='grey-2'
                extraClass={`mt-1 ${styles.infoText} mb-0`}
              >{`${formatCount(this.point.percentage, 1)}% (${
                this.point.y
              } of ${this.point.stackTotal})`}</Text>
            </>
          );
        },
      },
      plotOptions: {
        area: {
          stacking: 'normal',
          lineWidth: 2,
          marker: {
            symbol: 'circle',
          },
        },
      },
      series: data,
    };
  }, [categories, data, frequency, height, spacing, cardSize]);

  const [options, setOptions] = useState(getChartOptions());

  useEffect(() => {
    setOptions(getChartOptions());
  }, [cardSize, getChartOptions]);

  return (
    <>
      {legendsPosition === 'top' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data.map((d) => d.name)}
          colors={data.map((d) => d.color)}
          showFullLegends={false}
        />
      ) : null}
      <div className={styles.areaChart}>
        <HighchartsReact highcharts={Highcharts} options={options} />
      </div>
      {legendsPosition === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data.map((d) => d.name)}
          colors={data.map((d) => d.color)}
          showFullLegends={true}
        />
      ) : null}
    </>
  );
}

export default StackedAreaChart;
