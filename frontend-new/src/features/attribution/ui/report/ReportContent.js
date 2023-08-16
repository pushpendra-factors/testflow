import React, { useState, useEffect, useMemo, useContext } from 'react';
import { useSelector } from 'react-redux';
import { Spin } from 'antd';
import isArray from 'lodash/isArray';

import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  EACH_USER_TYPE,
  QUERY_TYPE_WEB,
  CHART_TYPE_BARCHART
} from 'Utils/constants';
import { getChartTypeMenuItems } from 'Utils/chart.helpers';
import { Text, SVG } from 'Components/factorsComponents';

import ReportTitle from 'Views/CoreQuery/AnalysisResultsPage/ReportTitle';
import CalendarRow from 'Views/CoreQuery/AnalysisResultsPage/CalendarRow';
import AttributionsResult from 'Views/CoreQuery/AttributionsResult';
import CampaignMetricsDropdown from 'Views/CoreQuery/AnalysisResultsPage/CampaignMetricsDropdown';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import {
  getChartType,
  shouldShowChartConfigOptions
} from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';

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
  let metricsDropdown = <div className='mr-0' />;

  const [attributionTableFilters, setAttributionTableFilters] = useState([]);

  const {
    coreQueryState: { chartTypes }
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
    }
    setChartTypeMenuItems(
      getChartTypeMenuItems(
        queryType,
        breakdown?.length,
        attributionsState?.attrQueries? attributionsState?.attrQueries : queries,
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
      isArray(attributionsState.attrQueries) &&
      attributionsState.attrQueries.length > 1
    ) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={attributionsState.attrQueries.map((q) => q.label)}
          currValue={currMetricsValue}
          onChange={setCurrMetricsValue}
        />
      );
    }
  }

  if (resultState.data) {
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
        appliedFilters={attributionTableFilters}
        setAttributionTableFilters={setAttributionTableFilters}
        v1={true}
      />
    );
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
            kpis={attributionsState?.attrQueries? attributionsState?.attrQueries: queries}
          />
        </div>
      </>

      <div className='mt-12'>{content}</div>
    </>
  );
}

export default ReportContent;
