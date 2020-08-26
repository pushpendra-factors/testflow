import React, { useRef, useCallback, useEffect } from 'react';
import c3 from 'c3';
import * as d3 from 'd3';
import { xScale } from 'd3';
import styles from './index.module.scss';

function UngroupedChart({ eventsData, groups, chartData, chartColors }) {

    const chartRef = useRef(null);

    const colors = ['#014694', '#008BAE', '#52C07C', '#F1C859', '#EEAC4C', '#DE7542'];

    const drawChart = useCallback(() => {
        const data = [
            { event: "Cart Updated", value: 100 },
            { event: "Checkout", value: 60 },
            { event: "Applied Coupon", value: 30 },
            { event: "Removed Coupon", value: 20 },
            { event: "Applied Second Time", value: 15 },
            { event: "Paid", value: 10 }
        ];
        const svg = d3.select("#chart");
        const margin = { top: 20, right: 30, bottom: 30, left: 50 };
        const width = +svg.attr("width") - margin.left - margin.right;
        const height = +svg.attr("height") - margin.top - margin.bottom;

        const tooltip = d3.select(chartRef.current).append("div").attr("class", "toolTip");

        const xScale = d3.scaleBand()
            .rangeRound([0, width]).padding(0.5)
            .domain(data.map(d => d.event));

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
            .data(data)
            .enter().append("rect")
            .attr("class", d => {
                return `bar`
            })
            .attr('fill', (d, index) => {
                return colors[index];
            })
            .attr("x", d => xScale(d.event))
            .attr("y", d => yScale(d.value))
            .attr("width", xScale.bandwidth())
            .attr("height", d => height - yScale(d.value))
            .on('mousemove', d => {
                tooltip
                    .style("left", d3.event.pageX - 50 + "px")
                    .style("top", d3.event.pageY - 80 + "px")
                    .style("display", "inline-block")
                    .html((d.event) + ' ' + (d.value) + '%');
            })
            .on('mouseover', (d, i, nodes) => {
                nodes.forEach((node, index) => {
                    if (index !== i) {
                        d3.select(node).attr('class', 'bar opaque')
                    }
                })
            })
            .on('mouseout', (d, i, nodes) => {
                tooltip
                    .style("display", "none")
                nodes.forEach((node, index) => {
                    if (index !== i) {
                        d3.select(node).attr('class', 'bar')
                    }
                })
            })

        // Add polygons
        g.selectAll(".area")
            .data(data)
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

                    console.log(x1, x3)

                    return `${x1},${y1} ${x2},${y2} ${x3},${y3} ${x4},${y4} ${x1},${y1}`;
                }
            });
    }, []);

    useEffect(() => {
        drawChart();
    }, [drawChart]);

    return (
        <div ref={chartRef} className={styles.ungroupedChart}>
            <svg width='1552' height="400" id="chart"></svg>
        </div>
    )
}

export default UngroupedChart;