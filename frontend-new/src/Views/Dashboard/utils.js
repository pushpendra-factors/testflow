import moment from 'moment';
import {
  getEventsData,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
} from '../../reducers/coreQuery/services';
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  NAMED_QUERY,
  LOCAL_STORAGE_ITEMS,
  ATTRIBUTION_METRICS,
} from '../../utils/constants';
import {
  getItemFromLocalStorage,
  setItemToLocalStorage,
} from '../../utils/dataFormatter';
import { DashboardDefaultDateRangeFormat } from '../CoreQuery/utils';

export const getDataFromServer = (
  query,
  unitId,
  dashboardId,
  durationObj,
  refresh,
  activeProjectId
) => {
  if (query.query.query_group) {
    const isCampaignQuery =
      query.query.cl && query.query.cl === QUERY_TYPE_CAMPAIGN;
    let queryGroup = query.query.query_group;
    if (durationObj.from && durationObj.to) {
      queryGroup = queryGroup.map((elem) => {
        return {
          ...elem,
          fr: moment(durationObj.from).startOf('day').utc().unix(),
          to: moment(durationObj.to).endOf('day').utc().unix(),
          gbt: elem.gbt
            ? isCampaignQuery
              ? 'date'
              : durationObj.frequency
            : '',
        };
      });
    } else {
      queryGroup = queryGroup.map((elem) => {
        return {
          ...elem,
          fr: moment().startOf('week').utc().unix(),
          to:
            moment().format('dddd') !== 'Sunday'
              ? moment().subtract(1, 'day').endOf('day').utc().unix()
              : moment().utc().unix(),
          gbt: elem.gbt
            ? isCampaignQuery
              ? 'date'
              : durationObj.frequency
            : '',
        };
      });
    }
    if (isCampaignQuery) {
      return getCampaignsData(
        activeProjectId,
        { query_group: queryGroup, cl: QUERY_TYPE_CAMPAIGN },
        {
          refresh,
          unit_id: unitId,
          id: dashboardId,
        }
      );
    } else {
      return getEventsData(activeProjectId, queryGroup, {
        refresh,
        unit_id: unitId,
        id: dashboardId,
      });
    }
  } else if (query.query.cl && query.query.cl === QUERY_TYPE_ATTRIBUTION) {
    let attributionQuery = query.query;
    if (durationObj.from && durationObj.to) {
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          from: moment(durationObj.from).startOf('day').utc().unix(),
          to: moment(durationObj.to).endOf('day').utc().unix(),
        },
      };
    } else {
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          from: moment().startOf('week').utc().unix(),
          to:
            moment().format('dddd') !== 'Sunday'
              ? moment().subtract(1, 'day').endOf('day').utc().unix()
              : moment().utc().unix(),
        },
      };
    }
    return getAttributionsData(activeProjectId, attributionQuery, {
      refresh,
      unit_id: unitId,
      id: dashboardId,
    });
  } else {
    let funnelQuery = query.query;
    if (durationObj.from && durationObj.to) {
      funnelQuery = {
        ...funnelQuery,
        fr: moment(durationObj.from).startOf('day').utc().unix(),
        to: moment(durationObj.to).endOf('day').utc().unix(),
      };
    } else {
      funnelQuery = {
        ...funnelQuery,
        fr: moment().startOf('week').utc().unix(),
        to:
          moment().format('dddd') !== 'Sunday'
            ? moment().subtract(1, 'day').endOf('day').utc().unix()
            : moment().utc().unix(),
      };
    }
    return getFunnelData(activeProjectId, funnelQuery, {
      refresh,
      unit_id: unitId,
      id: dashboardId,
    });
  }
};

export const getWebAnalyticsRequestBody = (units, durationObj) => {
  const query = {};
  const namedUnits = units.filter((unit) => unit.query.type === NAMED_QUERY);
  const customGroupUnits = units.filter(
    (unit) => unit.query.type !== NAMED_QUERY
  );

  query.units = namedUnits.map((unit) => {
    return {
      query_name: unit.query.qname,
      unit_id: unit.id,
    };
  });

  query.custom_group_units = customGroupUnits.map((unit) => {
    const usefulQuery = { ...unit.query };
    delete usefulQuery.type;
    delete usefulQuery.cl;
    return {
      unit_id: unit.id,
      ...usefulQuery,
    };
  });

  if (durationObj.from && durationObj.to) {
    if (durationObj?.dateType === 'now' || durationObj?.dateType === 'today') {
      query.from = moment(durationObj.from).utc().unix();
      query.to = moment(durationObj.to).utc().unix();
    } else {
      query.from = moment(durationObj.from).startOf('day').utc().unix();
      query.to = moment(durationObj.to).endOf('day').utc().unix();
    }
  } else {
    query.from = moment().startOf('week').utc().unix();
    query.to =
      moment().format('dddd') !== 'Sunday'
        ? moment().subtract(1, 'day').endOf('day').utc().unix()
        : moment().utc().unix();
  }
  // query.from = 1601490600;
  // query.to = 1604168999;
  return query;
};

export const getDashboardDateRange = () => {
  const lastAppliedDuration = getItemFromLocalStorage(
    LOCAL_STORAGE_ITEMS.DASHBOARD_DURATION
  );
  if (lastAppliedDuration) {
    return JSON.parse(lastAppliedDuration);
  }
  setItemToLocalStorage(
    LOCAL_STORAGE_ITEMS.DASHBOARD_DURATION,
    JSON.stringify(DashboardDefaultDateRangeFormat)
  );
  return {
    ...DashboardDefaultDateRangeFormat,
  };
};

export const getSavedAttributionMetrics = (metrics) => {
  const result = ATTRIBUTION_METRICS.map((am) => {
    const possibleHeaders = am.header.split(' OR ');
    const currentMetric = metrics.filter((m) => {
      const headers = m.header.split(' OR ');
      const intersection = possibleHeaders.filter(
        (h) => headers.indexOf(h) > -1
      );
      return intersection.length > 0;
    });
    return {
      ...am,
      enabled: currentMetric.length ? currentMetric[0].enabled : am.enabled,
    };
  });
  return result;
};
