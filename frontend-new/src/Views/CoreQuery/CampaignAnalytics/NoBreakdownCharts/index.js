import React, { useState, useEffect } from "react";
import { formatData, formatDataInLineChartFormat } from "./utils";
import ChartHeader from "../../../../components/SparkLineChart/ChartHeader";
import SparkChart from "../../../../components/SparkLineChart/Chart";
import {
  generateColors,
  numberWithCommas,
} from "../../../../utils/dataFormatter";
import LineChart from "../../../../components/LineChart";
import NoBreakdownTable from "./NoBreakdownTable";
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
} from "../../../../utils/constants";

function NoBreakdownCharts({ chartType, data, arrayMapper, section }) {
  const [chartsData, setChartsData] = useState([]);

  useEffect(() => {
    const formattedData = formatData(data, arrayMapper);
    setChartsData(formattedData);
  }, [data, arrayMapper]);

  if (!chartsData.length) {
    return (
      <div className="mt-4 flex justify-center items-center w-full h-64 ">
        No Data Found!
      </div>
    );
  }

  const table = (
    <div className="mt-12 w-full">
      <NoBreakdownTable
        chartType={chartType}
        chartsData={chartsData}
        isWidgetModal={false}
      />
    </div>
  );

  let chart = null;

  if (chartType === CHART_TYPE_SPARKLINES) {
    if (chartsData.length === 1) {
      chart = (
        <div className="flex items-center justify-center w-full">
          <div className="w-1/4">
            <ChartHeader
              bgColor="#4D7DB4"
              query={chartsData[0].name}
              total={numberWithCommas(chartsData[0].total)}
            />
          </div>
          <div className="w-3/4">
            <SparkChart
              frequency="date"
              page="campaigns"
              event={chartsData[0].mapper}
              chartData={chartsData[0].dataOverTime}
              chartColor="#4D7DB4"
            />
          </div>
        </div>
      );
    }

    if (chartsData.length > 1) {
      const appliedColors = generateColors(chartsData.length);
      chart = (
        <div className="flex items-center flex-wrap justify-center w-full">
          {chartsData.map((chartData, index) => {
            return (
              <div
                style={{ minWidth: "300px" }}
                key={chartData.index}
                className="w-1/3 mt-4 px-4"
              >
                <div className="flex flex-col">
                  <ChartHeader
                    total={numberWithCommas(chartData.total)}
                    query={chartData.name}
                    bgColor={appliedColors[index]}
                  />
                  <div className="mt-8">
                    <SparkChart
                      frequency="date"
                      page="campaigns"
                      event={chartData.mapper}
                      chartData={chartData.dataOverTime}
                      chartColor={appliedColors[index]}
                    />
                  </div>
                </div>
              </div>
            );
          })}
        </div>
      );
    }
  } else if (chartType === CHART_TYPE_LINECHART) {
    const lineChartData = formatDataInLineChartFormat(chartsData);
    const appliedColors = generateColors(chartsData.length);
    chart = (
      <LineChart
        frequency="date"
        chartData={lineChartData}
        hiddenEvents={[]}
        setHiddenEvents={() => {}}
        appliedColors={appliedColors}
        queries={chartsData.map((elem) => elem.name)}
        arrayMapper={arrayMapper.filter(
          (elem) => chartsData.findIndex((d) => d.index === elem.index) > -1
        )}
        isDecimalAllowed={false}
        section={section}
      />
    );
  }

  return (
    <div className="flex items-center justify-center flex-col">
      {chart}
      {table}
    </div>
  );
}

export default NoBreakdownCharts;
