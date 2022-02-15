import MomentTz from 'Components/MomentTz';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_CAMPAIGN,
  QUERY_TYPE_KPI,
  QUERY_TYPE_PROFILE,
} from '../../utils/constants';

export const getQuery = ({ queryType, requestQuery }) => {
  const startOfWeek = MomentTz().startOf('week').utc().unix();
  const todayNow = MomentTz().utc().unix();

  if (queryType === QUERY_TYPE_FUNNEL) {
    return {
      ...requestQuery,
      fr: startOfWeek,
      to: todayNow,
    };
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    return {
      ...requestQuery,
      query: {
        ...requestQuery.query,
        from: startOfWeek,
        to: todayNow,
      },
    };
  }

  if (queryType === QUERY_TYPE_EVENT) {
    return {
      query_group: requestQuery.map((q) => {
        return {
          ...q,
          fr: startOfWeek,
          to: todayNow,
          gbt: q.gbt ? 'date' : '',
        };
      }),
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
          gbt: q.gbt ? 'date' : '',
        };
      }),
    };
  }

  if (queryType === QUERY_TYPE_KPI) {
    return {
      ...requestQuery,
      qG: requestQuery.qG.map((q) => {
        return {
          ...q,
          fr: startOfWeek,
          to: todayNow,
          gbt: q.gbt ? 'date' : '',
        };
      }),
    };
  }

  if (queryType === QUERY_TYPE_PROFILE) {
    return {
      ...requestQuery,
    };
  }
};
