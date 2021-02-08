import React, { useCallback } from "react";
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
} from "../CoreQuery/utils";
import { useHistory } from "react-router-dom";
import {
  QUERY_TYPE_EVENT,
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_WEB,
  EACH_USER_TYPE,
  TOTAL_EVENTS_CRITERIA,
  DASHBOARD_MODAL,
} from "../../utils/constants";
import WebsiteAnalyticsTable from "./WebsiteAnalytics/WebsiteAnalyticsTable";
import ReportContent from "../CoreQuery/AnalysisResultsPage/ReportContent";

function ActiveUnitContent({
  unit,
  resultState,
  durationObj,
  handleDurationChange,
  setwidgetModal
}) {
  const history = useHistory();

  let equivalentQuery;
  if (unit.query.query) {
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
  } else {
    if (unit.query.cl && unit.query.cl === QUERY_TYPE_WEB) {
      equivalentQuery = {
        queryType: QUERY_TYPE_WEB,
      };
    }
  }

  const { queryType } = equivalentQuery;
  let breakdown,
    events = [],
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
        eventName: q.label,
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
      date_range: durationObj,
    };
  }

  let content = null;

  if (queryType === QUERY_TYPE_WEB) {
    content = (
      <WebsiteAnalyticsTable
        tableData={resultState.data[unit.id]}
        isWidgetModal={true}
        modalTitle={unit.title}
      />
    );
  }

  const handleEditQuery = useCallback(() => {
    history.push({
      pathname: "/analyse",
      state: {
        query: { ...unit.query, settings: unit.settings },
        global_search: true,
      },
    });
  }, [history, unit]);

  return (
    <div className="p-4">
      <ReportContent
        queryType={queryType}
        resultState={resultState}
        setDrawerVisible={handleEditQuery}
        queries={events.map((q) => q.label)}
        breakdown={breakdown}
        handleDurationChange={handleDurationChange}
        arrayMapper={arrayMapper}
        queryOptions={{ date_range: durationObj }}
        attributionsState={attributionsState}
        breakdownType={EACH_USER_TYPE}
        campaignState={{ ...equivalentQuery, date_range: durationObj }}
        eventPage={TOTAL_EVENTS_CRITERIA}
        section={DASHBOARD_MODAL}
        queryTitle={unit.title}
        onReportClose={setwidgetModal}
      />
    </div>
  );
}

export default ActiveUnitContent;
