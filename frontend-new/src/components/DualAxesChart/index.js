import React, { useCallback, useRef, useEffect } from "react";
import * as d3 from "d3";

function DualAxesChart({ chartData, queries, title = "chart" }) {
  const chartRef = useRef(null);

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
  }, [title]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return (
    <div className="w-full bar-chart">
      <div ref={chartRef}></div>
      {/* {queries && queries.length > 1 ? (
        <div className="mt-4">
          <ChartLegends events={queries} chartData={chartData} />
        </div>
      ) : null} */}
    </div>
  );
}

export default DualAxesChart;
