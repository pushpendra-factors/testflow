import React from 'react';
import AttributionsChart from './AttributionsChart';

function Attributions({
  chartType,
  attributionsState,
  resultState,
  unit,
  section,
  durationObj,
}) {
  const {
    eventGoal,
    touchpoint,
    models,
    linkedEvents,
    attr_dimensions,
    content_groups,
  } = attributionsState;

  return (
    <AttributionsChart
      event={eventGoal.label}
      linkedEvents={linkedEvents}
      touchpoint={touchpoint}
      durationObj={durationObj}
      data={
        resultState.data.result ? resultState.data.result : resultState.data
      }
      isWidgetModal={false}
      attribution_method={models[0]}
      attribution_method_compare={models[1]}
      section={section}
      attr_dimensions={attr_dimensions}
      content_groups={content_groups}
      currMetricsValue={0}
      chartType={chartType}
      cardSize={unit.cardSize}
      unitId={unit.id}
    />
  );
}

export default Attributions;
