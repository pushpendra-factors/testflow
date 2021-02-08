import React from "react";
import ChartHeader from "./ChartHeader";
import SparkChart from "./Chart";
import { numberWithCommas } from "../../utils/dataFormatter";

function SparkLineChart({
  queries,
  chartsData,
  appliedColors,
  page,
  resultState,
  frequency,
  arrayMapper,
  height,
  cardSize = 1
}) {
  if (queries.length > 1) {
    return (
      <div className="flex items-center flex-wrap justify-center w-full">
        {queries.map((q, index) => {
          const mapper = arrayMapper.find(
            (elem) => elem.eventName === q && elem.index === index
          ).mapper;
          let total = 0;
          const data = chartsData.map((elem) => {
            return {
              date: elem.date,
              [mapper]: elem[mapper],
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
              className="w-1/3 mt-4 px-4"
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
                    event={mapper}
                    chartData={data}
                    chartColor={appliedColors[index]}
                    height={height}
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
      <div className={`flex items-center justify-center w-full ${!cardSize ? "flex-col" : ""}`}>
        <div className={cardSize? "w-1/4" : ''}>
          <ChartHeader bgColor="#4D7DB4" query={queries[0]} total={total} />
        </div>
        <div className="w-3/4">
          <SparkChart
            frequency={frequency}
            page={page}
            event={
              arrayMapper.find((elem) => elem.eventName === queries[0]).mapper
            }
            chartData={chartsData}
            chartColor="#4D7DB4"
            height={height}
          />
        </div>
      </div>
    );
  }
}

export default SparkLineChart;
