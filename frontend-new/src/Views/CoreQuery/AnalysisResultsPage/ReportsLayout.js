import React from "react";
import AnalysisHeader from "./AnalysisHeader";
import ReportContent from "./ReportContent";

function ReportsLayout({ queryType, setShowResult, requestQuery, breakdownType, ...rest }) {
  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={setShowResult.bind(this, false)}
        queryType={queryType}
        breakdownType={breakdownType}
      />
      <div className="mt-24 px-20">
        <ReportContent queryType={queryType} {...rest} />
      </div>
    </>
  );
}

export default ReportsLayout;
