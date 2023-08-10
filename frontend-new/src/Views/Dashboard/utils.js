import MomentTz from 'Components/MomentTz';
import {
  getEventsData,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
  getProfileData,
  getKPIData,
  getAttributionsDataV1
} from '../../reducers/coreQuery/services';
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  NAMED_QUERY,
  ATTRIBUTION_METRICS,
  PREDEFINED_DATES,
  QUERY_TYPE_PROFILE,
  QUERY_TYPE_KPI
} from '../../utils/constants';
import {
  getItemFromLocalStorage,
  setItemToLocalStorage
} from '../../utils/localStorage.helpers';
import {
  DashboardDefaultDateRangeFormat,
  DefaultDateRangeFormat
} from '../CoreQuery/utils';
import { DASHBOARD_KEYS } from '../../constants/localStorage.constants';
import { getValidGranularityOptionsFromDaysDiff } from '../../utils/dataFormatter';

const formatFilters = (pr) => {
  return pr.map((p) => {
    if (p.ty === 'datetime') {
      return {
        ...p,
        va: p.va
      };
    }
    return p;
  });
};

export const getValidGranularityForSavedQueryWithSavedGranularity = ({
  durationObj,
  savedFrequency
}) => {
  // its possible that user may have saved gbt as hour but if the duration selected is more than 1 day we should not use gbt as hour
  if (!savedFrequency) {
    return 'date';
  }
  if (durationObj.from && durationObj.to) {
    const fr = MomentTz(durationObj.from).startOf('day').utc().unix();
    const to = MomentTz(durationObj.to).endOf('day').utc().unix();
    const daysDiff = MomentTz(to * 1000).diff(fr * 1000, 'days');
    const validGranularityOptions = getValidGranularityOptionsFromDaysDiff({
      daysDiff
    });
    const gbt =
      validGranularityOptions.indexOf(savedFrequency) > -1
        ? savedFrequency
        : validGranularityOptions[0];
    return gbt;
  }
  return 'date';
};

export const getDataFromServer = (
  query,
  unitId,
  dashboardId,
  durationObj,
  refresh,
  activeProjectId,
  v1 = false
) => {
  if (query.query.query_group) {
    const isCampaignQuery =
      query.query.cl && query.query.cl === QUERY_TYPE_CAMPAIGN;
    let queryGroup = query.query.query_group;
    queryGroup = queryGroup.map((elem) => {
      const obj = {
        ...elem,
        fr: MomentTz(durationObj.from).startOf('day').utc().unix(),
        to: MomentTz(durationObj.to).endOf('day').utc().unix(),
        gbt: elem.gbt ? durationObj.frequency : ''
      };
      if (!isCampaignQuery) {
        obj.ewp = obj.ewp.map((e) => {
          const pr = formatFilters(e.pr || []);
          return {
            ...e,
            pr
          };
        });
        obj.gup = formatFilters(obj.gup || []);
      }
      return obj;
    });
    if (isCampaignQuery) {
      return getCampaignsData(
        activeProjectId,
        { query_group: queryGroup, cl: QUERY_TYPE_CAMPAIGN },
        {
          refresh,
          unit_id: unitId,
          id: dashboardId
        },
        false
      );
    } else {
      return getEventsData(
        activeProjectId,
        queryGroup,
        {
          refresh,
          unit_id: unitId,
          id: dashboardId
        },
        false
      );
    }
  } else if (query.query.cl && query.query.cl === QUERY_TYPE_KPI) {
    const fr = MomentTz(durationObj.from).startOf('day').utc().unix();
    const to = MomentTz(durationObj.to).endOf('day').utc().unix();

    const KPIQuery = {
      ...query.query,
      qG: query.query.qG.map((q) => {
        return {
          ...q,
          fr,
          to,
          gbt: q.gbt ? durationObj.frequency : ''
        };
      })
    };

    return getKPIData(
      activeProjectId,
      KPIQuery,
      {
        refresh,
        unit_id: unitId,
        id: dashboardId
      },
      false
    );
  } else if (query.query.cl && query.query.cl === QUERY_TYPE_ATTRIBUTION) {
    let attributionQuery = query.query;
    if (durationObj.from && durationObj.to) {
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          from: MomentTz(durationObj.from).startOf('day').utc().unix(),
          to: MomentTz(durationObj.to).endOf('day').utc().unix()
        }
      };
    } else {
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          from: MomentTz().startOf('week').utc().unix(),
          to:
            MomentTz().format('dddd') !== 'Sunday'
              ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
              : MomentTz().utc().unix()
        }
      };
    }

    // synchronising from and to when the query has kpi query group for hubspot and salesforce group type
    if (attributionQuery?.query?.kpi_query_group) {
      const fr = attributionQuery.query.from;
      const to = attributionQuery.query.to;
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          kpi_query_group: {
            ...attributionQuery.query.kpi_query_group,
            qG: attributionQuery.query.kpi_query_group?.qG?.map((q) => {
              return {
                ...q,
                fr,
                to,
                gbt: q.gbt ? durationObj.frequency : ''
              };
            })
          }
        }
      };
    } else if (attributionQuery?.query?.kpi_queries?.length) {
      const fr = attributionQuery.query.from;
      const to = attributionQuery.query.to;
      attributionQuery.query.kpi_queries = attributionQuery.query.kpi_queries.map((kpiQ, index) => {
        return {
          ...kpiQ,
          kpi_query_group: {
            ...attributionQuery.query.kpi_queries[index].kpi_query_group,
            qG: attributionQuery.query.kpi_queries[index].kpi_query_group.qG.map((q) => {
              return {
                ...q,
                fr,
                to,
                gbt: q.gbt ? q.gbt : ''
              };
            })
          }
        };
      })
    }

    return v1
      ? getAttributionsDataV1(
          activeProjectId,
          attributionQuery,
          {
            refresh,
            unit_id: unitId,
            id: dashboardId
          },
          false
        )
      : getAttributionsData(
          activeProjectId,
          attributionQuery,
          {
            refresh,
            unit_id: unitId,
            id: dashboardId
          },
          false
        );
  } else if (query.query.cl && query.query.cl === QUERY_TYPE_PROFILE) {
    const profileQuery = query.query;
    return getProfileData(
      activeProjectId,
      profileQuery,
      {
        refresh,
        unit_id: unitId,
        id: dashboardId
      },
      false
    );
  } else {
    let funnelQuery = query.query;
    if (durationObj.from && durationObj.to) {
      funnelQuery = {
        ...funnelQuery,
        fr: MomentTz(durationObj.from).startOf('day').utc().unix(),
        to: MomentTz(durationObj.to).endOf('day').utc().unix()
      };
    } else {
      funnelQuery = {
        ...funnelQuery,
        fr: MomentTz().startOf('week').utc().unix(),
        to:
          MomentTz().format('dddd') !== 'Sunday'
            ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
            : MomentTz().utc().unix()
      };
      funnelQuery.ewp = funnelQuery.ewp.map((e) => {
        const pr = formatFilters(e.pr || []);
        return {
          ...e,
          pr
        };
      });
      funnelQuery.gup = formatFilters(funnelQuery.gup || []);
    }
    return getFunnelData(
      activeProjectId,
      funnelQuery,
      {
        refresh,
        unit_id: unitId,
        id: dashboardId
      },
      false
    );
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
      unit_id: unit.id
    };
  });

  query.custom_group_units = customGroupUnits.map((unit) => {
    const usefulQuery = { ...unit.query };
    delete usefulQuery.type;
    delete usefulQuery.cl;
    return {
      unit_id: unit.id,
      ...usefulQuery
    };
  });

  if (durationObj.from && durationObj.to) {
    if (durationObj?.dateType === 'now' || durationObj?.dateType === 'today') {
      query.from = MomentTz(durationObj.from).utc().unix();
      query.to = MomentTz(durationObj.to).utc().unix();
    } else {
      query.from = MomentTz(durationObj.from).startOf('day').utc().unix();
      query.to = MomentTz(durationObj.to).endOf('day').utc().unix();
    }
  } else {
    query.from = MomentTz().startOf('week').utc().unix();
    query.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
  }
  // query.from = 1601490600;
  // query.to = 1604168999;
  return query;
};

