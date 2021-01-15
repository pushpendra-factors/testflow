import moment from "moment";
import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_CAMPAIGN,
} from "../../utils/constants";

export const labelsObj = {
  totalEvents: "Event Count",
  totalUsers: "User Count",
  activeUsers: "User Count",
  frequency: "Count",
};

export const presentationObj = {
  pb: "barchart",
  pl: "linechart",
  pt: "table",
  pc: "sparklines",
};

const constantObj = {
  each: "each_given_event",
  any: "any_given_event",
  all: "all_given_event",
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
  query.ty = "unique_users";

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
    if (opt.eventIndex) {
      return {
        pr: opt.property,
        en: opt.prop_category,
        pty: opt.prop_type,
        ena: opt.eventName,
        eni: opt.eventIndex,
      };
    } else {
      return {
        pr: opt.property,
        en: opt.prop_category,
        pty: opt.prop_type,
        ena: opt.eventName,
      };
    }
  });
  query.ec = "any_given_event";
  query.tz = "Asia/Kolkata";
  return query;
};

export const getQuery = (
  activeTab,
  groupBy,
  queries,
  breakdownType = "each",
  dateRange
) => {
  const query = {};
  query.cl = QUERY_TYPE_EVENT;
  query.ty = parseInt(activeTab) === 0 ? "events_occurrence" : "unique_users";

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

  if (activeTab === "2") {
    query.ewp = [
      {
        na: "$session",
        pr: [],
      },
    ];
    query.gbt = "";
  } else {
    query.ewp = getEventsWithProperties(queries);
    query.gbt = breakdownType === "each" ? dateRange.frequency || "date" : "";

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
        gbpReq["grn"] = "day";
      }

      return gbpReq;
    });
  }
  query.ec = activeTab === "2" ? constantObj.each : constantObj[breakdownType];
  query.tz = "Asia/Kolkata";
  if (breakdownType === "each") {
    if (activeTab === "2") {
      return [query];
    } else {
      return [query, { ...query, gbt: "" }];
    }
  } else {
    return [query];
  }
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
  from: moment().startOf('week'),
  to: moment(),
  frequency:
    moment().format("dddd") === "Sunday" || moment().format("dddd") === "Monday"
      ? "hour"
      : "date",
};

const getFilters = (filters) => {
  const result = [];
  filters.forEach((filter) => {
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
  });
  return result;
};

const getFiltersState = (appliedFilter) => {
  const diffProps = [];
  appliedFilter.forEach((filter) => {
    const doesExist = diffProps.findIndex(
      (elem) => elem.op === filter.op && elem.pr === filter.pr
    );
    if (doesExist === -1) {
      diffProps.push({
        op: filter.op,
        pr: filter.pr,
      });
    }
  });

  const result = diffProps.map((elem) => {
    const propFilters = appliedFilter.filter(
      (filter) => filter.pr === elem.pr && filter.op === elem.op
    );
    const values = propFilters.map((filter) => filter.va);
    return {
      values,
      operator: reverseOperatorMap[elem.op],
      props: [elem.pr, propFilters[0].ty, propFilters[0].en],
    };
  });
  return result;
};

export const getAttributionQuery = (
  eventGoal,
  touchpoint,
  models,
  window,
  linkedEvents,
  dateRange = {}
) => {
  const eventFilters = getFilters(eventGoal.filters);
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
  const result = {
    queryType: QUERY_TYPE_ATTRIBUTION,
    eventGoal: {
      label: requestQuery.ce.na,
      filters: getFiltersState(requestQuery.ce.pr),
    },
    touchpoint: requestQuery.attribution_key,
    models: [requestQuery.attribution_methodology],
    window: requestQuery.lbw,
  };

  if (requestQuery.attribution_methodology_c) {
    result.models.push(requestQuery.attribution_methodology_c);
  }

  if (requestQuery.lfe && requestQuery.lfe.length) {
    result["linkedEvents"] = requestQuery.lfe.map((le) => {
      return {
        label: le.na,
        filters: getFiltersState(le.pr),
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
  const query = {
    channel,
    select_metrics,
    group_by: group_by.map((elem) => {
      return {
        name: elem.prop_category,
        property: elem.property,
      };
    }),
    filters,
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
  const result = {
    queryType: QUERY_TYPE_CAMPAIGN,
    camp_channels: requestQuery.channel,
    camp_measures: requestQuery.select_metrics,
    camp_groupBy: requestQuery.group_by.map((gb) => {
      return {
        prop_category: gb.name,
        property: gb.property,
      };
    }),
  };

  return result;
};

export const getSessionsQuery = (query) => {
  const user = query.query_group.map((elem) => {
    return {
      ...elem,
      ty: "unique_users",
    };
  });
  const session = [{
    cl: "events",
    ec: "each_given_event",
    ewp: [
      {
        na: "$session",
        pr: [],
      },
    ],
    fr: moment().startOf("week").utc().unix(),
    to: moment().utc().unix(),
    gbt: "",
    ty: "unique_users",
    tz: "Asia/Kolkata",
  }];
  return [...user, ...session];
};

export const getFrequencyQuery = (query) => {
  const event = query.query_group.map((elem) => {
    return {
      ...elem,
      ty: "events_occurrence",
    };
  });
  const user = query.query_group.map((elem) => {
    return {
      ...elem,
      ty: "unique_users",
    };
  });
  return [...event, ...user];
};
