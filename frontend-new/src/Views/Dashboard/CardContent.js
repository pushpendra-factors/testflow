import React from "react";
import { Spin } from "antd";
import {
  getStateQueryFromRequestQuery,
  presentationObj,
  getAttributionStateFromRequestQuery,
} from "../CoreQuery/utils";
import EventsAnalytics from "./EventsAnalytics";
import Funnels from "./Funnels";
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
} from "../../utils/constants";
import Attributions from "./Attributions";

function CardContent({ unit, resultState, setwidgetModal, durationObj }) {
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
    if (unit.query.query.query_group) {
      equivalentQuery = getStateQueryFromRequestQuery(
        unit.query.query.query_group[0]
      );
    } else if (unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_ATTRIBUTION) {
      equivalentQuery = getAttributionStateFromRequestQuery(unit.query.query.query);
    } else {
      equivalentQuery = getStateQueryFromRequestQuery(unit.query.query);
    }

    let breakdown,
      events,
      eventsMapper = {},
      reverseEventsMapper = {},
      arrayMapper = [],
      attributionsState;
    const { queryType } = equivalentQuery;
    if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_FUNNEL) {
      breakdown = [
        ...equivalentQuery.breakdown.event,
        ...equivalentQuery.breakdown.global,
      ];
      events = [...equivalentQuery.events];
      events.forEach((q, index) => {
        eventsMapper[`${q.label}`] = `event${index + 1}`;
        reverseEventsMapper[`event${index + 1}`] = q.label;
        arrayMapper.push({
          eventName: q,
          index,
          mapper: `event${index + 1}`,
        });
      });
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      attributionsState = {
        eventGoal: equivalentQuery.eventGoal,
        touchpoint: equivalentQuery.touchpoint,
        models: equivalentQuery.models,
        linkedEvents: equivalentQuery.linkedEvents,
      };
    }

    let dashboardPresentation = "pl";

    try {
      dashboardPresentation = unit.settings.chart;
    } catch (err) {
      console.log(err);
    }

    if (queryType === QUERY_TYPE_FUNNEL) {
      content = (
        <Funnels
          breakdown={breakdown}
          events={events.map((elem) => elem.label)}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          title={unit.id}
          eventsMapper={eventsMapper}
          reverseEventsMapper={reverseEventsMapper}
          unit={unit}
          setwidgetModal={setwidgetModal}
        />
      );
    }

    if (queryType === QUERY_TYPE_EVENT) {
      content = (
        <EventsAnalytics
          durationObj={durationObj}
          breakdown={breakdown}
          events={events.map((elem) => elem.label)}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          title={unit.id}
          eventsMapper={eventsMapper}
          reverseEventsMapper={reverseEventsMapper}
          unit={unit}
          setwidgetModal={setwidgetModal}
          arrayMapper={arrayMapper}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      content = (
        <Attributions unit={unit} title={unit.id} resultState={resultState} setwidgetModal={setwidgetModal} attributionsState={attributionsState} chartType={presentationObj[dashboardPresentation]} />
      );
    }
  }

  return <>{content}</>;
}

export default CardContent;
