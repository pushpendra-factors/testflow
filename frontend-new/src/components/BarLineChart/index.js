import React, { useCallback, useRef, useEffect } from "react";
import styles from "./index.module.scss";
import * as d3 from "d3";
import { getMaxYpoint } from "../BarChart/utils";
import ChartLegends from "./ChartLegends";

function BarLineChart({ chartData }) {
  const chartRef = useRef(null);

  const drawChart = useCallback(() => {
    d3.select(chartRef.current).html("");
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
      .getBoundingClientRect().width;
    const margin = { top: 20, right: 70, bottom: 40, left: 70 };
    const svg = d3
      .select(chartRef.current)
      .append("svg")
      .attr("width", availableWidth)
      .attr("height", 300);
    const width = +svg.attr("width") - margin.left - margin.right;
    const height = +svg.attr("height") - margin.top - margin.bottom;
    const xScale = d3
      .scaleBand()
      .rangeRound([0, width])
      .padding(0.1)
      .domain(
        chartData.map(function (d) {
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
      .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    // axis-x
    g.append("g")
      .attr("class", `axis axis--x ${styles.xAxis}`)
      .attr("transform", "translate(0," + height + ")")
      .call(d3.axisBottom(xScale));

    // axis-y
    g.append("g")
      .attr("class", `axis axis--y ${styles.y1}`)
      .call(d3.axisLeft(yScaleLeft).ticks(5));
    g.append("g")
      .attr("class", `axis axis--y ${styles.y2}`)
      .attr("transform", "translate( " + width + ", 0 )")
      .call(d3.axisRight(yScaleRight).ticks(5));

    g.append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", 0 - margin.left)
      .attr("x", 0 - height / 2)
      .attr("dy", "1em")
      .style("text-anchor", "middle")
      .attr("class", styles.yAxisLables)
      .text("Unique users");

    g.append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", 0 + width + 50)
      .attr("x", 0 - height / 2)
      .attr("dy", "1em")
      .style("text-anchor", "middle")
      .attr("class", styles.yAxisLables)
      .text("Cost per Conversion");

    var bar = g.selectAll("rect").data(chartData).enter().append("g");

    // bar chart
    bar
      .append("rect")
      .attr("x", function (d) {
        return xScale(d[0]);
      })
      .attr("y", function (d) {
        return yScaleLeft(d[2]);
      })
      .attr("width", xScale.bandwidth())
      .attr("height", function (d) {
        return height - yScaleLeft(d[2]);
      })
      .style("fill", "#4d7db4");

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
      .append("path")
      .style("fill", "none")
      .style("stroke", "#D4787D")
      .style("stroke-width", 3)
      .attr("d", line(chartData)); // 11. Calls the line generator

    bar
      .append("circle") // Uses the enter().append() method
      .style("fill", "#D4787D")
      .style("stroke", "#D4787D")
      .style("stroke-width", 4)
      .attr("cx", function (d, i) {
        return xScale(d[0]) + xScale.bandwidth() / 2;
      })
      .attr("cy", function (d) {
        return yScaleRight(d[1]);
      })
      .attr("r", 5);
  }, [chartData]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return (
    <div className="w-full bar-chart">
      <div ref={chartRef}></div>
      <ChartLegends />
    </div>
  );
}

export default BarLineChart;
