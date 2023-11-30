import React, { useRef, useCallback, useEffect, useMemo } from 'react';
import ReactDOMServer from 'react-dom/server';
import * as d3 from 'd3';
import styles from './styles.module.scss';
import {
  BAR_CHART_XAXIS_TICK_LENGTH,
  DASHBOARD_MODAL,
  FUNNELS_COUNT,
  FUNNEL_CHART_MARGIN
} from '../../../../utils/constants';
import {
  generateColors,
  calculatePercentage,
  formatCount
} from '../../../../utils/dataFormatter';
import { getOverAllDuration, getStepDuration } from '../utils';
import {
  Text,
  SVG,
  Number as NumFormat
} from '../../../../components/factorsComponents';
import LegendsCircle from '../../../../styles/components/LegendsCircle';
// import truncateURL from 'Utils/truncateURL';

function Chart({
  chartData,
  title = 'chart',
  cardSize = 1,
  arrayMapper,
  height: widgetHeight,
  section,
  durations,
  showXAxisLabels = true,
  showYAxisLabels = true,
  margin = FUNNEL_CHART_MARGIN,
  showStripes = false
}) {
  const chartRef = useRef(null);
  const tooltipRef = useRef(null);

  const renderedData = chartData.slice(0, FUNNELS_COUNT[cardSize]);
  const colors = generateColors(renderedData.length);

  const overallDuration = useMemo(() => {
    return getOverAllDuration(durations);
  }, [durations]);

  const drawChart = useCallback(() => {
    const availableWidth = d3
      .select(chartRef.current)
      .node()
      ?.getBoundingClientRect().width;
    const tooltip = d3.select(tooltipRef.current);
    d3.select(chartRef.current)
      .html('')
      .append('svg')
      .attr('width', availableWidth)
      .attr('height', widgetHeight || 420)
      .attr('id', `funnel-ungrouped-svg-${title}`);
    const svg = d3.select(`#funnel-ungrouped-svg-${title}`),
      width = +svg.attr('width') - margin.left,
      height = +svg.attr('height') - margin.top - margin.bottom,
      g = svg
        .append('g')
        .attr('transform', 'translate(' + margin.left + ',' + margin.top + ')');

    const x = d3.scaleBand().rangeRound([0, width]).padding(0.15);

    const y = d3.scaleLinear().rangeRound([height, 0]);

    const showTooltip = (d, index) => {
      const label = arrayMapper.find(
        (elem) => elem.mapper === d.event
      ).displayName;

      let padY = index ? 200 : 100;
      let padX = 10;

      if (section === DASHBOARD_MODAL) {
        padY += 25;
        padX = -10;
      }

      let stepTime;

      if (index) {
        stepTime = getStepDuration(durations, index - 1, index);
      }

      tooltip
        .style('left', d3.event.pageX + padX + 'px')
        .style('top', d3.event.pageY - padY + 'px')
        .style('display', 'inline-block')
        .html(
          ReactDOMServer.renderToString(
            <>
              <Text
                type='title'
                weight='medium'
                color='grey-2'
                lineHeight='small'
                extraClass='text-xs mb-0'
              >
                {label}
              </Text>
              <div
                className={`flex items-center mt-2 ${
                  index ? 'compareElem' : ''
                }`}
              >
                <LegendsCircle extraClass='mr-1' color={colors[index]} />
                <Text
                  extraClass='mr-1 mb-0 text-base'
                  lineHeight='medium'
                  type='title'
                  weight='bold'
                >
                  <NumFormat number={d.netCount} />
                </Text>
                <Text
                  extraClass='mr-1 mb-0 text-base'
                  lineHeight='medium'
                  color='grey'
                  type='title'
                  weight='medium'
                >
                  ({d.value}%)
                </Text>
              </div>
              {index ? (
                <div className='pt-4'>
                  <div className='flex flex-col'>
                    <Text
                      type='title'
                      color='grey'
                      weight='bold'
                      extraClass='text-xs mb-0'
                      lineHeight='small'
                    >
                      From previous step:
                    </Text>
                  </div>
                  <div className='flex justify-between items-center mt-2'>
                    <div className='flex flex-col items-start'>
                      <div className='flex items-center'>
                        <SVG name='clock' fill='#8692A3' />
                        <Text
                          type='title'
                          color='grey-2'
                          weight='medium'
                          extraClass='text-xs mb-0 ml-1 mt-1'
                          lineHeight='1'
                        >
                          {stepTime}
                        </Text>
                      </div>
                      <Text
                        type='title'
                        color='grey'
                        weight='medium'
                        extraClass='text-xs mb-0 mt-1'
                      >
                        TIME TAKEN
                      </Text>
                    </div>
                    <div className='flex flex-col items-start'>
                      <div className='flex items-center'>
                        <Text
                          type='title'
                          color='grey-2'
                          weight='medium'
                          extraClass='text-xs mb-0 mr-1'
                          lineHeight='1'
                        >
                          {formatCount(
                            100 -
                              calculatePercentage(
                                chartData[index].netCount,
                                chartData[index - 1].netCount
                              ),
                            1
                          )}
                          %
                        </Text>
                        <SVG name='dropoff' fill='#8692A3' />
                      </div>
                      <Text
                        type='title'
                        color='grey'
                        weight='medium'
                        extraClass='text-xs mb-0 mt-1'
                      >
                        DROP-OFF
                      </Text>
                    </div>
                  </div>
                </div>
              ) : null}
            </>
          )
        );
    };

    const hideTooltip = (d) => {
      tooltip.style('display', 'none');
    };

    x.domain(
      renderedData.map(function (d) {
        return d.event;
      })
    );
    y.domain([
      0,
      d3.max(renderedData, function (d) {
        return Number(d.value);
      })
    ]);
    const yAxisGrid = d3.axisLeft(y).tickSize(-width).tickFormat('').ticks(5);

    const infoDivHeight = 40;
    const infoDivWidth = 56;
    const overallInfoDivWidth = 75;
    const overallInfoDivHeight = 75;

    g.append('g')
      .attr('class', 'y-axis-grid')
      .call(yAxisGrid)
      .selectAll('line')
      .attr('stroke', '#E7E9ED');

    if (showXAxisLabels) {
      g.append('g')
        .attr('class', 'x-axis')
        .attr('transform', 'translate(0,' + height + ')')
        .call(
          d3.axisBottom(x).tickFormat((d) => {
            const label = arrayMapper.find(
              (elem) => elem.mapper === d
            ).displayName;
            // const urlTruncatedlabel = truncateURL(label);
            const urlTruncatedlabel = label;
            if (
              urlTruncatedlabel.length > BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
            ) {
              return (
                urlTruncatedlabel.slice(
                  0,
                  BAR_CHART_XAXIS_TICK_LENGTH[cardSize]
                ) + '...'
              );
            }
            return urlTruncatedlabel;
          })
        );
    }

    if (showYAxisLabels) {
      g.append('g')
        .attr('class', 'y-axis')
        .call(
          d3
            .axisLeft(y)
            .tickFormat((d) => {
              return d + '%';
            })
            .ticks(5)
        )
        .append('text')
        .attr('fill', '#000')
        .attr('transform', 'rotate(-90)')
        .attr('y', 6)
        .attr('dy', '0.71em');
    }

    if (showStripes) {
      g.append('defs')
        .append('pattern')
        .attr('id', `pattern-${title}`)
        .attr('patternUnits', 'userSpaceOnUse')
        .attr('width', 4)
        .attr('height', 4)
        .append('path')
        .attr('d', 'M-1,1 l2,-2 M0,4 l4,-4 M3,5 l2,-2')
        .attr('stroke', '#ffffff')
        .attr('stroke-width', 0.5);
    }

    g.selectAll('.bar')
      .data(renderedData)
      .enter()
      .append('rect')
      .attr('class', 'bar')
      .attr('fill', (_, i) => {
        return colors[i];
      })
      .attr('x', function (d) {
        return x(d.event);
      })
      .attr('y', function (d) {
        return y(Number(d.value));
      })
      .attr('width', x.bandwidth())
      .attr('height', function (d) {
        return height - y(Number(d.value));
      })
      .on('mousemove', function (d, index) {
        showTooltip(d, index);
      })
      .on('mouseout', function () {
        hideTooltip();
      });

    if (showStripes) {
      g.selectAll('.pattern')
        .data(renderedData)
        .enter()
        .append('rect')
        .attr('class', 'pattern')
        .attr('x', function (d) {
          return x(d.event);
        })
        .attr('y', function (d) {
          return y(Number(d.value));
        })
        .attr('width', x.bandwidth())
        .attr('height', function (d) {
          return height - y(Number(d.value));
        })
        .attr('fill', `url(#pattern-${title})`)
        .on('mousemove', function (d, index) {
          showTooltip(d, index);
        })
        .on('mouseout', function () {
          hideTooltip();
        });
    }

    g.selectAll('.area')
      .data(renderedData)
      .enter()
      .append('polygon')
      .attr('fill', (_, i) => {
        return `url(#funnel-ungrouped-gradient-${title}-${i})`;
      })
      .attr('class', 'area')
      .attr('points', (d, i, nodes) => {
        if (i > 0) {
          const dPrev = d3.select(nodes[i - 1]).datum();

          const x1 = x(d.event);
          const y1 = y(Number(d.value));

          const x2 = x1;
          const y2 = y(Number(dPrev.value));

          const x3 = x(d.event) + x.bandwidth();
          const y3 = y2;

          const x4 = x3;
          const y4 = y1;

          return `${x1},${y1} ${x2},${y2} ${x3},${y3} ${x4},${y4} ${x1},${y1}`;
        }
      })
      .on('mousemove', function (d, index) {
        showTooltip(d, index);
      })
      .on('mouseout', function (d) {
        hideTooltip();
      });

    svg
      .append('foreignObject')
      .attr('x', () => {
        return width + margin.left - 80;
      })
      .attr('y', (d) => {
        return y(100) + margin.top;
      })
      .attr('width', overallInfoDivWidth)
      .attr('height', overallInfoDivHeight)
      .append('xhtml:div')
      .attr(
        'class',
        'overallInfoDiv flex flex-col flex-1 pt-2 justify-between items-center'
      )
      .html(
        ReactDOMServer.renderToString(
          <>
            <Text
              type='title'
              color='grey-2'
              weight='bold'
              extraClass='mb-0 percent'
            >
              {chartData[chartData.length - 1].value}%
            </Text>
            <Text
              type='title'
              color='grey'
              weight='medium'
              extraClass='mb-0 duration'
            >
              {overallDuration}
            </Text>
            <Text
              type='title'
              extraClass='label w-full flex items-center justify-center mb-0'
              weight='bold'
              color='white'
            >
              OVERALL
            </Text>
          </>
        )
      );

    if (cardSize !== 2) {
      g.selectAll('.infoDiv')
        .data(renderedData)
        .enter()
        .append('foreignObject')
        .attr('x', function (d) {
          return x(d.event) + x.bandwidth() / 2 - infoDivWidth / 2;
        })
        .attr('y', (d) => {
          if (y(0) - y(Number(d.value)) > infoDivHeight / 2) {
            return y(Number(d.value)) - infoDivHeight / 2;
          } else {
            return y(Number(d.value)) - infoDivHeight;
          }
        })
        .attr('width', infoDivWidth)
        .attr('height', infoDivHeight)
        .append('xhtml:div')
        .attr(
          'class',
          'infoDiv flex flex-col items-center h-full justify-center bg-white'
        )
        .html((d, index) => {
          return ReactDOMServer.renderToString(
            <>
              <Text
                type='title'
                weight='medium'
                color='grey-2'
                extraClass='text-xs mb-0 percent'
              >
                {index
                  ? calculatePercentage(
                      chartData[index].netCount,
                      chartData[index - 1].netCount,
                      1
                    ) + '%'
                  : '100%'}
              </Text>
              <Text
                type='title'
                weight='medium'
                color='grey'
                extraClass='text-xs mb-0 count'
              >
                <NumFormat shortHand={true} number={d.netCount} />
              </Text>
            </>
          );
        })
        .on('mousemove', function (d, index) {
          showTooltip(d, index);
        })
        .on('mouseout', function (d) {
          hideTooltip();
        });
    }
  }, [
    arrayMapper,
    cardSize,
    chartData,
    title,
    widgetHeight,
    renderedData,
    colors,
    overallDuration,
    section,
    durations,
    margin,
    showXAxisLabels,
    showYAxisLabels,
    showStripes
  ]);

  useEffect(() => {
    drawChart();
  }, [drawChart, cardSize]);

  return (
    <div className='w-full'>
      <div ref={chartRef} className={styles.ungroupedChart}></div>
      <svg width='0' height='0'>
        <defs>
          {colors.map((color, index) => {
            return (
              <linearGradient
                key={index}
                id={`funnel-ungrouped-gradient-${title}-${index}`}
                x1='.5'
                x2='.5'
                y2='1'
              >
                <stop stopColor={color} stopOpacity='0.5' />
                <stop offset='1' stopColor={color} stopOpacity='0.1' />
              </linearGradient>
            );
          })}
        </defs>
      </svg>
      <div ref={tooltipRef} className={styles.ungroupedFunnelTooltip}></div>
    </div>
  );
}

export default Chart;
