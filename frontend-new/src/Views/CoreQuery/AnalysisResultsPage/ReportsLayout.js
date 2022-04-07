import React, {
  useCallback,
  useContext,
  useEffect,
  useState,
  useRef,
} from 'react';
import { Button } from 'antd';
import { ErrorBoundary } from 'react-error-boundary';
import { connect, useDispatch, useSelector } from 'react-redux';

import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_CAMPAIGN,
} from 'Utils/constants';
import { EMPTY_ARRAY } from 'Utils/global';

import AnalysisHeader from './AnalysisHeader';
import ReportContent from './ReportContent';
import WeeklyInsights from '../WeeklyInsights';

import {
  FaErrorComp,
  FaErrorLog,
  SVG,
} from '../../../components/factorsComponents';
import QueryComposer from '../../../components/QueryComposer';
import KPIComposer from '../../../components/KPIComposer';
import AttrQueryComposer from '../../../components/AttrQueryComposer';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { fetchWeeklyIngishts } from '../../../reducers/insights';
import ProfileComposer from '../../../components/ProfileComposer';
import { updateQuery } from '../../../reducers/coreQuery/services';
import { apiChartAnnotations } from '../../../utils/constants';
import { QUERY_UPDATED } from '../../../reducers/types';
import { getChartChangedKey } from './analysisResultsPage.helpers';

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
  breakdown,
  attributionsState,
  campaignState,
  composerFunctions,
  updateChartTypes,
  queryOptions,
  ...rest
}) {
  const dispatch = useDispatch();

  const savedQueries = useSelector((state) =>
    _.get(state, 'queries.data', EMPTY_ARRAY)
  );
  const { active_project } = useSelector((state) => state.global);

  const {
    setNavigatedFromDashboard,
    coreQueryState: { chartTypes },
  } = useContext(CoreQueryContext);

  const renderedCompRef = useRef(null);

  const [activeTab, setActiveTab] = useState(1);

  const [queryOpen, setQueryOpen] = useState(true);

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
          queries={composerFunctions.queries}
          runQuery={composerFunctions.runQuery}
          eventChange={composerFunctions.queryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={composerFunctions.setExtraOptions}
          runFunnelQuery={composerFunctions.runFunnelQuery}
          activeKey={composerFunctions.activeKey}
          collapse={composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return (
        <AttrQueryComposer
          runAttributionQuery={composerFunctions.runAttributionQuery}
          collapse={composerFunctions.showResult}
          queryOptions={composerFunctions.queryOptions}
          setQueryOptions={composerFunctions.setExtraOptions}
          setCollapse={() => setQueryOpen(false)}
        />
      );
    }

    if (queryType === QUERY_TYPE_KPI) {
      return (
        <KPIComposer
          queries={composerFunctions.queries}
          setQueryOptions={composerFunctions.setExtraOptions}
          eventChange={composerFunctions.queryChange}
          queryType={queryType}
          setQueryOptions={composerFunctions.setExtraOptions}
          activeKey={composerFunctions.activeKey}
          collapse={composerFunctions.showResult}
          setCollapse={() => setQueryOpen(false)}
          handleRunQuery={composerFunctions.runKPIQuery}
          setQueries={composerFunctions.setQueries}
          queryOptions={composerFunctions.queryOptions}
          selectedMainCategory={composerFunctions.selectedMainCategory}
          setSelectedMainCategory={composerFunctions.setSelectedMainCategory}
          KPIConfigProps={composerFunctions.KPIConfigProps}
        />
      );
    }

    if (queryType === QUERY_TYPE_PROFILE) {
      return (
        <ProfileComposer
          queries={composerFunctions.profileQueries}
          setQueries={composerFunctions.setProfileQueries}
          runProfileQuery={composerFunctions.runProfileQuery}
          eventChange={composerFunctions.profileQueryChange}
          queryType={queryType}
          queryOptions={queryOptions}
          setQueryOptions={composerFunctions.setExtraOptions}
          collapse={composerFunctions.showResult}
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

  const handleChartTypeChange = useCallback(
    ({ key, callUpdateService = true }) => {
      const changedKey = getChartChangedKey({
        queryType,
        breakdown,
        campaignGroupBy: campaignState.group_by,
        attributionModels: attributionsState.models,
      });

      updateChartTypes({
        ...chartTypes,
        [queryType]: {
          ...chartTypes[queryType],
          [changedKey]: key,
        },
      });

      if (savedQueryId && callUpdateService) {
        const queryGettingUpdated = savedQueries.find(
          (elem) => elem.id === savedQueryId
        );

        const settings = {
          ...queryGettingUpdated.settings,
          chart: apiChartAnnotations[key],
        };

        const reqBody = {
          title: queryGettingUpdated.title,
          settings,
        };

        updateQuery(active_project.id, savedQueryId, reqBody);

        dispatch({
          type: QUERY_UPDATED,
          queryId: savedQueryId,
          payload: reqBody,
        });
      }
    },
    [
      queryType,
      updateChartTypes,
      breakdown,
      chartTypes,
      campaignState.group_by,
      attributionsState.models,
      savedQueryId,
      savedQueries,
    ]
  );

  return (
    <>
      <AnalysisHeader
        requestQuery={requestQuery}
        onBreadCrumbClick={handleBreadCrumbClick}
        queryType={queryType}
        queryTitle={queryTitle}
        setQuerySaved={setQuerySaved}
        breakdownType={breakdownType}
        changeTab={setActiveTab}
        activeTab={activeTab}
        getCurrentSorter={getCurrentSorter}
        savedQueryId={savedQueryId}
        breakdown={breakdown}
        attributionsState={attributionsState}
        campaignState={campaignState}
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
                  breakdown={breakdown}
                  attributionsState={attributionsState}
                  campaignState={campaignState}
                  savedQueryId={savedQueryId}
                  handleChartTypeChange={handleChartTypeChange}
                  queryOptions={queryOptions}
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
