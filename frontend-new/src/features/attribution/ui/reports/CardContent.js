import React, { useMemo } from 'react';
import { useSelector } from 'react-redux';
import { get } from 'lodash';
import NoDataChart from 'Components/NoDataChart';
import {
  apiChartAnnotations,
  CHART_TYPE_TABLE,
  presentationObj,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE
} from 'Utils/constants';
import {
  getAttributionStateFromRequestQuery,
  getCampaignStateFromRequestQuery,
  getKPIStateFromRequestQuery,
  getProfileQueryFromRequestQuery,
  getStateQueryFromRequestQuery
} from 'Views/CoreQuery/utils';
import AttributionChart from './AttributionChart';
import {
  DEFAULT_DASHBOARD_PRESENTATION,
  DASHBOARD_PRESENTATION_KEYS
} from 'Components/SaveQuery/saveQuery.constants';

function CardContent({ durationObj, unit, attributionMetrics }) {
  let content = null;

  // this var is just for test remove it replace it with resultState.error in below ifcondition for NO Data Chart
  let dummyError = false;

  const { attr_dimensions: attrDimensions, content_groups: contentGroups } =
    useSelector((state) => state.coreQuery);
  const { config: kpiConfig } = useSelector((state) => state.kpi);

  const equivalentQuery = useMemo(() => {
    if (unit.query.query.query_group) {
      const isCampaignQuery =
        unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_CAMPAIGN;
      if (isCampaignQuery) {
        return getCampaignStateFromRequestQuery(
          unit.query.query.query_group[0]
        );
      } else {
        return getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
      }
    } else if (unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_KPI) {
      return getKPIStateFromRequestQuery(unit.query.query, kpiConfig);
    } else if (
      unit.query.query.cl &&
      unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
    ) {
      return getAttributionStateFromRequestQuery(
        unit.query.query.query,
        attrDimensions,
        contentGroups,
        kpiConfig
      );
    } else if (
      unit.query.query.cl &&
      unit.query.query.cl === QUERY_TYPE_PROFILE
    ) {
      return getProfileQueryFromRequestQuery(unit.query.query);
    } else {
      return getStateQueryFromRequestQuery(unit.query.query);
    }
  }, [unit.query.query, attrDimensions, contentGroups, kpiConfig]);

  const { queryType } = equivalentQuery;

  const attributionsState = useMemo(() => {
    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      return {
        eventGoal: equivalentQuery.eventGoal,
        touchpoint: equivalentQuery.touchpoint,
        models: equivalentQuery.models,
        linkedEvents: equivalentQuery.linkedEvents,
        attr_dimensions: equivalentQuery.attr_dimensions,
        content_groups: equivalentQuery.content_groups,
        queryOptions: { group_analysis: equivalentQuery.analyze_type },
        attrQueries: equivalentQuery.attrQueries
      };
    }
  }, [equivalentQuery, queryType]);

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

  //will be replace with resultState created in the widgetCard component and replace all dummyResult variables with resultState
  const dummyResultState = {
    loading: false,
    error: false,
    data: {
      cache_meta: {
        from: 1664632800,
        last_computed_at: 1665426618,
        preset: '',
        refreshed_at: 1665426618,
        timezone: 'Australia/Sydney',
        to: 1665233999
      },
      headers: [
        'Campaign',
        'AdGroup',
        'Impressions',
        'Clicks',
        'Spend',
        'CTR(%)',
        'Average CPC',
        'CPM',
        'ClickConversionRate(%)',
        'Opportunity Created - Users',
        'Cost Per Conversion',
        'Compare - Users',
        'Compare Cost Per Conversion',
        'key'
      ],
      meta: {
        currency: '',
        metrics: null,
        query: {
          agEn: '',
          agFn: '',
          agPr: '',
          agTy: '',
          cl: '',
          ec: '',
          ewp: null,
          fr: 0,
          gbp: null,
          gbt: null,
          gup: null,
          ovp: false,
          see: 0,
          sse: 0,
          to: 0,
          ty: '',
          tz: ''
        }
      },
      query: null,
      rows: [
        [
          'Grand Total',
          'Grand Total',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          101,
          0,
          0,
          0,
          'Grand Total'
        ],
        [
          '$none',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          52,
          0,
          0,
          0,
          '$none:-:$none:-:$none'
        ],
        [
          'Traditional_media',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          13,
          0,
          0,
          0,
          '$none:-:Traditional_media:-:$none'
        ],
        [
          'Context_marketing',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          11,
          0,
          0,
          0,
          '$none:-:Context_marketing:-:$none'
        ],
        [
          'Seasonal_push',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          8,
          0,
          0,
          0,
          '$none:-:Seasonal_push:-:$none'
        ],
        [
          'Email_marketing',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          6,
          0,
          0,
          0,
          '$none:-:Email_marketing:-:$none'
        ],
        [
          'Brand_awareness',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          4,
          0,
          0,
          0,
          '$none:-:Brand_awareness:-:$none'
        ],
        [
          'Product_launch',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          4,
          0,
          0,
          0,
          '$none:-:Product_launch:-:$none'
        ],
        [
          'Rebranding_campaign',
          '$none',
          0,
          0,
          0,
          0,
          0,
          0,
          0,
          3,
          0,
          0,
          0,
          '$none:-:Rebranding_campaign:-:$none'
        ]
      ]
    },
    apiCallStatus: {
      required: true,
      message: null
    }
  };

  if (dummyError) {
    content = (
      <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataChart />
      </div>
    );
  }

  if (dummyResultState.data) {
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

    content = (
      <AttributionChart
        event={eventGoal.label}
        linkedEvents={linkedEvents}
        touchpoint={touchpoint}
        durationObj={durationObj}
        data={
          dummyResultState.data.result
            ? dummyResultState.data.result
            : dummyResultState.data
        }
        isWidgetModal={false}
        attribution_method={models[0]}
        attribution_method_compare={models[1]}
        section={''}
        attr_dimensions={attr_dimensions}
        content_groups={content_groups}
        currMetricsValue={0}
        chartType={'stackedbarchartt'}
        // replace above type with below
        // chartType={presentationObj[dashboardPresentation]}
        cardSize={unit.cardSize}
        unitId={unit.id}
        attrQueries={attrQueries}
        queryOptions={queryOptions}
        attributionMetrics={attributionMetrics}
      />
    );
  }

  return <>{content}</>;
}

export default CardContent;
