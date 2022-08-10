import React, { useRef, useEffect, useCallback } from 'react';
import cx from 'classnames';
import ReactDOMServer from 'react-dom/server';
import PropTypes from 'prop-types';
import * as d3 from 'd3';
import moment from 'moment';
import {
  numberWithCommas,
  formatCount,
  addQforQuarter
} from '../../utils/dataFormatter';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import { METRIC_TYPES } from '../../utils/constants';
import { Number as NumFormat, SVG, Text } from '../factorsComponents';
import styles from './index.module.scss';

function SparkChart({
  chartData,
  chartColor,
  page,
  event,
  frequency,
  height: widgetHeight,
  title,
  metricType,
  comparisonEnabled
}) {
  const chartRef = useRef(null);

  const bisectDate = d3.bisector(function (d) {
    return d.date;
  }).left;

  const drawChart = useCallback(() => {
    const margin = {
      top: 10,
      right: 10,
      bottom: 30,
      left: 10
    };
    const width = d3
      .select(chartRef.current)
      .node()
      ?.getBoundingClientRect().width;
    const height = widgetHeight || 180;

    // append the svg object to the body of the page
    const svg = d3
      .select(chartRef.current)
      .html('')
      .append('svg')
      .attr('width', width)
      .attr('height', height + margin.top + margin.bottom)
      .append('g')
      .attr('transform', 'translate(' + margin.left + ',' + margin.top + ')');

    const tooltip = d3
      .select(chartRef.current)
      .append('div')
      .attr('class', 'tooltip')
      .style('display', 'none');

    const data = chartData;

    data.sort(function (a, b) {
      return new Date(a.date) - new Date(b.date);
    });

    const x = d3
      .scaleTime()
      .domain(
        d3.extent(data, function (d) {
          return d.date;
        })
      )
      .range([0, width]);

    const y = d3
      .scaleLinear()
      .domain(
        d3.extent(data, function (d) {
          return +d[event];
        })
      )
      .range([height, 0]);

    // Add the area
    svg
      .append('path')
      .datum(data)
      .attr('fill-opacity', 0.3)
      .attr('class', 'area')
      .attr('stroke', 'none')
      .attr(
        'd',
        d3
          .area()
          .x(function (d) {
            return x(d.date);
          })
          .y0(height)
          .y1(function (d) {
            return y(d[event]);
          })
      )
      .attr(
        'fill',
        `url(#area-gradient-${chartColor.substr(1)}-${page}-${title})`
      )
      .attr('stroke-width', 1);

    // Add the line
    svg
      .append('path')
      .datum(data)
      .attr('fill', 'none')
      .attr('stroke', chartColor)
      .attr('stroke-width', 2)
      .attr(
        'd',
        d3
          .line()
          .x(function (d) {
            return x(d.date);
          })
          .y(function (d) {
            return y(d[event]);
          })
      );

    svg
      .append('linearGradient')
      .attr('id', `area-gradient-${chartColor.substr(1)}-${page}-${title}`)
      .attr('gradientUnits', 'userSpaceOnUse')
      .attr('x1', '0%')
      .attr('y1', '0%')
      .attr('x2', '0%')
      .attr('y2', '100%')
      .selectAll('stop')
      .data([
        { offset: '0%', color: chartColor },
        { offset: '80%', color: 'white' }
      ])
      .enter()
      .append('stop')
      .attr('offset', function (d) {
        return d.offset;
      })
      .attr('stop-color', function (d) {
        return d.color;
      });

    const focus = svg
      .append('g')
      .attr('class', 'focus')
      .style('display', 'none');

    focus
      .append('circle')
      .attr('r', 5)
      .attr('stroke', chartColor)
      .attr('fill', '#fff')
      .attr('stroke-width', 2);

    svg
      .append('rect')
      .attr('class', 'overlay')
      .attr('width', width)
      .attr('height', height)
      .on('mouseover', function () {
        focus.style('display', null);
      })
      .on('mouseout', function () {
        tooltip.style('display', 'none');
        focus.style('display', 'none');
      })
      .on('mousemove', mousemove);

    function mousemove() {
      const x0 = x.invert(d3.mouse(this)[0]);
      const i = bisectDate(data, x0, 1);
      const d0 = data[i - 1];
      const d1 = data[i];
      const d = x0 - d0.date > d1.date - x0 ? d1 : d0;

      let left = d3.event.pageX + 20;
      if (left + 146 > document.documentElement.clientWidth) {
        left = d3.event.pageX - 166;
      }
      focus.attr(
        'transform',
        'translate(' + x(d.date) + ',' + y(d[event]) + ')'
      );
      tooltip.style('display', 'block');
      const format = getDateFormatForTimeSeriesChart({ frequency });

      const percentChange = comparisonEnabled
        ? ((d[event] - d.compareValue) / d.compareValue) * 100
        : 0;

      const changeIcon = comparisonEnabled ? (
        <SVG
          color={percentChange > 0 ? '#5ACA89' : '#FF0000'}
          name={percentChange > 0 ? 'arrowLift' : 'arrowDown'}
          size={16}
        />
      ) : null;

      tooltip
        .html(
          ReactDOMServer.renderToString(
            <div className="flex flex-col row-gap-2">
              <Text type="title" level={7} color="grey-2">
                {event}
              </Text>
              <div
                className={cx('flex flex-col')}
                style={
                  comparisonEnabled
                    ? {
                        borderLeft: `3px solid ${chartColor}`,
                        paddingLeft: '9px'
                      }
                    : {}
                }
              >
                <Text type="title" color="grey" level={7}>
                  {addQforQuarter(frequency) + moment(d.date).format(format)}
                </Text>
                <div className="flex items-center col-gap-1">
                  <Text
                    weight="bold"
                    type="title"
                    color="grey-6"
                    level={5}
                  >
                    {metricType != null
                      ? getFormattedKpiValue({ value: d[event], metricType })
                      : numberWithCommas(formatCount(d[event], 1))}
                  </Text>
                  {comparisonEnabled && (
                    <>
                      {changeIcon}
                      <Text level={7} type="title" color="grey">
                        <NumFormat number={Math.abs(percentChange)} />%
                      </Text>
                    </>
                  )}
                </div>
              </div>
              {comparisonEnabled && (
                <div className="flex flex-col pl-3">
                  <Text type="title" color="grey" level={7}>
                    {addQforQuarter(frequency) +
                      moment(d.compareDate).format(format)}
                  </Text>
                  <Text
                    weight="bold"
                    className="mt-0"
                    type="title"
                    color="grey-6"
                    level={5}
                  >
                    {metricType != null
                      ? getFormattedKpiValue({
                          value: d['compareValue'],
                          metricType
                        })
                      : numberWithCommas(formatCount(d['compareValue'], 1))}
                  </Text>
                </div>
              )}
            </div>
          )
        )
        .style('left', left + 'px')
        .style('top', d3.event.pageY - 40 + 'px');
    }
  }, [
    bisectDate,
    chartData,
    chartColor,
    event,
    page,
    frequency,
    widgetHeight,
    title
  ]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return <div className={styles.sparkChart} ref={chartRef} />;
}

export default SparkChart;

SparkChart.propTypes = {
  title: PropTypes.string,
  chartColor: PropTypes.string,
  event: PropTypes.string,
  frequency: PropTypes.string,
  height: PropTypes.number,
  metricType: PropTypes.oneOf([
    METRIC_TYPES.dateType,
    METRIC_TYPES.percentType
  ]),
  page: PropTypes.string,
  chartData: PropTypes.arrayOf(
    PropTypes.shape({
      date: PropTypes.instanceOf(Date)
    })
  ),
  comparisonEnabled: PropTypes.bool
};

SparkChart.defaultProps = {
  title: 'Chart',
  chartColor: '#4D7DB4',
  event: 'event',
  frequency: 'date',
  height: 180,
  metricType: undefined,
  page: 'KPI',
  chartData: [],
  comparisonEnabled: false
};
