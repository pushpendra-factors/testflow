import React, { useMemo, useState ,useEffect} from 'react';
import { Spin } from 'antd';
import { get } from 'lodash';
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
  getProfileQueryFromRequestQuery,
  getKPIStateFromRequestQuery
} from '../CoreQuery/utils';
import EventsAnalytics from './EventsAnalytics';
import Funnels from './Funnels';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  DASHBOARD_WIDGET_SECTION,
  REVERSE_USER_TYPES,
  presentationObj,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI,
  apiChartAnnotations,
  CHART_TYPE_TABLE
} from '../../utils/constants';
import Attributions from './Attributions';
import CampaignAnalytics from './CampaignAnalytics';
import NoDataChart from '../../components/NoDataChart';
import { useSelector } from 'react-redux';
import {
  FaErrorComp,
  SVG,
  Text,
  FaErrorLog
} from '../../components/factorsComponents';
import ProfileAnalysis from './ProfileAnalysis';
import KPIAnalysis from './KPIAnalysis';
import {
  DEFAULT_DASHBOARD_PRESENTATION,
  DASHBOARD_PRESENTATION_KEYS
} from '../../components/SaveQuery/saveQuery.constants';
import { ErrorBoundary } from 'react-error-boundary';
import NoDataInTimeRange from 'Components/NoDataInTimeRange';
import { getErrorMessage } from 'Utils/global';

function CardContent({ unit, resultState, durationObj }) {
  let content = null;
  const { eventNames, attr_dimensions, content_groups } = useSelector(
    (state) => state.coreQuery
  );
  const [errMsg,setErrMsg]=useState('');
  const { config: kpiConfig } = useSelector((state) => state.kpi);

  const equivalentQuery = useMemo(() => {
    if (unit.query.query.query_group) { 
      return getStateQueryFromRequestQuery(unit.query.query.query_group[0]);
    } else if (unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_KPI) {
      return getKPIStateFromRequestQuery(unit.query.query, kpiConfig);
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
    } else if (
      unit.query.query.cl &&
      unit.query.query.cl === QUERY_TYPE_PROFILE
    ) {
      return getProfileQueryFromRequestQuery(unit.query.query);
    } else {
      return getStateQueryFromRequestQuery(unit.query.query);
    }
  }, [unit.query.query, attr_dimensions, content_groups, kpiConfig]);

  const { queryType } = equivalentQuery;
  const breakdownType = useMemo(() => {
    if (queryType === QUERY_TYPE_EVENT) {
      return REVERSE_USER_TYPES[unit.query.query.query_group[0].ec];
    }
  }, [queryType, unit.query.query]);

  const events = useMemo(() => {
    if (
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_PROFILE
    ) {
      return equivalentQuery.events.map((elem) =>
        elem.alias ? elem.alias : eventNames[elem.label] || elem.label
      );
    }
    if (queryType === QUERY_TYPE_KPI) {
      return equivalentQuery.events;
    }
  }, [equivalentQuery.events, queryType]);

  const breakdown = useMemo(() => {
    if (
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_PROFILE ||
      queryType === QUERY_TYPE_KPI
    ) {
      return [
        ...equivalentQuery.breakdown.event,
        ...equivalentQuery.breakdown.global
      ];
    }
  }, [queryType, equivalentQuery.breakdown]);

  const arrayMapper = useMemo(() => {
    if (
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_KPI ||
      queryType === QUERY_TYPE_PROFILE
    ) {
      const am = [];
      equivalentQuery.events.forEach((q, index) => {
        am.push({
          eventName: q.alias ? q.alias : eventNames[q.label] || q.label,
          index,
          mapper: `event${index + 1}`,
          displayName: q.alias ? q.alias : eventNames[q.label] || q.label
        });
      });
      return am;
    }
  }, [equivalentQuery.events, eventNames, queryType]);

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

  useEffect(() => {
    const errorMessage = getErrorMessage(resultState);
    setErrMsg(errorMessage);
  }, [resultState]);

  const campaignState = useMemo(() => {
    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return {
        channel: unit.query.query.query_group[0].channel,
        filters: unit.query.query.query_group[0].filters,
        select_metrics: unit.query.query.query_group[0].select_metrics,
        group_by: unit.query.query.query_group[0].group_by
      };
    }
  }, [queryType, unit.query.query]);

  const campaignsArrayMapper = useMemo(() => {
    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return campaignState.select_metrics.map((metric, index) => {
        return {
          eventName: metric,
          index,
          mapper: `event${index + 1}`
        };
      });
    }
  }, [queryType, campaignState]);

  if (resultState.loading) {
    content = (
      <div className='flex justify-center items-center w-full h-full'>
        <Spin size='small' />
      </div>
    );
  }

  if (resultState.error) {

    return (
        <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataInTimeRange message={errMsg}/>
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

    if (queryType === QUERY_TYPE_FUNNEL) {
      content = (
        <Funnels
          breakdown={breakdown}
          events={events}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          unit={unit}
          arrayMapper={arrayMapper}
          section={DASHBOARD_WIDGET_SECTION}
        />
      );
    }

    if (queryType === QUERY_TYPE_EVENT) {
      content = (
        <EventsAnalytics
          durationObj={durationObj}
          breakdown={breakdown}
          events={events}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          unit={unit}
          arrayMapper={arrayMapper}
          section={DASHBOARD_WIDGET_SECTION}
          breakdownType={breakdownType}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      content = (
        <Attributions
          durationObj={durationObj}
          unit={unit}
          resultState={resultState}
          attributionsState={attributionsState}
          chartType={presentationObj[dashboardPresentation]}
          section={DASHBOARD_WIDGET_SECTION}
        />
      );
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      content = (
        <CampaignAnalytics
          unit={unit}
          resultState={resultState}
          campaignState={campaignState}
          chartType={presentationObj[dashboardPresentation]}
          arrayMapper={campaignsArrayMapper}
          section={DASHBOARD_WIDGET_SECTION}
          durationObj={durationObj}
        />
      );
    }

    if (queryType === QUERY_TYPE_PROFILE) {
      content = (
        <ProfileAnalysis
          queries={events}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          section={DASHBOARD_WIDGET_SECTION}
          breakdown={breakdown}
          unit={unit}
        />
      );
    }

    if (queryType === QUERY_TYPE_KPI) {
      content = (
        <KPIAnalysis
          kpis={events}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          section={DASHBOARD_WIDGET_SECTION}
          breakdown={breakdown}
          unit={unit}
          arrayMapper={arrayMapper}
          durationObj={durationObj}
        />
      );
    }
  }

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='small'
          title='Widget Error'
          subtitle='We are facing trouble loading this widget. Drop us a message on the in-app chat.'
          className="h-full"
        />
      }
      onError={FaErrorLog}
    >
      {content}
    </ErrorBoundary>
  );
}

export default CardContent;
