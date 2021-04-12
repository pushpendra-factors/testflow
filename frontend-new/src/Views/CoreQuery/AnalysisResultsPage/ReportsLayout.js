import React from "react";
import AnalysisHeader from "./AnalysisHeader";
import ReportContent from "./ReportContent";
import { FaErrorComp, FaErrorLog } from 'factorsComponents';
import {ErrorBoundary} from 'react-error-boundary';

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
      <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Analyse Results Error'} subtitle={'We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>
        <ReportContent
          breakdownType={breakdownType}
          queryTitle={querySaved}
          queryType={queryType}
          {...rest}
        />
        </ErrorBoundary>
      </div>
    </>
  );
}

export default ReportsLayout;
