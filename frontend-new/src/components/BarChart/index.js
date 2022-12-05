import React, { useRef, useCallback, useEffect, memo } from 'react';
import get from 'lodash/get';
import * as d3 from 'd3';
import { checkForWindowSizeChange } from '../../Views/CoreQuery/FunnelsResultPage/utils';
import { getMaxYpoint, getBarChartLeftMargin } from './utils';
import { numberWithCommas, generateColors } from '../../utils/dataFormatter';
import { CHART_COLOR_1 } from '../../constants/color.constants';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  REPORT_SECTION,
  DASHBOARD_MODAL,
  DASHBOARD_WIDGET_SECTION,
  BAR_COUNT,
  FONT_FAMILY
} from '../../utils/constants';
import TopLegends from '../GroupedBarChart/TopLegends';
import { getFormattedKpiValue } from '../../Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
import styles from '../../Views/CoreQuery/FunnelsResultPage/UngroupedChart/index.module.scss';

function BarChart({
  chartData,
  queries,
  title = 'chart',
  height: widgetHeight,
  section,
  cardSize = 1
}) {
  const tooltip = useRef(null);
  const chartRef = useRef(null);

  const getLabel = useCallback(
    (str, position = 'tick') => {
      let label = str.split(';')[0];
      label = label
        .split(',')
        .filter((elem) => elem)
        .join(',');

      const tickLength = BAR_CHART_XAXIS_TICK_LENGTH[cardSize];
      if (label.length > tickLength && position === 'tick') {
        return label.substr(0, tickLength) + '...';
      }
      return label;
    },
    [cardSize]
  );

  const showTooltip = useCallback(
    (d, i) => {
      const nodes = d3.select(chartRef.current).selectAll('.bar').nodes();
      nodes.forEach((node, index) => {
        if (index !== i) {
          d3.select(node).attr('class', 'bar opaque');
        }
      });

      const nodePosition = d3.select(nodes[i]).node()?.getBoundingClientRect();
      let left = nodePosition.x + nodePosition.width / 2;

      // if user is hovering over the last bar
      if (left + 200 >= document.documentElement.clientWidth) {
        left = nodePosition.x + nodePosition.width / 2 - 200;
      }

      const scrollTop =
        window.pageYOffset !== undefined
          ? window.pageYOffset
          : (
              document.documentElement ||
              document.body.parentNode ||
              document.body
            ).scrollTop;
      const top = nodePosition.y + scrollTop;
      const toolTipHeight = d3
        .select('.toolTip')
        .node()
        ?.getBoundingClientRect().height;

      tooltip.current
        .html(
          `
            <div>${getLabel(d.label, 'tooltip')}</div>
            <div style="color: #0E2647;" class="mt-2 leading-5 text-base">
              <span class="font-semibold">
                ${
                  get(d, 'metricType', null)
                    ? getFormattedKpiValue({
                        value: d.value,
                        metricType: get(d, 'metricType', null)
                      })
                    : numberWithCommas(d.value)
                }
              </span>
            </div>
          `
        )
        .style('opacity', 1)
        .style('left', left + 'px')
        .style('top', top - toolTipHeight + 5 + 'px');
    },
    [getLabel]
  );

  const hideTooltip = useCallback(() => {
    const nodes = d3.select(chartRef.current).selectAll('.bar').nodes();
    nodes.forEach((node) => {
      d3.select(node).attr('class', 'bar');
    });
    tooltip.current.style('opacity', 0);
  }, []);

  const drawChart = useCallback(() => {
    const arc = (r, sign) =>
      r
        ? `a${r * sign[0]},${r * sign[1]} 0 0 1 ${r * sign[2]},${r * sign[3]}`
        : '';

    function roundedRect(x, y, width, height, r) {
      const R = [
        Math.min(r[0], height, width),
        Math.min(r[1], height, width),
        Math.min(r[2], height, width),
        Math.min(r[3], height, width)
      ];

      return `M${x + R[0]},${y}h${width - R[0] - R[1]}${arc(
        R[1],
        [1, 1, 1, 1]
      )}v${height - R[1] - R[2]}${arc(R[2], [1, 1, -1, 1])}h${
        -width + R[2] + R[3]
      }${arc(R[3], [1, 1, -1, -1])}v${-height + R[3] + R[0]}${arc(
        R[0],
        [1, 1, 1, -1]
      )}z`;
    }

    const availableWidth = d3
      .select(chartRef.current)
      .node()
      ?.getBoundingClientRect().width;
    d3.select(chartRef.current)
      .html('')
      .append('svg')
      .attr('width', availableWidth)
      .attr('height', widgetHeight || 300)
      .attr('id', `chart-${title}`);
    const svg = d3.select(`#chart-${title}`);
    const max = getMaxYpoint(
      Math.max(...chartData.map((elem) => parseInt(elem.value)))
    );
    const margin = {
      top: 10,
      right: 0,
      bottom: 30,
      left: getBarChartLeftMargin(max)
    };
    const width = +svg.attr('width') - margin.left - margin.right;
    const height = +svg.attr('height') - margin.top - margin.bottom;

    const minBarHeight = 0.05 * height;

    tooltip.current = d3
      .select(chartRef.current)
      .append('div')
      .attr('class', 'toolTip')
      .style('opacity', 0)
      .style('transition', '0.5s');

    const xScale = d3
      .scaleBand()
      .rangeRound([0, width])
      .paddingOuter(0.15)
      .paddingInner(0.1)
      .domain(chartData.slice(0, BAR_COUNT[cardSize]).map((d) => d.label));

    const yScale = d3.scaleLinear().rangeRound([height, 0]).domain([0, max]);

    const yAxisGrid = d3
      .axisLeft(yScale)
      .tickSize(-width)
      .tickFormat('')
      .ticks(5);

    const g = svg
      .append('g')
      .attr('transform', `translate(${margin.left},${margin.top})`);

    g.append('g')
      .attr('class', 'y axis-grid')
      .call(yAxisGrid)
      .selectAll('line')
      .attr('stroke', '#E7E9ED');

    g.append('g')
      .attr('class', 'axis axis--x')
      .attr('transform', `translate(0,${height})`)
      .call(
        d3.axisBottom(xScale).tickFormat((d) => {
          return getLabel(d);
        })
      );

    g.append('g')
      .attr('class', 'axis axis--y')
      .call(
        d3
          .axisLeft(yScale)
          .tickFormat((d) => {
            return d;
          })
          .ticks(5)
      );

    const bars = g
      .selectAll('.bar')
      .data(chartData.slice(0, BAR_COUNT[cardSize]))
      .enter()
      .append('g')
      .attr('class', 'bar')
      .on('mousemove', (d, i) => {
        showTooltip(d, i);
      })
      .on('mouseout', () => {
        hideTooltip();
      });

    bars
      .append('path')
      .attr('d', (d) =>
        roundedRect(
          xScale(d.label),
          height - yScale(d.value) > minBarHeight
            ? yScale(d.value)
            : height - minBarHeight,
          xScale.bandwidth(),
          yScale(0) - yScale(d.value),
          [5, 5, 0, 0]
        )
      )
      .style('fill', (d) => {
        return d.color ? d.color : CHART_COLOR_1;
      });
    // .append('rect')
    // .attr('class', () => {
    //   return 'bar';
    // })
    // .attr('fill', (d) => {
    //   return d.color ? d.color : CHART_COLOR_1;
    // })
    // .attr('x', (d) => xScale(d.label))
    // .attr('y', (d) => {
    //   return height - yScale(d.value) > minBarHeight
    //     ? yScale(d.value)
    //     : height - minBarHeight;
    // })
    // .attr('width', xScale.bandwidth())
    // .attr('height', (d) => {
    //   return height - yScale(d.value) > minBarHeight
    //     ? height - yScale(d.value)
    //     : minBarHeight;
    // })

    bars
      .append('text')
      .text((d) => numberWithCommas(Number(d?.value?.toFixed(2) || 0)))
      .attr('x', (d) => xScale(d.label) + xScale.bandwidth() / 2)
      .attr('y', (d) => {
        const yValue =
          height - yScale(d.value) > minBarHeight
            ? yScale(d.value)
            : height - minBarHeight;
        return yValue - 5 > 0 ? yValue - 5 : yValue + 15;
      })
      .attr('class', 'bar-chart-label')
      .attr('text-anchor', 'middle');

    // g.selectAll(".bar")
    //   .transition()
    //   .duration(500)
    //   .attr("y", function (d) { console.log(yScale(d.value)); return yScale(d.value); })
    //   .attr("height", function (d) { return height - yScale(d.value); })
    //   .delay(function (d, i) { console.log(i); return (i * 1000) })

    d3.select(chartRef.current)
      .select('.axis.axis--x')
      .selectAll('.tick')
      .select('text')
      .attr('dy', '16px');
  }, [
    chartData,
    showTooltip,
    hideTooltip,
    title,
    widgetHeight,
    getLabel,
    cardSize
  ]);

  const displayChart = useCallback(() => {
    drawChart();
  }, [drawChart]);

  useEffect(() => {
    window.addEventListener('resize', () =>
      checkForWindowSizeChange(displayChart)
    );
    return () => {
      window.removeEventListener('resize', () =>
        checkForWindowSizeChange(displayChart)
      );
    };
  }, [displayChart]);

  useEffect(() => {
    displayChart();
  }, [displayChart, cardSize]);

  let legendColors = {};

  if (queries && queries.length > 1) {
    const appliedColors = generateColors(queries.length);
    legendColors = queries.map((_, index) => {
      return appliedColors[index];
    });
  }

  return (
    <div className='w-full bar-chart'>
      {queries && queries.length > 1 && section === DASHBOARD_WIDGET_SECTION ? (
        <TopLegends
          cardSize={cardSize}
          showAllLegends={true}
          showFullLengthLegends={true}
          legends={queries}
          colors={legendColors}
        />
      ) : null}
      <div ref={chartRef} className={styles.ungroupedChart}></div>
      {queries &&
      queries.length > 1 &&
      (section === REPORT_SECTION || section === DASHBOARD_MODAL) ? (
        <div className='mt-4'>
          <TopLegends
            cardSize={cardSize}
            showAllLegends={true}
            showFullLengthLegends={true}
            legends={queries}
            colors={legendColors}
          />
        </div>
      ) : null}
    </div>
  );
}

export default memo(BarChart);
