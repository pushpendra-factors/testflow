import React, { useCallback, useRef, useEffect } from "react";
import c3 from "c3";
import * as d3 from "d3";
import moment from "moment";
import styles from "./index.module.scss";
import ChartLegends from "./ChartLegends";
import { numberWithCommas } from "../../utils/dataFormatter";

function LineChart({
  chartData,
  appliedColors,
  queries,
  setHiddenEvents,
  hiddenEvents,
  isDecimalAllowed,
  frequency,
  arrayMapper,
  cardSize = 1,
}) {
  const chartRef = useRef(null);

  const xAxisValues = chartData.find((elem) => elem[0] === "x").slice(1);
  xAxisValues.sort(function (left, right) {
    return moment.utc(left).diff(moment.utc(right));
  });

  let xAxisCount = Math.ceil(xAxisValues.length / 2);

  if (xAxisCount > 10) {
    xAxisCount = 10;
  }

  if (!cardSize && xAxisCount > 5) {
    xAxisCount = 5;
  }

  const modVal = Math.ceil(xAxisValues.length / xAxisCount);

  const finalXaxisValues = [];
  let j = 1;

  for (let i = 0; i < xAxisCount && j < xAxisValues.length; i++) {
    finalXaxisValues.push(xAxisValues[1 + i * modVal]);
    j = 1 + (i + 1) * modVal;
  }

  const colors = {};

  queries.forEach((_, index) => {
    const key = arrayMapper.find((m) => m.index === index).mapper;
    colors[key] = appliedColors[index];
  });

  const focusHoveredLines = useCallback((name) => {
    d3.select(chartRef.current)
      .selectAll(".c3-chart-line.c3-target")
      .nodes()
      .forEach((node) => {
        node.classList.add("c3-defocused");
      });
    d3.select(chartRef.current)
      .select(`.c3-chart-line.c3-target.c3-target-${name.split(" ").join("-")}`)
      .nodes()
      .forEach((node) => {
        node.classList.remove("c3-defocused");
        node.classList.add("c3-focused");
      });
  }, []);

  const focusAllLines = useCallback(() => {
    d3.select(chartRef.current)
      .selectAll(".c3-chart-line.c3-target")
      .nodes()
      .forEach((node) => {
        node.classList.remove("c3-defocused");
        node.classList.remove("c3-focused");
      });
  }, []);

  const modifyCirclesCSS = useCallback(() => {
    const eventChartNames = arrayMapper.map((elem) => elem.mapper);
    eventChartNames.forEach((name) => {
      d3.select(chartRef.current)
        .selectAll(`g.c3-circles-${name}`)
        .selectAll("circle.c3-shape.c3-circle")
        .style("stroke", colors[name]);
    });
  }, [colors, arrayMapper]);

  const drawChart = useCallback(() => {
    c3.generate({
      bindto: chartRef.current,
      size: {
        height: 300,
      },
      padding: {
        left: 30,
        bottom: 24,
        right: 10,
      },
      transition: {
        duration: 500,
      },
      data: {
        x: "x",
        xFormat: "%Y-%m-%d %H-%M",
        columns: chartData,
        colors,
        onmouseover: (d) => {
          focusHoveredLines(d.name);
        },
        onmouseout: () => {
          focusAllLines();
        },
      },
      axis: {
        x: {
          type: "timeseries",
          tick: {
            values: finalXaxisValues,
            format: (d) => {
              return frequency === "hour"
                ? moment(d).format("MMM D, h A")
                : moment(d).format("MMM D");
            },
          },
        },
        y: {
          tick: {
            count: 6,
            format(d) {
              if (!isDecimalAllowed) {
                return parseInt(d);
              } else {
                return parseFloat(d.toFixed(2));
              }
            },
          },
        },
      },
      onrendered: () => {
        d3.select(chartRef.current)
          .select(".c3-axis.c3-axis-x")
          .selectAll(".tick")
          .select("tspan")
          .attr("dy", "16px");
      },
      legend: {
        show: false,
      },
      grid: {
        y: {
          show: true,
        },
      },
      tooltip: {
        grouped: false,
        contents: (d) => {
          const data = d[0];
          let label = arrayMapper.find((elem) => elem.mapper === data.name)
            .eventName;
          label = label
            .split(",")
            .filter((elem) => elem)
            .join(",");
          return `   
              <div class="toolTip">
                  <div class="font-semibold">${
                    frequency === "hour"
                      ? moment(data.x).format("h A, MMM D, YYYY")
                      : moment(data.x).format("MMM D, YYYY")
                  }</div>
                  <div class="my-2">${label}</div>
                  <div class="flex items-center justify-start">
                      <div class="mr-1" style="background-color:${
                        colors[data.name]
                      };width:16px;height:16px;border-radius:8px"></div>
                      <div style="color:#0E2647;font-size:18px;line-height:24px">${numberWithCommas(
                        data.value
                      )}</div>
                  </div>
              </div>
            `;
        },
      },
    });
  }, [
    chartData,
    finalXaxisValues,
    colors,
    arrayMapper,
    focusHoveredLines,
    focusAllLines,
    isDecimalAllowed,
    frequency,
  ]);

  const displayChart = useCallback(() => {
    drawChart();
    modifyCirclesCSS();
  }, [drawChart, modifyCirclesCSS]);

  useEffect(() => {
    displayChart();
  }, [displayChart]);

  return (
    <div className="flex flex-col w-full">
      <div className={styles.lineChart} ref={chartRef} />
      <ChartLegends
        colors={colors}
        events={queries}
        focusHoveredLines={focusHoveredLines}
        focusAllLines={focusAllLines}
        setHiddenEvents={setHiddenEvents}
        hiddenEvents={hiddenEvents}
        arrayMapper={arrayMapper}
      />
    </div>
  );
}

export default LineChart;
