import React, { useRef, useCallback, useEffect } from 'react';
import * as d3 from 'd3';
import ReactDOMServer from 'react-dom/server';
import styles from './styles.module.scss';
import {
  calculatePercentage,
  generateColors,
  formatCount,
  formatDuration,
} from '../../../../utils/dataFormatter';
import {
  REPORT_SECTION,
  DASHBOARD_WIDGET_SECTION,
  DASHBOARD_MODAL,
  FUNNELS_COUNT,
  BAR_CHART_XAXIS_TICK_LENGTH,
} from '../../../../utils/constants';
import ChartLegends from './ChartLegends';
import {
  Text,
  SVG,
  Number as NumFormat,
} from '../../../../components/factorsComponents';
import LegendsCircle from '../../../../styles/components/LegendsCircle';

function Chart({
  eventsData,
  groups,
  title = 'chart',
  arrayMapper,
  height: widgetHeight,
  section,
  cardSize = 1,
  durations,
}) {
  const chartRef = useRef(null);
  const tooltipRef = useRef(null);
  const renderedData = groups.slice(0, FUNNELS_COUNT[cardSize]);
  const keys = arrayMapper
    .map((elem) => elem.mapper)
    .filter((elem) => Object.keys(renderedData[0]).indexOf(elem) > -1);
  const colors = generateColors(keys.length);

  const durationMetric = durations.metrics.find(
    (elem) => elem.title === 'MetaStepTimeInfo'
  );
  const firstEventIdx = durationMetric.headers.findIndex(
    (elem) => elem === 'step_0_1_time'
  );

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
      .attr('id', `funnel-grouped-svg-${title}`);
    const svg = d3.select(`#funnel-grouped-svg-${title}`),
      margin = {
        top: 10,
        right: 20,
        bottom: 30,
        left: 40,
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
    y.domain([
      0,
      d3.max(renderedData, function (d) {
        return d3.max(keys, function (key) {
          return Number(d[key]);
        });
      }),
    ]).nice();

    const yAxisGrid = d3.axisLeft(y).tickSize(-width).tickFormat('').ticks(5);

    const infoDivHeight = 45;
    const infoDivWidth = 75;

    const showTooltip = (data) => {
      const nonConvertedName = groups.find((g) => g.name === data.group)
        ?.nonConvertedName;
      const currGrp = groups.find((g) => g.name === data.group);
      const durationGrp = durationMetric.rows.find(
        (elem) => elem.slice(0, firstEventIdx).join(', ') === nonConvertedName
      );
      const firstEventData = eventsData[0];
      const currEventData = eventsData.find((elem) => elem.name === data.key);
      const prevEventData =
        currEventData.index > 1
          ? eventsData.find((elem) => elem.index === currEventData.index - 1)
          : null;

      let timeTaken;

      if (prevEventData) {
        const durationIdx = durationMetric.headers.findIndex(
          (elem) =>
            elem ===
            `step_${prevEventData.index - 1}_${prevEventData.index}_time`
        );
        timeTaken = durationGrp
          ? formatDuration(Number(durationGrp[durationIdx]))
          : '0s';
      }

      let padY = prevEventData ? 250 : 200;
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
                  {data.group.includes('$no_group') ? 'Overall' : data.group}
                </Text>
                <Text type='title' color='grey-2' extraClass='mb-0'>
                  {currGrp.value} Overall Conversion
                </Text>
              </div>
              <div className='pt-3'>
                {prevEventData ? (
                  <Text
                    type='title'
                    color='grey'
                    weight='bold'
                    extraClass='text-xs mb-0'
                    lineHeight='small'
                  >
                    Between steps:
                  </Text>
                ) : null}
                {prevEventData ? (
                  <div className={`flex items-center mt-1`}>
                    <LegendsCircle
                      extraClass='mr-1'
                      color={colors[prevEventData.index - 1]}
                    />
                    <Text
                      extraClass='mr-1 mb-0 text-base'
                      lineHeight='medium'
                      type='title'
                      weight='bold'
                    >
                      <NumFormat number={prevEventData.data[data.group]} />
                    </Text>
                    <Text
                      extraClass='mr-1 mb-0 text-base'
                      lineHeight='medium'
                      color='grey'
                      type='title'
                      weight='medium'
                    >
                      (
                      {calculatePercentage(
                        prevEventData.data[data.group],
                        firstEventData.data[data.group]
                      )}
                      %)
                    </Text>
                  </div>
                ) : null}
                <div className={`flex items-center mt-1`}>
                  <LegendsCircle
                    extraClass='mr-1'
                    color={colors[currEventData.index - 1]}
                  />
                  <Text
                    extraClass='mr-1 mb-0 text-base'
                    lineHeight='medium'
                    type='title'
                    weight='bold'
                  >
                    <NumFormat number={currEventData.data[data.group]} />
                  </Text>
                  <Text
                    extraClass='mr-1 mb-0 text-base'
                    lineHeight='medium'
                    color='grey'
                    type='title'
                    weight='medium'
                  >
                    (
                    {calculatePercentage(
                      currEventData.data[data.group],
                      firstEventData.data[data.group]
                    )}
                    %)
                  </Text>
                </div>
                {prevEventData ? (
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
                          {timeTaken}
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
                                currEventData.data[data.group],
                                prevEventData.data[data.group]
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
                ) : null}
              </div>
            </>
          )
        );
    };

    const hideTooltip = () => {
      tooltip.html(null);
      tooltip.style('display', 'none');
    };

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
          return { key: key, value: Number(d[key]), group: d.name };
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

    base
      .selectAll('.area')
      .data(function (d) {
        return keys.map(function (key) {
          return { key: key, value: Number(d[key]), group: d.name };
        });
      })
      .enter()
      .append('polygon')
      .attr('fill', (_, i) => {
        return `url(#funnel-grouped-gradient-${title}-${i})`;
      })
      .attr('class', 'area')
      .attr('points', (d, i, nodes) => {
        if (i > 0) {
          const dPrev = d3.select(nodes[i - 1]).datum();
          const X1 = x1(d.key);
          const Y1 = y(Number(d.value));

          const X2 = X1;
          const Y2 = y(Number(dPrev.value));

          const X3 = x1(d.key) + x1.bandwidth();
          const Y3 = Y2;

          const X4 = X3;
          const Y4 = Y1;

          return `${X1},${Y1} ${X2},${Y2} ${X3},${Y3} ${X4},${Y4} ${X1},${Y1}`;
        }
      })
      .on('mousemove', (d) => {
        showTooltip(d);
      })
      .on('mouseout', () => {
        hideTooltip();
      });

    if (cardSize !== 2) {
      g.selectAll('.infoDiv')
        .data(renderedData)
        .enter()
        .append('foreignObject')
        .attr('x', function (d, index) {
          return x0(d.name) + x0.bandwidth() - infoDivWidth;
        })
        .attr('y', 1)
        .attr('width', infoDivWidth)
        .attr('height', infoDivHeight)
        .append('xhtml:div')
        .attr(
          'class',
          'infoDiv flex flex-col items-center h-full justify-center bg-white w-full'
        )
        .html((d) => {
          const nonConvertedName = groups.find((g) => g.name === d.name)
            ?.nonConvertedName;
          const durationGrp = durationMetric.rows.find(
            (elem) =>
              elem.slice(0, firstEventIdx).join(', ') === nonConvertedName
          );
          const durationVals = durationGrp
            ? durationGrp.slice(firstEventIdx)
            : [];
          let total = 0;
          durationVals.forEach((val) => {
            total += Number(val);
          });
          return ReactDOMServer.renderToString(
            <>
              <Text
                type='title'
                weight='medium'
                color='grey-2'
                extraClass='text-xs mb-0 percent'
              >
                {d[`event${keys.length}`]}%
              </Text>
              <Text
                type='title'
                weight='medium'
                color='grey'
                extraClass='text-xs mb-0 count'
              >
                {formatDuration(total)}
              </Text>
            </>
          );
        });
    }

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
            return d + '%';
          })
          .ticks(5)
      );
  }, [
    colors,
    keys,
    renderedData,
    title,
    widgetHeight,
    eventsData,
    groups,
    cardSize,
    section,
    durationMetric.headers,
    durationMetric.rows,
    firstEventIdx,
  ]);

  useEffect(() => {
    drawChart();
  }, [drawChart, cardSize]);

  return (
    <div className='w-full'>
      {section === DASHBOARD_WIDGET_SECTION ? (
        <ChartLegends
          colors={colors}
          legends={keys}
          arrayMapper={arrayMapper}
          cardSize={cardSize}
          section={section}
        />
      ) : null}
      <div ref={chartRef} className={styles.groupedChart}></div>
      <svg width='0' height='0'>
        <defs>
          {colors.map((color, index) => {
            return (
              <linearGradient
                key={index}
                id={`funnel-grouped-gradient-${title}-${index}`}
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
      <div ref={tooltipRef} className={styles.groupedFunnelTooltip}></div>
      {section === REPORT_SECTION || section === DASHBOARD_MODAL ? (
        <ChartLegends
          colors={colors}
          legends={keys}
          arrayMapper={arrayMapper}
          cardSize={cardSize}
          section={section}
        />
      ) : null}
    </div>
  );
}

export default React.memo(Chart);
