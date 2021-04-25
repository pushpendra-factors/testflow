import React, { useCallback, useRef, useEffect } from 'react';
import styles from './index.module.scss';
import barStyles from '../../Views/CoreQuery/FunnelsResultPage/UngroupedChart/index.module.scss';
import * as d3 from 'd3';
import { getMaxYpoint } from '../BarChart/utils';
import ChartLegends from './ChartLegends';
import { formatCount, numberWithCommas } from '../../utils/dataFormatter';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  REPORT_SECTION,
  DASHBOARD_WIDGET_SECTION,
  BARLINE_COUNT,
} from '../../utils/constants';
import TopLegends from './TopLegends';

function BarLineChart({
  chartData,
  title = 'chart',
  responseRows,
  responseHeaders,
  visibleIndices,
  height: widgetHeight,
  section,
  cardSize = 1,
  legends,
}) {
  const chartRef = useRef(null);
  const tooltip = useRef(null);

  const showTooltip = useCallback(
    (d, i, chartType) => {
      const nodes = d3.select(chartRef.current).selectAll('.bar').nodes();
      nodes.forEach((node, index) => {
        if (index !== i) {
          d3.select(node).attr('class', 'bar opaque');
        }
      });
      let nodePosition, left, top;
      const scrollTop =
        window.pageYOffset !== undefined
          ? window.pageYOffset
          : (
              document.documentElement ||
              document.body.parentNode ||
              document.body
            ).scrollTop;
      if (chartType === "bar") {
        nodePosition = d3.select(nodes[i]).node()?.getBoundingClientRect();
        left = nodePosition.x + nodePosition.width / 2;
        // // if user is hovering over the last bar
        if (left + 200 >= document.documentElement.clientWidth) {
          left = nodePosition.x + nodePosition.width / 2 - 200;
        }
        top = nodePosition.y + scrollTop;
      } else {
        let identifier;
        if (d[0] === '$none') {
          identifier = `id-${title}-none-${i}`;
        } else {
          identifier = `id-${title}-${d[0].split(' ').join('-')}-${i}`;
        }
        nodePosition = d3
          .select(`#${identifier}`)
          .node()
          ?.getBoundingClientRect();
        left = nodePosition.x + 20;
        if (left + 200 >= document.documentElement.clientWidth) {
          left = nodePosition.x + nodePosition.width / 2 - 200;
        }
        top = nodePosition.y - 10 + scrollTop;
      }

      const impressionsIdx = responseHeaders.indexOf('Impressions');
      const clicksIdx = responseHeaders.indexOf('Clicks');
      const spendIdx = responseHeaders.indexOf('Spend');
      const visitorsIdx = responseHeaders.indexOf('Website Visitors');
      const rowIndex = visibleIndices[i];

      const toolTipHeight = d3.select(".toolTip").node()?.getBoundingClientRect()
        .height;

      tooltip.current.html(
        `
        <div style="border-bottom: 1px solid #E7E9ED;">
          <div class="pb-2" style="color: #3E516C;font-size: 14px;line-height: 24px;font-weight: 500;">${
            d[0]
          }</div>
        </div>
        <div style="border-bottom: 1px solid #E7E9ED;" class="py-2">
          <div style="font-weight: 600;font-size: 12px;line-height: 16px;" class="mt-2 flex justify-between">
            <div style="color: #4D7DB4">CONVERSIONS</div>
            <div style="color: #3E516C;">${numberWithCommas(
              formatCount(d[2], 1)
            )}</div>
          </div>
          <div style="font-weight: 600;font-size: 12px;line-height: 16px;" class="mt-2 flex justify-between">
            <div style="color: #D4787D">COST PER CONVERSION</div>
            <div style="color: #3E516C;">${numberWithCommas(
              formatCount(d[1], 1)
            )}</div>
          </div>
        </div>
        <div style="font-size: 12px;line-height: 18px;color: #3E516C;">
          <div class="flex justify-between pt-2">
            <div>Impressions</div>
            <div>${numberWithCommas(
              formatCount(responseRows[rowIndex][impressionsIdx], 1)
            )}</div>
          </div>
          <div class="flex justify-between pt-2">
            <div>Clicks</div>
            <div>${numberWithCommas(
              formatCount(responseRows[rowIndex][clicksIdx], 1)
            )}</div>
          </div>
          <div class="flex justify-between pt-2">
            <div>Spend</div>
            <div>${numberWithCommas(
              formatCount(responseRows[rowIndex][spendIdx], 1)
            )}</div>
          </div>
          <div class="flex justify-between pt-2">
            <div>Visitors</div>
            <div>${numberWithCommas(
              formatCount(responseRows[rowIndex][visitorsIdx], 1)
            )}</div>
          </div>
        </div>
                `
      );
      tooltip.current
        .style('left', left + 'px')
        .style('top', top - toolTipHeight + 5 + 'px')
        .style('opacity', '1');
    },
    [responseHeaders, responseRows, title, visibleIndices]
  );

  const hideTooltip = useCallback(() => {
    const nodes = d3.select(chartRef.current).selectAll('.bar').nodes();
    nodes.forEach((node) => {
      d3.select(node).attr('class', 'bar');
    });
    tooltip.current.style('opacity', 0);
  }, []);

  const drawChart = useCallback(() => {
    d3.select(chartRef.current).html('');
    tooltip.current = d3
      .select(chartRef.current)
      .append('div')
      .attr('class', 'toolTip')
      .style('opacity', '0')
      .style('transition', '0.5s');
    const valuesLeft = [];
    chartData.forEach((cd) => {
      valuesLeft.push(cd[2]);
    });
    const maxLeft = getMaxYpoint(Math.max(...valuesLeft));
    const valuesRight = [];
    chartData.forEach((cd) => {
      valuesRight.push(cd[1]);
    });
    const maxRight = getMaxYpoint(Math.max(...valuesRight));
    const availableWidth = d3
      .select(chartRef.current)
      .node()
      ?.getBoundingClientRect().width;
    const margin = { top: 20, right: 70, bottom: 40, left: 80 };
    const svg = d3
      .select(chartRef.current)
      .append('svg')
      .attr('width', availableWidth)
      .attr('height', widgetHeight || 300);
    const width = +svg.attr('width') - margin.left - margin.right;
    const height = +svg.attr('height') - margin.top - margin.bottom;
    const minBarHeight = 0.05 * height;
    const xScale = d3
      .scaleBand()
      .rangeRound([0, width])
      .padding(0.1)
      .domain(
        chartData.slice(0, BARLINE_COUNT[cardSize]).map(function (d) {
          return d[0];
        })
      );
    const yScaleLeft = d3
      .scaleLinear()
      .rangeRound([height, 0])
      .domain([0, maxLeft]);
    const yScaleRight = d3
      .scaleLinear()
      .rangeRound([height, 0])
      .domain([0, maxRight]);
    var g = svg
      .append('g')
      .attr('transform', 'translate(' + margin.left + ',' + margin.top + ')');

    // axis-x
    g.append('g')
      .attr('class', `axis axis--x ${styles.xAxis}`)
      .attr('transform', 'translate(0,' + height + ')')
      .call(
        d3.axisBottom(xScale).tickFormat((label) => {
          if (label.length > BAR_CHART_XAXIS_TICK_LENGTH[cardSize]) {
            return (
              label.substr(0, BAR_CHART_XAXIS_TICK_LENGTH[cardSize]) + '...'
            );
          }
          return label;
        })
      );

    // axis-y
    g.append('g')
      .attr('class', `axis axis--y ${styles.y1}`)
      .call(d3.axisLeft(yScaleLeft).ticks(5));
    g.append('g')
      .attr('class', `axis axis--y ${styles.y2}`)
      .attr('transform', 'translate( ' + width + ', 0 )')
      .call(d3.axisRight(yScaleRight).ticks(5));

    g.append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 0 - margin.left)
      .attr('x', 0 - height / 2)
      .attr('dy', '1em')
      .style('text-anchor', 'middle')
      .attr('class', styles.yAxisLables)
      .text('Unique users');

    g.append('text')
      .attr('transform', 'rotate(-90)')
      .attr('y', 0 + width + 50)
      .attr('x', 0 - height / 2)
      .attr('dy', '1em')
      .style('text-anchor', 'middle')
      .attr('class', styles.yAxisLables)
      .text('Cost per Conversion');

    var bar = g
      .selectAll('rect')
      .data(chartData.slice(0, BARLINE_COUNT[cardSize]))
      .enter()
      .append('g');

    // bar chart
    bar
      .append('rect')
      .attr('x', function (d) {
        return xScale(d[0]);
      })
      .attr('y', function (d) {
        return height - yScaleLeft(d[2]) > minBarHeight
          ? yScaleLeft(d[2])
          : height - minBarHeight;
      })
      .attr('class', 'bar')
      .attr('width', xScale.bandwidth())
      .attr('height', function (d) {
        return height - yScaleLeft(d[2]) > minBarHeight
          ? height - yScaleLeft(d[2])
          : minBarHeight;
      })
      .style('fill', '#4d7db4')
      .on('mousemove', (d, i) => {
        showTooltip(d, i, 'bar');
      })
      .on('mouseout', () => {
        hideTooltip();
      });

    // labels on the bar chart
    // bar
    //   .append("text")
    //   .attr("dy", "1.3em")
    //   .attr("x", function (d) {
    //     return xScale(d[0]) + xScale.bandwidth() / 2;
    //   })
    //   .attr("y", function (d) {
    //     return yScale(d[2]);
    //   })
    //   .attr("text-anchor", "middle")
    //   .text(function (d) {
    //     return d[2];
    //   });

    // line chart
    var line = d3
      .line()
      .x(function (d, i) {
        return xScale(d[0]) + xScale.bandwidth() / 2;
      })
      .y(function (d) {
        return yScaleRight(d[1]);
      });

    bar
      .append('path')
      .style('fill', 'none')
      .style('stroke', '#D4787D')
      .style('stroke-width', 3)
      .attr('d', line(chartData.slice(0, BARLINE_COUNT[cardSize]))); // 11. Calls the line generator

    bar
      .append('circle') // Uses the enter().append() method
      .style('fill', '#D4787D')
      .style('stroke', '#D4787D')
      .style('stroke-width', 4)
      .attr('id', (d, i) => {
        if (d[0] === '$none') {
          return `id-${title}-none-${i}`;
        }
        return `id-${title}-${d[0].split(' ').join('-')}-${i}`;
      })
      .attr('cx', function (d, i) {
        return xScale(d[0]) + xScale.bandwidth() / 2;
      })
      .attr('cy', function (d) {
        return yScaleRight(d[1]);
      })
      .attr('r', 5)
      .on('mousemove', (d, i) => {
        showTooltip(d, i, 'circle');
      })
      .on('mouseout', () => {
        hideTooltip();
      });
  }, [chartData, hideTooltip, showTooltip, title, widgetHeight, cardSize]);

  useEffect(() => {
    drawChart();
  }, [drawChart, cardSize]);

  return (
    <div className='w-full bar-chart'>
      {section === DASHBOARD_WIDGET_SECTION ? (
        <TopLegends legends={legends} cardSize={cardSize} />
      ) : null}
      <div className={barStyles.ungroupedChart} ref={chartRef}></div>
      {section === REPORT_SECTION ? <ChartLegends legends={legends} /> : null}
    </div>
  );
}

export default BarLineChart;
