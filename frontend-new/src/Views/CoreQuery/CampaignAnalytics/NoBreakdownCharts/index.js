import React, { useState, useEffect } from "react";
import { formatData, formatDataInLineChartFormat } from "./utils";
import ChartHeader from "../../../../components/SparkLineChart/ChartHeader";
import SparkChart from "../../../../components/SparkLineChart/Chart";
import { generateColors } from "../../../../utils/dataFormatter";
import LineChart from "../../../../components/LineChart";
import NoBreakdownTable from "./NoBreakdownTable";
import { CHART_TYPE_SPARKLINES } from "../../../../utils/constants";
import ChartTypeDropdown from "../../../../components/ChartTypeDropdown";

function NoBreakdownCharts({
  chartType,
  data,
  arrayMapper,
  isWidgetModal,
  setChartType,
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

  const table = (
    <div className="mt-16">
      <NoBreakdownTable
        chartType={chartType}
        chartsData={chartsData}
        isWidgetModal={isWidgetModal}
      />
    </div>
  );

  const menuItems = [
    {
      key: "sparklines",
      onClick: setChartType,
      name: "Sparkline",
    },
    {
      key: "linechart",
      onClick: setChartType,
      name: "Line Chart",
    },
  ];

  const typeDropdown = (
    <div className="flex items-center w-full mt-4 justify-end">
      <ChartTypeDropdown
        chartType={chartType}
        menuItems={menuItems}
        onClick={(item) => {
          setChartType(item.key);
        }}
      />
    </div>
  );

  if (chartType === CHART_TYPE_SPARKLINES) {
    if (chartsData.length === 1) {
      return (
        <>
          {typeDropdown}
          <div className="flex items-center flex-wrap mt-4 justify-center">
            <div className="w-1/4">
              <ChartHeader
                bgColor="#4D7DB4"
                query={chartsData[0].name}
                total={chartsData[0].total}
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
          {table}
        </>
      );
    }

    if (chartsData.length > 1) {
      const appliedColors = generateColors(chartsData.length);
      return (
        <>
          {typeDropdown}
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
                      total={chartData.total}
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
          {table}
        </>
      );
    }
  } else {
    const lineChartData = formatDataInLineChartFormat(chartsData);
    const appliedColors = generateColors(chartsData.length);
    return (
      <>
        {typeDropdown}
        <div className="w-full flex items-center mt-4 justify-center">
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
          />
        </div>
        {table}
      </>
    );
  }
}

export default NoBreakdownCharts;
