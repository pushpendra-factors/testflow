import React from "react";
import { QUERY_TYPE_ATTRIBUTION } from "../../../utils/constants";
import ResultsHeader from "../ResultsHeader";
import Header from "../../AppLayout/Header";
import { Spin } from "antd";
import AttributionsChart from "./AttributionsChart";

function AttributionsResult({
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  resultState
}) {
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

  if(resultState.data) {
    content = (
      <div className="mt-48 mb-8 fa-container">
        <AttributionsChart
          data={resultState.data}
        />
      </div>
    )
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
            className="text-base font-medium pb-4"
            style={{ color: "#8692A3" }}>
            Leads count (as unique users) vs Opportunities (as sum of opportunity value) - Last Touch
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
