import React from "react";
import AttributionsChart from "./AttributionsChart";
import GroupedAttributionsChart from "./GroupedAttributionsChart";

function AttributionsResult({
  resultState,
  attributionsState,
}) {
  let content = null;

  const { eventGoal, touchpoint, models, linkedEvents } = attributionsState;

  if (models.length === 1) {
    content = (
      <AttributionsChart
        event={eventGoal.label}
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        data={
          resultState.data.result ? resultState.data.result : resultState.data
        }
        isWidgetModal={false}
        attribution_method={models[0]}
      />
    );
  }

  if (models.length === 2) {
    content = (
      <GroupedAttributionsChart
        event={eventGoal.label}
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        data={
          resultState.data.result ? resultState.data.result : resultState.data
        }
        isWidgetModal={false}
        attribution_method={models[0]}
        attribution_method_compare={models[1]}
      />
    );
  }

  return (
    <>
      {content}
    </>
  );
}

export default AttributionsResult;
