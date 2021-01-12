import React, { useState, useEffect } from "react";
import {
  formatData,
  formatDataInLineChartFormat,
} from "../../CoreQuery/CampaignAnalytics/BreakdownCharts/utils";
import BarChart from "../../../components/BarChart";
import BreakdownTable from "../../CoreQuery/CampaignAnalytics/BreakdownCharts/BreakdownTable";
import LineChart from "../../../components/LineChart";
import { generateColors } from "../../../utils/dataFormatter";
import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_LINECHART,
} from "../../../utils/constants";

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  isWidgetModal,
}) {
  const [chartsData, setChartsData] = useState([]);
  const [currentEventIndex, setCurrentEventIndex] = useState(0);
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
        No Data Found!
      </div>
    );
  }

  if (chartType === CHART_TYPE_BARCHART) {
    return (
      <div className="flex items-center flex-wrap mt-4 justify-center">
        <BarChart chartData={visibleProperties} />
      </div>
    );
  } else if (chartType === CHART_TYPE_LINECHART) {
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
    return (
      <>
        <div className="flex items-center flex-wrap mt-4 justify-center">
          <LineChart
            frequency="date"
            chartData={lineChartData}
            hiddenEvents={[]}
            setHiddenEvents={() => {}}
            appliedColors={appliedColors}
            queries={visibleProperties.map((v) => v.label)}
            arrayMapper={mapper}
            isDecimalAllowed={false}
          />
        </div>
      </>
    );
  } else {
    return (
      <BreakdownTable
        currentEventIndex={currentEventIndex}
        chartType={chartType}
        chartsData={chartsData}
        breakdown={breakdown}
        arrayMapper={arrayMapper}
        isWidgetModal={isWidgetModal}
        responseData={data}
        visibleProperties={visibleProperties}
        maxAllowedVisibleProperties={maxAllowedVisibleProperties}
        setVisibleProperties={setVisibleProperties}
      />
    )
  }
}

export default BreakdownCharts;
