import React, { useState, useEffect, useCallback } from 'react';
import { Spin } from 'antd';
import { runQuery, getFunnelData } from '../../reducers/coreQuery/services';
import { useSelector } from 'react-redux';
import { initialState, getStateQueryFromRequestQuery, presentationObj } from '../CoreQuery/utils';
import EventsAnalytics from './EventsAnalytics';
import Funnels from './Funnels';

function CardContent({ unit }) {
  const [resultState, setResultState] = useState(initialState);
  const { active_project } = useSelector(state => state.global);

  const getData = useCallback(async () => {
    try {
      setResultState({
        ...initialState,
        loading: true
      });
      if (typeof unit.query.query === 'string') {
        const res = await getFunnelData(active_project.id, JSON.parse(unit.query.query));
        setResultState({
          ...initialState,
          data: res.data
        });
      } else {
        const res = await runQuery(active_project.id, [unit.query.query.query_group[0]]);
        setResultState({
          ...initialState,
          data: res.data.result_group[0]
        });
      }
    } catch (err) {
      console.log(err);
      console.log(err.response);
      setResultState({
        ...initialState,
        error: true
      });
    }
  }, [active_project.id, unit.query]);

  useEffect(() => {
    getData();
  }, [getData]);

  let content = null;

  if (resultState.loading) {
    content = (
			<div className="flex justify-center items-center w-full h-64">
				<Spin size="small" />
			</div>
    );
  }

  if (resultState.error) {
    content = (
			<div className="flex justify-center items-center w-full h-64">
				Something went wrong!
			</div>
    );
  }

  if (resultState.data) {
    let equivalentQuery;
    if (typeof unit.query.query === 'string') {
      equivalentQuery = getStateQueryFromRequestQuery(JSON.parse(unit.query.query));
    } else {
      equivalentQuery = getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
    }

    const breakdown = [...equivalentQuery.breakdown.event, ...equivalentQuery.breakdown.global];
    const events = [...equivalentQuery.events];
    const queryType = equivalentQuery.queryType;

    const eventsMapper = {};
    const reverseEventsMapper = {};

    events.forEach((q, index) => {
      eventsMapper[`${q.label}`] = `event${index + 1}`;
      reverseEventsMapper[`event${index + 1}`] = q.label;
    });

    if (queryType === 'funnel') {
      content = (
				<Funnels
					breakdown={breakdown}
					events={events.map(elem => elem.label)}
					resultState={resultState}
					chartType={presentationObj[unit.presentation]}
					title={unit.id}
					eventsMapper={eventsMapper}
					reverseEventsMapper={reverseEventsMapper}
				/>
      );
    }

    if (queryType === 'event') {
      content = (
				<EventsAnalytics
					breakdown={breakdown}
					events={events.map(elem => elem.label)}
					resultState={resultState}
					chartType={presentationObj[unit.presentation]}
					title={unit.id}
					eventsMapper={eventsMapper}
					reverseEventsMapper={reverseEventsMapper}
				/>
      );
    }
  }

  return (
    <>
			{content}
    </>
  );
}

export default CardContent;
