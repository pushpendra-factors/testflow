import React, { useRef, useEffect, useCallback } from 'react';
import * as d3 from 'd3';
import moment from 'moment';
import styles from './index.module.scss';
import {
  numberWithCommas,
  formatCount,
  addQforQuarter,
  formatDuration
} from '../../utils/dataFormatter';
import { METRIC_TYPES } from '../../utils/constants';
import { getDateFormatForTimeSeriesChart } from '../../utils/chart.helpers';

function SparkChart({
  chartData,
  chartColor,
  page,
  event,
  frequency,
  height: widgetHeight,
  title = 'chart',
  metricType = null
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
      tooltip
        .html(
          `<div class="font-semibold">${
            addQforQuarter(frequency) + moment(d.date).format(format)
          }</div><div class="font-normal mt-1">${
            metricType === METRIC_TYPES.dateType
              ? formatDuration(d[event])
              : numberWithCommas(formatCount(d[event], 1))
          }</div>`
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
