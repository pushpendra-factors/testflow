import React from "react";
import {
  QUERY_TYPE_ATTRIBUTION,
  ATTRIBUTION_METHODOLOGY,
} from "../../../utils/constants";
import styles from "../FunnelsResultPage/index.module.scss";
import ResultsHeader from "../ResultsHeader";
import Header from "../../AppLayout/Header";
import { Spin } from "antd";
import AttributionsChart from "./AttributionsChart";
import GroupedAttributionsChart from "./GroupedAttributionsChart";

function AttributionsResult({
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  resultState,
  setDrawerVisible,
  attributionsState,
}) {
  let content = null;

  const { eventGoal, touchpoint, models, linkedEvents } = attributionsState;

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
    content = (
      <div className="mt-48 mb-8 fa-container">
        {models.length === 1 ? (
          <AttributionsChart
            event={eventGoal.label}
            linkedEvents={linkedEvents}
            touchpoint={touchpoint}
            data={resultState.data}
            isWidgetModal={false}
            attribution_method={models[0]}
          />
        ) : null}
        {models.length === 2 ? (
          <GroupedAttributionsChart
            event={eventGoal.label}
            linkedEvents={linkedEvents}
            touchpoint={touchpoint}
            data={resultState.data}
            isWidgetModal={false}
            attribution_method={models[0]}
            attribution_method_compare={models[1]}
          />
        ) : null}
      </div>
    );
  }

  return (
    <>
      <Header>
        <ResultsHeader
          setShowResult={setShowResult}
          requestQuery={requestQuery}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
          queryType={QUERY_TYPE_ATTRIBUTION}
        />

        <div className="pt-4">
          <div
            className="app-font-family text-3xl font-semibold"
            style={{ color: "#8692A3" }}
          >
            Untitled Analysis
          </div>
          <div
            className={`text-base font-medium pb-1 cursor-pointer ${styles.eventsText}`}
            style={{ color: "#8692A3" }}
            onClick={setDrawerVisible.bind(this, true)}
          >
            {eventGoal.label} (unique users)
          </div>
        </div>

        {/* <div className="py-4">
          <FiltersInfo
            durationObj={durationObj}
            handleDurationChange={handleDurationChange}
            breakdown={breakdown}
          />
        </div> */}
      </Header>
      {content}
    </>
  );
}

export default AttributionsResult;
