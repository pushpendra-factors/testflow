import React, { useRef, useCallback, useEffect } from 'react';
import * as d3 from 'd3';
import styles from './index.module.scss';
import { checkForWindowSizeChange, calculatePercentage, generateColors } from '../utils';

function UngroupedChart({ chartData }) {

    const chartRef = useRef(null);

    const appliedColors = generateColors(chartData.length);

    let tooltip = useRef(null);

    const showTooltip = useCallback((d, x, y) => {
        tooltip.current
            .style("opacity", 1)
        tooltip.current
            .html(`
                    <div>
                        <div>www.chargebee.com/subscription-management/create-manage-plans</div>
                        <div class="mt-2"><span class="font-semibold">${d.netCount}</span> (${d.value}%)</div>
                    </div>
                `)
            .style("left", x + 25 + "px")
            .style("top", y - 80 + "px")
    }, [])

    const hideTooltip = useCallback(() => {
        tooltip.current
            .style("opacity", 0);
    }, [])

    const showChangePercentage = useCallback(() => {
        const barNodes = d3.select(chartRef.current).selectAll('.bar').nodes();
        const xAxis = d3.select(chartRef.current).select('.axis.axis--x').node();
        const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
        barNodes.forEach((node, index) => {
            const positionCurrentBar = node.getBoundingClientRect();

            // show values inside bar

            document.getElementById(`value${index}`).style.left = positionCurrentBar.left + 'px';
            document.getElementById(`value${index}`).style.width = positionCurrentBar.width + 'px';

            if (positionCurrentBar.height < 30) {
                document.getElementById(`value${index}`).style.top = (positionCurrentBar.top - 30 + scrollTop) + 'px';
            } else {
                document.getElementById(`value${index}`).style.top = (positionCurrentBar.top + scrollTop + 5) + 'px';
                document.getElementById(`value${index}`).style.color = 'white';
            }

            //show change percentages in grey polygon areas
            if (index < (barNodes.length - 1)) {
                const positionNextBar = barNodes[index + 1].getBoundingClientRect();
                document.getElementById(`change${index}`).style.left = positionCurrentBar.right + 'px';
                document.getElementById(`change${index}`).style.width = (positionNextBar.left - positionCurrentBar.right) + 'px';
                document.getElementById(`change${index}`).style.top = (xAxis.getBoundingClientRect().top - xAxis.getBoundingClientRect().height - 30 + scrollTop) + 'px';
            }
        });
    }, []);

    const showOverAllConversionPercentage = useCallback(() => {

        //place percentage text in the chart
        const barNodes = d3.select(chartRef.current).selectAll('.bar').nodes();
        const lastBarNode = barNodes[barNodes.length - 1];
        const lastBarPosition = lastBarNode.getBoundingClientRect();
        const yGridLines = d3.select(chartRef.current).select('.y.axis-grid').selectAll('g.tick').nodes();
        const topGridLine = yGridLines[yGridLines.length - 1];
        const top = topGridLine.getBoundingClientRect().y;
        const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
        const conversionText = document.getElementById('conversionText');
        conversionText.style.left = `${lastBarPosition.x}px`;
        conversionText.style.width = `${lastBarPosition.width}px`;
        conversionText.style.top = `${top + scrollTop}px`;

        // show vertical grid line
        const verticalLine = document.getElementById("overAllConversionLine");
        verticalLine.style.left = `${lastBarPosition.x + lastBarPosition.width - 1}px`;
        const bottomGridLine = yGridLines[0];
        const bottom = bottomGridLine.getBoundingClientRect().y;
        const height = bottom - top;
        verticalLine.style.height = `${height}px`;
        verticalLine.style.top = `${top + scrollTop}px`;
    }, []);

    const drawChart = useCallback(() => {
        const availableWidth = d3.select(chartRef.current).node().getBoundingClientRect().width;
        d3.select(chartRef.current).html('').append('svg').attr('width', availableWidth).attr('height', 400).attr('id', 'chart')
        const svg = d3.select("#chart");
        const margin = { top: 30, right: 0, bottom: 30, left: 40 };
        const width = +svg.attr("width") - margin.left - margin.right;
        const height = +svg.attr("height") - margin.top - margin.bottom;

        tooltip.current = d3.select(chartRef.current).append("div").attr("class", "toolTip").style("opacity", 0);

        const xScale = d3.scaleBand()
            .rangeRound([0, width])
            .paddingOuter(0.15)
            .paddingInner(0.3)
            .domain(chartData.map(d => d.event));

        const yScale = d3.scaleLinear()
            .rangeRound([height, 0])
            .domain([0, 100])

        const yAxisGrid = d3.axisLeft(yScale).tickSize(-width).tickFormat('').ticks(5);

        const g = svg.append("g")
            .attr("transform", `translate(${margin.left},${margin.top})`);

        g.append('g')
            .attr('class', 'y axis-grid')
            .call(yAxisGrid)
            .selectAll('line')
            .attr('stroke', '#E7E9ED')

        g.append("g")
            .attr("class", "axis axis--x")
            .attr("transform", `translate(0,${height})`)
            .call(d3.axisBottom(xScale));

        g.append("g")
            .attr("class", "axis axis--y")
            .call(d3.axisLeft(yScale).tickFormat(d => {
                return d + '%'
            }).ticks(5));

        g.selectAll(".bar")
            .data(chartData)
            .enter()
            .append("rect")
            .attr("class", d => {
                return `bar`
            })
            .attr('fill', (d, index) => {
                return appliedColors[index];
            })
            .attr("x", d => xScale(d.event))
            .attr("y", d => yScale(d.value))
            .attr("width", xScale.bandwidth())
            .attr("height", d => height - yScale(d.value))
            .on('mousemove', d => {
                showTooltip(d, d3.event.pageX, d3.event.pageY);
            })
            .on('mouseover', (d, i, nodes) => {
                nodes.forEach((node, index) => {
                    if (index !== i) {
                        d3.select(node).attr('class', 'bar opaque')
                    }
                })
            })
            .on('mouseout', (d, i, nodes) => {
                hideTooltip();
                nodes.forEach((node, index) => {
                    if (index !== i) {
                        d3.select(node).attr('class', 'bar')
                    }
                })
            })
        d3.select(chartRef.current).select(".axis.axis--x").selectAll('.tick').select('text').attr("dy", "16px");

        // Add polygons
        g.selectAll(".area")
            .data(chartData)
            .enter().append("polygon")
            .attr("class", "area")
            .attr("points", (d, i, nodes) => {
                if (i < nodes.length - 1) {
                    const dNext = d3.select(nodes[i + 1]).datum();

                    const x1 = xScale(d.event) + xScale.bandwidth();
                    const y1 = height;

                    const x2 = x1;
                    const y2 = yScale(d.value);

                    const x3 = xScale(dNext.event);
                    const y3 = yScale(dNext.value);

                    const x4 = x3;
                    const y4 = height;

                    return `${x1},${y1} ${x2},${y2} ${x3},${y3} ${x4},${y4} ${x1},${y1}`;
                }
            });
    }, [chartData, showTooltip, hideTooltip, appliedColors]);

    const displayChart = useCallback(() => {
        drawChart();
        showChangePercentage();
        showOverAllConversionPercentage();
    }, [drawChart, showChangePercentage, showOverAllConversionPercentage])

    useEffect(() => {
        window.addEventListener("resize", () => checkForWindowSizeChange(displayChart));
        return () => {
            window.removeEventListener("resize", () => checkForWindowSizeChange(displayChart));
        }
    }, [displayChart]);

    useEffect(() => {
        displayChart();
    }, [displayChart]);



    const percentChanges = chartData.slice(1).map((elem, index) => {
        return calculatePercentage(chartData[index].netCount - elem.netCount, chartData[index].netCount);
    });

    return (
        <div className="ungrouped-chart">

            <div id="overAllConversionLine" className={`absolute border-l border-solid ${styles.overAllConversionLine}`}></div>

            <div style={{ transition: '2s' }} id="conversionText" className="absolute flex justify-end pr-1">
                <div className={styles.conversionText}>
                    <div className="font-semibold flex justify-end">{chartData[chartData.length - 1].value}%</div>
                    <div>Conversion</div>
                </div>
            </div>

            {chartData.map((d, index) => {
                return (
                    <div onMouseOut={hideTooltip} onMouseMove={(e) => showTooltip(d, e.screenX, e.screenY)} className={`${styles.valueText} absolute font-bold flex justify-center`} id={`value${index}`} key={d.event + index}>{d.netCount}</div>
                )
            })}
            {percentChanges.map((change, index) => {
                return (
                    <div className={`absolute flex flex-col items-center ${styles.changePercents}`} id={`change${index}`} key={index}>
                        <div className="flex justify-center">
                            <svg width="18" height="18" viewBox="0 0 18 18" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path fillRule="evenodd" clipRule="evenodd" d="M2.64045 1.27306C2.37516 0.788663 1.76742 0.61104 1.28302 0.876329C0.798626 1.14162 0.621004 1.74936 0.886293 2.23376L5.29748 10.2882C5.55527 10.7589 6.13864 10.9422 6.6193 10.7036L9.91856 9.06529L13.0943 14.1005L11.318 15.1145C10.9341 15.3337 11.0026 15.9067 11.4273 16.0292L15.4142 17.1791C15.6796 17.2556 15.9567 17.1025 16.0332 16.8372L17.2161 12.736C17.3405 12.3047 16.8776 11.9407 16.4878 12.1632L14.8329 13.108L11.1285 7.23455C10.8548 6.80069 10.2973 6.64423 9.83789 6.87235L6.59021 8.48501L2.64045 1.27306Z" fill="#8692A3" />
                            </svg>
                        </div>
                        <div className="flex justify-center">{change}%</div>
                    </div>
                )
            })}
            <div ref={chartRef} className={styles.ungroupedChart}>
            </div>
        </div>
    )
}

export default UngroupedChart;