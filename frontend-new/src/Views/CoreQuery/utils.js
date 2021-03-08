import moment from "moment";
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
} from "../../utils/constants";

export const labelsObj = {
  [TOTAL_EVENTS_CRITERIA]: "Event Count",
  [TOTAL_USERS_CRITERIA]: "User Count",
  [ACTIVE_USERS_CRITERIA]: "User Count",
  [FREQUENCY_CRITERIA]: "Count",
};

export const presentationObj = {
  pb: "barchart",
  pl: "linechart",
  pt: "table",
  pc: "sparklines",
};

export const initialState = { loading: false, error: false, data: null };

export const initialResultState = [1, 2, 3, 4].map(() => {
  return initialState;
});

const operatorMap = {
  "=": "equals",
  "!=": "notEqual",
  contains: "contains",
  "not contains": "notContains",
  "<": "lesserThan",
  "<=": "lesserThanOrEqual",
  ">": "greaterThan",
  ">=": "greaterThanOrEqual",
};

const reverseOperatorMap = {
  equals: "=",
  notEqual: "!=",
  contains: "contains",
  notContains: "not contains",
  lesserThan: "<",
  lesserThanOrEqual: "<=",
  greaterThan: ">",
  greaterThanOrEqual: ">=",
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
            lop: !index ? "AND" : "OR",
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: val,
          });
        });
      } else {
        filterProps.push({
          en: fil.props[2],
          lop: "AND",
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

export const getFunnelQuery = (groupBy, queries, dateRange) => {
  const query = {};
  query.cl = QUERY_TYPE_FUNNEL;
  query.ty = TYPE_UNIQUE_USERS;

  const period = {};
  if (dateRange.from && dateRange.to) {
    period.from = moment(dateRange.from).startOf("day").utc().unix();
    period.to = moment(dateRange.to).endOf("day").utc().unix();
  } else {
    period.from = moment().startOf("week").utc().unix();
    period.to =
      moment().format("dddd") !== "Sunday"
        ? moment().subtract(1, "day").endOf("day").utc().unix()
        : moment().utc().unix();
  }

  query.fr = period.from;
  query.to = period.to;

  query.ewp = getEventsWithProperties(queries);
  query.gbt = "";

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
    if (opt.prop_type === "datetime") {
      opt.grn ? (appGbp["grn"] = opt.grn) : (appGbp["grn"] = "day");
    }
    if (opt.prop_type === "numerical") {
      opt.gbty ? (appGbp["gbty"] = opt.gbty) : (appGbp["gbty"] = "");
    }
    return appGbp;
  });
  query.ec = "any_given_event";
  query.tz = "Asia/Kolkata";
  return query;
};

export const getQuery = (
  groupBy,
  queries,
  result_criteria,
  user_type,
  dateRange
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
    period.from = moment(dateRange.from).startOf("day").utc().unix();
    period.to = moment(dateRange.to).endOf("day").utc().unix();
  } else {
    period.from = moment().startOf("week").utc().unix();
    period.to =
      moment().format("dddd") !== "Sunday"
        ? moment().subtract(1, "day").endOf("day").utc().unix()
        : moment().utc().unix();
  }

  query.fr = period.from;
  query.to = period.to;

  query.ewp = getEventsWithProperties(queries);
  query.gbt = user_type === EACH_USER_TYPE ? dateRange.frequency : "";

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
    if (opt.prop_type === "datetime") {
      opt.grn ? (gbpReq["grn"] = opt.grn) : (gbpReq["grn"] = "day");
    }
    if (opt.prop_type === "numerical") {
      opt.gbty ? (gbpReq["gbty"] = opt.gbty) : (gbpReq["gbty"] = "");
    }
    return gbpReq;
  });
  query.ec = constantObj[user_type];
  query.tz = "Asia/Kolkata";
  const sessionsQuery = {
    cl: QUERY_TYPE_EVENT,
    ty: TYPE_UNIQUE_USERS,
    fr: period.from,
    to: period.to,
    ewp: [
      {
        na: "$session",
        pr: [],
      },
    ],
    gbt: "",
    ec: constantObj.each,
    tz: "Asia/Kolkata",
  };
  if (result_criteria === ACTIVE_USERS_CRITERIA) {
    return [query, { ...query, gbt: "" }, sessionsQuery];
  } else if (result_criteria === FREQUENCY_CRITERIA) {
    return [
      query,
      { ...query, gbt: "" },
      { ...query, ty: TYPE_UNIQUE_USERS },
      { ...query, ty: TYPE_UNIQUE_USERS, gbt: "" },
    ];
  }
  if (user_type === ANY_USER_TYPE || user_type === ALL_USER_TYPE) {
    return [query];
  }
  return [query, { ...query, gbt: "" }];
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
  const str = userObj.slice(0, userObj.length - 1).join(",");
  const eventIdx = eventData.findIndex(
    (elem) => elem.slice(0, elem.length - 1).join(",") === str
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
    res.data.result_group[0].headers.indexOf("error") > -1
  ) {
    return true;
  }
  return false;
};

export const numberWithCommas = (x) => {
  return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ",");
};

