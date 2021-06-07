import React from "react";
import AttributionsChart from "./AttributionsChart";
import GroupedAttributionsChart from "./GroupedAttributionsChart";

function AttributionsResult({
  resultState,
  compareResult,
  attributionsState,
  section,
  durationObj,
  cmprDuration,
  currMetricsValue
}) {
  let content = null;

  const { eventGoal, touchpoint, models, linkedEvents, attr_dimensions } = attributionsState;

  if (models.length === 1) {
    content = (
      <AttributionsChart
        event={eventGoal.label}
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        durationObj={durationObj}
        cmprDuration={cmprDuration}
        data={
          resultState.data.result ? resultState.data.result : resultState.data
        }
        data2={
          compareResult?.data ? compareResult.data : null
        }
        isWidgetModal={false}
        attribution_method={models[0]}
        section={section}
        attr_dimensions={attr_dimensions}
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
        data2={
          compareResult?.data ? compareResult.data : null
        }
        isWidgetModal={false}
        attribution_method={models[0]}
        attribution_method_compare={models[1]}
        section={section}
        currMetricsValue={currMetricsValue}
        durationObj={durationObj}
        cmprDuration={cmprDuration}
        attr_dimensions={attr_dimensions}
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
