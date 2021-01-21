import React, { useState, useEffect } from "react";
import { useDispatch } from "react-redux";
import moment from "moment";
import {
  QUERY_TYPE_CAMPAIGN,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_BARCHART,
} from "../../../utils/constants";
import styles from "../FunnelsResultPage/index.module.scss";
import ResultsHeader from "../ResultsHeader";
import Header from "../../AppLayout/Header";
import { Spin } from "antd";
import NoBreakdownCharts from "./NoBreakdownCharts";
import BreakdownCharts from "./BreakdownCharts";
import { SET_CAMP_MEASURES } from "../../../reducers/coreQuery/actions";

function CampaignAnalytics({
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  resultState,
  setDrawerVisible,
  arrayMapper,
  campaignState,
  title = "chart",
  isWidgetModal,
}) {
  const { group_by: breakdown } = campaignState;
  const [chartType, setChartType] = useState(null);
  const dispatch = useDispatch();

  useEffect(() => {
    return () => {
      dispatch({ type: SET_CAMP_MEASURES, payload: [] });
    };
  }, [dispatch]);

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
      <div
        className={`${
          isWidgetModal ? "mt-8" : "mt-48"
        } flex justify-center items-center w-full h-64`}
      >
        <Spin size="large" />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div
        className={`${
          isWidgetModal ? "mt-8" : "mt-48"
        } flex justify-center items-center w-full h-64`}
      >
        Something went wrong!
      </div>
    );
  }

  let chart = null;

  if (breakdown.length) {
    chart = (
      <BreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
        setChartType={setChartType}
        breakdown={breakdown}
        isWidgetModal={isWidgetModal}
        title={title}
      />
    );
  } else {
    chart = (
      <NoBreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={resultState.data}
        setChartType={setChartType}
        isWidgetModal={isWidgetModal}
      />
    );
  }

  if (resultState.data) {
    content = (
      <div className={`${isWidgetModal ? "mt-8" : "mt-40 fa-container"} mb-8`}>
        {chart}
      </div>
    );
  }

  return (
    <>
      {!isWidgetModal ? (
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
              {querySaved ||
                `Untitled Analysis ${moment().format("DD/MM/YYYY")}`}
            </div>
            <div
              className={`text-base font-medium pb-1 cursor-pointer ${styles.eventsText}`}
              style={{ color: "#3E516C" }}
              onClick={setDrawerVisible.bind(this, true)}
            >
              {arrayMapper.map((elem) => elem.eventName).join(", ")}
            </div>
          </div>
        </Header>
      ) : null}
      {content}
    </>
  );
}

export default CampaignAnalytics;
