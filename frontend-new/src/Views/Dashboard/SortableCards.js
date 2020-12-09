import React, { useCallback, useRef } from "react";
import { ReactSortable } from "react-sortablejs";
import WidgetCard from "./WidgetCard";
import { useSelector, useDispatch } from "react-redux";
import { UNITS_ORDER_CHANGED } from "../../reducers/types";
import { updateDashboard } from "../../reducers/dashboard/services";
import { getRequestForNewState } from "../../reducers/dashboard/utils";

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
      dispatch({ type: UNITS_ORDER_CHANGED, payload: newState });
      clearTimeout(timerRef.current);
      timerRef.current = setTimeout(() => {
        updateDashboard(active_project.id, activeDashboard.id, {
          units_position: body,
        });
      }, 1000);
    },
    [activeDashboard.id, active_project.id, dispatch]
  );

  const activeUnits = activeDashboardUnits.data.filter(
    (elem) => savedQueries.findIndex((sq) => sq.id === elem.query_id) > -1
  );

  return (
    <ReactSortable
      className="flex flex-wrap"
      list={activeUnits}
      setList={onDrop}
    >
      {activeUnits.map((item) => {
        const savedQuery = savedQueries.find((sq) => sq.id === item.query_id);
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
  );
}

export default SortableCards;
