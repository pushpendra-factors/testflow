import React from "react";
import AnalysisHeader from "./AnalysisHeader";
import ReportContent from "./ReportContent";

function ReportsLayout({
  queryType,
  setShowResult,
  requestQuery,
  querySaved,
  ...rest
}) {
  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={setShowResult.bind(this, false)}
        queryType={queryType}
        queryTitle={querySaved}
      />
      <div className="mt-24 px-20">
        <ReportContent queryTitle={querySaved} queryType={queryType} {...rest} />
      </div>
    </>
  );
}

export default ReportsLayout;
