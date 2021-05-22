import React from "react";
import SingleTouchPoint from "./SingleTouchPoint";
import DualTouchPoint from "./DualTouchPoint";

function Attributions({ chartType, attributionsState, resultState, unit, section }) {
  const { eventGoal, touchpoint, models, linkedEvents } = attributionsState;
  if (models.length === 1) {
    return (
      <SingleTouchPoint
        event={eventGoal.label}
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        data={resultState.data}
        isWidgetModal={false}
        attribution_method={models[0]}
				chartType={chartType}
        unit={unit}
        section={section}
      />
    );
  }

  if (models.length === 2) {
    return (
      <DualTouchPoint
        event={eventGoal.label}
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        data={resultState.data}
        isWidgetModal={false}
        attribution_method={models[0]}
				attribution_method_compare={models[1]}
				chartType={chartType}
        unit={unit}
        section={section}
      />
    );
  }

  return <div></div>;
}

export default Attributions;
