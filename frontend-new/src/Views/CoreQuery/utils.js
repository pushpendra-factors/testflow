import React from 'react';
import MomentTz from 'Components/MomentTz';
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
  TOTAL_EVENTS_CRITERIA,
  TYPE_EVENTS_OCCURRENCE,
  TYPE_UNIQUE_USERS,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  constantObj,
  ANY_USER_TYPE,
  ALL_USER_TYPE,
  EACH_USER_TYPE,
  TOTAL_USERS_CRITERIA,
  CHART_TYPE_BARCHART,
  CHART_TYPE_LINECHART,
  CHART_TYPE_TABLE,
  CHART_TYPE_SPARKLINES,
  CHART_TYPE_STACKED_AREA,
  CHART_TYPE_STACKED_BAR,
  apiChartAnnotations,
  INITIAL_SESSION_ANALYTICS_SEQ,
  MARKETING_TOUCHPOINTS,
  PREDEFINED_DATES,
} from '../../utils/constants';
import { Radio } from 'antd';

export const labelsObj = {
  [TOTAL_EVENTS_CRITERIA]: 'Event Count',
  [TOTAL_USERS_CRITERIA]: 'User Count',
  [ACTIVE_USERS_CRITERIA]: 'User Count',
  [FREQUENCY_CRITERIA]: 'Count',
};

export const initialState = {
  loading: false,
  error: false,
  data: null,
  apiCallStatus: { required: true, message: null },
};

export const initialResultState = [1, 2, 3, 4].map(() => {
  return initialState;
});

const operatorMap = {
  '=': 'equals',
  '!=': 'notEqual',
  contains: 'contains',
  'does not contain': 'notContains',
  '<': 'lesserThan',
  '<=': 'lesserThanOrEqual',
  '>': 'greaterThan',
  '>=': 'greaterThanOrEqual',
};

const reverseOperatorMap = {
  equals: '=',
  notEqual: '!=',
  contains: 'contains',
  notContains: 'does not contain',
  lesserThan: '<',
  lesserThanOrEqual: '<=',
  greaterThan: '>',
  greaterThanOrEqual: '>=',
};

const getEventsWithProperties = (queries) => {
  const ewps = [];
  queries.forEach((ev) => {
    const filterProps = [];
    ev.filters.forEach((fil) => {
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            en: fil.props[2],
            lop: !index ? 'AND' : 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: val,
          });
        });
      } else {
        filterProps.push({
          en: fil.props[2],
          lop: 'AND',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.values,
        });
      }
    });
    ewps.push({
      na: ev.label,
      pr: filterProps,
    });
  });
  return ewps;
};

const getGlobalFilters = (globalFilters = []) => {
  const filterProps = [];
  globalFilters.forEach((fil) => {
    if (Array.isArray(fil.values)) {
      fil.values.forEach((val, index) => {
        filterProps.push({
          en: 'user_g',
          lop: !index ? 'AND' : 'OR',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: val,
        });
      });
    } else {
      filterProps.push({
        en: 'user_g',
        lop: 'AND',
        op: operatorMap[fil.operator],
        pr: fil.props[0],
        ty: fil.props[1],
        va: fil.values,
      });
    }
  });

  return filterProps;
};

