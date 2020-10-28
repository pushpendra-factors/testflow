import moment from 'moment';

export const labelsObj = {
  totalEvents: 'Event Count',
  totalUsers: 'User Count',
  activeUsers: 'User Count',
  frequency: 'Count'
};

const constantObj = {
  each: 'each_given_event',
  any: 'any_given_event',
  all: 'all_given_event'
};

export const initialState = { loading: false, error: false, data: null };

export const initialResultState = [1, 2, 3, 4].map(() => {
  return initialState;
});

const getEventsWithProperties = (queries) => {
  const ewps = [];
  queries.forEach(ev => {
    ewps.push({
      na: ev.label,
      pr: []
    });
  });
  return ewps;
};

export const getFunnelQuery = (groupBy, queries) => {
  const query = {};
  query.cl = 'funnel';
  query.ty = 'unique_users';

  const period = {
    from: moment().subtract(7, 'days').startOf('day').utc().unix(),
    to: moment().utc().unix()
  };

  query.fr = period.from;
  query.to = period.to;

  query.ewp = getEventsWithProperties(queries);
  query.gbt = '';

  const appliedGroupBy = [...groupBy.event, ...groupBy.global];

  query.gbp = appliedGroupBy
    .map(opt => {
      if (opt.eventIndex) {
        return {
          pr: opt.property,
          en: opt.prop_category,
          pty: opt.prop_type,
          ena: opt.eventName,
          eni: opt.eventIndex
        };
      } else {
        return {
          pr: opt.property,
          en: opt.prop_category,
          pty: opt.prop_type,
          ena: opt.eventName
        };
      }
    });
  query.ec = 'any_given_event';
  query.tz = 'Asia/Kolkata';
  return query;
}

export const getQuery = (activeTab, queryType, groupBy, queries, breakdownType = 'each') => {
  const query = {};
  query.cl = queryType === 'event' ? 'events' : 'funnel';
  query.ty = parseInt(activeTab) === 1 ? 'unique_users' : 'events_occurrence';

  const period = {
    from: moment().subtract(7, 'days').startOf('day').utc().unix(),
    to: moment().utc().unix()
  };

  query.fr = period.from;
  query.to = period.to;

  if (activeTab === '2') {
    query.ewp = [
      {
        na: '$session',
        pr: []
      }
    ];
    query.gbt = '';
  } else {
    query.ewp = getEventsWithProperties(queries);
    query.gbt = 'date';

    const appliedGroupBy = [...groupBy.event, ...groupBy.global];

    query.gbp = appliedGroupBy
      .map(opt => {
        if (opt.eventIndex) {
          return {
            pr: opt.property,
            en: opt.prop_category,
            pty: opt.prop_type,
            ena: opt.eventName,
            eni: opt.eventIndex
          };
        } else {
          return {
            pr: opt.property,
            en: opt.prop_category,
            pty: opt.prop_type,
            ena: opt.eventName
          };
        }
      });
  }
  query.ec = constantObj[breakdownType];
  query.tz = 'Asia/Kolkata';
  return query;
};

export const calculateFrequencyData = (eventData, userData, appliedBreakdown) => {
  if (appliedBreakdown.length) {
    return calculateFrequencyDataForBreakdown(eventData, userData);
  } else {
    return calculateFrequencyDataForNoBreakdown(eventData, userData);
  }
};

export const calculateFrequencyDataForNoBreakdown = (eventData, userData) => {
  const rows = eventData.rows.map((elem, index) => {
    const eventVals = elem.slice(1).map((e, idx) => {
      if (!e) return e;
      const eVal = e / userData.rows[index][idx + 1];
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [elem[0], ...eventVals];
  });
  const result = { ...eventData, rows };
  return result;
};

const getEventIdx = (eventData, userObj) => {
  const str = userObj.slice(0, userObj.length - 1).join(',');
  const eventIdx = eventData.findIndex(elem => elem.slice(0, elem.length - 1).join(',') === str);
  return eventIdx;
};

export const calculateFrequencyDataForBreakdown = (eventData, userData) => {
  const rows = userData.rows.map(userObj => {
    const eventIdx = getEventIdx(eventData.rows, userObj);
    let eventObj = null;
    if (eventIdx > -1) {
      eventObj = eventData.rows[eventIdx];
    }
    let eVal = 0;
    if (eventObj && eventObj[eventObj.length - 1] && userObj[userObj.length - 1]) {
      eVal = eventObj[eventObj.length - 1] / userObj[userObj.length - 1];
      eVal = eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    }
    return [...userObj.slice(0, userObj.length - 1), eVal];
  });
  const result = { ...userData, rows };
  return result;
};

export const calculateActiveUsersData = (userData, sessionData, appliedBreakdown) => {
  if (appliedBreakdown.length) {
    return calculateActiveUsersDataForBreakdown(userData, sessionData);
  } else {
    return calculateActiveUsersDataForNoBreakdown(userData, sessionData);
  }
};

const calculateActiveUsersDataForNoBreakdown = (userData, sessionData) => {
  const rows = userData.rows.map((elem) => {
    const eventVals = elem.slice(1).map((e) => {
      if (!e) return e;
      const eVal = sessionData.rows[0][1] / e;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [elem[0], ...eventVals];
  });
  const result = { ...userData, rows };
  return result;
};

const calculateActiveUsersDataForBreakdown = (userData, sessionData) => {
  const rows = userData.rows.map((elem) => {
    const eventVals = elem.slice(elem.length - 1).map((e) => {
      if (!e) return e;
      const eVal = sessionData.rows[0][1] / e;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [...elem.slice(0, elem.length - 1), ...eventVals];
  });
  const result = { ...userData, rows };
  return result;
};

export const hasApiFailed = (res) => {
  if (res.data && res.data.result_group && res.data.result_group[0] && res.data.result_group[0].headers && (res.data.result_group[0].headers.indexOf('error') > -1)) {
    return true;
  }
  return false;
};

export const numberWithCommas = (x) => {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
};

export const formatApiData = (data, appliedBreakdown) => {
  if (!appliedBreakdown.length) {
    return data;
  }

  const result = { ...data };

  if (result.headers.indexOf('event_name') !== 1) {
    const idx = result.headers.indexOf('event_name');
    if (idx === -1) {
      return null;
    } else {
      result.headers = [result.headers[0], 'event_name', ...result.headers.slice(1, idx)];
      result.rows = result.rows.map(row => {
        return [row[0], row[idx], ...row.slice(1, idx)];
      });
    }
  }
  return result;
};
