import React, { useCallback, useEffect, memo } from 'react';
import styles from './styles.module.scss';
// import { Text, Number as NumFormat } from '../factorsComponents';
// import styles from './styles.module.scss';
// import ReactDOMServer from 'react-dom/server';
// import moment from 'moment';
import Highcharts from 'highcharts';
import { high_charts_scatter_plot_default_spacing } from '../../utils/constants';
// import LegendsCircle from '../../styles/components/LegendsCircle';
// import { formatCount, generateColors } from '../../utils/dataFormatter';
// import TopLegends from '../GroupedBarChart/TopLegends';

function ScatterPlotChart({
  series,
  yAxisTitle = 'Unique Users',
  xAxisTitle = 'Cost Per Conversion',
  generateTooltip,
  spacing = high_charts_scatter_plot_default_spacing,
  chartId = 'areaChartContainer',
  cardSize = 1,
  height = null,
  // legendsPosition = 'bottom',
  // showAllLegends = false,
}) {
  // const colors = generateColors(data.length);
  const drawChart = useCallback(() => {
    Highcharts.chart(chartId, {
      chart: {
        type: 'scatter',
        height,
        spacing:
          cardSize !== 1 ? high_charts_scatter_plot_default_spacing : spacing,
        style: {
          fontFamily: "'Work Sans', sans-serif",
        },
      },
      title: {
        text: undefined,
      },
      credits: {
        enabled: false,
      },
      tooltip: {
        shared: false,
        backgroundColor: 'white',
        borderWidth: 0,
        borderRadius: 12,
        padding: 16,
        useHTML: true,
        formatter: function () {
          return generateTooltip(this.point.index);
        },
      },
      xAxis: {
        gridLineDashStyle: 'Dash',
        gridLineWidth: 1,
        gridLineColor: '#D9D9D9',
        labels: {
          style: {
            color: '#8692A3',
            fontSize: '12px',
          },
        },
        title: {
          margin: 10,
          enabled: true,
          text: xAxisTitle,
          style: {
            color: '#3E516C',
            fontSize: '14px',
          },
        },
      },
      yAxis: [
        {
          gridLineDashStyle: 'Dash',
          gridLineColor: '#D9D9D9',
          gridLineWidth: 1,
          min: 0,
          labels: {
            style: {
              color: '#8692A3',
              fontSize: '12px',
            },
          },
          title: {
            margin: 20,
            text: yAxisTitle,
            style: {
              color: '#3E516C',
              fontSize: '14px',
            },
          },
        },
      ],
      legend: {
        enabled: false,
      },
      plotOptions: {
        scatter: {
          marker: {
            symbol: 'circle',
            radius: 5,
            states: {
              hover: {
                enabled: true,
              },
            },
          },
          states: {
            hover: {
              marker: {
                enabled: false,
              },
            },
          },
        },
      },
      series,
    });
  }, [series, xAxisTitle, yAxisTitle, chartId, height, generateTooltip]);

  useEffect(() => {
    drawChart();
  }, [cardSize, drawChart]);

  return (
    <>
      {/* {legendsPosition === 'top' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data.map((d) => d.name)}
          colors={colors}
          showAllLegends={showAllLegends}
        />
      ) : null} */}
      <div id={chartId} className={styles.scatterPlotChart}></div>
      {/* {legendsPosition === 'bottom' ? (
        <TopLegends
          cardSize={cardSize}
          legends={data.map((d) => d.name)}
          colors={colors}
          showAllLegends={showAllLegends}
        />
      ) : null} */}
    </>
  );
}

export default memo(ScatterPlotChart);