export const getFunnelQuery = (
  groupBy,
  queries,
  session_analytics_seq,
  dateRange,
  globalFilters = []
) => {
  const query = {};
  query.cl = QUERY_TYPE_FUNNEL;
  query.ty = TYPE_UNIQUE_USERS;

  const period = {};
  if (dateRange.from && dateRange.to) {
    period.from = MomentTz(dateRange.from).startOf('day').utc().unix();
    period.to = MomentTz(dateRange.to).endOf('day').utc().unix();
  } else {
    period.from = MomentTz().startOf('week').utc().unix();
    period.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
  }

  query.fr = period.from;
  query.to = period.to;

  query.ewp = getEventsWithProperties(queries);
  query.gbt = '';

  const appliedGroupBy = [...groupBy.event, ...groupBy.global];
  query.gbp = appliedGroupBy.map((opt) => {
    let appGbp = {};
    if (opt.eventIndex) {
      appGbp = {
        pr: opt.property,
        en: opt.prop_category,
        pty: opt.prop_type,
        ena: opt.eventName,
        eni: opt.eventIndex,
      };
    } else {
      appGbp = {
        pr: opt.property,
        en: opt.prop_category,
        pty: opt.prop_type,
        ena: opt.eventName,
      };
    }
    if (opt.prop_type === 'datetime') {
      opt.grn ? (appGbp['grn'] = opt.grn) : (appGbp['grn'] = 'day');
    }
    if (opt.prop_type === 'numerical') {
      opt.gbty ? (appGbp['gbty'] = opt.gbty) : (appGbp['gbty'] = '');
    }
    return appGbp;
  });
  query.gup = getGlobalFilters(globalFilters);
  // if (session_analytics_seq.start && session_analytics_seq.end) {
  //   query.sse = session_analytics_seq.start;
  //   query.see = session_analytics_seq.end;
  // }
  query.ec = 'any_given_event';
  query.tz = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  return query;
};

export const getQuery = (
  groupBy,
  queries,
  result_criteria,
  user_type,
  dateRange,
  globalFilters = []
) => {
  const query = {};
  query.cl = QUERY_TYPE_EVENT;
  query.ty =
    result_criteria === TOTAL_EVENTS_CRITERIA ||
    result_criteria === FREQUENCY_CRITERIA
      ? TYPE_EVENTS_OCCURRENCE
      : TYPE_UNIQUE_USERS;

  const period = {};
  if (dateRange.from && dateRange.to) {
    period.from = MomentTz(dateRange.from).utc().unix();
    period.to = MomentTz(dateRange.to).utc().unix();
  } else {
    period.from = MomentTz().startOf('week').utc().unix();
    period.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').utc().unix()
        : MomentTz().utc().unix();
  }

  query.fr = period.from;
  query.to = period.to;

  query.ewp = getEventsWithProperties(queries);
  query.gup = getGlobalFilters(globalFilters);
  query.gbt = user_type === EACH_USER_TYPE ? dateRange.frequency : '';

  const appliedGroupBy = [...groupBy.event, ...groupBy.global];

  query.gbp = appliedGroupBy.map((opt) => {
    let gbpReq = {};
    if (opt.eventIndex) {
      gbpReq = {
        pr: opt.property,
        en: opt.prop_category,
        pty: opt.prop_type,
        ena: opt.eventName,
        eni: opt.eventIndex,
      };
    } else {
      gbpReq = {
        pr: opt.property,
        en: opt.prop_category,
        pty: opt.prop_type,
        ena: opt.eventName,
      };
    }
    if (opt.prop_type === 'datetime') {
      opt.grn ? (gbpReq['grn'] = opt.grn) : (gbpReq['grn'] = 'day');
    }
    if (opt.prop_type === 'numerical') {
      opt.gbty ? (gbpReq['gbty'] = opt.gbty) : (gbpReq['gbty'] = '');
    }
    return gbpReq;
  });
  query.ec = constantObj[user_type];
  query.tz = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  const sessionsQuery = {
    cl: QUERY_TYPE_EVENT,
    ty: TYPE_UNIQUE_USERS,
    fr: period.from,
    to: period.to,
    ewp: [
      {
        na: '$session',
        pr: [],
      },
    ],
    gup: [],
    gbt: '',
    ec: constantObj.each,
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
  };
  if (result_criteria === ACTIVE_USERS_CRITERIA) {
    return [query, { ...query, gbt: '' }, sessionsQuery];
  } else if (result_criteria === FREQUENCY_CRITERIA) {
    return [
      query,
      { ...query, gbt: '' },
      { ...query, ty: TYPE_UNIQUE_USERS },
      { ...query, ty: TYPE_UNIQUE_USERS, gbt: '' },
    ];
  }
  if (user_type === ANY_USER_TYPE || user_type === ALL_USER_TYPE) {
    return [query];
  }
  return [query, { ...query, gbt: '' }];
};