export const getDashboardDateRange = () => {
  const lastAppliedDuration = JSON.parse(
    getItemFromLocalStorage(DASHBOARD_KEYS.DASHBOARD_DURATION)
  );
  if (lastAppliedDuration) {
    const dateType = lastAppliedDuration.dateType;
    switch (dateType) {
      case PREDEFINED_DATES.TODAY: {
        return {
          ...lastAppliedDuration,
          from: MomentTz().startOf('day'),
          to: MomentTz().endOf('day')
        };
      }
      case PREDEFINED_DATES.YESTERDAY: {
        return {
          ...lastAppliedDuration,
          from: MomentTz().subtract(1, 'day').startOf('day'),
          to: MomentTz().subtract(1, 'day').endOf('day')
        };
      }
      case PREDEFINED_DATES.THIS_WEEK: {
        return {
          ...DefaultDateRangeFormat
        };
      }
      case PREDEFINED_DATES.LAST_WEEK: {
        return {
          ...DashboardDefaultDateRangeFormat
        };
      }
      case PREDEFINED_DATES.LAST_MONTH: {
        return {
          ...lastAppliedDuration,
          from: MomentTz().subtract(1, 'month').startOf('month'),
          to: MomentTz().subtract(1, 'month').endOf('month')
        };
      }
      case PREDEFINED_DATES.THIS_MONTH: {
        if (MomentTz().format('D') === '1') {
          return {
            ...lastAppliedDuration,
            from: MomentTz().subtract(1, 'day').startOf('month'),
            to: MomentTz().subtract(1, 'day').endOf('month'),
            dateType: PREDEFINED_DATES.LAST_MONTH
          };
        } else {
          return {
            ...lastAppliedDuration,
            from: MomentTz().startOf('month'),
            to: MomentTz().subtract(1, 'day').endOf('day')
          };
        }
      }
      default:
        return lastAppliedDuration;
    }
  }
  setItemToLocalStorage(
    DASHBOARD_KEYS.DASHBOARD_DURATION,
    JSON.stringify(DashboardDefaultDateRangeFormat)
  );
  return {
    ...DashboardDefaultDateRangeFormat
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
      enabled: currentMetric.length ? currentMetric[0].enabled : am.enabled
    };
  });
  return result;
};
