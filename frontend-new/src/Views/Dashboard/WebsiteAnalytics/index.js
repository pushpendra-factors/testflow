import React, { useEffect, useState, useCallback } from "react";
import { getWebAnalyticsRequestBody } from "../utils";
import { initialState } from "../../CoreQuery/utils";
import { useSelector } from "react-redux";
import { getWebAnalyticsData } from "../../../reducers/coreQuery/services";
import { Spin } from "antd";
import TableUnits from "./TableUnits";
import CardUnit from "./CardUnit";

function WebsiteAnalytics({
  webAnalyticsUnits,
  savedQueries,
  setwidgetModal,
  showDeleteWidgetModal,
  refreshClicked,
  setRefreshClicked,
  durationObj,
}) {
  const { active_project } = useSelector((state) => state.global);
  const [resultState, setResultState] = useState(initialState);
  const fetchData = useCallback(
    async (refresh = false) => {
      try {
        const reqBody = getWebAnalyticsRequestBody(
          webAnalyticsUnits,
          durationObj
        );
        setResultState({ ...initialState, loading: true });
        const dashboardId = webAnalyticsUnits[0].dashboard_id;
        const response = await getWebAnalyticsData(
          active_project.id,
          reqBody,
          dashboardId,
          refresh
        );
        setResultState({ ...initialState, data: response.data.result });
      } catch (err) {
        console.log(err);
        setResultState({ ...initialState, error: true });
      }
    },
    [active_project.id, durationObj, webAnalyticsUnits]
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  if (resultState.loading) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        <Spin size="large" />
      </div>
    );
  }

  if (resultState.error) {
    return (
      <div className="flex justify-center items-center w-full h-64">
        Something went wrong!
      </div>
    );
  }

  if (resultState.data) {
    const tableUnits = webAnalyticsUnits.filter(
      (unit) => unit.presentation === "pt"
    );
    const cardUnits = webAnalyticsUnits.filter(
      (unit) => unit.presentation === "pc"
    );

    return (
      <>
        {cardUnits.length ? (
          <CardUnit cardUnits={cardUnits} data={resultState.data} />
        ) : null}
        {tableUnits.length ? (
          <TableUnits tableUnits={tableUnits} data={resultState.data} />
        ) : null}
      </>
    );
  }

  return null;
}

export default WebsiteAnalytics;
