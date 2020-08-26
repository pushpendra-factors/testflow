import React, { useRef, useCallback, useEffect } from 'react';
import c3 from 'c3';
import * as d3 from 'd3';
import styles from './index.module.scss';

function GroupedChart({ eventsData, groups, chartData, chartColors }) {
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
            bindto: chartRef.current,
            data: {
                columns: chartData,
                type: 'bar',
                colors: chartColors,
            },
            transition: {
                duration: 1000
            },
            bar: {
                space: 0.05,
            },
            axis: {
                x: {
                    type: 'category',
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
                position: function (data, width, height, element) {
                    const top = d3.mouse(element)[1] - 100;
                    const left = d3.mouse(element)[0] + 80;
                    return { top, left }
                },
                contents: d => {
                    let group = groups[d[0].index].name;
                    let eventIndex = eventsData.findIndex(elem => elem.name === d[0].id);
                    let event = eventsData.find(elem => elem.name === d[0].id);
                    let eventsOutput;
                    if (!eventIndex) {
                        eventsOutput = (
                            `
                                <div class="flex justify-between my-2">
                                    <div class="font-bold" style="color:${event.color}">Event ${event.index}</div>
                                    <div>${3}</div>
                                </div>
                            `
                        );
                    } else {
                        let prevEvent = eventsData[eventIndex - 1];
                        eventsOutput = (
                            `
                                <div class="my-2">
                                    <div class="flex justify-between">
                                        <div class="font-bold" style="color:${prevEvent.color}">Event ${prevEvent.index}</div>
                                        <div>${3}</div>
                                    </div>
                                    <div class="flex justify-between">
                                        <div class="font-bold" style="color:${event.color}">Event ${event.index}</div>
                                        <div>${3}</div>
                                    </div>
                                </div>
                                <hr />
                                <div class="my-2">
                                    <div>
                                        <svg width="17" height="17" viewBox="0 0 17 17" fill="none" xmlns="http://www.w3.org/2000/svg">
                                            <path fillRule="evenodd" clipRule="evenodd" d="M1.87727 0.574039C1.61198 0.0896421 1.00424 -0.08798 0.51984 0.177309C0.0354429 0.442598 -0.142179 1.05034 0.12311 1.53473L4.5343 9.58922C4.79208 10.0599 5.37545 10.2432 5.85612 10.0045L9.15537 8.36627L12.3311 13.4015L10.5548 14.4155C10.1709 14.6347 10.2394 15.2077 10.6641 15.3302L14.6511 16.4801C14.9164 16.5566 15.1935 16.4035 15.27 16.1382L16.4529 12.037C16.5773 11.6057 16.1144 11.2417 15.7246 11.4642L14.0697 12.409L10.3653 6.53552C10.0916 6.10167 9.53412 5.94521 9.07471 6.17333L5.82702 7.78599L1.87727 0.574039Z" fill="#8692A3"/>
                                        </svg>
                                    </div>
                                    <div>33% drop from ${prevEvent.index}-${event.index}</div>
                                </div>
                            `
                        );
                    }
                    return (
                        `
                            <div class="bg-white px-4 rounded-md shadow-md border-2 text-xs">
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

    useEffect(() => {
        drawChart();
        showVerticalGridLines();
        showConverionRates();
    }, [drawChart, showVerticalGridLines, showConverionRates]);

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
                                <div style={{fontSize: eventsData.length > 2 ? '18px' : '14px'}} className={styles.conversionText}>
                                    <div className="font-bold flex justify-end">{elem.conversion_rate}</div>
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