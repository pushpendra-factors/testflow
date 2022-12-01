import MomentTz from 'Components/MomentTz';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
  EACH_USER_TYPE
} from '../../utils/constants';

export const getQuery = ({ queryType, requestQuery, user_type }) => {
  const startOfWeek = MomentTz().startOf('week').utc().unix();
  const todayNow = MomentTz().utc().unix();

  if (queryType === QUERY_TYPE_FUNNEL) {
    return {
      ...requestQuery,
      fr: startOfWeek,
      to: todayNow
    };
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    return {
      ...requestQuery,
      query: {
        ...requestQuery.query,
        from: startOfWeek,
        to: todayNow
      }
    };
  }

  if (queryType === QUERY_TYPE_EVENT) {
    return {
      query_group: requestQuery.map((q) => {
        return {
          ...q,
          fr: q.fr,
          to: q.to,
          gbt: user_type === EACH_USER_TYPE ? q?.gbt : '' // when user_type is ANY_USER_TYPE/ALL_USER_TYPE then gbt is sent as blank. Please make sure this use case is met.
        };
      })
    };
  }

  if (queryType === QUERY_TYPE_CAMPAIGN) {
    return {
      ...requestQuery,
      query_group: requestQuery.query_group.map((q) => {
        return {
          ...q,
          fr: startOfWeek,
          to: todayNow,
          gbt: q.gbt ? q.gbt : ''
        };
      })
    };
  }

  if (queryType === QUERY_TYPE_KPI) {
    return {
      ...requestQuery,
      qG: requestQuery.qG.map((q) => {
        return {
          ...q,
          fr: q.fr,
          to: q.to,
          gbt: q.gbt ? q.gbt : ''
        };
      })
    };
  }

  if (queryType === QUERY_TYPE_PROFILE) {
    return {
      ...requestQuery
    };
  }
};
