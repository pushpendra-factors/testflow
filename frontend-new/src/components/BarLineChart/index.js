import React, { useCallback, useRef, useEffect } from "react";
import styles from "./index.module.scss";
import * as d3 from "d3";
import { getMaxYpoint } from "../BarChart/utils";

function BarLineChart({ chartData, queries }) {
  const chartRef = useRef(null);

  const drawChart = useCallback(() => {
    const values = [];
    chartData.forEach((cd) => {
      values.push(cd[1], cd[2]);
    });
    const max = getMaxYpoint(Math.max(...values));
    const availableWidth = d3
      .select(chartRef.current)
      .node()
      .getBoundingClientRect().width;
    const margin = { top: 20, right: 70, bottom: 30, left: 70 };
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
    const yScale = d3.scaleLinear().rangeRound([height, 0]).domain([0, max]);
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
      .call(d3.axisLeft(yScale).ticks(5));
    g.append("g")
      .attr("class", `axis axis--y ${styles.y2}`)
      .attr("transform", "translate( " + width + ", 0 )")
      .call(d3.axisRight(yScale).ticks(5));

    g.append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", 0 - margin.left)
      .attr("x", 0 - height / 2)
      .attr("dy", "1em")
      .style("text-anchor", "middle")
      .attr('class', styles.yAxisLables)
      .text("Unique users");
    
    g.append("text")
      .attr("transform", "rotate(-90)")
      .attr("y", 0 + width + 50)
      .attr("x", 0 - height / 2)
      .attr("dy", "1em")
      .style("text-anchor", "middle")
      .attr('class', styles.yAxisLables)
      .text("Cost per Conversions");

    var bar = g.selectAll("rect").data(chartData).enter().append("g");

    // bar chart
    bar
      .append("rect")
      .attr("x", function (d) {
        return xScale(d[0]);
      })
      .attr("y", function (d) {
        return yScale(d[2]);
      })
      .attr("width", xScale.bandwidth())
      .attr("height", function (d) {
        return height - yScale(d[2]);
      })
      .attr("class", styles.bar);

    // labels on the bar chart
    bar
      .append("text")
      .attr("dy", "1.3em")
      .attr("x", function (d) {
        return xScale(d[0]) + xScale.bandwidth() / 2;
      })
      .attr("y", function (d) {
        return yScale(d[2]);
      })
      .attr("text-anchor", "middle")
      .text(function (d) {
        return d[2];
      });

    // line chart
    var line = d3
      .line()
      .x(function (d, i) {
        return xScale(d[0]) + xScale.bandwidth() / 2;
      })
      .y(function (d) {
        return yScale(d[1]);
      });

    bar
      .append("path")
      .attr("class", styles.line) // Assign a class for styling
      .attr("d", line(chartData)); // 11. Calls the line generator

    bar
      .append("circle") // Uses the enter().append() method
      .attr("class", styles.dot) // Assign a class for styling
      .attr("cx", function (d, i) {
        return xScale(d[0]) + xScale.bandwidth() / 2;
      })
      .attr("cy", function (d) {
        return yScale(d[1]);
      })
      .attr("r", 5);
  }, [chartData]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return (
    <div className="w-full bar-chart">
      <div className={styles.ungroupedChart} ref={chartRef}></div>
      {/* {queries && queries.length > 1 ? (
        <div className="mt-4">
          <ChartLegends events={queries} chartData={chartData} />
        </div>
      ) : null} */}
    </div>
  );
}

export default BarLineChart;