export const formatApiData = (data, metrics) => {
  if (data.headers[0] !== "event_index") {
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
      if (pr.lop === "AND") {
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
  const queryType = requestQuery.cl;
  const breakdown = requestQuery.gbp.map((opt) => {
    return {
      property: opt.pr,
      prop_category: opt.en,
      prop_type: opt.pty,
      eventName: opt.ena,
      eventIndex: opt.eni ? opt.eni : 0,
    };
  });
  const event = breakdown.filter((b) => b.eventIndex);
  const global = breakdown.filter((b) => !b.eventIndex);
  const result = {
    events,
    queryType,
    breakdown: {
      event,
      global,
    },
  };
  return result;
};

export const DefaultDateRangeFormat = {
  from:
    moment().format("dddd") === "Sunday" || moment().format("dddd") === "Monday"
      ? moment().subtract(3, "days").startOf("week")
      : moment().startOf("week"),
  to:
    moment().format("dddd") === "Sunday" || moment().format("dddd") === "Monday"
      ? moment().subtract(3, "days").endOf("week")
      : moment().subtract(1, "day"),
  frequency: "date",
};

const getFilters = (filters) => {
  const result = [];
  filters.forEach((filter) => {
    if(filter.props[1] !== 'categorical') {
      result.push({
        en: filter.props[2],
        lop: "AND",
        op: operatorMap[filter.operator],
        pr: filter.props[0],
        ty: filter.props[1],
        va: filter.values,
      });
    }

    if(filter.props[1] === 'categorical') {
      filter.values.forEach((value, index) => {
        result.push({
          en: filter.props[2],
          lop: !index ? "AND" : "OR",
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
    if(filter.props[1] !== 'categorical') {
      result.push({
        attribution_key: touchpoint,
        lop: "AND",
        op: operatorMap[filter.operator],
        pr: filter.props[0],
        ty: filter.props[1],
        va: filter.values,
      });
    }

    if(filter.props[1] === 'categorical') {
      filter.values.forEach((value, index) => {
        result.push({
          attribution_key: touchpoint,
          lop: !index ? "AND" : "OR",
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
  touchpointFilters,
  queryType,
  models,
  window,
  linkedEvents,
  dateRange = {}
) => {

  //attribution_key_f Filters
  //query_type conv_time, interact_time [ConversionBased,EngagementBased];

  const eventFilters = getFilters(eventGoal.filters);
  let touchPointFiltersQuery = [];
  if(touchpointFilters.length) {
    touchPointFiltersQuery = getFiltersTouchpoints(touchpointFilters, touchpoint);
  }
  
  const query = {
    cl: QUERY_TYPE_ATTRIBUTION,
    meta: {
      metrics_breakdown: true,
    },
    query: {
      cm: ["Impressions", "Clicks", "Spend"],
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
    query.query.from = moment(dateRange.from).startOf("day").utc().unix();
    query.query.to = moment(dateRange.to).endOf("day").utc().unix();
  } else {
    query.query.from = moment().startOf("week").utc().unix();
    query.query.to =
      moment().format("dddd") !== "Sunday"
        ? moment().subtract(1, "day").endOf("day").utc().unix()
        : moment().utc().unix();
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
  return query;
};

export const getAttributionStateFromRequestQuery = (requestQuery) => {
  const filters = [];
  requestQuery.ce.pr.forEach((pr) => {
    if (pr.lop === "AND") {
      filters.push({
        operator: reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: [pr.va],
      });
    } else {
      filters[filters.length - 1].values.push(pr.va);
    }
  });

  const touchPointFilters = [];
  if (requestQuery.attribution_key_f) {
    requestQuery.attribution_key_f.forEach((pr) => {
      touchPointFilters.push({
        operator: reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: [pr.va],
      });
    })
  }
  

  const result = {
    queryType: QUERY_TYPE_ATTRIBUTION,
    eventGoal: {
      label: requestQuery.ce.na,
      filters,
    },
    touchpoint_filters: touchPointFilters,
    attr_query_type: requestQuery.query_type,
    touchpoint: requestQuery.attribution_key,
    models: [requestQuery.attribution_methodology],
    window: requestQuery.lbw,
  };

  if (requestQuery.attribution_methodology_c) {
    result.models.push(requestQuery.attribution_methodology_c);
  }

  if (requestQuery.lfe && requestQuery.lfe.length) {
    result["linkedEvents"] = requestQuery.lfe.map((le) => {
      const linkedFilters = [];
      le.pr.forEach((pr) => {
        if (pr.lop === "AND") {
          linkedFilters.push({
            operator: reverseOperatorMap[pr.op],
            props: [pr.pr, pr.ty, pr.en],
            values: [pr.va],
          });
        } else {
          linkedFilters[linkedFilters.length - 1].values.push(pr.va);
        }
      });
      return {
        label: le.na,
        filters: linkedFilters,
      };
    });
  } else {
    result["linkedEvents"] = [];
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
        logical_operator: !index ? "AND" : "OR",
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
    gbt: "date",
  };
  if (dateRange.from && dateRange.to) {
    query.fr = moment(dateRange.from).startOf("day").utc().unix();
    query.to = moment(dateRange.to).endOf("day").utc().unix();
  } else {
    query.fr = moment().startOf("week").utc().unix();
    query.to =
      moment().format("dddd") !== "Sunday"
        ? moment().subtract(1, "day").endOf("day").utc().unix()
        : moment().utc().unix();
  }
  return {
    query_group: [query, { ...query, gbt: "" }],
    cl: QUERY_TYPE_CAMPAIGN,
  };
};

export const getCampaignStateFromRequestQuery = (requestQuery) => {
  const camp_filters = [];
  requestQuery.filters.forEach((filter) => {
    if (filter.logical_operator === "AND") {
      camp_filters.push({
        operator: reverseOperatorMap[filter.condition],
        props: [filter.property, "", filter.name],
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
