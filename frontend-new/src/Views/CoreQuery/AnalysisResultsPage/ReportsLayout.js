import React from "react";
import AnalysisHeader from "./AnalysisHeader";
import ReportContent from "./ReportContent";

function ReportsLayout({
  queryType,
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  breakdownType,
  ...rest
}) {
  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={setShowResult.bind(this, false)}
        queryType={queryType}
        queryTitle={querySaved}
        setQuerySaved={setQuerySaved}
        breakdownType={breakdownType}
      />
      <div className="mt-24 px-20">
        <ReportContent
          breakdownType={breakdownType}
          queryTitle={querySaved}
          queryType={queryType}
          {...rest}
        />
      </div>
    </>
  );
}

export default ReportsLayout;
