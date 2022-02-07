import React, {
  useCallback,
  useContext,
  useEffect,
  useState,
  useRef,
} from 'react';
import AnalysisHeader from './AnalysisHeader';
import ReportContent from './ReportContent';
import {
  FaErrorComp,
  FaErrorLog,
  SVG,
} from '../../../components/factorsComponents';
import QueryComposer from '../../../components/QueryComposer';
import KPIComposer from '../../../components/KPIComposer';
import AttrQueryComposer from '../../../components/AttrQueryComposer';
import { ErrorBoundary } from 'react-error-boundary';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import WeeklyInsights from '../WeeklyInsights';
import { fetchWeeklyIngishts } from '../../../reducers/insights';
import { connect, useDispatch } from 'react-redux';
import ProfileComposer from '../../../components/ProfileComposer';

import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
} from '../../../utils/constants';

import { Button } from 'antd';

function ReportsLayout({
  queryType,
  setShowResult,
  requestQuery,
  queryTitle,
  setQuerySaved,
  breakdownType,
  activeProject,
  fetchWeeklyIngishts,
  savedQueryId,
  ...rest
}) {
  const renderedCompRef = useRef(null);
  const { setNavigatedFromDashboard } = useContext(CoreQueryContext);
  const [activeTab, setActiveTab] = useState(1);

  const [queryOpen, setQueryOpen] = useState(true);
  // const [insights, setInsights] = useState(null);
  const dispatch = useDispatch();

  const handleBreadCrumbClick = useCallback(() => {
    setShowResult(false);
    setNavigatedFromDashboard(false);
  }, [setNavigatedFromDashboard, setShowResult]);

  const getCurrentSorter = useCallback(() => {
    if (renderedCompRef.current && renderedCompRef.current.currentSorter) {
      return renderedCompRef.current.currentSorter;
    }
    return [];
  }, []);

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

  useEffect(() => {
    if (requestQuery) {
      setQueryOpen(false);
    } else {
      setQueryOpen(true);
    }
  }, [requestQuery]);

  const renderComposer = () => {
    if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
      return (
        <QueryComposer
          queries={rest.composerFunctions.queries}
          runQuery={rest.composerFunctions.runQuery}
          eventChange={rest.composerFunctions.queryChange}
          queryType={queryType}
          queryOptions={rest.queryOptions}
          setQueryOptions={rest.composerFunctions.setExtraOptions}
          runFunnelQuery={rest.composerFunctions.runFunnelQuery}
          activeKey={rest.composerFunctions.activeKey}
          collapse={rest.composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return (
        <AttrQueryComposer
          runAttributionQuery={rest.composerFunctions.runAttributionQuery}
          collapse={rest.composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }

    if (queryType === QUERY_TYPE_KPI) {
      return (
        <KPIComposer
          // queries={rest.composerFunctions.queries}
          // runQuery={rest.composerFunctions.runQuery}
          // eventChange={rest.composerFunctions.queryChange}
          // queryType={queryType}
          // queryOptions={rest.queryOptions}
          // setQueryOptions={rest.composerFunctions.setExtraOptions}
          // runFunnelQuery={rest.composerFunctions.runFunnelQuery}
          // activeKey={rest.composerFunctions.activeKey}
          // collapse={rest.composerFunctions.showResult}
          // setCollapse={() => setQueryOpen(false)}

          queries={rest.composerFunctions.queries}
          setQueryOptions={rest.composerFunctions.setExtraOptions}
          eventChange={rest.composerFunctions.queryChange}
          queryType={queryType}
          setQueryOptions={rest.composerFunctions.setExtraOptions}
          activeKey={rest.composerFunctions.activeKey}
          collapse={rest.composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
          handleRunQuery={rest.composerFunctions.runKPIQuery}
          setQueries={rest.composerFunctions.setQueries}
          queryOptions={rest.composerFunctions.queryOptions}
          selectedMainCategory={rest.composerFunctions.selectedMainCategory}
          setSelectedMainCategory={
            rest.composerFunctions.setSelectedMainCategory
          }
          KPIConfigProps={rest.composerFunctions.KPIConfigProps}
        />
      );
    }

    if (queryType === QUERY_TYPE_PROFILE) {
      return (
        <ProfileComposer
          queries={rest.composerFunctions.profileQueries}
          setQueries={rest.composerFunctions.setProfileQueries}
          runProfileQuery={rest.composerFunctions.runProfileQuery}
          eventChange={rest.composerFunctions.profileQueryChange}
          queryType={queryType}
          queryOptions={rest.queryOptions}
          setQueryOptions={rest.composerFunctions.setExtraOptions}
          collapse={rest.composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }
  };

  const renderQueryComposerNew = () => {
    if (
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_ATTRIBUTION ||
      queryType === QUERY_TYPE_KPI ||
      queryType === QUERY_TYPE_PROFILE
    ) {
      return (
        <div
          className={`query_card_cont ${
            queryOpen ? `query_card_open` : `query_card_close`
          }`}
          onClick={(e) => !queryOpen && setQueryOpen(true)}
        >
          <div className={`query_composer`}>{renderComposer()}</div>
          <Button size={'large'} className={`query_card_expand`}>
            <SVG name={'expand'} size={20}></SVG>Expand
          </Button>
        </div>
      );
    }
    return null;
  };

  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={handleBreadCrumbClick}
        queryType={queryType}
        queryTitle={queryTitle}
        setQuerySaved={setQuerySaved}
        breakdownType={breakdownType}
        changeTab={changeTab}
        activeTab={activeTab}
        getCurrentSorter={getCurrentSorter}
        savedQueryId={savedQueryId}
      />
      <div className='mt-24 px-8'>
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
              {renderQueryComposerNew()}
              {requestQuery && (
                <ReportContent
                  breakdownType={breakdownType}
                  queryType={queryType}
                  renderedCompRef={renderedCompRef}
                  {...rest}
                />
              )}
            </>
          )}

          {Number(activeTab) === 2 && (
            <WeeklyInsights requestQuery={requestQuery} queryType={queryType} />
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
