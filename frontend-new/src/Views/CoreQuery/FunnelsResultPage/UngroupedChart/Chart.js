import React, { useRef, useCallback, useEffect } from "react";
import * as d3 from "d3";
import styles from "./index.module.scss";
import { checkForWindowSizeChange } from "../utils";
import {
  calculatePercentage,
  generateColors,
} from "../../../../utils/dataFormatter";

function Chart({ chartData, title = "chart", cardSize = 1, arrayMapper }) {
  const chartRef = useRef(null);
  const tooltip = useRef(null);
  const appliedColors = generateColors(chartData.length);

  const showTooltip = useCallback(
    (d, i) => {
      const nodes = d3.select(chartRef.current).selectAll(".bar").nodes();
      nodes.forEach((node, index) => {
        if (index !== i) {
          d3.select(node).attr("class", "bar opaque");
        }
      });

      const nodePosition = d3.select(nodes[i]).node().getBoundingClientRect();
      let left = nodePosition.x + nodePosition.width / 2;

      // if user is hovering over the last bar
      if (left + 200 >= document.documentElement.clientWidth) {
        left = nodePosition.x + nodePosition.width / 2 - 200;
      }

      const scrollTop =
        window.pageYOffset !== undefined
          ? window.pageYOffset
          : (
              document.documentElement ||
              document.body.parentNode ||
              document.body
            ).scrollTop;
      const top = nodePosition.y + scrollTop;
      const toolTipHeight = d3.select(".toolTip").node().getBoundingClientRect()
        .height;

      tooltip.current
        .html(
          `
              <div>${arrayMapper[i].eventName}</div>
              <div style="color: #0E2647;" class="mt-2 leading-5 text-base"><span class="font-semibold">${d.netCount}</span> (${d.value}%)</div>
            `
        )
        .style("opacity", 1)
        .style("left", left + "px")
        .style("top", top - toolTipHeight + 5 + "px");
    },
    [arrayMapper]
  );

  const hideTooltip = useCallback(() => {
    const nodes = d3.select(chartRef.current).selectAll(".bar").nodes();
    nodes.forEach((node) => {
      d3.select(node).attr("class", "bar");
    });
    tooltip.current.style("opacity", 0);
  }, []);

  const showChangePercentage = useCallback(() => {
    const barNodes = d3.select(chartRef.current).selectAll(".bar").nodes();
    const xAxis = d3.select(chartRef.current).select(".axis.axis--x").node();
    const scrollTop =
      window.pageYOffset !== undefined
        ? window.pageYOffset
        : (
            document.documentElement ||
            document.body.parentNode ||
            document.body
          ).scrollTop;
    barNodes.forEach((node, index) => {
      const positionCurrentBar = node.getBoundingClientRect();
      // show change percentages in grey polygon areas
      if (index < barNodes.length - 1) {
        const positionNextBar = barNodes[index + 1].getBoundingClientRect();
        document.getElementById(`${title}-change${index}`).style.left =
          positionCurrentBar.right + "px";
        document.getElementById(`${title}-change${index}`).style.width =
          positionNextBar.left - positionCurrentBar.right + "px";
        document.getElementById(`${title}-change${index}`).style.top =
          xAxis.getBoundingClientRect().top -
          xAxis.getBoundingClientRect().height -
          30 +
          scrollTop +
          "px";
      }
    });
  }, [title]);

  const showOverAllConversionPercentage = useCallback(() => {
    // place percentage text in the chart
    const barNodes = d3.select(chartRef.current).selectAll(".bar").nodes();
    const lastBarNode = barNodes[barNodes.length - 1];
    const lastBarPosition = lastBarNode.getBoundingClientRect();
    const yGridLines = d3
      .select(chartRef.current)
      .select(".y.axis-grid")
      .selectAll("g.tick")
      .nodes();
    const topGridLine = yGridLines[yGridLines.length - 1];
    const secondLastGridLine = yGridLines[yGridLines.length - 2];
    const top = topGridLine.getBoundingClientRect().y;
    const height =
      secondLastGridLine.getBoundingClientRect().y -
      topGridLine.getBoundingClientRect().y;
    const scrollTop =
      window.pageYOffset !== undefined
        ? window.pageYOffset
        : (
            document.documentElement ||
            document.body.parentNode ||
            document.body
          ).scrollTop;
    const conversionText = document.getElementById(`conversionText-${title}`);
    conversionText.style.left = `${lastBarPosition.x}px`;
    conversionText.style.height = `${height}px`;
    conversionText.style.width = `${lastBarPosition.width}px`;
    conversionText.style.top = `${top + scrollTop}px`;
  }, [title]);

  const drawChart = useCallback(() => {
    const availableWidth = d3
      .select(chartRef.current)
      .node()
      .getBoundingClientRect().width;
    d3.select(chartRef.current)
      .html("")
      .append("svg")
      .attr("width", availableWidth)
      .attr("height", 300)
      .attr("id", `chart-${title}`);
    const svg = d3.select(`#chart-${title}`);
    const margin = {
      top: 30,
      right: 0,
      bottom: 30,
      left: 40,
    };
    const width = +svg.attr("width") - margin.left - margin.right;
    const height = +svg.attr("height") - margin.top - margin.bottom;

    tooltip.current = d3
      .select(chartRef.current)
      .append("div")
      .attr("class", "toolTip")
      .style("opacity", 0)
      .style("transition", "0.5s");
    const xScale = d3
      .scaleBand()
      .rangeRound([0, width])
      .paddingOuter(0.15)
      .paddingInner(0.3)
      .domain(chartData.map((d) => d.event));

    const yScale = d3.scaleLinear().rangeRound([height, 0]).domain([0, 100]);

    const yAxisGrid = d3
      .axisLeft(yScale)
      .tickSize(-width)
      .tickFormat("")
      .ticks(5);

    const g = svg
      .append("g")
      .attr("transform", `translate(${margin.left},${margin.top})`);

    g.append("g")
      .attr("class", "y axis-grid")
      .call(yAxisGrid)
      .selectAll("line")
      .attr("stroke", "#E7E9ED");

    g.append("g")
      .attr("class", "axis axis--x")
      .attr("transform", `translate(0,${height})`)
      .call(
        d3.axisBottom(xScale).tickFormat((d, i) => {
          return arrayMapper[i].eventName;
        })
      );

    g.append("g")
      .attr("class", "axis axis--y")
      .call(
        d3
          .axisLeft(yScale)
          .tickFormat((d) => {
            return d + "%";
          })
          .ticks(5)
      );

    g.selectAll(".bar")
      .data(chartData)
      .enter()
      .append("rect")
      .attr("class", () => {
        return "bar";
      })
      .attr("fill", (_, index) => {
        return appliedColors[index];
      })
      .attr("x", (d) => xScale(d.event))
      .attr("y", (d) => yScale(d.value))
      .attr("width", xScale.bandwidth())
      .attr("height", (d) => height - yScale(d.value))
      .on("mousemove", (d, i) => {
        showTooltip(d, i);
      })
      .on("mouseout", () => {
        hideTooltip();
      });

    d3.select(chartRef.current)
      .select(".axis.axis--x")
      .selectAll(".tick")
      .select("text")
      .attr("dy", "16px");

    g.selectAll(".text")
      .data(chartData)
      .enter()
      .append("text")
      .attr("class", "text")
      .text(function (d) {
        return d.netCount;
      })
      .attr("x", function (d) {
        return xScale(d.event) + xScale.bandwidth() / 2;
      })

      .attr("y", function (d) {
        return yScale(d.value) < 200 ? yScale(d.value) + 20 : 220;
      })
      .attr("class", "font-bold")
      .attr("fill", function (d) {
        return yScale(d.value) < 200 ? "white" : "black";
      })
      .attr("text-anchor", "middle");

    g.selectAll(".vLine")
      .data(chartData)
      .enter()
      .append("line")
      .attr("class", "vLine")
      .attr("x1", function (d) {
        return xScale(d.event) + xScale.bandwidth();
      })
      .attr("y1", function () {
        return 0;
      })
      .attr("x2", function (d) {
        return xScale(d.event) + xScale.bandwidth();
      })
      .attr("y2", height)
      .style("stroke-width", function (_, i) {
        return i === chartData.length - 1 ? 1 : 0;
      })
      .style("stroke", "#B7BEC8")
      .style("fill", "none");

    // Add polygons
    g.selectAll(".area")
      .data(chartData)
      .enter()
      .append("polygon")
      .attr("class", "area")
      .text(function (d) {
        return d.netCount;
      })
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
  }, [chartData, showTooltip, hideTooltip, appliedColors, title, arrayMapper]);

  const displayChart = useCallback(() => {
    drawChart();
    showChangePercentage();
    showOverAllConversionPercentage();
  }, [drawChart, showChangePercentage, showOverAllConversionPercentage]);

  useEffect(() => {
    window.addEventListener("resize", () =>
      checkForWindowSizeChange(displayChart)
    );
    return () => {
      window.removeEventListener("resize", () =>
        checkForWindowSizeChange(displayChart)
      );
    };
  }, [displayChart]);

  useEffect(() => {
    displayChart();
  }, [displayChart]);

  const percentChanges = chartData.slice(1).map((elem, index) => {
    return calculatePercentage(
      chartData[index].netCount - elem.netCount,
      chartData[index].netCount
    );
  });

  return (
    <div id={`${title}-ungroupedChart`} className="ungrouped-chart">
      <div
        id={`conversionText-${title}`}
        className="absolute flex items-center justify-end pr-1"
      >
        <div className={styles.conversionText}>
          <div className="font-semibold flex justify-end">
            {chartData[chartData.length - 1].value}%
          </div>
          {cardSize ? <div className="font-normal">Conversion</div> : null}
        </div>
      </div>

      {percentChanges.map((change, index) => {
        return (
          <div
            className={`absolute flex justify-center items-center ${styles.changePercents}`}
            id={`${title}-change${index}`}
            key={index}
          >
            <div className="flex items-center justify-center mr-1">
              {cardSize ? (
                <svg
                  width="16"
                  height="16"
                  viewBox="0 0 24 24"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                >
                  <g clipPath="url(#clip0)">
                    <path
                      d="M13.8306 19.0713C14.3815 19.0713 14.8281 18.6247 14.8281 18.0737C14.8281 17.5232 14.3822 17.0768 13.8316 17.0762L8.34395 17.0702L19.0708 6.34337C19.4613 5.95285 19.4613 5.31968 19.0708 4.92916C18.6802 4.53863 18.0471 4.53863 17.6565 4.92916L6.92974 15.656V10.1683C6.92974 9.6146 6.47931 9.16632 5.92565 9.16828C5.37474 9.17022 4.92863 9.61737 4.92863 10.1683L4.92862 18.0713C4.92863 18.6236 5.37634 19.0713 5.92863 19.0713L13.8306 19.0713Z"
                      fill="#8692A3"
                    />
                  </g>
                  <defs>
                    <clipPath id="clip0">
                      <rect width="24" height="24" fill="white" />
                    </clipPath>
                  </defs>
                </svg>
              ) : null}
            </div>
            <div
              className={`leading-4 flex justify-center ${
                !cardSize ? "text-xs" : ""
              }`}
            >
              {change}%
            </div>
          </div>
        );
      })}
      <div ref={chartRef} className={styles.ungroupedChart}></div>
    </div>
  );
}

export default Chart;
