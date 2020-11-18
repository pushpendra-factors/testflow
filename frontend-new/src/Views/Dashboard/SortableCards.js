import React from 'react';
import { ReactSortable } from 'react-sortablejs';
import WidgetCard from './WidgetCard';
import { useSelector, useDispatch } from 'react-redux';
import { UNITS_ORDER_CHANGED } from '../../reducers/types';
import { updateDashboard } from '../../reducers/dashboard/services';
import { getRequestForNewState } from '../../reducers/dashboard/utils';

function SortableCards() {
  const dispatch = useDispatch();

  const { active_project } = useSelector(state => state.global);
  const { data: savedQueries } = useSelector(state => state.queries);
  const { activeDashboardUnits, activeDashboard } = useSelector(state => state.dashboard);

  const onDrop = (newState) => {
    const body = getRequestForNewState(newState);
    updateDashboard(active_project.id, activeDashboard.id, { units_position: body });
    dispatch({ type: UNITS_ORDER_CHANGED, payload: newState });
  };

  const activeUnits = activeDashboardUnits.data.filter(elem => savedQueries.findIndex(sq => sq.id === elem.query_id) > -1);

  return (
    <ReactSortable className="flex flex-wrap" list={activeUnits} setList={onDrop}>
      {activeUnits.map((item) => {
        const savedQuery = savedQueries.find(sq => sq.id === item.query_id);
        return (
          <WidgetCard
            key={item.id}
            unit={{ ...item, query: savedQuery }}
            onDrop={onDrop}
          />
        );
      })}
    </ReactSortable>
  );
}

export default SortableCards;
