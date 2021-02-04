import React, { useState, useEffect } from "react";
import {
  formatData,
  formatDataInLineChartFormat,
} from "../../CoreQuery/CampaignAnalytics/NoBreakdownCharts/utils";
import ChartHeader from "../../../components/SparkLineChart/ChartHeader";
import SparkChart from "../../../components/SparkLineChart/Chart";
import { generateColors, numberWithCommas } from "../../../utils/dataFormatter";
import LineChart from "../../../components/LineChart";
import NoBreakdownTable from "../../CoreQuery/CampaignAnalytics/NoBreakdownCharts/NoBreakdownTable";
import {
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
} from "../../../utils/constants";

function NoBreakdownCharts({
  chartType,
  data,
  arrayMapper,
  isWidgetModal,
  unit,
  section,
  setwidgetModal,
}) {
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

  let tableContent = null;

  if (chartType === CHART_TYPE_TABLE) {
    tableContent = (
      <div
        onClick={() => setwidgetModal({ unit, data })}
        style={{ color: "#5949BC" }}
        className="mt-3 font-medium text-base cursor-pointer flex justify-end item-center"
      >
        Show More &rarr;
      </div>
    );
  }

  let chartContent = null;

  if (chartType === CHART_TYPE_SPARKLINES) {
    if (chartsData.length === 1) {
      chartContent = (
        <div
          className={`flex items-center mt-4 flex-wrap justify-center ${
            !unit.cardSize ? "flex-col" : ""
          }`}
        >
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
      chartContent = (
        <div className="flex items-center flex-wrap mt-4 justify-center">
          {chartsData.map((chartData, index) => {
            return (
              <div
                style={{ minWidth: "300px" }}
                key={chartData.index}
                className="w-1/3 mt-4 px-1"
              >
                <div className="flex flex-col">
                  <ChartHeader
                    total={numberWithCommas(chartData.total)}
                    query={chartData.name}
                    bgColor={appliedColors[index]}
                  />
                  <div className="mt-4">
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
    chartContent = (
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
        height={200}
        cardSize={unit.cardSize}
      />
    );
  } else {
    chartContent = (
      <NoBreakdownTable
        chartType={chartType}
        chartsData={chartsData}
        isWidgetModal={isWidgetModal}
      />
    );
  }

  return (
    <div
      style={{
        boxShadow:
          chartType === CHART_TYPE_SPARKLINES ||
          chartType === CHART_TYPE_LINECHART
            ? "inset 0px 1px 0px rgba(0, 0, 0, 0.1)"
            : "",
      }}
      className="w-full px-6"
    >
      {chartContent}
      {tableContent}
    </div>
  );
}

export default NoBreakdownCharts;
