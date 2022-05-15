import React, { useState, useEffect, useMemo, useContext } from 'react';
import { Spin } from 'antd';

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
  ReverseProfileMapper,
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
import { getChartType } from './analysisResultsPage.helpers';
import { getKpiLabel } from '../KPIAnalysis/kpiAnalysis.helpers';

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
  handleChartTypeChange,
}) {
  let content = null,
    queryDetail = null,
    durationObj = {},
    groupAnalysis = '',
    metricsDropdown = <div className='mr-0'></div>;

  const {
    coreQueryState: { chartTypes },
  } = useContext(CoreQueryContext);

  const chartType = useMemo(() => {
    return getChartType({
      breakdown,
      chartTypes,
      queryType,
      campaignGroupBy: campaignState.group_by,
      attributionModels: attributionsState.models,
    });
  }, [
    breakdown,
    campaignState.group_by,
    chartTypes,
    queryType,
    attributionsState.models,
  ]);

  const [currMetricsValue, setCurrMetricsValue] = useState(0);
  const [chartTypeMenuItems, setChartTypeMenuItems] = useState([]);

  useEffect(() => {
    let items = [];

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      items = getChartTypeMenuItems(queryType, campaignState.group_by.length);
    }

    if (queryType === QUERY_TYPE_KPI) {
      items = getChartTypeMenuItems(queryType, breakdown.length);
    }

    if (
      (queryType === QUERY_TYPE_EVENT && breakdownType === EACH_USER_TYPE) ||
      queryType === QUERY_TYPE_FUNNEL
    ) {
      items = getChartTypeMenuItems(queryType, breakdown.length, queries);
    }
    if (queryType === QUERY_TYPE_PROFILE) {
      items = getChartTypeMenuItems(queryType, breakdown.length, queries);
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      items = getChartTypeMenuItems(queryType);
    }
    setChartTypeMenuItems(items.length > 1 ? items : []);
  }, [queryType, campaignState.group_by, breakdown, breakdownType, queries]);

  if (resultState.loading) {
    content = (
      <div className='h-64 flex items-center justify-center w-full'>
        <Spin size={'large'} />
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
    queryDetail = arrayMapper
      .map((elem) => {
        return elem.eventName;
      })
      .join(', ');
  }

  if (queryType === QUERY_TYPE_KPI) {
    durationObj = queryOptions.date_range;
    queryDetail = arrayMapper
      .map((elem) => {
        return elem.eventName;
      })
      .join(', ');
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
          isWidgetModal={true}
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
          />
        </div>
      </>

      <div className='mt-12'>{content}</div>
    </>
  );
}

export default ReportContent;
