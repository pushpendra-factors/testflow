import React, { useEffect, useState } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { ReactSortable } from 'react-sortablejs';
import WidgetCard from './WidgetCard';

function SortableCards({ activeUnits, durationObj }) {
  const { active_project } = useSelector((state) => state.global);
  const { data: savedQueries } = useSelector((state) => state.queries);

  const dispatch = useDispatch();

  // dummy function to test dummyData values reordering
  const onDrop = (newState) => {
    console.log('onDrop function called----', newState);
  };

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
