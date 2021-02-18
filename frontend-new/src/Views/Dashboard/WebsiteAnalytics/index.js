import React, { useEffect, useState, useCallback } from "react";
import { getWebAnalyticsRequestBody } from "../utils";
import { initialState } from "../../CoreQuery/utils";
import { useSelector } from "react-redux";
import { getWebAnalyticsData } from "../../../reducers/coreQuery/services";
import { Spin } from "antd";
import TableUnits from "./TableUnits";
import CardUnit from "./CardUnit";
import NoDataChart from 'Components/NoDataChart';

function WebsiteAnalytics({
  webAnalyticsUnits,
  setwidgetModal,
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
        <NoDataChart />
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
          <CardUnit resultState={resultState} setwidgetModal={setwidgetModal} cardUnits={cardUnits} data={resultState.data} />
        ) : null}
        {tableUnits.length ? (
          <TableUnits resultState={resultState} setwidgetModal={setwidgetModal} tableUnits={tableUnits} data={resultState.data} />
        ) : null}
      </>
    );
  }

  return null;
}

export default WebsiteAnalytics;
