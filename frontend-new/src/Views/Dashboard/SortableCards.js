import React from 'react';
import { ReactSortable } from 'react-sortablejs';
import WidgetCard from './WidgetCard';
import { useSelector, useDispatch } from 'react-redux';
import { UNITS_ORDER_CHANGED } from '../../reducers/types';
import { updateDashboard } from '../../reducers/dashboard/services';

function SortableCards() {

    const dispatch = useDispatch();

    const { active_project } = useSelector(state => state.global);
    const { data: savedQueries } = useSelector(state => state.queries);
    const { activeDashboardUnits, activeDashboard } = useSelector(state => state.dashboard);

    const onDrop = (newState) => {
        const body = {};
        newState.forEach((elem, index) => {
            console.log(elem);
            body[elem.id] = {
                position: index,
                size: elem.cardSize
            }
        });
        console.log(body);
        updateDashboard(active_project.id, activeDashboard.id, { units_position: body })
        dispatch({ type: UNITS_ORDER_CHANGED, payload: newState })
    };

    return (
        <ReactSortable className="flex flex-wrap" list={activeDashboardUnits.data} setList={onDrop}>
            {activeDashboardUnits.data.map((item) => {
                const savedQuery = savedQueries.find(sq => sq.id === item.query_id);
                if (savedQuery) {
                    return (
                        <WidgetCard
                            key={item.id}
                            unit={{ ...item, query: savedQuery }}
                        />
                    );
                } else {
                    return null;
                }

            })}
        </ReactSortable>
    )
}

export default SortableCards;