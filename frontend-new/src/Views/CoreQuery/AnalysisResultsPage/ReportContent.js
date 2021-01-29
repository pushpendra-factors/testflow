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
} from "../../../utils/constants";
import { Spin } from "antd";
import FunnelsResultPage from "../FunnelsResultPage";
import CalendarRow from "./CalendarRow";
import AttributionsResult from "../AttributionsResult";
import CampaignAnalytics from "../CampaignAnalytics";
import { getChartTypeMenuItems } from "../../../utils/dataFormatter";
import CampaignMetricsDropdown from "./CampaignMetricsDropdown";

function ReportContent({
  resultState,
  queryType,
  setDrawerVisible,
  queries,
  breakdown,
  requestQuery,
  querySaved,
  setQuerySaved,
  queryOptions,
  savedChartType = null,
  handleDurationChange,
  campaignState,
  arrayMapper,
  attributionsState,
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
    setChartTypeMenuItems(items);
  }, [queryType, campaignState.group_by]);

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
    }
  }, [savedChartType, queryType, campaignState.group_by]);

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

  if (queryType === QUERY_TYPE_FUNNEL) {
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
  }

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    queryDetail = campaignState.select_metrics.join(", ");
    durationObj = campaignState.date_range;
    if (campaignState.select_metrics.length > 1) {
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
    if (queryType === QUERY_TYPE_FUNNEL) {
      content = (
        <FunnelsResultPage
          setDrawerVisible={setDrawerVisible}
          queries={queries}
          resultState={resultState}
          breakdown={breakdown}
          requestQuery={requestQuery}
          querySaved={querySaved}
          setQuerySaved={setQuerySaved}
          arrayMapper={arrayMapper}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      content = (
        <AttributionsResult
          resultState={resultState}
          attributionsState={attributionsState}
        />
      );
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      const campaignsArrayMapper = campaignState.select_metrics.map(
        (metric, index) => {
          return {
            eventName: metric,
            index,
            mapper: `event${index + 1}`,
          };
        }
      );
      content = (
        <CampaignAnalytics
          resultState={resultState}
          arrayMapper={campaignsArrayMapper}
          campaignState={campaignState}
          chartType={chartType}
          currMetricsValue={currMetricsValue}
        />
      );
    }
  }

  return (
    <>
      <ReportTitle
        setDrawerVisible={setDrawerVisible}
        title={""}
        queryDetail={queryDetail}
      />
      <div className="mt-6">
        <CalendarRow
          handleDurationChange={handleDurationChange}
          durationObj={durationObj}
          handleChartTypeChange={handleChartTypeChange}
          chartTypeMenuItems={chartTypeMenuItems}
          chartType={chartType}
          metricsDropdown={metricsDropdown}
        />
      </div>
      <div className="mt-12">{content}</div>
    </>
  );
}

export default ReportContent;
