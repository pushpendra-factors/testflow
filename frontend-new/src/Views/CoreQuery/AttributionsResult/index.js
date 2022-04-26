import React from 'react';
import AttributionsChart from './AttributionsChart';

function AttributionsResult({
  resultState,
  attributionsState,
  section,
  durationObj,
  currMetricsValue,
  renderedCompRef,
  chartType,
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
      currMetricsValue={currMetricsValue}
      ref={renderedCompRef}
      chartType={chartType}
    />
  );
}

export default AttributionsResult;
