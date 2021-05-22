import React, { useCallback, useContext } from 'react';
import AnalysisHeader from './AnalysisHeader';
import ReportContent from './ReportContent';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';

function ReportsLayout({
  queryType,
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  breakdownType,
  ...rest
}) {
  const { setNavigatedFromDashboard } = useContext(CoreQueryContext);

  const handleBreadCrumbClick = useCallback(() => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
  }, [setNavigatedFromDashboard, setShowResult]);

  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={handleBreadCrumbClick}
        queryType={queryType}
        queryTitle={querySaved}
        setQuerySaved={setQuerySaved}
        breakdownType={breakdownType}
      />
      <div className='mt-24 px-20'>
        <ErrorBoundary
          fallback={
            <FaErrorComp
              size={'medium'}
              title={'Analyse Results Error'}
              subtitle={
                'We are facing trouble loading Analyse results. Drop us a message on the in-app chat.'
              }
            />
          }
          onError={FaErrorLog}
        >
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
