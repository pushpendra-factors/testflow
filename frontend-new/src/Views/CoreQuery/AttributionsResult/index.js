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
  queryOptions,
  appliedFilters,
  setAttributionTableFilters,
  v1=false
}) {
  const {
    eventGoal,
    touchpoint,
    models,
    linkedEvents,
    attr_dimensions: attrDimensions,
    content_groups: contentGroups,
    attrQueries
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
      attr_dimensions={attrDimensions}
      content_groups={contentGroups}
      currMetricsValue={currMetricsValue}
      renderedCompRef={renderedCompRef}
      chartType={chartType}
      queryOptions={queryOptions}
      attrQueries={attrQueries}
      appliedFilters={appliedFilters}
      setAppliedFilters={setAttributionTableFilters}
      v1={v1}
    />
  );
}

export default AttributionsResult;
