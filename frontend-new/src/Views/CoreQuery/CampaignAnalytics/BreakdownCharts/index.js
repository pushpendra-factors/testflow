import React, { useState, useEffect } from "react";
import { formatData, formatDataInLineChartFormat, formatDataInHighChartsFormat } from "./utils";
import { CHART_TYPE_BARCHART, CHART_TYPE_LINECHART, DASHBOARD_MODAL, CHART_TYPE_STACKED_AREA, CHART_TYPE_STACKED_BAR } from "../../../../utils/constants";
import BarChart from "../../../../components/BarChart";
import BreakdownTable from "./BreakdownTable";
import LineChart from "../../../../components/LineChart";
import { generateColors } from "../../../../utils/dataFormatter";
import NoDataChart from 'Components/NoDataChart';
import StackedAreaChart from "../../../../components/StackedAreaChart";
import StackedBarChart from "../../../../components/StackedBarChart";

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  title = "chart",
  currentEventIndex,
  section
}) {
  const [chartsData, setChartsData] = useState([]);
  const [visibleProperties, setVisibleProperties] = useState([]);
  const maxAllowedVisibleProperties = 5;

  useEffect(() => {
    const formattedData = formatData(
      data,
      arrayMapper,
      breakdown,
      currentEventIndex
    );
    setVisibleProperties(formattedData.slice(0, maxAllowedVisibleProperties));
    setChartsData(formattedData);
  }, [data, arrayMapper, currentEventIndex, breakdown]);

  if (!chartsData.length) {
    return (
      <div className="mt-4 flex justify-center items-center w-full h-64 ">
        <NoDataChart />
      </div>
    );
  }

  const table = (
    <div className="mt-12 w-full">
      <BreakdownTable
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        chartsData={chartsData}
        breakdown={breakdown}
        arrayMapper={arrayMapper}
        isWidgetModal={section === DASHBOARD_MODAL}
        responseData={data}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        setVisibleProperties={setVisibleProperties}
      />
    </div>
  );

  let chart = null;

  if (chartType === CHART_TYPE_BARCHART) {
    chart = <BarChart section={section} title={title} chartData={visibleProperties} />;
  } else if(chartType === CHART_TYPE_STACKED_AREA) {
    const { categories, highchartsData } = formatDataInHighChartsFormat(
      data.result_group[0],
      arrayMapper,
      currentEventIndex,
      visibleProperties
    );
    chart = (
      <div className="w-full">
        <StackedAreaChart
          frequency="date"
          categories={categories}
          data={highchartsData}
        />
      </div>
    ); 
  } else if(chartType === CHART_TYPE_STACKED_BAR) {
    const { categories, highchartsData } = formatDataInHighChartsFormat(
      data.result_group[0],
      arrayMapper,
      currentEventIndex,
      visibleProperties
    );
    chart = (
      <div className="w-full">
        <StackedBarChart
          frequency="date"
          categories={categories}
          data={highchartsData}
        />
      </div>
    ); 
  } else if(chartType === CHART_TYPE_LINECHART) {
    const mapper = visibleProperties.map((v, index) => {
      return {
        index: index,
        mapper: `event${index + 1}`,
        eventName: v.label,
      };
    });
    const lineChartData = formatDataInLineChartFormat(
      visibleProperties,
      data,
      breakdown,
      currentEventIndex,
      arrayMapper,
      mapper
    );
    const appliedColors = generateColors(visibleProperties.length);
    chart = (
      <LineChart
        frequency="date"
        chartData={lineChartData}
        hiddenEvents={[]}
        setHiddenEvents={() => {}}
        appliedColors={appliedColors}
        queries={visibleProperties.map((v) => v.label)}
        arrayMapper={mapper}
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

export default BreakdownCharts;
