import React from "react";
import SingleTouchPoint from "./SingleTouchPoint";
import DualTouchPoint from "./DualTouchPoint";

function Attributions({ chartType, attributionsState, resultState, setwidgetModal, title }) {
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
        setwidgetModal={setwidgetModal}
        title={title}
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
				setwidgetModal={setwidgetModal}
        chartType={chartType}
        title={title}
      />
    );
  }

  return <div></div>;
}

export default Attributions;
