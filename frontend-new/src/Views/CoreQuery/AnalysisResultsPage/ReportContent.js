import React, { useState, useEffect, useMemo, useContext } from 'react';
import { useSelector } from 'react-redux';
import { Spin } from 'antd';
import get from 'lodash/get';
import isArray from 'lodash/isArray';

import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  EACH_USER_TYPE,
  QUERY_TYPE_WEB,
  CHART_TYPE_BARCHART,
  ReverseProfileMapper
} from 'Utils/constants';
import { toLetters } from 'Utils/dataFormatter';
import { getChartTypeMenuItems } from 'Utils/chart.helpers';
import { Text, SVG } from 'Components/factorsComponents';

import ReportTitle from './ReportTitle';
import FunnelsResultPage from '../FunnelsResultPage';
import CalendarRow from './CalendarRow';
import AttributionsResult from '../AttributionsResult';
import CampaignAnalytics from '../CampaignAnalytics';
import CampaignMetricsDropdown from './CampaignMetricsDropdown';
import EventsAnalytics from '../EventsAnalytics';
import WebsiteAnalyticsTable from '../../Dashboard/WebsiteAnalytics/WebsiteAnalyticsTable';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import KPIAnalysis from '../KPIAnalysis';
import ProfilesResultPage from '../ProfilesResultPage';
import {
  getChartType,
  shouldShowChartConfigOptions
} from './analysisResultsPage.helpers';
import { getKpiLabel } from '../KPIAnalysis/kpiAnalysis.helpers';
import { ATTRIBUTION_GROUP_ANALYSIS_KEYS } from '../AttributionsResult/attributionsResult.constants';
import NoDataChart from 'Components/NoDataChart';
const nodata = (
  <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
    <NoDataChart />
  </div>
);
function ReportContent({
  resultState,
  queryType,
  setDrawerVisible,
  queries,
  breakdown,
  queryOptions,
  handleDurationChange,
  campaignState,
  arrayMapper,
  attributionsState,
  breakdownType,
  queryTitle,
  eventPage,
  section,
  onReportClose,
  runAttrCmprQuery,
  campaignsArrayMapper,
  handleGranularityChange,
  renderedCompRef,
  handleChartTypeChange
}) {
  let content = null;
  let queryDetail = null;
  let durationObj = {};
  let groupAnalysis = '';
  let metricsDropdown = <div className='mr-0' />;

  const {
    coreQueryState: { chartTypes, comparison_data }
  } = useContext(CoreQueryContext);

  const { attrQueries } = useSelector((state) => state.coreQuery);

  const chartType = useMemo(
    () =>
      getChartType({
        breakdown,
        chartTypes,
        queryType,
        campaignGroupBy: campaignState.group_by,
        attributionModels: attributionsState.models
      }),
    [
      breakdown,
      campaignState.group_by,
      chartTypes,
      queryType,
      attributionsState.models
    ]
  );

  const [currMetricsValue, setCurrMetricsValue] = useState(0);
  const [chartTypeMenuItems, setChartTypeMenuItems] = useState([]);
  const [secondAxisKpiIndices, setSecondAxisKpiIndices] = useState([]);

  useEffect(() => {
    if (queryType === QUERY_TYPE_EVENT && breakdownType !== EACH_USER_TYPE) {
      setChartTypeMenuItems([]);
      return;
    }
    setChartTypeMenuItems(
      getChartTypeMenuItems(
        queryType,
        breakdown?.length,
        queries,
        attributionsState?.touchpoint
      )
    );
  }, [
    queryType,
    breakdown,
    breakdownType,
    queries,
    attributionsState?.touchpoint
  ]);

  if (resultState.loading) {
    content = (
      <div className='h-64 flex items-center justify-center w-full'>
        <Spin size='large' />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className='h-64 flex items-center justify-center w-full'>
        Something Went Wrong!
      </div>
    );
  }

  if (resultState.apiCallStatus && !resultState.apiCallStatus.required) {
    content = (
      <div className='h-64 flex flex-col items-center justify-center w-full'>
        <SVG name='nodata' />
        <Text type='title' color='grey' extraClass='mb-0'>
          {resultState.apiCallStatus.message}
        </Text>
      </div>
    );
  }

  if (queryType === QUERY_TYPE_WEB) {
    durationObj = queryOptions.date_range;
    queryDetail = 'Website Analytics';
  }

  if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
    durationObj = queryOptions.date_range;
    queryDetail = arrayMapper.map((elem) => elem.eventName).join(', ');
  }

  if (queryType === QUERY_TYPE_KPI) {
    durationObj = queryOptions.date_range;
    queryDetail = arrayMapper.map((elem) => elem.eventName).join(', ');
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    queryDetail = attributionsState.eventGoal.label;
    durationObj = attributionsState.date_range;
    if (
      attributionsState.models.length === 2 &&
      chartType === CHART_TYPE_BARCHART
    ) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={['Conversions', 'Cost Per Conversion']}
          currValue={currMetricsValue}
          onChange={setCurrMetricsValue}
        />
      );
    }
    if (
      attributionsState.models.length === 1 &&
      isArray(attrQueries) &&
      attrQueries.length > 1 &&
      get(
        queryOptions,
        'group_analysis',
        ATTRIBUTION_GROUP_ANALYSIS_KEYS.USERS
      ) !== ATTRIBUTION_GROUP_ANALYSIS_KEYS.USERS
    ) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={attrQueries.map((q) => q.label)}
          currValue={currMetricsValue}
          onChange={setCurrMetricsValue}
        />
      );
    }
  }

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    queryDetail = campaignState.select_metrics.join(', ');
    durationObj = campaignState.date_range;
    if (
      campaignState.select_metrics.length > 1 &&
      campaignState.group_by.length
    ) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={campaignState.select_metrics}
          currValue={currMetricsValue}
          onChange={setCurrMetricsValue}
        />
      );
    }
  }

  if (queryType === QUERY_TYPE_KPI && breakdown.length && queries.length > 1) {
    metricsDropdown = (
      <CampaignMetricsDropdown
        metrics={queries.map((q) => getKpiLabel(q))}
        currValue={currMetricsValue}
        onChange={setCurrMetricsValue}
      />
    );
  }

  if (queryType === QUERY_TYPE_PROFILE) {
    durationObj = queryOptions.date_range;
    groupAnalysis = queryOptions.group_analysis;
    if (queries.length > 1 && breakdown.length) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={queries.map(
            (_, index) =>
              `${
                ReverseProfileMapper[queries[index]]
                  ? ReverseProfileMapper[queries[index]][groupAnalysis]
                  : queries[index]
              } (${toLetters(index)})`
          )}
          currValue={currMetricsValue}
          onChange={setCurrMetricsValue}
        />
      );
    }
  }

  if (resultState.data) {
    if (queryType === QUERY_TYPE_WEB) {
      content = (
        <WebsiteAnalyticsTable
          tableData={resultState.data}
          section={section}
          isWidgetModal
        />
      );
    }

    if (queryType === QUERY_TYPE_FUNNEL) {
      content = (
        <FunnelsResultPage
          queries={queries}
          resultState={resultState}
          breakdown={breakdown}
          arrayMapper={arrayMapper}
          section={section}
          durationObj={durationObj}
          renderedCompRef={renderedCompRef}
          chartType={chartType}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      if (comparison_data.error) {
        content = nodata;
      } else {
        content = (
          <AttributionsResult
            resultState={resultState}
            durationObj={durationObj}
            attributionsState={attributionsState}
            section={section}
            currMetricsValue={currMetricsValue}
            renderedCompRef={renderedCompRef}
            chartType={chartType}
            queryOptions={queryOptions}
          />
        );
      }
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      content = (
        <CampaignAnalytics
          resultState={resultState}
          arrayMapper={campaignsArrayMapper}
          campaignState={campaignState}
          chartType={chartType}
          currMetricsValue={currMetricsValue}
          section={section}
          durationObj={durationObj}
          renderedCompRef={renderedCompRef}
        />
      );
    }

    if (queryType === QUERY_TYPE_EVENT) {
      content = (
        <EventsAnalytics
          resultState={resultState}
          arrayMapper={arrayMapper}
          chartType={chartType}
          breakdown={breakdown}
          queries={queries}
          page={eventPage}
          durationObj={durationObj}
          breakdownType={breakdownType}
          section={section}
          renderedCompRef={renderedCompRef}
        />
      );
    }

    if (queryType === QUERY_TYPE_KPI) {
      content = (
        <KPIAnalysis
          resultState={resultState}
          kpis={queries}
          breakdown={breakdown}
          section={section}
          currMetricsValue={currMetricsValue}
          durationObj={durationObj}
          chartType={chartType}
          renderedCompRef={renderedCompRef}
          secondAxisKpiIndices={secondAxisKpiIndices}
        />
      );
    }

    if (queryType === QUERY_TYPE_PROFILE) {
      content = (
        <ProfilesResultPage
          queries={queries}
          groupAnalysis={groupAnalysis}
          resultState={resultState}
          chartType={chartType}
          section={section}
          breakdown={breakdown}
          currMetricsValue={currMetricsValue}
          renderedCompRef={renderedCompRef}
        />
      );
    }
  }

  return (
    <>
      <>
        {queryType === QUERY_TYPE_CAMPAIGN || queryType === QUERY_TYPE_WEB ? (
          <ReportTitle
            setDrawerVisible={setDrawerVisible}
            title={queryTitle}
            queryDetail={queryDetail}
            section={section}
            onReportClose={onReportClose}
            queryType={queryType}
            apiCallStatus={resultState.apiCallStatus}
          />
        ) : null}
        <div className='mt-6'>
          <CalendarRow
            queryType={queryType}
            handleDurationChange={handleDurationChange}
            durationObj={durationObj}
            handleChartTypeChange={handleChartTypeChange}
            chartTypeMenuItems={chartTypeMenuItems}
            chartType={chartType}
            metricsDropdown={metricsDropdown}
            triggerAttrComparision={runAttrCmprQuery}
            handleGranularityChange={handleGranularityChange}
            section={section}
            setSecondAxisKpiIndices={setSecondAxisKpiIndices}
            secondAxisKpiIndices={secondAxisKpiIndices}
            showChartConfigOptions={shouldShowChartConfigOptions({
              queryType,
              breakdown,
              chartType
            })}
            kpis={queries}
          />
        </div>
      </>

      <div className='mt-12'>{content}</div>
    </>
  );
}

export default ReportContent;
