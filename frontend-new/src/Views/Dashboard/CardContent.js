import React, { useMemo } from 'react';
import { Spin } from 'antd';
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
  getCampaignStateFromRequestQuery,
  getProfileQueryFromRequestQuery,
  getKPIStateFromRequestQuery,
} from '../CoreQuery/utils';
import EventsAnalytics from './EventsAnalytics';
import Funnels from './Funnels';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  DASHBOARD_WIDGET_SECTION,
  reverse_user_types,
  presentationObj,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI,
} from '../../utils/constants';
import Attributions from './Attributions';
import CampaignAnalytics from './CampaignAnalytics';
import NoDataChart from '../../components/NoDataChart';
import { useSelector } from 'react-redux';
import { SVG, Text } from '../../components/factorsComponents';
import ProfileAnalysis from './ProfileAnalysis';
import KPIAnalysis from './KPIAnalysis';

function CardContent({ unit, resultState, durationObj }) {
  let content = null;
  const { eventNames, attr_dimensions } = useSelector(
    (state) => state.coreQuery
  );
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
        attr_dimensions
      );
    } else if (
      unit.query.query.cl &&
      unit.query.query.cl === QUERY_TYPE_PROFILE
    ) {
      return getProfileQueryFromRequestQuery(unit.query.query);
    } else {
      return getStateQueryFromRequestQuery(unit.query.query);
    }
  }, [unit.query.query, attr_dimensions]);

  const { queryType } = equivalentQuery;
  const breakdownType = useMemo(() => {
    if (queryType === QUERY_TYPE_EVENT) {
      return reverse_user_types[unit.query.query.query_group[0].ec];
    }
  }, [queryType, unit.query.query]);

  const events = useMemo(() => {
    if (
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_PROFILE ||
      queryType === QUERY_TYPE_KPI
    ) {
      return equivalentQuery.events.map((elem) =>
        elem.alias ? elem.alias : eventNames[elem.label] || elem.label
      );
    }
  }, [equivalentQuery.events, queryType]);

  const breakdown = useMemo(() => {
    if (
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType == QUERY_TYPE_PROFILE ||
      queryType === QUERY_TYPE_KPI
    ) {
      return [
        ...equivalentQuery.breakdown.event,
        ...equivalentQuery.breakdown.global,
      ];
    }
  }, [queryType, equivalentQuery.breakdown]);

  const arrayMapper = useMemo(() => {
    if (
      queryType === QUERY_TYPE_EVENT ||
      queryType === QUERY_TYPE_FUNNEL ||
      queryType === QUERY_TYPE_KPI
    ) {
      const am = [];
      equivalentQuery.events.forEach((q, index) => {
        am.push({
          eventName: q.alias ? q.alias : eventNames[q.label] || q.label,
          index,
          mapper: `event${index + 1}`,
          displayName: q.alias ? q.alias : eventNames[q.label] || q.label,
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
      };
    }
  }, [equivalentQuery, queryType]);

  const campaignState = useMemo(() => {
    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return {
        channel: unit.query.query.query_group[0].channel,
        filters: unit.query.query.query_group[0].filters,
        select_metrics: unit.query.query.query_group[0].select_metrics,
        group_by: unit.query.query.query_group[0].group_by,
      };
    }
  }, [queryType, unit.query.query]);

  const campaignsArrayMapper = useMemo(() => {
    if (queryType === QUERY_TYPE_CAMPAIGN) {
      return campaignState.select_metrics.map((metric, index) => {
        return {
          eventName: metric,
          index,
          mapper: `event${index + 1}`,
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
    content = (
      <div className='flex justify-center items-center w-full h-full'>
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
    const dashboardPresentation =
      unit.query && unit.query.settings && unit.query.settings.chart
        ? unit.query.settings.chart
        : 'pt';

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
          queries={events}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          section={DASHBOARD_WIDGET_SECTION}
          breakdown={breakdown}
          unit={unit}
          arrayMapper={arrayMapper}
        />
      );
    }
  }

  return <>{content}</>;
}

export default CardContent;
