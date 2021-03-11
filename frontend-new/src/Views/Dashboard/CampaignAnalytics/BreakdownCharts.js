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
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_BAR_CHART_HEIGHT,
  DASHBOARD_WIDGET_LINE_CHART_HEIGHT,
} from "../../../utils/constants";
import NoDataChart from '../../../components/NoDataChart';

function BreakdownCharts({
  arrayMapper,
  chartType,
  breakdown,
  data,
  isWidgetModal,
  setwidgetModal,
  unit,
  section,
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
  }, [data, arrayMapper, currentEventIndex, breakdown, maxAllowedVisibleProperties]);

  if (!chartsData.length) {
    return (
      <div className="mt-4 flex justify-center items-center w-full h-64 ">
        <NoDataChart />
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

  if (chartType === CHART_TYPE_BARCHART) {
    chartContent = (
      <BarChart
        section={section}
        height={DASHBOARD_WIDGET_BAR_CHART_HEIGHT}
        title={unit.id}
        chartData={visibleProperties}
        cardSize={unit.cardSize}
      />
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
    chartContent = (
      <>
        <LineChart
          frequency="date"
          chartData={lineChartData}
          hiddenEvents={[]}
          setHiddenEvents={() => {}}
          appliedColors={appliedColors}
          queries={visibleProperties.map((v) => v.label)}
          arrayMapper={mapper}
          isDecimalAllowed={false}
          cardSize={unit.cardSize}
          section={section}
          height={DASHBOARD_WIDGET_LINE_CHART_HEIGHT}
        />
      </>
    );
  } else {
    chartContent = (
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
    );
  }

  return (
    <div className={`w-full px-6 flex flex-1 flex-col  justify-center`}>
      {chartContent}
      {tableContent}
    </div>
  );
}

export default BreakdownCharts;
