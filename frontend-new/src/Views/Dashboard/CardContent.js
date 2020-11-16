import React, { useState, useEffect, useCallback } from 'react';
import { Spin } from 'antd';
import { runQuery, getFunnelData } from '../../reducers/coreQuery/services';
import { useSelector, useDispatch } from 'react-redux';
import { initialState, getStateQueryFromRequestQuery, presentationObj } from '../CoreQuery/utils';
import EventsAnalytics from './EventsAnalytics';
import Funnels from './Funnels';
import { DASHBOARD_UNIT_DATA_LOADED } from '../../reducers/types';

function CardContent({ unit }) {
  const [resultState, setResultState] = useState(initialState);
  const { active_project } = useSelector(state => state.global);
  const dispatch = useDispatch();

  const getData = useCallback(async (refresh = false) => {
    try {
      setResultState({
        ...initialState,
        loading: true
      });
      if (typeof unit.query.query === 'string') {
        let res;
        if (refresh) {
          res = await getFunnelData(active_project.id, JSON.parse(unit.query.query));
        } else {
          res = await getFunnelData(active_project.id, JSON.parse(unit.query.query), { refresh: false, unit_id: unit.id, id: unit.dashboard_id });
        }
        // const res = await getFunnelData(active_project.id, JSON.parse(unit.query.query));
        let resultantData = null;
        if (res.data.result) {
          // cached data
          resultantData = res.data.result;
        } else {
          // refreshed data
          resultantData = res.data;
        }
        setResultState({
          ...initialState,
          data: resultantData
        });
      } else {
        let res;
        if (refresh) {
          res = await runQuery(active_project.id, [unit.query.query.query_group[0]]);
        } else {
          res = await runQuery(active_project.id, [unit.query.query.query_group[0]], { refresh: false, unit_id: unit.id, id: unit.dashboard_id });
        }
        let resultantData = null;
        if (res.data.result) {
          // cached data
          resultantData = res.data.result.result_group[0];
        } else {
          // refreshed data
          resultantData = res.data.result.result_group[0];
        }
        setResultState({
          ...initialState,
          data: resultantData
        });
      }
      dispatch({ type: DASHBOARD_UNIT_DATA_LOADED });
    } catch (err) {
      console.log(err);
      console.log(err.response);
      setResultState({
        ...initialState,
        error: true
      });
    }
  }, [active_project.id, unit.query, dispatch, unit.id, unit.dashboard_id]);

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

    let dashboardPresentation = 'pl';

    try {
      dashboardPresentation = unit.settings.chart;
    } catch (err) {
      console.log(err);
    }

    if (queryType === 'funnel') {
      content = (
        <Funnels
          breakdown={breakdown}
          events={events.map(elem => elem.label)}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
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
          chartType={presentationObj[dashboardPresentation]}
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
