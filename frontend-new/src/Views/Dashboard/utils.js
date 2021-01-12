import moment from "moment";
import {
  runQuery,
  getFunnelData,
  getAttributionsData,
  getCampaignsData,
} from "../../reducers/coreQuery/services";
import {
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
} from "../../utils/constants";

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
          fr: moment(durationObj.from).startOf("day").utc().unix(),
          to: moment(durationObj.to).endOf("day").utc().unix(),
          gbt: elem.gbt
            ? isCampaignQuery
              ? "date"
              : durationObj.frequency
            : "",
        };
      });
    } else {
      queryGroup = queryGroup.map((elem) => {
        return {
          ...elem,
          fr: moment().startOf("week").utc().unix(),
          to:
            moment().format("dddd") !== "Sunday"
              ? moment().subtract(1, "day").endOf("day").utc().unix()
              : moment().utc().unix(),
          gbt: elem.gbt
            ? isCampaignQuery
              ? "date"
              : durationObj.frequency
            : "",
        };
      });
    }
    if (isCampaignQuery) {
      if (refresh) {
        return getCampaignsData(activeProjectId, {
          query_group: queryGroup,
          cl: QUERY_TYPE_CAMPAIGN,
        });
      } else {
        return getCampaignsData(
          activeProjectId,
          { query_group: queryGroup, cl: QUERY_TYPE_CAMPAIGN },
          {
            refresh,
            unit_id: unitId,
            id: dashboardId,
          }
        );
      }
    } else {
      if (refresh) {
        return runQuery(activeProjectId, queryGroup);
      } else {
        return runQuery(activeProjectId, queryGroup, {
          refresh,
          unit_id: unitId,
          id: dashboardId,
        });
      }
    }
  } else if (query.query.cl && query.query.cl === QUERY_TYPE_ATTRIBUTION) {
    let attributionQuery = query.query;
    if (durationObj.from && durationObj.to) {
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          from: moment(durationObj.from).startOf("day").utc().unix(),
          to: moment(durationObj.to).endOf("day").utc().unix(),
        },
      };
    } else {
      attributionQuery = {
        ...attributionQuery,
        query: {
          ...attributionQuery.query,
          from: moment().startOf("week").utc().unix(),
          to:
            moment().format("dddd") !== "Sunday"
              ? moment().subtract(1, "day").endOf("day").utc().unix()
              : moment().utc().unix(),
        },
      };
    }
    return getAttributionsData(activeProjectId, attributionQuery, {
      refresh: false,
      unit_id: unitId,
      id: dashboardId,
    });
  } else {
    let funnelQuery = query.query;
    if (durationObj.from && durationObj.to) {
      funnelQuery = {
        ...funnelQuery,
        fr: moment(durationObj.from).startOf("day").utc().unix(),
        to: moment(durationObj.to).endOf("day").utc().unix(),
      };
    } else {
      funnelQuery = {
        ...funnelQuery,
        fr: moment().startOf("week").utc().unix(),
        to:
          moment().format("dddd") !== "Sunday"
            ? moment().subtract(1, "day").endOf("day").utc().unix()
            : moment().utc().unix(),
      };
    }
    return getFunnelData(activeProjectId, funnelQuery, {
      refresh: false,
      unit_id: unitId,
      id: dashboardId,
    });
  }
};
