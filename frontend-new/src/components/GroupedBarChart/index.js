import React, { useRef, useEffect, useCallback } from "react";
import c3 from "c3";
import styles from "../../Views/CoreQuery/FunnelsResultPage/GroupedChart/index.module.scss";

function GroupedBarChart({
  categories,
  chartData,
  colors,
  responseRows,
  responseHeaders,
  visibleIndices,
  method1,
  method2,
  event,
}) {
  const chartRef = useRef(null);
  
  const drawChart = useCallback(() => {
    c3.generate({
      size: {
        height: 300,
      },
      padding: {
        left: 50,
        bottom: 20,
      },
      bindto: chartRef.current,
      data: {
        columns: chartData,
        type: "bar",
        colors
      },
      legend: {
        show: true,
      },
      transition: {
        duration: 1000,
      },
      bar: {
        space: 0.05,
      },
      axis: {
        x: {
          type: "category",
          tick: {
            multiline: true,
            multilineMax: 3,
          },
          categories: categories,
        },
        y: {
          tick: {
            count: 5,
            format(d) {
              return parseInt(d);
            },
          },
        },
      },
      grid: {
        y: {
          show: true,
        },
      },
      tooltip: {
        contents: (d) => {
          const userIdx = responseHeaders.indexOf(`${event} - Users`);
          const compareUsersIdx = responseHeaders.indexOf(`Compare - Users`);
          const impressionsIdx = responseHeaders.indexOf("Impressions");
          const clicksIdx = responseHeaders.indexOf("Clicks");
          const spendIdx = responseHeaders.indexOf("Spend");
          const visitorsIdx = responseHeaders.indexOf("Website Visitors");
          const rowIndex = visibleIndices[d[0].index];
          return `<div style="width:200px;border:1px solid #E7E9ED; border-radius: 12px; box-shadow: 0px 6px 20px rgba(0, 0, 0, 0.2);" class="p-4 bg-white flex flex-col">
										<div style="border-bottom: 1px solid #E7E9ED;">
											<div class="pb-2" style="color: #3E516C;font-size: 14px;line-height: 24px;font-weight: 500;">${
                        categories[d[0].index]
                      }</div>
										</div>
										<div style="border-bottom: 1px solid #E7E9ED;" class="py-2">
											<div style="font-weight: 600;font-size: 10px;line-height: 16px;color: #8692A3;">CONVERSIONS</div>
											<div style="font-weight: 600;font-size: 12px;line-height: 16px;" class="mt-2 flex justify-between">
												<div style="color: #4D7DB4;">${method1}</div>
												<div style="color: #3E516C;">${responseRows[rowIndex][userIdx]}</div>
											</div>
											<div style="font-weight: 600;font-size: 12px;line-height: 16px;" class="mt-2 flex justify-between">
												<div style="color: #4CBCBD;">${method2}</div>
												<div style="color: #3E516C;">${responseRows[rowIndex][compareUsersIdx]}</div>
											</div>
										</div>
										<div style="font-size: 12px;line-height: 18px;color: #3E516C;">
											<div class="flex justify-between pt-2">
												<div>Impressions</div>
												<div>${responseRows[rowIndex][impressionsIdx]}</div>
											</div>
											<div class="flex justify-between pt-2">
												<div>Clicks</div>
												<div>${responseRows[rowIndex][clicksIdx]}</div>
											</div>
											<div class="flex justify-between pt-2">
												<div>Spend</div>
												<div>${responseRows[rowIndex][spendIdx]}</div>
											</div>
											<div class="flex justify-between pt-2">
												<div>Visitors</div>
												<div>${responseRows[rowIndex][visitorsIdx]}</div>
											</div>
										</div>
									</div>`;
        },
      },
    });
  }, [
    categories,
    chartData,
    colors,
    event,
    method1,
    method2,
    responseHeaders,
    responseRows,
    visibleIndices,
  ]);

  useEffect(() => {
    drawChart();
  }, [drawChart]);

  return (
    <div className={`w-full bar-chart ${styles.groupedChart}`}>
      <div ref={chartRef}></div>
    </div>
  );
}

export default GroupedBarChart;
