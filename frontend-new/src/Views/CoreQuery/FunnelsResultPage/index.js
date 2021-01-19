import React from "react";
import Header from "../../AppLayout/Header";
import ResultsHeader from "../ResultsHeader";
import EventsInfo from "./EventsInfo";
import { Spin } from "antd";
import ResultantChart from "./ResultantChart";
import FiltersInfo from "../FiltersInfo";
import { QUERY_TYPE_FUNNEL } from "../../../utils/constants";

function FunnelsResultPage({
  queries,
  setDrawerVisible,
  resultState,
  breakdown,
  requestQuery,
  setShowResult,
  querySaved,
  setQuerySaved,
  handleDurationChange,
  durationObj,
  arrayMapper,
}) {
  let content = null;

  if (resultState.loading) {
    content = (
      <div className="mt-40 flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className="mt-40 flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

  if (resultState.data) {
    content = (
      <div className="mt-40 mb-8 fa-container">
        <ResultantChart
          queries={queries}
          setDrawerVisible={setDrawerVisible}
          resultState={resultState}
          breakdown={breakdown}
          requestQuery={requestQuery}
          setShowResult={setShowResult}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
          arrayMapper={arrayMapper}
        />
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
          queryType={QUERY_TYPE_FUNNEL}
        />
        <EventsInfo setDrawerVisible={setDrawerVisible} queries={queries} />
        <div className="py-4">
          <FiltersInfo
            durationObj={durationObj}
            handleDurationChange={handleDurationChange}
            breakdown={breakdown}
          />
        </div>
      </Header>
      {content}
    </>
  );
}

export default FunnelsResultPage;
