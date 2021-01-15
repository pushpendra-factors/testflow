import React, { useCallback } from "react";
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
} from "../CoreQuery/utils";
import ResultTab from "../CoreQuery/EventsAnalytics/ResultTab";
import ResultantChart from "../CoreQuery/FunnelsResultPage/ResultantChart";
import { Text, SVG } from "../../components/factorsComponents";
import { Button, Divider, Spin } from "antd";
import styles from "./index.module.scss";
import FiltersInfo from "../CoreQuery/FiltersInfo";
import { useHistory } from "react-router-dom";
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
} from "../../utils/constants";
import AttributionsChart from "../CoreQuery/AttributionsResult/AttributionsChart";
import GroupedAttributionsChart from "../CoreQuery/AttributionsResult/GroupedAttributionsChart";
import CampaignAnalytics from "../CoreQuery/CampaignAnalytics";

function ActiveUnitContent({
  unit,
  resultState,
  setwidgetModal,
  durationObj,
  handleDurationChange,
}) {
  const history = useHistory();

  let equivalentQuery;
  if (unit.query.query.query_group) {
    if (unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_CAMPAIGN) {
      equivalentQuery = {
        ...unit.query.query.query_group[0],
        queryType: QUERY_TYPE_CAMPAIGN,
      };
    } else {
      equivalentQuery = getStateQueryFromRequestQuery(
        unit.query.query.query_group[0]
      );
    }
  } else if (
    unit.query.query.cl &&
    unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
  ) {
    equivalentQuery = getAttributionStateFromRequestQuery(
      unit.query.query.query
    );
  } else {
    equivalentQuery = getStateQueryFromRequestQuery(unit.query.query);
  }

  const { queryType } = equivalentQuery;
  let breakdown,
    events,
    eventsMapper = {},
    reverseEventsMapper = {},
    arrayMapper = [],
    attributionsState;

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

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    arrayMapper = equivalentQuery.select_metrics.map((metric, index) => {
      return {
        eventName: metric,
        index,
        mapper: `event${index + 1}`,
      };
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

  let content = null;

  if (queryType === QUERY_TYPE_EVENT) {
    content = (
      <ResultTab
        queries={events.map((elem) => elem.label)}
        eventsMapper={eventsMapper}
        reverseEventsMapper={reverseEventsMapper}
        breakdown={breakdown}
        queryType={queryType}
        isWidgetModal={true}
        page="totalEvents"
        durationObj={durationObj}
        handleDurationChange={handleDurationChange}
        resultState={[resultState]}
        index={0}
        arrayMapper={arrayMapper}
        title={`modal${unit.id}`}
      />
    );
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    if (resultState.loading) {
      content = (
        <div className="flex justify-center items-center w-full h-64">
          <Spin size="large" />
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
      const { eventGoal, touchpoint, models, linkedEvents } = attributionsState;
      if (models.length === 1) {
        content = (
          <AttributionsChart
            data={resultState.data}
            isWidgetModal={true}
            title={unit.title}
            event={eventGoal.label}
            linkedEvents={linkedEvents}
            touchpoint={touchpoint}
            attribution_method={models[0]}
          />
        );
      } else if (models.length === 2) {
        content = (
          <GroupedAttributionsChart
            event={eventGoal.label}
            linkedEvents={linkedEvents}
            touchpoint={touchpoint}
            data={resultState.data}
            isWidgetModal={true}
            attribution_method={models[0]}
            attribution_method_compare={models[1]}
          />
        );
      }
    }
  }

  if (queryType === QUERY_TYPE_FUNNEL) {
    let subcontent = null;

    if (resultState.loading) {
      subcontent = (
        <div className="flex justify-center items-center w-full h-64">
          <Spin size="large" />
        </div>
      );
    }

    if (resultState.error) {
      subcontent = (
        <div className="flex justify-center items-center w-full h-64">
          Something went wrong!
        </div>
      );
    }

    if (resultState.data) {
      subcontent = (
        <ResultantChart
          isWidgetModal={true}
          queries={events.map((elem) => elem.label)}
          breakdown={breakdown}
          eventsMapper={eventsMapper}
          reverseEventsMapper={reverseEventsMapper}
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
          resultState={resultState}
        />
      );
    }

    content = (
      <>
        <FiltersInfo
          durationObj={durationObj}
          handleDurationChange={handleDurationChange}
          breakdown={breakdown}
        />
        {subcontent}
      </>
    );
  }

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    content = (
      <CampaignAnalytics
        resultState={resultState}
        arrayMapper={arrayMapper}
        campaignState={equivalentQuery}
        isWidgetModal={true}
        title={`modal${unit.id}`}
        setDrawerVisible={() => {}}
      />
    );
  }

  const handleEditQuery = useCallback(() => {
    history.push({
      pathname: "/core-analytics",
      state: {
        query: { ...unit.query, settings: unit.settings },
        global_search: true,
      },
    });
  }, [history, unit]);

  return (
    <div className="p-4">
      <div className="flex flex-col">
        <div className="flex justify-between items-center">
          <Text extraClass="m-0" type={"title"} level={3} weight={"bold"}>
            {unit.title}
          </Text>
          <div className="flex items-center">
            <Button
              onClick={handleEditQuery}
              style={{ display: "flex" }}
              className="flex items-center mr-2"
              size="small"
            >
              Edit Query
            </Button>
            <Button
              style={{ display: "flex" }}
              className="flex items-center"
              size={"small"}
              type="text"
              onClick={setwidgetModal.bind(this, false)}
            >
              <SVG size={24} name="times"></SVG>
            </Button>
          </div>
        </div>
        {queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_FUNNEL ? (
          <div className="flex">
            {equivalentQuery.events.map((event, index) => {
              return (
                <div key={index} className="flex items-center mr-1 mt-3">
                  <div className={styles.eventCharacter}>
                    {String.fromCharCode(index + 65)}
                  </div>
                  <div className={styles.eventName}>{event.label}</div>
                </div>
              );
            })}
          </div>
        ) : null}
      </div>
      <Divider />
      {content}
    </div>
  );
}

export default ActiveUnitContent;
