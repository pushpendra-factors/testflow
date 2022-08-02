import React, { useRef, useEffect, useCallback } from 'react';
import * as d3 from 'd3';
import ReactDOMServer from 'react-dom/server';
import styles from './styles.module.scss';
import {
  REPORT_SECTION,
  DASHBOARD_WIDGET_SECTION,
  BAR_CHART_XAXIS_TICK_LENGTH,
  DASHBOARD_MODAL,
  FUNNELS_COUNT,
} from '../../utils/constants';
import TopLegends from './TopLegends';
import { getBarChartLeftMargin, getMaxYpoint } from '../BarChart/utils';
import { Text, Number as NumFormat } from '../factorsComponents';

function GroupedBarChart({
  chartData,
  colors,
  metricsData,
  method1,
  method2,
  height: widgetHeight,
  section,
  title = 'chart',
  cardSize = 1,
  allValues,
  legends,
  tooltipTitle,
  attributionMethodsMapper,
}) {
  const renderedData = chartData.slice(0, FUNNELS_COUNT[cardSize]);
  const keys = Object.keys(renderedData[0]).slice(1);
  const chartRef = useRef(null);
  const tooltipRef = useRef(null);

  const drawChart = useCallback(() => {
    const max = getMaxYpoint(Math.max(...allValues));
    const availableWidth = d3
      .select(chartRef.current)
      .node()
      ?.getBoundingClientRect().width;
    const tooltip = d3.select(tooltipRef.current);

    const showTooltip = (data) => {
      const row = metricsData.find((elem) => elem.category === data.group);
      let padY = 300;
      let padX = 10;

      if (section === DASHBOARD_MODAL) {
        padY += 25;
        padX = -10;
      }
      tooltip
        .style('left', d3.event.pageX + padX + 'px')
        .style('top', d3.event.pageY - padY + 'px')
        .style('display', 'inline-block')
        .html(
          ReactDOMServer.renderToString(
            <>
              <div className='pb-3 groupInfo'>
                <Text
                  type='title'
                  weight='bold'
                  color='grey-8'
                  extraClass='mb-0'
                >
                  {data.group}
                </Text>
              </div>
              <div className='py-3 horizontal-border'>
                <Text
                  type='title'
                  weight='bold'
                  color='grey'
                  level={8}
                  extraClass='uppercase leading-4 mb-0'
                >
                  {tooltipTitle}
                </Text>
                <div className='flex justify-between mt-2'>
                  <Text
                    style={{ color: colors[0] }}
                    type='title'
                    extraClass='mb-0 leading-4'
                    weight='bold'
                    level={8}
                  >
                    {attributionMethodsMapper[method1]}
                  </Text>
                  <Text color='grey-2' type='title' extraClass='mb-0'>
                    <NumFormat number={data[method1]} />
                  </Text>
                </div>
                <div className='flex justify-between mt-2'>
                  <Text
                    style={{ color: colors[1] }}
                    type='title'
                    extraClass='mb-0 leading-4'
                    weight='bold'
                    level={8}
                  >
                    {attributionMethodsMapper[method2]}
                  </Text>
                  <Text color='grey-2' type='title' extraClass='mb-0'>
                    <NumFormat number={data[method2]} />
                  </Text>
                </div>
              </div>
              <div className='pt-3'>
                <div className='flex justify-between'>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    Impressions
                  </Text>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    <NumFormat number={row['Impressions']} />
                  </Text>
                </div>
                <div className='flex justify-between mt-2'>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    Clicks
                  </Text>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    <NumFormat number={row['Clicks']} />
                  </Text>
                </div>
                <div className='flex justify-between mt-2'>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    Spend
                  </Text>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    <NumFormat number={row['Spend']} />
                  </Text>
                </div>
                <div className='flex justify-between mt-2'>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    Sessions
                  </Text>
                  <Text
                    color='grey-2'
                    type='title'
                    extraClass='mb-0 leading-4'
                    level={8}
                  >
                    <NumFormat number={row['Sessions']} />
                  </Text>
                </div>
              </div>
            </>
          )
        );
    };

    const hideTooltip = (d) => {
      tooltip.style('display', 'none');
    };

    d3.select(chartRef.current)
      .html('')
      .append('svg')
      .attr('width', availableWidth)
      .attr('height', widgetHeight || 420)
      .attr('id', `funnel-grouped-svg-${title}`);
    const svg = d3.select(`#funnel-grouped-svg-${title}`),
      margin = {
        top: 10,
        right: 20,
        bottom: 30,
        left: getBarChartLeftMargin(max),
      },
      width = +svg.attr('width') - margin.left - margin.right,
      height = +svg.attr('height') - margin.top - margin.bottom,
      g = svg
        .append('g')
        .attr('transform', 'translate(' + margin.left + ',' + margin.top + ')');

    const x0 = d3.scaleBand().rangeRound([0, width]).padding(0.25);
    const x1 = d3.scaleBand().paddingInner(0.05);
    const y = d3.scaleLinear().rangeRound([height, 0]);
    const z = d3.scaleOrdinal().range(colors);

    x0.domain(
      renderedData.map(function (d) {
        return d.name;
      })
    );
    x1.domain(keys).rangeRound([0, x0.bandwidth()]);
    y.domain([0, max]).nice();

    const yAxisGrid = d3.axisLeft(y).tickSize(-width).tickFormat('').ticks(5);

    g.append('g')
      .attr('class', 'y-axis-grid')
      .call(yAxisGrid)
      .selectAll('line')
      .attr('stroke', '#E7E9ED');

    const base = g
      .append('g')
      .selectAll('g')
      .data(renderedData)
      .enter()
      .append('g')
      .attr('transform', function (d) {
        return 'translate(' + x0(d.name) + ',0)';
      });
    base
      .selectAll('rect')
      .data(function (d) {
        return keys.map(function (key) {
          return {
            key,
            value: Number(d[key]),
            group: d.name,
            [method1]: d[method1],
            [method2]: d[method2],
          };
        });
      })
      .enter()
      .append('rect')
      .attr('x', function (d) {
        return x1(d.key);
      })
      .attr('y', function (d) {
        return y(d.value);
      })
      .attr('width', x1.bandwidth())
      .attr('height', function (d) {
        return height - y(d.value);
      })
      .attr('fill', function (d) {
        return z(d.key);
      })
      .on('mousemove', (d) => {
        showTooltip(d);
      })
      .on('mouseout', () => {
        hideTooltip();
      });
    g.append('g')
      .attr('class', 'x-axis')
      .attr('transform', 'translate(0,' + height + ')')
      .call(
        d3.axisBottom(x0).tickFormat((d) => {
          let label;
          if (d.includes('$no_group')) {
            label = 'Overall';
          } else {
            label = d;
          }
          if (label.length > BAR_CHART_XAXIS_TICK_LENGTH[cardSize]) {
            return (
              label.substr(0, BAR_CHART_XAXIS_TICK_LENGTH[cardSize]) + '...'
            );
          }
          return label;
        })
      );

    g.append('g')
      .attr('class', 'y-axis')
      .call(
        d3
          .axisLeft(y)
          .tickFormat((d) => {
            return d;
          })
          .ticks(5)
          .tickSize(-width)
      );
  }, [
    tooltipTitle,
    allValues,
    cardSize,
    colors,
    keys,
    method1,
    method2,
    renderedData,
    metricsData,
    section,
    title,
    widgetHeight,
    attributionMethodsMapper,
  ]);

  useEffect(() => {
    drawChart();
  }, [drawChart, cardSize]);

  return (
    <div className='w-full'>
      {section === DASHBOARD_WIDGET_SECTION ? (
        <TopLegends
          showFullLengthLegends={cardSize === 1}
          legends={legends}
          cardSize={cardSize}
          colors={colors}
        />
      ) : null}
      <div ref={chartRef} className={styles.groupedChart}></div>
      {section === REPORT_SECTION || section === DASHBOARD_MODAL ? (
        <TopLegends
          showFullLengthLegends={cardSize === 1}
          legends={legends}
          cardSize={cardSize}
          colors={colors}
        />
      ) : null}
      <div ref={tooltipRef} className={styles.groupedChartTooltip}></div>
    </div>
  );
}

export default GroupedBarChart;