export const calculateFrequencyData = (
  eventData,
  userData,
  appliedBreakdown
) => {
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
    const idx = userData.metrics.rows.findIndex((r) => r[0] === elem[0]);
    if (!elem[2] || !userData.metrics.rows[idx][2]) return 0;
    const eVal = elem[2] / userData.metrics.rows[idx][2];
    return [elem[0], elem[1], parseFloat(eVal.toFixed(2))];
  });
  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics,
    },
  };
  return result;
};

const getEventIdx = (eventData, userObj) => {
  const str = userObj.slice(0, userObj.length - 1).join(',');
  const eventIdx = eventData.findIndex(
    (elem) => elem.slice(0, elem.length - 1).join(',') === str
  );
  return eventIdx;
};

export const calculateFrequencyDataForBreakdown = (eventData, userData) => {
  const rows = userData.rows.map((userObj) => {
    const eventIdx = getEventIdx(eventData.rows, userObj);
    let eventObj = null;
    if (eventIdx > -1) {
      eventObj = eventData.rows[eventIdx];
    }
    let eVal = 0;
    if (
      eventObj &&
      eventObj[eventObj.length - 1] &&
      userObj[userObj.length - 1]
    ) {
      eVal = eventObj[eventObj.length - 1] / userObj[userObj.length - 1];
      eVal = eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    }
    return [...userObj.slice(0, userObj.length - 1), eVal];
  });

  const metrics = userData.metrics.rows.map((userObj) => {
    const eventIdx = getEventIdx(eventData.metrics.rows, userObj);
    let eventObj = null;
    let eVal = 0;
    if (eventIdx > -1) {
      eventObj = eventData.metrics.rows[eventIdx];
    }
    if (
      eventObj &&
      userObj[userObj.length - 1] &&
      eventObj[eventObj.length - 1]
    ) {
      eVal = eventObj[eventObj.length - 1] / userObj[userObj.length - 1];
    }
    return [
      ...userObj.slice(0, userObj.length - 1),
      parseFloat(eVal.toFixed(2)),
    ];
  });

  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics,
    },
  };
  return result;
};

export const calculateActiveUsersData = (
  userData,
  sessionData,
  appliedBreakdown
) => {
  if (appliedBreakdown.length) {
    return calculateActiveUsersDataForBreakdown(userData, sessionData);
  } else {
    return calculateActiveUsersDataForNoBreakdown(userData, sessionData);
  }
};

const calculateActiveUsersDataForNoBreakdown = (userData, sessionData) => {
  const rows = userData.rows.map((elem) => {
    const eventVals = elem.slice(1).map((e) => {
      if (!e || !sessionData.rows[0][2]) return 0;
      const eVal = (e / sessionData.rows[0][2]) * 100;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [elem[0], ...eventVals];
  });

  const metrics = userData.metrics.rows.map((elem) => {
    if (!elem[2] || !sessionData.rows[0][2]) return 0;
    const eVal = (elem[2] / sessionData.rows[0][2]) * 100;
    return [elem[0], elem[1], parseFloat(eVal.toFixed(2))];
  });

  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics,
    },
  };

  return result;
};

