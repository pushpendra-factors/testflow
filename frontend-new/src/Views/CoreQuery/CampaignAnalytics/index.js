import React, { useState, useEffect } from "react";
import {
  QUERY_TYPE_CAMPAIGN,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_LINECHART,
  CHART_TYPE_BARCHART,
} from "../../../utils/constants";
import styles from "../FunnelsResultPage/index.module.scss";
import ResultsHeader from "../ResultsHeader";
import Header from "../../AppLayout/Header";
import { Spin } from "antd";
import NoBreakdownCharts from "./NoBreakdownCharts";
import BreakdownCharts from "./BreakdownCharts";

function CampaignAnalytics({
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  resultState,
  setDrawerVisible,
  arrayMapper,
  breakdown,
}) {
  const [chartType, setChartType] = useState(null);

  useEffect(() => {
    if (breakdown.length) {
      setChartType(CHART_TYPE_BARCHART);
    } else {
      setChartType(CHART_TYPE_SPARKLINES);
    }
  }, [breakdown]);

  let content = null;

  if (resultState.loading) {
    content = (
      <div className="mt-48 flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className="mt-48 flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

  if (resultState.data) {
    if (breakdown.length) {
      content = (
        <div className="mt-40 mb-8 fa-container">
          <BreakdownCharts
            arrayMapper={arrayMapper}
            chartType={chartType}
            data={resultState.data}
            setChartType={setChartType}
            breakdown={breakdown}
            isWidgetModal={false}
          />
        </div>
      );
    } else {
      content = (
        <div className="mt-40 mb-8 fa-container">
          <NoBreakdownCharts
            arrayMapper={arrayMapper}
            chartType={chartType}
            data={resultState.data}
            setChartType={setChartType}
            isWidgetModal={false}
          />
        </div>
      );
    }
  }

  return (
    <>
      <Header>
        <ResultsHeader
          setShowResult={setShowResult}
          requestQuery={requestQuery}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
          queryType={QUERY_TYPE_CAMPAIGN}
        />

        <div className="pt-4">
          <div
            className="app-font-family text-3xl font-semibold"
            style={{ color: "#8692A3" }}
          >
            {querySaved || "Untitled Report"}
          </div>
          <div
            className={`text-base font-medium pb-1 cursor-pointer ${styles.eventsText}`}
            style={{ color: "#8692A3" }}
            onClick={setDrawerVisible.bind(this, true)}
          >
            impressions
          </div>
        </div>
      </Header>
      {content}
    </>
  );
}

export default CampaignAnalytics;
