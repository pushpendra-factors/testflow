import moment from 'moment';

export const labelsObj = {
  totalEvents: 'Event Count',
  totalUsers: 'User Count',
  activeUsers: 'User Count',
  frequency: 'Count'
};

export const presentationObj = {
  pb: 'barchart',
  pl: 'linechart',
  pt: 'table',
  pc: 'sparklines'
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

const operatorMap = {
  '=': 'equals',
  '!=': 'notEqual',
  contains: 'contains',
  'not contains': 'notContains',
  '<': 'lesserThan',
  '<=': 'lesserThanOrEqual',
  '>': 'greaterThan',
  '>=': 'greaterThanOrEqual'
};

const getEventsWithProperties = (queries) => {
  const ewps = [];
  queries.forEach(ev => {
    const filterProps = [];
    ev.filters.forEach(fil => {
      let vals;
      if(Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            en: fil.props[2],
            lop: 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: val
          });
        })
      } else {
        vals = fil.values;
        filterProps.push({
          en: fil.props[2],
          lop: 'AND',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: vals
        });
      }
      
    });
    ewps.push({
      na: ev.label,
      pr: filterProps
    });
  });
  return ewps;
};

export const getFunnelQuery = (groupBy, queries, dateRange) => {
  const query = {};
  query.cl = 'funnel';
  query.ty = 'unique_users';

  const period = {};
  if (dateRange.from && dateRange.to) {
    period.from = moment(dateRange.from).startOf('day').utc().unix();
    period.to = moment(dateRange.to).endOf('day').utc().unix();
  } else {
    period.from = moment().startOf('week').utc().unix();
    period.to = moment().utc().unix();
  }

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
};

export const getQuery = (activeTab, groupBy, queries, breakdownType = 'each', dateRange) => {
  const query = {};
  query.cl = 'events';
  query.ty = parseInt(activeTab) === 1 ? 'unique_users' : 'events_occurrence';

  const period = {};
  if (dateRange.from && dateRange.to) {
    period.from = moment(dateRange.from).startOf('day').utc().unix();
    period.to = moment(dateRange.to).endOf('day').utc().unix();
  } else {
    period.from = moment().startOf('week').utc().unix();
    period.to = moment().utc().unix();
  }

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
    query.gbt = breakdownType === 'each' ? dateRange.frequency || 'date' : '';

    const appliedGroupBy = [...groupBy.event, ...groupBy.global];

    query.gbp = appliedGroupBy
      .map(opt => {
        let gbpReq = {};
        if (opt.eventIndex) {
          gbpReq = {
            pr: opt.property,
            en: opt.prop_category,
            pty: opt.prop_type,
            ena: opt.eventName,
            eni: opt.eventIndex
          }; 
        } else {
          gbpReq = {
            pr: opt.property,
            en: opt.prop_category,
            pty: opt.prop_type,
            ena: opt.eventName
          };
        }
        if(opt.prop_type === 'datetime') {
          gbpReq['grn'] = "day";
        }

        return gbpReq;
      });
  }
  query.ec = activeTab === '2' ? constantObj.each : constantObj[breakdownType];
  query.tz = 'Asia/Kolkata';
  if (breakdownType === 'each') {
    if (activeTab === '2') {
      return [query];
    } else {
      return [query, { ...query, gbt: '' }];
    }
  } else {
    return [query];
  }
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
  const metrics = eventData.metrics.rows.map((elem) => {
    const idx = userData.metrics.rows.findIndex(r => r[0] === elem[0]);
    if (!elem[1] || !userData.metrics.rows[idx][1]) return 0;
    const eVal = (elem[1] / userData.metrics.rows[idx][1]);
    return [elem[0], parseFloat(eVal.toFixed(2))];
  });
  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics
    }
  };
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

  const metrics = userData.metrics.rows.map(userObj => {
    const eventIdx = getEventIdx(eventData.metrics.rows, userObj);
    let eventObj = null;
    let eVal = 0;
    if (eventIdx > -1) {
      eventObj = eventData.metrics.rows[eventIdx];
    }
    if (eventObj && userObj[userObj.length - 1] && eventObj[eventObj.length - 1]) {
      eVal = (eventObj[eventObj.length - 1] / userObj[userObj.length - 1]);
    }
    return [...userObj.slice(0, userObj.length - 1), parseFloat(eVal.toFixed(2))];
  });

  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics
    }
  };
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
      if (!e || !sessionData.rows[0][1]) return 0;
      const eVal = e / sessionData.rows[0][1] * 100;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [elem[0], ...eventVals];
  });

  const metrics = userData.metrics.rows.map(elem => {
    if (!elem[1] || !sessionData.rows[0][1]) return 0;
    const eVal = (elem[1] / sessionData.rows[0][1]) * 100;
    return [elem[0], parseFloat(eVal.toFixed(2))];
  });
  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics
    }
  };
  return result;
};

const calculateActiveUsersDataForBreakdown = (userData, sessionData) => {
  const differentDates = new Set();
  userData.rows.forEach(ud => {
    differentDates.add(ud[0]);
  });
  const rows = userData.rows.map((elem) => {
    const eventVals = elem.slice(elem.length - 1).map((e) => {
      if (!e || !sessionData.rows[0][1]) return e;
      const eVal = e / sessionData.rows[0][1] * 100;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [...elem.slice(0, elem.length - 1), ...eventVals];
  });

  const metrics = userData.metrics.rows.map(elem => {
    if (!elem[elem.length - 1] || !sessionData.rows[0][1]) return 0;
    const eVal = (elem[elem.length - 1] / sessionData.rows[0][1]) * 100;
    return [...elem.slice(0, elem.length - 1), parseFloat(eVal.toFixed(2))];
  });
  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics
    }
  };
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

export const formatApiData = (data, metrics) => {
  return { ...data, metrics };
};

export const getStateQueryFromRequestQuery = (requestQuery) => {
  const events = requestQuery.ewp.map(e => {
    return {
      label: e.na,
      filters: []
    };
  });
  const queryType = requestQuery.cl;
  const breakdown = requestQuery.gbp.map(opt => {
    return {
      property: opt.pr,
      prop_category: opt.en,
      prop_type: opt.pty,
      eventName: opt.ena,
      eventIndex: opt.eni ? opt.eni : 0
    };
  });
  const event = breakdown.filter(b => b.eventIndex);
  const global = breakdown.filter(b => !b.eventIndex);
  const result = {
    events,
    queryType,
    breakdown: {
      event,
      global
    }
  };
  return result;
};

export const DefaultDateRangeFormat = {
  from: '',
  to: '',
  frequency: moment().format('dddd') === 'Sunday' ? 'hour' : 'date'
};