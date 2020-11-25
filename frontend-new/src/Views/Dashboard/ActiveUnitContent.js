import React from 'react';
import { getStateQueryFromRequestQuery } from '../CoreQuery/utils';
import ResultTab from '../EventsAnalytics/ResultTab.js';
import ResultantChart from '../CoreQuery/FunnelsResultPage/ResultantChart';

function ActiveUnitContent({ unit, unitData }) {
  let equivalentQuery;
  if (unit.query.query.query_group) {
    equivalentQuery = getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
  } else {
    equivalentQuery = getStateQueryFromRequestQuery(unit.query.query);
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

  let content = null;

  if (queryType === 'event') {
    content = (
            <ResultTab
                queries={events.map(elem => elem.label)}
                eventsMapper={eventsMapper}
                reverseEventsMapper={reverseEventsMapper}
                breakdown={breakdown}
                queryType={queryType}
                isWidgetModal={true}
                page="totalEvents"
                resultState={
                    [
                      {
                        loading: false,
                        data: unitData,
                        error: false
                      }
                    ]
                }
                index={0}
            />
    );
  }

  if (queryType === 'funnel') {
    content = (
            <div className="fa-container">
                <ResultantChart
                    modal={true}
                    queries={events.map(elem => elem.label)}
                    breakdown={breakdown}
                    eventsMapper={eventsMapper}
                    reverseEventsMapper={reverseEventsMapper}
                    resultState={
                        {
                          loading: false,
                          data: unitData,
                          error: false
                        }
                    }
                />
            </div>
    );
  }

  return (
        <div className="flex py-12 px-4">
            {content}
        </div>
  );
}

export default ActiveUnitContent;
