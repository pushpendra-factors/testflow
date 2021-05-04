import React from 'react';
import { Spin } from 'antd';
import {
  getStateQueryFromRequestQuery,
  getAttributionStateFromRequestQuery,
  getCampaignStateFromRequestQuery,
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
} from '../../utils/constants';
import Attributions from './Attributions';
import CampaignAnalytics from './CampaignAnalytics';
import NoDataChart from '../../components/NoDataChart';

function CardContent({ unit, resultState, setwidgetModal, durationObj }) {
  let content = null;

  if (resultState.loading) {
    content = (
      <div className='flex justify-center items-center w-full h-64'>
        <Spin size='small' />
      </div>
    );
  }

  if (resultState.error) {
    content = (
      <div className='flex justify-center items-center w-full h-64'>
        <NoDataChart />
      </div>
    );
  }

  if (resultState.data) {
    let equivalentQuery;
    if (unit.query.query.query_group) {
      const isCampaignQuery =
        unit.query.query.cl && unit.query.query.cl === QUERY_TYPE_CAMPAIGN;
      if (isCampaignQuery) {
        equivalentQuery = getCampaignStateFromRequestQuery(
          unit.query.query.query_group[0]
        );
      } else {
        equivalentQuery = getStateQueryFromRequestQuery(
          unit.query.query.query_group[0]
        );
      }
    } else if (
      unit.query.query.cl &&
      unit.query.query.cl === QUERY_TYPE_ATTRIBUTION
    ) {
      equivalentQuery = getAttributionStateFromRequestQuery(
        unit.query.query.query
      );
    } else {
      equivalentQuery = getStateQueryFromRequestQuery(unit.query.query);
    }

    let breakdown,
      events,
      arrayMapper = [],
      attributionsState,
      campaignState,
      breakdownType;

    const { queryType } = equivalentQuery;
    if (queryType === QUERY_TYPE_EVENT || queryType === QUERY_TYPE_FUNNEL) {
      breakdown = [
        ...equivalentQuery.breakdown.event,
        ...equivalentQuery.breakdown.global,
      ];
      events = [...equivalentQuery.events];
      events.forEach((q, index) => {
        arrayMapper.push({
          eventName: q.label,
          index,
          mapper: `event${index + 1}`,
        });
      });
    }

    if (queryType === QUERY_TYPE_EVENT) {
      breakdownType = reverse_user_types[unit.query.query.query_group[0].ec];
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      attributionsState = {
        eventGoal: equivalentQuery.eventGoal,
        touchpoint: equivalentQuery.touchpoint,
        models: equivalentQuery.models,
        linkedEvents: equivalentQuery.linkedEvents,
      };
    }

    if (queryType === QUERY_TYPE_CAMPAIGN) {
      campaignState = {
        channel: unit.query.query.query_group[0].channel,
        filters: unit.query.query.query_group[0].filters,
        select_metrics: unit.query.query.query_group[0].select_metrics,
        group_by: unit.query.query.query_group[0].group_by,
      };
      arrayMapper = campaignState.select_metrics.map((metric, index) => {
        return {
          eventName: metric,
          index,
          mapper: `event${index + 1}`,
        };
      });
    }

    let dashboardPresentation = 'pl';

    try {
      dashboardPresentation = unit.settings.chart;
    } catch (err) {
      console.log(err);
    }

    if (queryType === QUERY_TYPE_FUNNEL) {
      content = (
        <Funnels
          breakdown={breakdown}
          events={events.map((elem) => elem.label)}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          unit={unit}
          setwidgetModal={setwidgetModal}
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
          events={events.map((elem) => elem.label)}
          resultState={resultState}
          chartType={presentationObj[dashboardPresentation]}
          unit={unit}
          setwidgetModal={setwidgetModal}
          arrayMapper={arrayMapper}
          section={DASHBOARD_WIDGET_SECTION}
          breakdownType={breakdownType}
        />
      );
    }

    if (queryType === QUERY_TYPE_ATTRIBUTION) {
      content = (
        <Attributions
          unit={unit}
          resultState={resultState}
          setwidgetModal={setwidgetModal}
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
          setwidgetModal={setwidgetModal}
          campaignState={campaignState}
          chartType={presentationObj[dashboardPresentation]}
          arrayMapper={arrayMapper}
          section={DASHBOARD_WIDGET_SECTION}
        />
      );
    }
  }

  return <>{content}</>;
}

export default CardContent;
