import React, { useState, useEffect, useCallback } from 'react';
import { Spin } from 'antd';
import { runQuery } from '../../reducers/coreQuery/services';
import { useSelector } from 'react-redux';
import { initialState, getStateQueryFromRequestQuery, presentationObj } from '../CoreQuery/utils';
import MultipleEventsWithBreakdown from './MultipleEventsWithBreakdown';
import SingleEventSingleBreakdown from './SingleEventSingleBreakdown';
import SingleEventMultipleBreakdown from './SingleEventMultipleBreakdown';
import NoBreakdownCharts from './NoBreakdownCharts';

function CardContent({ unit }) {
  const [resultState, setResultState] = useState(initialState);
  const { active_project } = useSelector(state => state.global);

  const getData = useCallback(async () => {
    try {
      setResultState({
        ...initialState,
        loading: true
      });
      const res = await runQuery(active_project.id, [unit.query.query.query_group[0]]);
      setResultState({
        ...initialState,
        data: res.data.result_group[0]
      });
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
    const equivalentQuery = getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
    const breakdown = [...equivalentQuery.breakdown.event, ...equivalentQuery.breakdown.global];
    const events = [...equivalentQuery.events];
    // const queryType = equivalentQuery.queryType;

    if (events.length > 1 && breakdown.length) {
      content = (
                <MultipleEventsWithBreakdown
                    breakdownType="each"
                    queries={events.map(elem => elem.label)}
                    breakdown={breakdown}
                    resultState={resultState}
                    page="totalEvents"
                    chartType={presentationObj[unit.presentation]}
                    title={unit.id}
                />
      );
    }

    if (events.length === 1 && breakdown.length === 1) {
      content = (
                <SingleEventSingleBreakdown
                    breakdownType="each"
                    queries={events.map(elem => elem.label)}
                    breakdown={breakdown}
                    resultState={resultState}
                    page="totalEvents"
                    chartType={presentationObj[unit.presentation]}
                    title={unit.id}
                />
      );
    }

    if (events.length === 1 && breakdown.length > 1) {
      content = (
                <SingleEventMultipleBreakdown
                    breakdownType="each"
                    queries={events.map(elem => elem.label)}
                    breakdown={breakdown}
                    resultState={resultState}
                    page="totalEvents"
                    chartType={presentationObj[unit.presentation]}
                    title={unit.id}
                />
      );
    }

    if (!breakdown.length) {
      const eventsMapper = {};
      const reverseEventsMapper = {};

      events.forEach((q, index) => {
        eventsMapper[`${q.label}`] = `event${index + 1}`;
        reverseEventsMapper[`event${index + 1}`] = q.label;
      });

      content = (
                <NoBreakdownCharts
                    queries={events.map(elem => elem.label)}
                    eventsMapper={eventsMapper}
                    reverseEventsMapper={reverseEventsMapper}
                    resultState={resultState}
                    page="totalEvents"
                    chartType="linechart"
                    title={unit.id}
                />
      );
    }
  }

  return (
        <div className="card-content">
            {content}
        </div>
  );
}

export default CardContent;
