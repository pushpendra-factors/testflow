import React, { useRef, useCallback, useEffect } from 'react';
import c3 from 'c3';
import * as d3 from 'd3';
import styles from './index.module.scss';
import { checkForWindowSizeChange, calculatePercentage, generateColors } from '../utils';

function GroupedChart({ eventsData, groups, chartData }) {

    const appliedColors = generateColors(chartData.length);
    const chartColors = {};
    chartData.forEach((elem, index) => {
        chartColors[elem[0]] = appliedColors[index];
    });
    
    const chartRef = useRef(null);

    const showConverionRates = useCallback(() => {
        let yGridLines = d3.select(chartRef.current).select('g.c3-ygrids').selectAll('line').nodes();
        let top, secondTop, height;
        let topGridLine = yGridLines[yGridLines.length - 1];
        let secondTopGridLine = yGridLines[yGridLines.length - 2];
        top = topGridLine.getBoundingClientRect().y;
        secondTop = secondTopGridLine.getBoundingClientRect().y;
        height = secondTop - top;
        const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;

        groups.forEach((elem) => {
            document.getElementById(`conversion-text-${elem.name}`).style.top = `${top + scrollTop}px`;
            document.getElementById(`conversion-text-${elem.name}`).style.height = `${height}px`;
        });

        d3.select(chartRef.current).select('g.c3-axis-x').selectAll('g.tick').nodes()
            .forEach((elem, index) => {
                let position = elem.getBoundingClientRect();
                document.getElementById(`conversion-text-${groups[index].name}`).style.left = `${position.x}px`;
                let width = document.getElementById(groups[index].name).getBoundingClientRect().x - position.x;
                document.getElementById(`conversion-text-${groups[index].name}`).style.width = `${width}px`;
            })
    }, [groups]);

    const showVerticalGridLines = useCallback(() => {
        let yGridLines = d3.select(chartRef.current).select('g.c3-ygrids').selectAll('line').nodes();
        let top, bottom, height;
        let topGridLine = yGridLines[yGridLines.length - 1];
        top = topGridLine.getBoundingClientRect().y;
        let bottomGridLine = yGridLines[0];
        bottom = bottomGridLine.getBoundingClientRect().y;
        height = bottom - top;
        let lastBarClassNmae = eventsData[eventsData.length - 1].name.split(" ").join("-"); //this is an issue if someone disables the last legend item. Will figure out something for this.
        const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
        d3.select(chartRef.current).select(`g.c3-shapes-${lastBarClassNmae}`).selectAll('path').nodes()
            .forEach((elem, index) => {
                let position = elem.getBoundingClientRect();
                let verticalLine = document.getElementById(groups[index].name);
                verticalLine.style.left = `${position.x + position.width - 1}px`;
                verticalLine.style.height = `${height}px`;
                verticalLine.style.top = `${top + scrollTop}px`;
            });
    }, [groups, eventsData]);


    const drawChart = useCallback(() => {
        c3.generate({
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

                    //blur all the bars
                    d3.select(chartRef.current)
                        .selectAll(`.c3-shapes`)
                        .selectAll('path')
                        .style('opacity', '0.5')

                    let id = elemData.name;
                    if (!id) id = elemData.id;

                    const searchedClass = `c3-target-${id.split(" ").join("-")}`;
                    let hoveredIndex;

                    //style previous bar

                    const bars = d3.select(chartRef.current)
                        .selectAll('.c3-chart-bar.c3-target')
                        .nodes()

                    bars
                        .forEach((node, index) => {
                            if (node.getAttribute('class').split(" ").indexOf(searchedClass) > -1) {
                                hoveredIndex = index;
                            }
                        })

                    if (hoveredIndex !== 0) {
                        d3.select(bars[hoveredIndex - 1]).select(`.c3-shape-${elemData.index}`).style('opacity', 1)
                    }

                    // style hovered bar
                    d3.select(chartRef.current)
                        .selectAll(`.c3-shapes-${id.split(" ").join('-')}`)
                        .selectAll('path')
                        .nodes()
                        .forEach((node, index) => {
                            if (index === elemData.index) {
                                d3.select(node).style('opacity', 1)
                            } else {
                                d3.select(node).style('opacity', 0.5)
                            }
                        })
                },
                onmouseout: () => {
                    d3.select(chartRef.current)
                        .selectAll(`.c3-shapes`)
                        .selectAll('path')
                        .style('opacity', '1')
                },
            },
            onrendered: () => {
                d3.select(chartRef.current).select(".c3-axis.c3-axis-x").selectAll('.tick').select('tspan').attr("dy", "16px");
            },
            legend: {
                padding: 8,
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
                        multilineMax: 3,
                    },
                    categories: groups
                        .filter(elem => elem.is_visible)
                        .map(elem => elem.name),
                },
                y: {
                    max: 100,
                    tick: {
                        values: [0, 20, 40, 60, 80, 100],
                        format: (d) => {
                            if (d) {
                                return d + '%';
                            } else {
                                return d
                            }

                        },
                    }
                }
            },
            tooltip: {
                grouped: false,
                contents: d => {
                    const group = groups[d[0].index].name;
                    const eventIndex = eventsData.findIndex(elem => elem.name === d[0].id);
                    const event = eventsData.find(elem => elem.name === d[0].id);
                    const eventWeightage = calculatePercentage(event.data[group], eventsData[0].data[group])
                    let eventsOutput;
                    if (!eventIndex) {
                        eventsOutput = (
                            `
                                <div class="flex justify-between my-2">
                                    <div class="font-bold" style="color:${event.color}">Event ${event.index}</div>
                                    <div>${event.data[group]} (${eventWeightage}%)</div>
                                </div>
                            `
                        );
                    } else {
                        const prevEvent = eventsData[eventIndex - 1];
                        const prevEventWeightage = calculatePercentage(prevEvent.data[group], eventsData[0].data[group])
                        const difference = calculatePercentage(prevEvent.data[group] - event.data[group], prevEvent.data[group]);
                        eventsOutput = (
                            `
                                <div class="my-2">
                                    <div class="flex justify-between">
                                        <div class="font-bold" style="color:${prevEvent.color}">Event ${prevEvent.index}</div>
                                        <div><span class="font-semibold">${prevEvent.data[group]}</span> (${prevEventWeightage}%)</div>
                                    </div>
                                    <div class="flex justify-between">
                                        <div class="font-bold" style="color:${event.color}">Event ${event.index}</div>
                                        <div><span class="font-semibold">${event.data[group]}</span> (${eventWeightage}%)</div>
                                    </div>
                                </div>
                                <hr />
                                <div class="my-2">
                                    <div>
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
                                <div class="my-2">
                                    <div class="font-black">${group}</div>
                                    <div>${groups[d[0].index].conversion_rate} Overall Conversion</div>
                                </div>
                                <hr />
                                ${eventsOutput}
                            </div>
                        `
                    )
                }
            },
            grid: {
                y: {
                    show: true,
                },
            },
        });
    }, [chartColors, chartData, eventsData, groups]);


    const displayChart = useCallback(() => {
        drawChart();
        showVerticalGridLines();
        showConverionRates();
    }, [drawChart, showVerticalGridLines, showConverionRates])

    useEffect(() => {
        window.addEventListener("resize", () => checkForWindowSizeChange(displayChart), false);
        return () => {
            window.removeEventListener("resize", () => checkForWindowSizeChange(displayChart), false);
        }
    }, [displayChart]);

    useEffect(() => {
        displayChart();
    }, [displayChart]);

    const visibleEvents = eventsData.filter(elem => elem.display);

    return (
        <div className="grouped-chart">
            {
                groups
                    .map(elem => {
                        return (
                            <div className={`absolute border-l border-solid ${styles.verticalGridLines}`} key={elem.name} id={elem.name}></div>
                        );
                    })
            }
            {
                groups
                    .map(elem => {
                        return (
                            <div style={{ transition: '2s' }} key={elem.name} id={`conversion-text-${elem.name}`} className="absolute leading-5 text-base flex justify-end pr-1">
                                <div style={{ fontSize: visibleEvents.length > 2 ? '18px' : '14px' }} className={styles.conversionText}>
                                    <div className="font-semibold flex justify-end">{elem.conversion_rate}</div>
                                    <div>Conversion</div>
                                </div>
                            </div>
                        );
                    })
            }
            <div className={styles.groupedChart} ref={chartRef} />
        </div>
    )
}

export default GroupedChart;