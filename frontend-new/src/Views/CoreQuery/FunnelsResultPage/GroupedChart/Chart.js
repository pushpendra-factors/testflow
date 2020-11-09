/* eslint-disable */
import React, { useRef, useCallback, useEffect } from 'react';
import c3 from 'c3';
import * as d3 from 'd3';
import styles from './index.module.scss';
import { checkForWindowSizeChange, calculatePercentage, generateColors } from '../utils';

function Chart({ eventsData, groups, chartData, eventsMapper, reverseEventsMapper, title = "chart" }) {

  // console.log(eventsData)
  // console.log(groups)
  // console.log(chartData);

  const appliedColors = generateColors(chartData.length);
  const chartColors = {};
  chartData.forEach((elem, index) => {
    chartColors[elem[0]] = appliedColors[index];
  });

  const chartRef = useRef(null);

  const showConverionRates = useCallback(() => {
    const yGridLines = d3.select(chartRef.current).select('g.c3-ygrids').selectAll('line').nodes();
    let top, secondTop, height;
    const topGridLine = yGridLines[yGridLines.length - 1];
    const secondTopGridLine = yGridLines[yGridLines.length - 2];
    top = topGridLine.getBoundingClientRect().y;
    secondTop = secondTopGridLine.getBoundingClientRect().y;
    height = secondTop - top;
    const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;

    groups
      .forEach((elem) => {
        document.getElementById(`${title}-conversion-text-${elem.name}`).style.top = `${top + scrollTop}px`;
        document.getElementById(`${title}-conversion-text-${elem.name}`).style.height = `${height}px`;
      });

    d3.select(chartRef.current).select('g.c3-axis-x').selectAll('g.tick').nodes()
      .forEach((elem, index) => {
        const position = elem.getBoundingClientRect();
        document.getElementById(`${title}-conversion-text-${groups[index].name}`).style.left = `${position.x}px`;
        const width = document.getElementById(`${title}-${groups[index].name}`).getBoundingClientRect().x - position.x;
        document.getElementById(`${title}-conversion-text-${groups[index].name}`).style.width = `${width}px`;
      });
  }, [groups]);

  const showVerticalGridLines = useCallback(() => {
    const yGridLines = d3.select(chartRef.current).select('g.c3-ygrids').selectAll('line').nodes();
    let top, bottom, height;
    const topGridLine = yGridLines[yGridLines.length - 1];
    top = topGridLine.getBoundingClientRect().y;
    const bottomGridLine = yGridLines[0];
    bottom = bottomGridLine.getBoundingClientRect().y;
    height = bottom - top;
    const lastBarClassNmae = eventsData[eventsData.length - 1].name.split(' ').join('-'); // this is an issue if someone disables the last legend item. Will figure out something for this.
    const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
    d3.select(chartRef.current).select(`g.c3-shapes-${lastBarClassNmae}`).selectAll('path').nodes()
      .forEach((elem, index) => {
        const position = elem.getBoundingClientRect();
        const verticalLine = document.getElementById(`${title}-${groups[index].name}`);
        verticalLine.style.left = `${position.x + position.width - 1}px`;
        verticalLine.style.height = `${height}px`;
        verticalLine.style.top = `${top + scrollTop}px`;
      });
  }, [groups, eventsData]);

  const drawChart = useCallback(() => {
    const chart = c3.generate({
      size: {
        height: 400
      },
      padding: {
        left: 40,
        bottom: 24
      },
      bindto: chartRef.current,
      data: {
        columns: chartData,
        type: 'bar',
        colors: chartColors,
        onmouseover: (elemData) => {
          // blur all the bars
          d3.select(chartRef.current)
            .selectAll('.c3-shapes')
            .selectAll('path')
            .style('opacity', '0.3');

          let id = elemData.name;
          if (!id) id = elemData.id;

          const searchedClass = `c3-target-${id.split(' ').join('-')}`;
          let hoveredIndex;

          // style previous bar

          const bars = d3.select(chartRef.current)
            .selectAll('.c3-chart-bar.c3-target')
            .nodes();

          bars
            .forEach((node, index) => {
              if (node.getAttribute('class').split(' ').indexOf(searchedClass) > -1) {
                hoveredIndex = index;
              }
            });

          if (hoveredIndex !== 0) {
            d3.select(bars[hoveredIndex - 1]).select(`.c3-shape-${elemData.index}`).style('opacity', 1);
          }

          // style hovered bar
          d3.select(chartRef.current)
            .selectAll(`.c3-shapes-${id.split(' ').join('-')}`)
            .selectAll('path')
            .nodes()
            .forEach((node, index) => {
              if (index === elemData.index) {
                d3.select(node).style('opacity', 1);
              } else {
                d3.select(node).style('opacity', 0.3);
              }
            });
        },
        onmouseout: () => {
          d3.select(chartRef.current)
            .selectAll('.c3-shapes')
            .selectAll('path')
            .style('opacity', '1');
        }
      },
      onrendered: () => {
        d3.select(chartRef.current).select('.c3-axis.c3-axis-x').selectAll('.tick').select('tspan').attr('dy', '16px');
      },
      legend: {
        show: false
      },
      transition: {
        duration: 1000
      },
      bar: {
        space: 0.05,
        width: {
          ratio: 0.7
        }
      },
      axis: {
        x: {
          type: 'category',
          tick: {
            multiline: true,
            multilineMax: 3
          },
          categories: groups
            .filter(elem => elem.is_visible)
            .map(elem => elem.name)
        },
        y: {
          max: 100,
          tick: {
            values: [0, 20, 40, 60, 80, 100],
            format: (d) => {
              if (d) {
                return d + '%';
              } else {
                return d;
              }
            }
          }
        }
      },
      tooltip: {
        grouped: false,

        position: (d) => {
          const bars = d3.select(chartRef.current).select(`.c3-bars.c3-bars-${d[0].id.split(' ').join('-')}`).selectAll('path').nodes();
          const nodePosition = d3.select(bars[d[0].index]).node().getBoundingClientRect();
          let left = (nodePosition.x + (nodePosition.width / 2));
          // if user is hovering over the last bar
          if (left + 200 >= (document.documentElement.clientWidth)) {
            left = (nodePosition.x + (nodePosition.width / 2)) - 200;
          }

          const top = nodePosition.y;
          const toolTipHeight = d3.select('.toolTip').node().getBoundingClientRect().height;

          return { top: top - toolTipHeight + 5, left };
        },

        contents: d => {
          const group = groups[d[0].index].name;
          const eventIndex = eventsData.findIndex(elem => elem.name === d[0].id);
          const event = eventsData.find(elem => elem.name === d[0].id);
          const eventWeightage = calculatePercentage(event.data[group], eventsData[0].data[group]);
          let eventsOutput;
          if (!eventIndex) {
            eventsOutput = (
              `
                                <div class="flex justify-between mt-2">
                                    <div class="font-semibold leading-4" style="color:${chartColors[event.name]}">Event ${event.index}</div>
                                    <div class="leading-4"><span class="font-semibold">${event.data[group]}</span> (${eventWeightage}%)</div>
                                </div>
                            `
            );
          } else {
            const prevEvent = eventsData[eventIndex - 1];
            const prevEventWeightage = calculatePercentage(prevEvent.data[group], eventsData[0].data[group]);
            const difference = calculatePercentage(prevEvent.data[group] - event.data[group], prevEvent.data[group]);
            eventsOutput = (
              `
                                <div class="my-2">
                                    <div class="flex justify-between">
                                        <div class="font-semibold leading-4" style="color:${chartColors[prevEvent.name]}">Event ${prevEvent.index}</div>
                                        <div class="leading-4"><span class="font-semibold">${prevEvent.data[group]}</span> (${prevEventWeightage}%)</div>
                                    </div>
                                    <div class="flex justify-between mt-2">
                                        <div class="font-semibold leading-4" style="color:${chartColors[event.name]}">Event ${event.index}</div>
                                        <div class="leading-4"><span class="font-semibold">${event.data[group]}</span> (${eventWeightage}%)</div>
                                    </div>
                                </div>
                                <hr />
                                <div class="mt-3 flex">
                                    <div class="mr-2">
                                        <svg width="17" height="17" viewBox="0 0 17 17" fill="none" xmlns="http://www.w3.org/2000/svg">
                                            <path fillRule="evenodd" clipRule="evenodd" d="M1.87727 0.574039C1.61198 0.0896421 1.00424 -0.08798 0.51984 0.177309C0.0354429 0.442598 -0.142179 1.05034 0.12311 1.53473L4.5343 9.58922C4.79208 10.0599 5.37545 10.2432 5.85612 10.0045L9.15537 8.36627L12.3311 13.4015L10.5548 14.4155C10.1709 14.6347 10.2394 15.2077 10.6641 15.3302L14.6511 16.4801C14.9164 16.5566 15.1935 16.4035 15.27 16.1382L16.4529 12.037C16.5773 11.6057 16.1144 11.2417 15.7246 11.4642L14.0697 12.409L10.3653 6.53552C10.0916 6.10167 9.53412 5.94521 9.07471 6.17333L5.82702 7.78599L1.87727 0.574039Z" fill="#8692A3"/>
                                        </svg>
                                    </div>
                                    <div>${difference}% drop from ${prevEvent.index}-${event.index}</div>
                                </div>
                            `
            );
          }
          return (
            `
                            <div class="toolTip">
                                <div style="font-size:14px;color:#08172B;" class="font-semibold leading-5">${group}</div>
                                <div class="mb-2">${groups[d[0].index].conversion_rate} Overall Conversion</div>
                                <hr />
                                ${eventsOutput}
                            </div>
                        `
          );
        }
      },
      grid: {
        y: {
          show: true
        }
      }
    });
    d3.select(chartRef.current).insert('div', '.chart').attr('class', 'legend flex flex-wrap justify-center items-center').selectAll('span')
      .data(Object.values(eventsMapper))
      .enter().append('span')
      .attr('data-id', function (id) { return id; })
      .html(function (id) {
        return `<div class="flex items-center cursor-pointer"><div style="background-color: ${chart.color(id)};width:16px;height:16px;border-radius:8px"></div>
        <div class="px-2">${reverseEventsMapper[id]}</div></div>`;
      })
      // .each(function (id) {
      //   d3.select(this).style('background-color', chart.color(id));
      // })
      .on('mouseover', function (id) {
        chart.focus(id);
      })
      .on('mouseout', function (id) {
        chart.revert();
      })
      .on('click', function (id) {
        chart.toggle(id);
      });
  }, [chartColors, chartData, eventsData, groups]);

  const displayChart = useCallback(() => {
    drawChart();
    // showVerticalGridLines();
    // showConverionRates();
  }, [drawChart, showVerticalGridLines, showConverionRates]);

  useEffect(() => {
    window.addEventListener('resize', () => checkForWindowSizeChange(displayChart), false);
    return () => {
      window.removeEventListener('resize', () => checkForWindowSizeChange(displayChart), false);
    };
  }, [displayChart]);

  useEffect(() => {
    displayChart();
  }, [displayChart]);

  // const visibleEvents = eventsData.filter(elem => elem.display);

  return (
    <div className="grouped-chart">
      {/* {
        groups
          .map(elem => {
            return (
              <div className={`absolute border-l border-solid ${styles.verticalGridLines}`} key={elem.name} id={`${title}-${elem.name}`}></div>
            );
          })
      }
      {
        groups
          .map(elem => {
            return (
              <div style={{ transition: '2s' }} key={elem.name} id={`${title}-conversion-text-${elem.name}`} className="absolute z-10 leading-5 text-base flex justify-end pr-1">
                <div style={{ fontSize: '14px' }} className={styles.conversionText}>
                  <div className="font-semibold flex justify-end">{elem.conversion_rate}</div>
                  <div>Conversion</div>
                </div>
              </div>
            );
          })
      } */}
      <div className={styles.groupedChart} ref={chartRef} />
    </div>
  );
}

export default React.memo(Chart);
