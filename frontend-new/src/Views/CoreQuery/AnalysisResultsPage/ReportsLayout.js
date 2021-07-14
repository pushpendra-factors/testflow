import React, { useCallback, useContext, useEffect, useState } from 'react';
import AnalysisHeader from './AnalysisHeader';
import ReportContent from './ReportContent';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import WeeklyInsights from '../WeeklyInsights';
import { fetchWeeklyIngishts } from '../../../reducers/insights';
import { connect, useDispatch } from 'react-redux';

function ReportsLayout({
  queryType,
  setShowResult,
  requestQuery,
  querySaved,
  setQuerySaved,
  breakdownType,
  activeProject,
  fetchWeeklyIngishts,
  ...rest
}) {
  const { setNavigatedFromDashboard } = useContext(CoreQueryContext);
  const [activeTab, setActiveTab] = useState(1);
  // const [insights, setInsights] = useState(null);
  const dispatch = useDispatch();

  const handleBreadCrumbClick = useCallback(() => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
  }, [setNavigatedFromDashboard, setShowResult]);

  function changeTab(key) {
    // console.log('current tab is=-->>',key);
    setActiveTab(key);
  }

  useEffect(() => {
    return () => {
      dispatch({ type: 'SET_ACTIVE_INSIGHT', payload: false });
      dispatch({ type: 'RESET_WEEKLY_INSIGHTS', payload: false });
    };
  }, [dispatch, activeProject]);

  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={handleBreadCrumbClick}
        queryType={queryType}
        queryTitle={querySaved}
        setQuerySaved={setQuerySaved}
        breakdownType={breakdownType}
        changeTab={changeTab}
        activeTab={activeTab}
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
          {Number(activeTab) === 1 && (
            <>
              <ReportContent
                breakdownType={breakdownType}
                queryTitle={querySaved}
                queryType={queryType}
                {...rest}
              />
            </>
          )}

          {Number(activeTab) === 2 && (
            <>
              <WeeklyInsights
                requestQuery={requestQuery}
                queryType={queryType}
                queryTitle={querySaved}
              />
            </>
          )}
        </ErrorBoundary>
      </div>
    </>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  insights: state.insights,
});

export default connect(mapStateToProps, { fetchWeeklyIngishts })(ReportsLayout);
