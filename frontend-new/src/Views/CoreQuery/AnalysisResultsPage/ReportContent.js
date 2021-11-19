import React, {
  useState,
  useCallback,
  useEffect,
  useMemo,
  useContext,
} from 'react';
import ReportTitle from './ReportTitle';
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
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  presentationObj,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_STACKED_BAR,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
} from '../../../utils/constants';
import { Spin } from 'antd';
import FunnelsResultPage from '../FunnelsResultPage';
import CalendarRow from './CalendarRow';
import AttributionsResult from '../AttributionsResult';
import CampaignAnalytics from '../CampaignAnalytics';
import { getChartTypeMenuItems, toLetters } from '../../../utils/dataFormatter';
import CampaignMetricsDropdown from './CampaignMetricsDropdown';
import EventsAnalytics from '../EventsAnalytics';
import WebsiteAnalyticsTable from '../../Dashboard/WebsiteAnalytics/WebsiteAnalyticsTable';
import { CoreQueryContext } from '../../../contexts/CoreQueryContext';
import { Text, SVG } from '../../../components/factorsComponents';
import KPIAnalysis from '../KPIAnalysis';
import ProfilesResultPage from '../ProfilesResultPage';

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
  updateChartTypes,
  renderedCompRef,
}) {
  let content = null,
    queryDetail = null,
    durationObj = {},
    metricsDropdown = <div className='mr-0'></div>;

  // const KPIBreakdown = useMemo(() => {
  //   return breakdown.length ? ['$browser_version', '$source', '$medium'] : [];
  // }, [breakdown]);
  // const KPIQueries = useMemo(() => {
  //   return ['$session', '$form_submitted'];
  // }, []);

  const {
    coreQueryState: { chartTypes, navigatedFromDashboard },
  } = useContext(CoreQueryContext);

  const chartType = useMemo(() => {
    let key;
    if (queryType === QUERY_TYPE_FUNNEL) {
      key = breakdown.length ? 'breakdown' : 'no_breakdown';
      return chartTypes[queryType][key] === CHART_TYPE_TABLE
        ? CHART_TYPE_BARCHART
        : chartTypes[queryType][key];
    }
    if (queryType === QUERY_TYPE_KPI) {
      key = breakdown.length ? 'breakdown' : 'no_breakdown';
      return chartTypes[queryType][key] === CHART_TYPE_TABLE
        ? breakdown.length
          ? CHART_TYPE_BARCHART
          : CHART_TYPE_SPARKLINES
        : chartTypes[queryType][key];
    }
    if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_PROFILE) {
      key = breakdown.length ? 'breakdown' : 'no_breakdown';
      if (breakdown.length >= 1) {
        return chartTypes[queryType][key] === CHART_TYPE_TABLE
          ? CHART_TYPE_BARCHART
          : chartTypes[queryType][key];
      }
      return chartTypes[queryType][key] === CHART_TYPE_TABLE
        ? CHART_TYPE_SPARKLINES
        : chartTypes[queryType][key];
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      key = campaignState.group_by.length ? 'breakdown' : 'no_breakdown';
      if (campaignState.group_by.length >= 1) {
        return chartTypes[queryType][key] === CHART_TYPE_TABLE
          ? CHART_TYPE_BARCHART
          : chartTypes[queryType][key];
      }
      return chartTypes[queryType][key] === CHART_TYPE_TABLE
        ? CHART_TYPE_SPARKLINES
        : chartTypes[queryType][key];
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      key =
        attributionsState.models.length === 1
          ? 'single_touch_point'
          : 'dual_touch_point';
      return chartTypes[queryType][key] === CHART_TYPE_TABLE
        ? CHART_TYPE_BARCHART
        : chartTypes[queryType][key];
    }
  }, [
    breakdown,
    campaignState.group_by,
    chartTypes,
    queryType,
    attributionsState.models,
  ]);

  const [currMetricsValue, setCurrMetricsValue] = useState(0);
  const [chartTypeMenuItems, setChartTypeMenuItems] = useState([]);

  const handleChartTypeChange = useCallback(
    ({ key }) => {
      let changedKey;
      if (
        queryType === QUERY_TYPE_EVENT ||
        queryType === QUERY_TYPE_FUNNEL ||
        queryType === QUERY_TYPE_PROFILE ||
        queryType === QUERY_TYPE_KPI
      ) {
        changedKey = breakdown.length ? 'breakdown' : 'no_breakdown';
      }
      if (queryType === QUERY_TYPE_CAMPAIGN) {
        changedKey = campaignState.group_by.length
          ? 'breakdown'
          : 'no_breakdown';
      }
      if (queryType === QUERY_TYPE_ATTRIBUTION) {
        changedKey =
          attributionsState.models.length > 1
            ? 'dual_touch_point'
            : 'single_touch_point';
      }
      updateChartTypes({
        ...chartTypes,
        [queryType]: {
          ...chartTypes[queryType],
          [changedKey]: key,
        },
      });
    },
    [
      queryType,
      updateChartTypes,
      breakdown,
      campaignState.group_by,
      chartTypes,
      attributionsState.models,
    ]
  );

  useEffect(() => {
    if (navigatedFromDashboard?.query?.settings?.chart) {
      handleChartTypeChange({
        key: presentationObj[navigatedFromDashboard.query.settings.chart],
      });
    }
  }, [navigatedFromDashboard]);

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
        metrics={queries.map((q) => q)}
        currValue={currMetricsValue}
        onChange={setCurrMetricsValue}
      />
    );
  }

  if (queryType === QUERY_TYPE_PROFILE) {
    if (queries.length > 1 && breakdown.length) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={queries.map((_, index) => `All Users (${toLetters(index)})`)}
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
          queries={queries}
          breakdown={breakdown}
          section={section}
          currMetricsValue={currMetricsValue}
          durationObj={durationObj}
          chartType={chartType}
        />
      );
    }

    if (queryType === QUERY_TYPE_PROFILE) {
      content = (
        <ProfilesResultPage
          queries={queries}
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
        </div>{' '}
      </>

      <div className='mt-12'>{content}</div>
    </>
  );
}

export default ReportContent;
