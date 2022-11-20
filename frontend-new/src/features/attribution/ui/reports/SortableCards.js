import React, { useEffect, useState } from 'react';
import { useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { ReactSortable } from 'react-sortablejs';
import { getRequestForNewState } from 'Reducers/dashboard/utils';
import { getDashboardDateRange } from 'Views/Dashboard/utils';
import WidgetCard from './WidgetCard';

function SortableCards({ activeUnits }) {
  const { active_project } = useSelector((state) => state.global);
  const { data: savedQueries } = useSelector((state) => state.queries);

  const [durationObj, setDurationObj] = useState(getDashboardDateRange());
  const dispatch = useDispatch();

  // this is t0 test
  const [dummyData2, setDummy] = useState([{ id: 10 }, { id: 11 }, { id: 12 }]);

  // dummy function to test dummyData values reordering
  const onDrop = (newState) => {
    let testDummyArr = [];
    const body = getRequestForNewState(newState);
    for (let keys in body.position) {
      testDummyArr[body.position[keys]] = { id: Number(keys) };
    }
    setDummy(testDummyArr);
  };

  // reordering function for cards

  // const onDrop = useCallback(
  //   async (newState) => {
  //     const body = getRequestForNewState(newState);
  //     // changes unit in fronted
  //     // dispatch({
  //     //   type: ATTRIBUTION_UNITS_ORDER_CHANGED,
  //     //   payload: newState,
  //     //   units_position: body
  //     // });

  //     // updates bin backend
  //     // clearTimeout(timerRef.current);
  //     // timerRef.current = setTimeout(() => {
  //     //   updateDashboard(active_project.id, activeDashboard.id, {
  //     //     units_position: body
  //     //   });
  //     // }, 300);
  //   },
  //   [activeDashboard?.id, active_project.id, dispatch]
  // );

  return (
    <ReactSortable
      className='flex flex-wrap flex-col'
      list={activeUnits}
      // active units is the list of all cards and position
      setList={onDrop}
      // on drop changes the position drag and drop
    >
      {/* Widgets card collection will go here */}
      {activeUnits.map((item, index) => {
        const savedQuery = savedQueries.find((sq) => sq.id === item.query_id);
        return (
          <WidgetCard
            durationObj={durationObj}
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
