import React, { useState, useEffect, useMemo, useContext } from 'react';
import { Spin, Tabs } from 'antd';

import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  EACH_USER_TYPE,
  QUERY_TYPE_WEB,
  CHART_TYPE_LINECHART
} from 'Utils/constants';
import { getChartTypeMenuItems } from 'Utils/chart.helpers';
import { Text, SVG } from 'Components/factorsComponents';

import ReportTitle from 'Views/CoreQuery/AnalysisResultsPage/ReportTitle';
import CalendarRow from 'Views/CoreQuery/AnalysisResultsPage/CalendarRow';
import CampaignMetricsDropdown from 'Views/CoreQuery/AnalysisResultsPage/CampaignMetricsDropdown';
import EventsAnalytics from 'Views/CoreQuery/EventsAnalytics';
import WebsiteAnalyticsTable from 'Views/Dashboard/WebsiteAnalytics/WebsiteAnalyticsTable';
import { CoreQueryContext } from 'Context/CoreQueryContext';
import KPIAnalysis from 'Views/CoreQuery/KPIAnalysis';
import {
  getChartType,
  shouldShowChartConfigOptions
} from 'Views/CoreQuery/AnalysisResultsPage/analysisResultsPage.helpers';
import { getKpiLabel } from 'Views/CoreQuery/KPIAnalysis/kpiAnalysis.helpers';
const { TabPane } = Tabs;

function ReportContent({
  resultState,
  queryType,
  setDrawerVisible,
  queries,
  breakdown,
  updateAppliedBreakdown,
  runKPIQuery,
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
  querySaved
}) {
  let content = null;
  let queryDetail = null;
  let durationObj = {};
  let metricsDropdown = <div className='mr-0' />;

  const {
    coreQueryState: { chartTypes }
  } = useContext(CoreQueryContext);

  const chartType = useMemo(
    () =>
      getChartType({
        breakdown,
        chartTypes,
        queryType
      }),
    [
      breakdown,
      chartTypes,
      queryType
    ]
  );
  // const chartType= CHART_TYPE_LINECHART;

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
        breakdown.length,
        queries
      )
    );
  }, [
    queryType,
    breakdown,
    breakdownType,
    queries
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

  

  if (queryType === QUERY_TYPE_KPI) {
    durationObj = queryOptions.date_range;
    queryDetail = queryTitle;
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
          breakdown={breakdown.length ? [
            {
              property: breakdown?.[0]?.na,
              prop_type: 'categorical',
              display_name: breakdown?.[0]?.d_na,
            }
          ]: []}
          section={section}
          currMetricsValue={currMetricsValue}
          durationObj={durationObj}
          chartType={chartType}
          renderedCompRef={renderedCompRef}
          secondAxisKpiIndices={secondAxisKpiIndices}
        />
      );
    }

  }

  function handleBreakdownChange(key) {
    const result = querySaved?.g_by?.filter((item) => key === item.na);
    updateAppliedBreakdown(result);
    runKPIQuery(querySaved, result[0]);
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
        {querySaved?.g_by?.length ?
          <div>
            <Tabs onChange={handleBreakdownChange} activeKey={breakdown?.[0]?.na} type="card" tabBarGutter={4}>
              {querySaved?.g_by?.map((item) => {
                return (
                  <TabPane tab={item.d_na} key={item.na}>
                  </TabPane>
                )
              })}
            </Tabs>
          </div>
        : null}
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
