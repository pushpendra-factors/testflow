import React, { useEffect, useRef } from 'react';
import Highcharts from 'highcharts';
import { Text } from 'Components/factorsComponents';
import cx from 'classnames';
import ReactDOMServer from 'react-dom/server';
import { DataMap } from '..';

interface SparklineChartProps {
  data: DataMap;
}

const SparklineChart: React.FC<SparklineChartProps> = ({ data }) => {
  const chartRef = useRef<HTMLDivElement>(null);
  const chartInstanceRef = useRef<Highcharts.Chart | null>(null);

  useEffect(() => {
    if (chartRef.current) {
      const chartOptions: Highcharts.Options = {
        chart: {
          height: 280,
          style: {
            fontFamily: 'Inter'
          }
        },
        title: {
          text: undefined
        },
        legend: {
          enabled: false
        },
        xAxis: {
          categories: Object.keys(data || {}),
          labels: {
            enabled: true
          },
          tickWidth: 0,
          lineWidth: 0
        },
        yAxis: {
          title: {
            text: null
          },
          labels: {
            formatter: function () {
              return this.value + '%';
            }
          },
          max: 100,
          min: 0
        },
        plotOptions: {
          area: {
            color: {
              linearGradient: {
                x1: 0,
                y1: 0,
                x2: 0,
                y2: 1
              },
              stops: [
                [0, 'rgba(0, 120, 255, 1)'],
                [1, 'rgba(255, 255, 255, 1)']
              ]
            },
            marker: {
              radius: 2
            },
            lineWidth: 1
          }
        },
        series: [
          {
            name: 'Engagement Score',
            data: Object.values(data || {}),
            type: 'area',
            lineWidth: 2,
            marker: {
              enabled: false
            }
          }
        ],
        tooltip: {
          backgroundColor: 'white',
          borderWidth: 1,
          borderRadius: 12,
          borderColor: '#00000040',
          shadow: true,
          useHTML: true,
          formatter() {
            return ReactDOMServer.renderToString(
              <div
                className='flex flex-col row-gap-2 p-2'
                style={{ minWidth: '120px' }}
              >
                <Text
                  type='title'
                  level={7}
                  color='grey-2'
                  extraClass='m-0'
                  weight='medium'
                >
                  {this.point.series.name}
                </Text>
                <div className={cx('flex flex-col')}>
                  <Text type='title' color='grey' level={7} extraClass='m-0'>
                    {this.point.category}
                  </Text>
                  <div className='flex items-center'>
                    <Text
                      weight='bold'
                      type='title'
                      color='grey-6'
                      level={5}
                      extraClass='m-0'
                    >
                      {this.point.y}
                    </Text>
                  </div>
                </div>
              </div>
            );
          }
        },
        credits: {
          enabled: false
        }
      };

      if (!chartInstanceRef.current) {
        chartInstanceRef.current = new Highcharts.Chart(
          chartRef.current,
          chartOptions
        );
      } else {
        chartInstanceRef.current.update(chartOptions);
      }
    }
  }, [data]);

  return <div ref={chartRef} />;
};

export default SparklineChart;