const calculateActiveUsersDataForBreakdown = (userData, sessionData) => {
  const differentDates = new Set();
  userData.rows.forEach((ud) => {
    differentDates.add(ud[1]);
  });
  const rows = userData.rows.map((elem) => {
    const eventVals = elem.slice(elem.length - 1).map((e) => {
      if (!e || !sessionData.rows[0][2]) return e;
      const eVal = (e / sessionData.rows[0][2]) * 100;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [...elem.slice(0, elem.length - 1), ...eventVals];
  });

  const metrics = userData.metrics.rows.map((elem) => {
    if (!elem[elem.length - 1] || !sessionData.rows[0][2]) return 0;
    const eVal = (elem[elem.length - 1] / sessionData.rows[0][2]) * 100;
    return [...elem.slice(0, elem.length - 1), parseFloat(eVal.toFixed(2))];
  });
  const result = {
    ...userData,
    rows,
    metrics: {
      ...userData.metrics,
      rows: metrics,
    },
  };
  return result;
};

export const hasApiFailed = (res) => {
  if (
    res.data &&
    res.data.result_group &&
    res.data.result_group[0] &&
    res.data.result_group[0].headers &&
    res.data.result_group[0].headers.indexOf('error') > -1
  ) {
    return true;
  }
  return false;
};

export const numberWithCommas = (x) => {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
};

export const formatApiData = (data, metrics) => {
  if (data.headers[0] !== 'event_index') {
    const order = data.meta.metrics[0].rows.map((elem) => elem[1]);
    const rowData = data.rows;
    for (let i = 0; i < rowData.length; i++) {
      const originalOrder = rowData[i].slice(1);
      const newOrder = [];
      for (let j = 0; j < order.length; j++) {
        const idx = order.indexOf(j);
        newOrder.push(originalOrder[idx]);
      }
      rowData[i] = [rowData[i][0], ...newOrder];
    }
  }
  return { ...data, metrics };
};

export const getStateQueryFromRequestQuery = (requestQuery) => {
  const events = requestQuery.ewp.map((e) => {
    const filters = [];
    e.pr.forEach((pr) => {
      if (pr.lop === 'AND') {
        filters.push({
          operator: reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.en],
          values: [pr.va],
        });
      } else {
        filters[filters.length - 1].values.push(pr.va);
      }
    });
    return {
      label: e.na,
      filters,
    };
  });

  const globalFilters = [];

  if (requestQuery && requestQuery.gup && Array.isArray(requestQuery.gup)) {
    requestQuery.gup.forEach((pr) => {
      if (pr.lop === 'AND') {
        globalFilters.push({
          operator: reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.en],
          values: [pr.va],
        });
      } else {
        globalFilters[globalFilters.length - 1].values.push(pr.va);
      }
    });
  }

  const queryType = requestQuery.cl;
  const session_analytics_seq = INITIAL_SESSION_ANALYTICS_SEQ;
  // if (requestQuery.cl && requestQuery.cl === QUERY_TYPE_FUNNEL) {
  //   if (requestQuery.sse && requestQuery.see) {
  //     session_analytics_seq.start = requestQuery.sse;
  //     session_analytics_seq.end = requestQuery.see;
  //   }
  // }
  const breakdown = requestQuery.gbp.map((opt) => {
    return {
      property: opt.pr,
      prop_category: opt.en,
      prop_type: opt.pty,
      eventName: opt.ena,
      eventIndex: opt.eni ? opt.eni : 0,
      grn: opt.grn,
      gbty: opt.gbty,
    };
  });
  const event = breakdown
    .filter((b) => b.eventIndex)
    .map((b, index) => {
      return {
        ...b,
        overAllIndex: index,
      };
    });
  const global = breakdown
    .filter((b) => !b.eventIndex)
    .map((b, index) => {
      return {
        ...b,
        overAllIndex: index,
      };
    });
  const result = {
    events,
    queryType,
    session_analytics_seq,
    globalFilters,
    breakdown: {
      event,
      global,
    },
  };
  return result;
};

export const DefaultDateRangeFormat = {
  from:
    MomentTz().format('dddd') === 'Sunday'
      ? MomentTz().subtract(1, 'day').startOf('week')
      : MomentTz().startOf('week'),
  to:
    MomentTz().format('dddd') === 'Sunday'
      ? MomentTz().subtract(1, 'day').endOf('week')
      : MomentTz().subtract(1, 'day').endOf('day'),
  frequency: MomentTz().format('dddd') === 'Monday' ? 'hour' : 'date',
  dateType:
    MomentTz().format('dddd') === 'Sunday'
      ? PREDEFINED_DATES.LAST_WEEK
      : PREDEFINED_DATES.THIS_WEEK,
};

export const DashboardDefaultDateRangeFormat = {
  from: MomentTz().subtract(7, 'days').startOf('week'),
  to: MomentTz().subtract(7, 'days').endOf('week'),
  frequency: 'date',
  dateType: PREDEFINED_DATES.LAST_WEEK,
};

