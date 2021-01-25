import React, { useCallback, useRef } from "react";
import { ReactSortable } from "react-sortablejs";
import WidgetCard from "./WidgetCard";
import { useSelector, useDispatch } from "react-redux";
import { UNITS_ORDER_CHANGED } from "../../reducers/types";
import { updateDashboard } from "../../reducers/dashboard/services";
import { getRequestForNewState } from "../../reducers/dashboard/utils";
import { QUERY_TYPE_WEB } from "../../utils/constants";
import WebsiteAnalytics from "./WebsiteAnalytics";

function SortableCards({
  setwidgetModal,
  durationObj,
  showDeleteWidgetModal,
  refreshClicked,
  setRefreshClicked,
}) {
  const dispatch = useDispatch();
  const timerRef = useRef(null);

  const { active_project } = useSelector((state) => state.global);
  const { data: savedQueries } = useSelector((state) => state.queries);
  const { activeDashboardUnits, activeDashboard } = useSelector(
    (state) => state.dashboard
  );

  const onDrop = useCallback(
    async (newState) => {
      const body = getRequestForNewState(newState);
      dispatch({ type: UNITS_ORDER_CHANGED, payload: newState, units_position: body });
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => {
        updateDashboard(active_project.id, activeDashboard.id, {
          units_position: body,
        });
      }, 300);
    },
    [activeDashboard.id, active_project.id, dispatch]
  );

  const activeUnits = activeDashboardUnits.data.filter(
    (elem) =>
      savedQueries.findIndex(
        (sq) => sq.id === elem.query_id && sq.query.cl !== QUERY_TYPE_WEB
      ) > -1
  );

  const webAnalyticsUnits = activeDashboardUnits.data.filter(
    (elem) =>
      savedQueries.findIndex(
        (sq) => sq.id === elem.query_id && sq.query.cl === QUERY_TYPE_WEB
      ) > -1
  );

  return (
    <>
      {activeUnits.length ? (
        <ReactSortable
          className="flex flex-wrap"
          list={activeUnits}
          setList={onDrop}
        >
          {activeUnits.map((item) => {
            const savedQuery = savedQueries.find(
              (sq) => sq.id === item.query_id
            );
            return (
              <WidgetCard
                durationObj={durationObj}
                key={item.id}
                unit={{ ...item, query: savedQuery }}
                onDrop={onDrop}
                setwidgetModal={setwidgetModal}
                showDeleteWidgetModal={showDeleteWidgetModal}
                refreshClicked={refreshClicked}
                setRefreshClicked={setRefreshClicked}
              />
            );
          })}
        </ReactSortable>
      ) : null}
      {webAnalyticsUnits.length ? (
        <WebsiteAnalytics
          durationObj={durationObj}
          webAnalyticsUnits={webAnalyticsUnits}
          setwidgetModal={setwidgetModal}
        />
      ) : null}
    </>
  );
}

export default SortableCards;
