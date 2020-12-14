import React from "react";
import ChartHeader from "./ChartHeader";
import SparkChart from "./Chart";
import { numberWithCommas } from "../../utils/dataFormatter";

function SparkLineChart({
  queries,
  chartsData,
  parentClass,
  appliedColors,
  eventsMapper,
  page,
  resultState,
  frequency,
}) {
  if (queries.length > 1) {
    return (
      <div className={parentClass}>
        {queries.map((q, index) => {
          let total = 0;
          const data = chartsData.map((elem) => {
            return {
              date: elem.date,
              [eventsMapper[q]]: elem[eventsMapper[q]],
            };
          });
          const queryRow = resultState.data.metrics.rows.find(
            (elem) => elem[0] === index
          );
          total = queryRow ? queryRow[2] : 0;
          total =
            total % 1 !== 0
              ? parseFloat(total.toFixed(2))
              : numberWithCommas(total);

          return (
            <div
              style={{ minWidth: "300px" }}
              key={q + index}
              className="w-1/3 mt-4 px-1"
            >
              <div className="flex flex-col">
                <ChartHeader
                  total={total}
                  query={q}
                  bgColor={appliedColors[index]}
                />
                <div className="mt-8">
                  <SparkChart
                    frequency={frequency}
                    page={page}
                    event={eventsMapper[q]}
                    chartData={data}
                    chartColor={appliedColors[index]}
                  />
                </div>
              </div>
            </div>
          );
        })}
      </div>
    );
  } else {
    let total = resultState.data.metrics.rows.find((elem) => elem[0] === 0)[2];
    total =
      total % 1 !== 0 ? parseFloat(total.toFixed(2)) : numberWithCommas(total);

    return (
      <div className={parentClass}>
        <div className="w-1/4">
          <ChartHeader bgColor="#4D7DB4" query={queries[0]} total={total} />
        </div>
        <div className="w-3/4">
          <SparkChart
            frequency={frequency}
            page={page}
            event={eventsMapper[queries[0]]}
            chartData={chartsData}
            chartColor="#4D7DB4"
          />
        </div>
      </div>
    );
  }
}

export default SparkLineChart;