const getFilters = (filters) => {
  const result = [];
  filters.forEach((filter) => {
    if (filter.props[1] !== 'categorical') {
      result.push({
        en: filter.props[2],
        lop: 'AND',
        op: operatorMap[filter.operator],
        pr: filter.props[0],
        ty: filter.props[1],
        va: filter.values,
      });
    }

    if (filter.props[1] === 'categorical') {
      filter.values.forEach((value, index) => {
        result.push({
          en: filter.props[2],
          lop: !index ? 'AND' : 'OR',
          op: operatorMap[filter.operator],
          pr: filter.props[0],
          ty: filter.props[1],
          va: value,
        });
      });
    }
  });
  return result;
};

const getFiltersTouchpoints = (filters, touchpoint) => {
  const result = [];
  filters.forEach((filter) => {
    if (filter.props[1] !== 'categorical') {
      result.push({
        attribution_key: touchpoint,
        lop: 'AND',
        op: operatorMap[filter.operator],
        pr: filter.props[0],
        ty: filter.props[1],
        va: filter.values,
      });
    }

    if (filter.props[1] === 'categorical') {
      filter.values.forEach((value, index) => {
        result.push({
          attribution_key: touchpoint,
          lop: !index ? 'AND' : 'OR',
          op: operatorMap[filter.operator],
          pr: filter.props[0],
          ty: filter.props[1],
          va: value,
        });
      });
    }
  });
  return result;
};

export const getAttributionQuery = (
  eventGoal,
  touchpoint,
  attr_dimensions,
  touchpointFilters,
  queryType,
  models,
  window,
  linkedEvents,
  dateRange = {}
) => {
  const eventFilters = getFilters(eventGoal.filters);
  let touchPointFiltersQuery = [];
  if (touchpointFilters.length) {
    touchPointFiltersQuery = getFiltersTouchpoints(
      touchpointFilters,
      touchpoint
    );
  }

  const query = {
    cl: QUERY_TYPE_ATTRIBUTION,
    meta: {
      metrics_breakdown: true,
    },
    query: {
      cm: ['Impressions', 'Clicks', 'Spend'],
      ce: {
        na: eventGoal.label,
        pr: eventFilters,
      },
      attribution_key: touchpoint,
      attribution_key_f: touchPointFiltersQuery,
      query_type: queryType,
      attribution_methodology: models[0],
      lbw: window,
    },
  };
  if (dateRange.from && dateRange.to) {
    query.query.from = MomentTz(dateRange.from).startOf('day').utc().unix();
    query.query.to = MomentTz(dateRange.to).endOf('day').utc().unix();
  } else {
    query.query.from = MomentTz().startOf('week').utc().unix();
    query.query.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
  }
  if (models[1]) {
    query.query.attribution_methodology_c = models[1];
  }
  if (linkedEvents.length) {
    query.query.lfe = linkedEvents.map((le) => {
      const linkedEventFilters = getFilters(le.filters);
      return {
        na: le.label,
        pr: linkedEventFilters,
      };
    });
  }
  const attribution_key_dimensions = attr_dimensions
    .filter((d) => d.touchPoint === touchpoint && d.enabled && d.type === 'key')
    .map((d) => d.header);
  const attribution_key_custom_dimensions = attr_dimensions
    .filter(
      (d) => d.touchPoint === touchpoint && d.enabled && d.type === 'custom'
    )
    .map((d) => d.header);

  if (touchpoint !== MARKETING_TOUCHPOINTS.SOURCE) {
    query.query.attribution_key_dimensions = attribution_key_dimensions;
    query.query.attribution_key_custom_dimensions = attribution_key_custom_dimensions;
  }

  return query;
};

