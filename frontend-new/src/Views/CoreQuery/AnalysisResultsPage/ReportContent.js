import React, { useState, useCallback, useEffect } from "react";
import ReportTitle from "./ReportTitle";
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  CHART_TYPE_LINECHART,
  CHART_TYPE_BARCHART,
  CHART_TYPE_SPARKLINES,
  EACH_USER_TYPE,
  QUERY_TYPE_WEB,
} from "../../../utils/constants";
import { Spin } from "antd";
import FunnelsResultPage from "../FunnelsResultPage";
import CalendarRow from "./CalendarRow";
import AttributionsResult from "../AttributionsResult";
import CampaignAnalytics from "../CampaignAnalytics";
import { getChartTypeMenuItems } from "../../../utils/dataFormatter";
import CampaignMetricsDropdown from "./CampaignMetricsDropdown";
import EventsAnalytics from "../EventsAnalytics";
import WebsiteAnalyticsTable from "../../Dashboard/WebsiteAnalytics/WebsiteAnalyticsTable";

function ReportContent({
  resultState,
  queryType,
  setDrawerVisible,
  queries,
  breakdown,
  queryOptions,
  savedChartType = null,
  handleDurationChange,
  campaignState,
  arrayMapper,
  attributionsState,
  breakdownType,
  queryTitle,
  eventPage,
  section,
  onReportClose,
  cmprDuration,
  runAttrCmprQuery,
  cmprResultState,
  campaignsArrayMapper
}) {
  let content = null,
    queryDetail = null,
    durationObj = {},
    metricsDropdown = <div className="mr-2">Data from</div>;

  const [currMetricsValue, setCurrMetricsValue] = useState(0);
  const [chartTypeMenuItems, setChartTypeMenuItems] = useState([]);
  const [chartType, setChartType] = useState(CHART_TYPE_LINECHART);

  const handleChartTypeChange = useCallback(({ key }) => {
    setChartType(key);
  }, []);

  useEffect(() => {
    let items = [];
    if (queryType === QUERY_TYPE_CAMPAIGN) {
      items = getChartTypeMenuItems(
        queryType,
        campaignState.group_by.length > 0
      );
    }
    if (queryType === QUERY_TYPE_EVENT && breakdownType === EACH_USER_TYPE) {
      items = getChartTypeMenuItems(queryType, breakdown.length > 0);
    }
    setChartTypeMenuItems(items);
  }, [queryType, campaignState.group_by, breakdown, breakdownType]);

  useEffect(() => {
    if (savedChartType) {
      setChartType(savedChartType);
    } else {
      if (queryType === QUERY_TYPE_CAMPAIGN) {
        if (campaignState.group_by.length) {
          setChartType(CHART_TYPE_BARCHART);
        } else {
          setChartType(CHART_TYPE_SPARKLINES);
        }
      }
      if (queryType === QUERY_TYPE_EVENT) {
        if (breakdown.length) {
          setChartType(CHART_TYPE_BARCHART);
        } else {
          setChartType(CHART_TYPE_SPARKLINES);
        }
      }
    }
  }, [savedChartType, queryType, campaignState.group_by, breakdown]);

  if (resultState.loading) {
    content = (
      <div className="h-64 flex items-center justify-center w-full">
        <Spin size={"large"} />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className="h-64 flex items-center justify-center w-full">
        Something Went Wrong!
      </div>
    );
  }

  if (queryType === QUERY_TYPE_WEB) {
    durationObj = queryOptions.date_range;
    queryDetail = "Website Analytics";
  }

  if (queryType === QUERY_TYPE_FUNNEL || queryType === QUERY_TYPE_EVENT) {
    durationObj = queryOptions.date_range;
    queryDetail = arrayMapper
      .map((elem) => {
        return elem.eventName;
      })
      .join(", ");
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    queryDetail = attributionsState.eventGoal.label;
    durationObj = attributionsState.date_range;
    if (attributionsState.models.length === 2) {
      metricsDropdown = (
        <CampaignMetricsDropdown
          metrics={["Conversions", "Cost Per Conversion"]}
          currValue={currMetricsValue}
          onChange={setCurrMetricsValue}
        />
      );
    }
  }

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    queryDetail = campaignState.select_metrics.join(", ");
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
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      content = (
        <AttributionsResult
          resultState={resultState}
          compareResult={cmprResultState}
          durationObj={durationObj}
          cmprDuration={cmprDuration}
          attributionsState={attributionsState}
          section={section}
          currMetricsValue={currMetricsValue}
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
        />
      );
    }
  }

  return (
    <>
      <ReportTitle
        setDrawerVisible={setDrawerVisible}
        title={queryTitle}
        queryDetail={queryDetail}
        section={section}
        onReportClose={onReportClose}
        queryType={queryType}
      />
      <div className="mt-6">
        <CalendarRow
          queryType={queryType}
          handleDurationChange={handleDurationChange}
          durationObj={durationObj}
          handleChartTypeChange={handleChartTypeChange}
          chartTypeMenuItems={chartTypeMenuItems}
          chartType={chartType}
          metricsDropdown={metricsDropdown}
          triggerAttrComparision={runAttrCmprQuery}
          cmprResultState={cmprResultState}
        />
      </div>
      <div className="mt-12">{content}</div>
    </>
  );
}

export default ReportContent;
