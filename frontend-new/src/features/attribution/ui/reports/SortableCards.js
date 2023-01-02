import React, { useCallback, useRef } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { ReactSortable } from 'react-sortablejs';
import { isEqual } from 'lodash';
import WidgetCard from './WidgetCard';
import { getRequestForNewState } from 'Reducers/dashboard/utils';
import { updateDashboard } from 'Reducers/dashboard/services';
import { ATTRIBUTION_DASHBOARD_UNITS_UPDATED } from 'Attribution/state/action.constants';

function SortableCards({ activeUnits, durationObj, showDeleteWidgetModal }) {
  const timerRef = useRef(null);
  const { active_project: activeProject } = useSelector(
    (state) => state.global
  );
  const { data: savedQueries } = useSelector(
    (state) => state.attributionDashboard.attributionQueries
  );

  const { id: attributionDashboardId } = useSelector(
    (state) => state.attributionDashboard.dashboard
  );

  const dispatch = useDispatch();

  const onDrop = useCallback(
    async (newState) => {
      if (!isEqual(activeUnits, newState)) {
        const body = getRequestForNewState(newState);
        dispatch({
          type: ATTRIBUTION_DASHBOARD_UNITS_UPDATED,
          payload: newState,
          units_position: body
        });
        clearTimeout(timerRef.current);
        timerRef.current = setTimeout(() => {
          updateDashboard(activeProject.id, attributionDashboardId, {
            units_position: body
          });
        }, 300);
      }
    },
    [attributionDashboardId, activeProject.id, dispatch]
  );

  return (
    <ReactSortable
      className='flex flex-wrap flex-col'
      list={activeUnits}
      setList={onDrop}
    >
      {activeUnits.map((item, index) => {
        const savedQuery = savedQueries.find((sq) => sq.id === item.query_id);
        return (
          <WidgetCard
            durationObj={durationObj}
            key={item.id}
            unit={{ ...item, query: savedQuery }}
            showDeleteWidgetModal={showDeleteWidgetModal}
          />
        );
      })}
    </ReactSortable>
  );
}

export default SortableCards;