export const getAttributionStateFromRequestQuery = (
  requestQuery,
  initial_attr_dimensions
) => {
  const filters = [];
  requestQuery.ce.pr.forEach((pr) => {
    if (pr.lop === 'AND') {
      let val = pr.ty === 'categorical' ? [pr.va] : pr.va;
      filters.push({
        operator: reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: val,
      });
    } else if (pr.ty === 'categorical') {
      filters[filters.length - 1].values.push(pr.va);
    }
  });

  const touchPointFilters = [];
  if (requestQuery.attribution_key_f) {
    requestQuery.attribution_key_f.forEach((pr) => {
      if (pr.lop === 'AND') {
        let val = pr.ty === 'categorical' ? [pr.va] : pr.va;
        touchPointFilters.push({
          operator: reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.attribution_key],
          values: val,
        });
      } else if (pr.ty === 'categorical') {
        touchPointFilters[touchPointFilters.length - 1].values.push(pr.va);
      }
    });
  }

  const touchpoint = requestQuery.attribution_key;
  const attr_dimensions = initial_attr_dimensions.map((dimension) => {
    if (dimension.touchPoint === touchpoint) {
      return {
        ...dimension,
        enabled: !requestQuery.attribution_key_dimensions
          ? dimension.defaultValue
          : requestQuery.attribution_key_dimensions?.indexOf(dimension.header) >
              -1 ||
            requestQuery.attribution_key_custom_dimensions?.indexOf(
              dimension.header
            ) > -1,
      };
    }
    return dimension;
  });

  const result = {
    queryType: QUERY_TYPE_ATTRIBUTION,
    eventGoal: {
      label: requestQuery.ce.na,
      filters,
    },
    touchpoint_filters: touchPointFilters,
    attr_query_type: requestQuery.query_type,
    touchpoint,
    attr_dimensions,
    models: [requestQuery.attribution_methodology],
    window: requestQuery.lbw,
  };

  if (requestQuery.attribution_methodology_c) {
    result.models.push(requestQuery.attribution_methodology_c);
  }

  if (requestQuery.lfe && requestQuery.lfe.length) {
    result['linkedEvents'] = requestQuery.lfe.map((le) => {
      const linkedFilters = [];
      le.pr.forEach((pr) => {
        if (pr.lop === 'AND') {
          let val = pr.ty === 'categorical' ? [pr.va] : pr.va;
          linkedFilters.push({
            operator: reverseOperatorMap[pr.op],
            props: [pr.pr, pr.ty, pr.en],
            values: val,
          });
        } else if (pr.ty === 'categorical') {
          linkedFilters[linkedFilters.length - 1].values.push(pr.va);
        }
      });
      return {
        label: le.na,
        filters: linkedFilters,
      };
    });
  } else {
    result['linkedEvents'] = [];
  }
  return result;
};

export const getCampaignsQuery = (
  channel,
  select_metrics,
  filters,
  group_by,
  dateRange = {}
) => {
  const appliedFilters = [];

  filters.forEach((filter) => {
    filter.values.forEach((value, index) => {
      appliedFilters.push({
        name: filter.props[2],
        property: filter.props[0],
        condition: operatorMap[filter.operator],
        logical_operator: !index ? 'AND' : 'OR',
        value,
      });
    });
  });

  const query = {
    channel,
    select_metrics,
    group_by: group_by.map((elem) => {
      return {
        name: elem.prop_category,
        property: elem.property,
      };
    }),
    filters: appliedFilters,
    gbt: dateRange.frequency,
  };
  if (dateRange.from && dateRange.to) {
    query.fr = MomentTz(dateRange.from).startOf('day').utc().unix();
    query.to = MomentTz(dateRange.to).endOf('day').utc().unix();
  } else {
    query.fr = MomentTz().startOf('week').utc().unix();
    query.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
  }
  return {
    query_group: [query, { ...query, gbt: '' }],
    cl: QUERY_TYPE_CAMPAIGN,
  };
};

export const getCampaignStateFromRequestQuery = (requestQuery) => {
  const camp_filters = [];
  requestQuery.filters.forEach((filter) => {
    if (filter.logical_operator === 'AND') {
      camp_filters.push({
        operator: reverseOperatorMap[filter.condition],
        props: [filter.property, '', filter.name],
        values: [filter.value],
      });
    } else {
      camp_filters[camp_filters.length - 1].values.push(filter.value);
    }
  });
  const result = {
    queryType: QUERY_TYPE_CAMPAIGN,
    camp_channels: requestQuery.channel,
    camp_measures: requestQuery.select_metrics,
    camp_filters,
    camp_groupBy: requestQuery.group_by.map((gb) => {
      return {
        prop_category: gb.name,
        property: gb.property,
      };
    }),
  };

  return result;
};

