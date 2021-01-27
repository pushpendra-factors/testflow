import React from "react";
import ReportTitle from "./ReportTitle";
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
} from "../../../utils/constants";
import { Spin } from "antd";
import FunnelsResultPage from "../FunnelsResultPage";
import CalendarRow from "./CalendarRow";
import AttributionsResult from "../AttributionsResult";

function ReportContent({
  resultState,
  queryType,
  setDrawerVisible,
  queries,
  breakdown,
  requestQuery,
  querySaved,
  setQuerySaved,
  queryOptions,
  handleDurationChange,
  arrayMapper,
  attributionsState,
}) {
  let content = null,
    queryDetail = null;
  if (resultState.loading) {
    content = (
      <div className="h-64 flex items-center justify-center w-full">
        <Spin size={"large"} />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className="h-64 flex items-center justify-center w-full">
        Something Went Wrong!
      </div>
    );
  }

  if (queryType === QUERY_TYPE_FUNNEL) {
    queryDetail = arrayMapper
      .map((elem) => {
        return elem.eventName;
      })
      .join(", ");
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    queryDetail = attributionsState.eventGoal.label;
  }

  if (resultState.data) {
    if (queryType === QUERY_TYPE_FUNNEL) {
      content = (
        <FunnelsResultPage
          setDrawerVisible={setDrawerVisible}
          queries={queries}
          resultState={resultState}
          breakdown={breakdown}
          requestQuery={requestQuery}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
          arrayMapper={arrayMapper}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      content = (
        <AttributionsResult
          resultState={resultState}
          attributionsState={attributionsState}
        />
      );
    }
  }

  return (
    <>
      <ReportTitle
        setDrawerVisible={setDrawerVisible}
        title={""}
        queryDetail={queryDetail}
      />
      <div className="mt-6">
        <CalendarRow durationObj={queryOptions.date_range} />
      </div>
      <div className="mt-12">{content}</div>
    </>
  );
}

export default ReportContent;
