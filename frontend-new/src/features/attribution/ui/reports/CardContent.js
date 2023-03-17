import React, { useMemo } from 'react';
import { useSelector } from 'react-redux';
import { Spin } from 'antd';
import { get } from 'lodash';
import NoDataChart from 'Components/NoDataChart';
import {
  apiChartAnnotations,
  CHART_TYPE_TABLE,
  DASHBOARD_WIDGET_SECTION,
  presentationObj,
  QUERY_TYPE_ATTRIBUTION
} from 'Utils/constants';
import {
  getStateQueryFromRequestQuery
} from 'Views/CoreQuery/utils';

import {getAttributionStateFromRequestQuery} from 'Attribution/utils';

import AttributionsChart from 'Views/Dashboard/Attributions/AttributionsChart';
import {
  DEFAULT_DASHBOARD_PRESENTATION,
  DASHBOARD_PRESENTATION_KEYS
} from 'Components/SaveQuery/saveQuery.constants';
import { SVG, Text } from 'Components/factorsComponents';

function CardContent({ unit, resultState, durationObj }) {
  let content = null;
  const { eventNames, attr_dimensions, content_groups } = useSelector(
    (state) => state.coreQuery
  );
  const { config: kpiConfig } = useSelector((state) => state.kpi);

  const equivalentQuery = useMemo(() => {
    if (unit.query.query.query_group) {
      return getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
    } else if (
      unit.query.query.cl &&
      unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
    ) {
      return getAttributionStateFromRequestQuery(
        unit.query.query.query,
        attr_dimensions,
        content_groups,
        kpiConfig
      );
    }
  }, [unit.query.query, attr_dimensions, content_groups, kpiConfig]);

  const attributionsState = useMemo(() => {
    return {
      eventGoal: equivalentQuery.eventGoal,
      touchpoint: equivalentQuery.touchpoint,
      models: equivalentQuery.models,
      linkedEvents: [],
      attr_dimensions: equivalentQuery.attr_dimensions,
      content_groups: equivalentQuery.content_groups,
      queryOptions: { group_analysis: 'all' },
      attrQueries: equivalentQuery.attrQueries
    };
  }, [equivalentQuery]);

  if (resultState.loading) {
    content = (
      <div className='flex justify-center items-center w-full h-full'>
        <Spin size='small' />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  if (resultState.apiCallStatus && !resultState.apiCallStatus.required) {
    content = (
      <div className='flex justify-center flex-col items-center w-full h-full px-2 text-center'>
        <SVG name='nodata' />
        <Text type='title' color='grey' extraClass='mb-0'>
          {resultState.apiCallStatus.message}
        </Text>
      </div>
    );
  }

  if (resultState.data) {
    const reportSelectedChart = get(
      unit,
      'query.settings.chart',
      apiChartAnnotations[CHART_TYPE_TABLE]
    );

    const selectedDashboardPresentation = get(
      unit,
      'query.settings.dashboardPresentation',
      DEFAULT_DASHBOARD_PRESENTATION
    );

    const dashboardPresentation =
      selectedDashboardPresentation === DASHBOARD_PRESENTATION_KEYS.CHART
        ? reportSelectedChart
        : apiChartAnnotations[CHART_TYPE_TABLE];

    const {
      eventGoal,
      touchpoint,
      models,
      linkedEvents,
      attr_dimensions,
      content_groups,
      attrQueries,
      queryOptions
    } = attributionsState;

    content = (
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
        section={DASHBOARD_WIDGET_SECTION}
        attr_dimensions={attr_dimensions}
        content_groups={content_groups}
        currMetricsValue={0}
        chartType={presentationObj[dashboardPresentation]}
        cardSize={unit.cardSize}
        unitId={unit.id}
        attrQueries={attrQueries}
        queryOptions={queryOptions}
      />
    );

    // if (queryType === QUERY_TYPE_ATTRIBUTION) {
    //   content = (
    //     <Attributions
    //       durationObj={durationObj}
    //       unit={unit}
    //       resultState={resultState}
    //       attributionsState={attributionsState}
    //       chartType={presentationObj[dashboardPresentation]}
    //       section={DASHBOARD_WIDGET_SECTION}
    //     />
    //   );
    // }
  }
  return <>{content}</>;
}

export default CardContent;