export const getSaveChartOptions = (queryType, requestQuery) => {
  if (queryType === QUERY_TYPE_FUNNEL) {
    return (
      <>
        <Radio value={apiChartAnnotations[CHART_TYPE_BARCHART]}>
          Display Funnel Chart
        </Radio>
        <Radio value={apiChartAnnotations[CHART_TYPE_TABLE]}>
          Display Table
        </Radio>
      </>
    );
  }
  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    return (
      <>
        <Radio value={apiChartAnnotations[CHART_TYPE_BARCHART]}>
          Display Bar Chart
        </Radio>
        <Radio value={apiChartAnnotations[CHART_TYPE_TABLE]}>
          Display Table
        </Radio>
      </>
    );
  }
  if (queryType === QUERY_TYPE_CAMPAIGN) {
    const commons = (
      <>
        <Radio value={apiChartAnnotations[CHART_TYPE_LINECHART]}>
          Display Line Chart
        </Radio>
        <Radio value={apiChartAnnotations[CHART_TYPE_TABLE]}>
          Display Table
        </Radio>
      </>
    );
    if (!requestQuery.query_group[0].group_by.length) {
      return (
        <>
          <Radio value={apiChartAnnotations[CHART_TYPE_SPARKLINES]}>
            Display Spark Line Chart
          </Radio>
          {commons}
        </>
      );
    } else {
      return (
        <>
          <Radio value={apiChartAnnotations[CHART_TYPE_BARCHART]}>
            Display Bar Chart
          </Radio>
          <Radio value={apiChartAnnotations[CHART_TYPE_STACKED_AREA]}>
            Display Stacked Area Chart
          </Radio>
          <Radio value={apiChartAnnotations[CHART_TYPE_STACKED_BAR]}>
            Display Stacked Bar Chart
          </Radio>
          {commons}
        </>
      );
    }
  }

  if (queryType === QUERY_TYPE_EVENT) {
    if (requestQuery[0].ec === constantObj[EACH_USER_TYPE]) {
      const commons = (
        <>
          <Radio value={apiChartAnnotations[CHART_TYPE_LINECHART]}>
            Display Line Chart
          </Radio>
          <Radio value={apiChartAnnotations[CHART_TYPE_TABLE]}>
            Display Table
          </Radio>
        </>
      );
      if (!requestQuery[0].gbp.length) {
        return (
          <>
            <Radio value={apiChartAnnotations[CHART_TYPE_SPARKLINES]}>
              Display Spark Line Chart
            </Radio>
            {commons}
          </>
        );
      } else {
        return (
          <>
            <Radio value={apiChartAnnotations[CHART_TYPE_BARCHART]}>
              Display Bar Chart
            </Radio>
            <Radio value={apiChartAnnotations[CHART_TYPE_STACKED_AREA]}>
              Display Stacked Area Chart
            </Radio>
            <Radio value={apiChartAnnotations[CHART_TYPE_STACKED_BAR]}>
              Display Stacked Bar Chart
            </Radio>
            {commons}
          </>
        );
      }
    } else {
      const commons = (
        <>
          <Radio value={apiChartAnnotations[CHART_TYPE_TABLE]}>
            Display Table
          </Radio>
        </>
      );
      if (!requestQuery[0].gbp.length) {
        return (
          <>
            <Radio value='pc'>Display Count</Radio>
            {commons}
          </>
        );
      } else {
        return (
          <>
            <Radio value='pb'>Display Bar Chart</Radio>
            {commons}
          </>
        );
      }
    }
  }
};

export const isComparisonEnabled = (queryType, events, groupBy, models) => {
  if (queryType === QUERY_TYPE_FUNNEL) {
    const newAppliedBreakdown = [...groupBy.event, ...groupBy.global];
    return newAppliedBreakdown.length === 0;
  }
  // if (queryType === QUERY_TYPE_EVENT) {
  //   const newAppliedBreakdown = [...groupBy.event, ...groupBy.global];
  //   return !(events.length > 1 && newAppliedBreakdown.length > 0);
  // }
  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    if (models.length === 1) {
      return true;
    }
  }
  return false;
};