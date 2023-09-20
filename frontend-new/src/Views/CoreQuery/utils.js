import get from 'lodash/get';
import lowerCase from 'lodash/lowerCase';
import startCase from 'lodash/startCase';

import { EMPTY_ARRAY, generateRandomKey, groupFilters } from 'Utils/global';
import { formatFilterDate, isDateInMilliSeconds } from 'Utils/dataFormatter';
import MomentTz from 'Components/MomentTz';
import _ from 'lodash';

import { AttributionQueryV1 } from 'Attribution/state/classes';

import {
  QUERY_TYPE_FUNNEL,
  QUERY_TYPE_EVENT,
  QUERY_TYPE_ATTRIBUTION,
  QUERY_TYPE_KPI,
  TOTAL_EVENTS_CRITERIA,
  TYPE_EVENTS_OCCURRENCE,
  TYPE_UNIQUE_USERS,
  ACTIVE_USERS_CRITERIA,
  FREQUENCY_CRITERIA,
  EVENT_QUERY_USER_TYPE,
  ANY_USER_TYPE,
  ALL_USER_TYPE,
  EACH_USER_TYPE,
  TOTAL_USERS_CRITERIA,
  INITIAL_SESSION_ANALYTICS_SEQ,
  MARKETING_TOUCHPOINTS,
  PREDEFINED_DATES,
  QUERY_TYPE_PROFILE,
  QUERY_OPTIONS_DEFAULT_VALUE
} from 'Utils/constants';
import {
  CORE_QUERY_INITIAL_STATE,
  FILTER_TYPES,
  INITIAL_STATE
} from './constants';

import {
  operatorMap,
  reverseOperatorMap,
  reverseDateOperatorMap
} from 'Utils/operatorMapping';

export const initialState = INITIAL_STATE;

export const labelsObj = {
  [TOTAL_EVENTS_CRITERIA]: 'Event Count',
  [TOTAL_USERS_CRITERIA]: 'User Count',
  [ACTIVE_USERS_CRITERIA]: 'User Count',
  [FREQUENCY_CRITERIA]: 'Count'
};

export const formatFiltersForQuery = (filters, scope = 'event') => {
  const formattedFilters = [];
  const groupByRef = {};
  let count = 0;
  filters.forEach((filter) => {
    let { ref } = filter;
    if (!ref) {
      ref = count++;
    }
    if (!groupByRef[ref]) {
      groupByRef[ref] = [];
    }
    groupByRef[ref].push(filter);
  });

  for (const ref in groupByRef) {
    const filterGroup = groupByRef[ref];
    filterGroup.forEach((filter, i) => {
      const entityInit = filter.props[3] === 'group' ? 'user' : filter.props[3];
      const entity = scope === 'global' && entityInit ? 'user_g' : entityInit;
      const values = Array.isArray(filter.values)
        ? filter.values
        : [filter.values];
      const operator = operatorMap[filter.operator];
      values.forEach((value, j) => {
        const valueLop = i === 0 && j === 0 ? 'AND' : 'OR';
        const filterStruct = {
          en: entity,
          grpn: filter.props[0],
          lop: valueLop,
          op: operator,
          pr: filter.props[1],
          ty: filter.props[2],
          va: value
        };
        if (values.length > 1) {
          filterStruct.lop = valueLop;
        }
        formattedFilters.push(filterStruct);
      });
    });
  }
  return formattedFilters;
};

export const formatBreakdown = (opt) => {
  const breakdown = {
    pr: opt.property,
    en: opt.prop_category === 'group' ? 'user' : opt.prop_category,
    pty: opt.prop_type,
    grpn: opt.groupName
  };
  if (opt.eventName) {
    breakdown.ena = opt.eventName;
  }
  if (opt.eventIndex) {
    breakdown.eni = opt.eventIndex;
  }

  if (opt.prop_type === 'datetime') {
    breakdown.grn = opt.grn || 'day';
  }

  if (opt.prop_type === 'numerical') {
    breakdown.gbty = opt.gbty || '';
  }

  return breakdown;
};

export const formatBreakdownsForQuery = (breakdownArr) =>
  breakdownArr.map((breakdown) => formatBreakdown(breakdown));

export const getEventsWithProperties = (queries) => {
  return queries.map((ev) => {
    const filterProps = formatFiltersForQuery(ev.filters);
    return {
      an: ev.alias,
      na: ev.label,
      grpa: ev.group,
      pr: filterProps
    };
  });
};

export const getQuery = (
  groupBy,
  queries,
  resultCriteria,
  userType,
  dateRange,
  globalFilters = [],
  groupAnalysis
) => {
  const query = {
    cl: QUERY_TYPE_EVENT,
    ty:
      resultCriteria === TOTAL_EVENTS_CRITERIA
        ? TYPE_EVENTS_OCCURRENCE
        : TYPE_UNIQUE_USERS,
    grpa: groupAnalysis,
    ewp: getEventsWithProperties(queries),
    gup: formatFiltersForQuery(globalFilters, 'global'),
    gbt: userType === EACH_USER_TYPE ? dateRange.frequency : '',
    gbp: formatBreakdownsForQuery([...groupBy.event, ...groupBy.global]),
    ec: EVENT_QUERY_USER_TYPE[userType],
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata'
  };

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

  if (userType === ANY_USER_TYPE || userType === ALL_USER_TYPE) {
    return [query];
  }

  return [query, { ...query, gbt: '' }];
};

export const getFunnelQuery = (
  groupBy,
  queries,
  session_analytics_seq,
  dateRange,
  globalFilters = [],
  eventsCondition,
  groupAnalysis,
  conversionDurationNumber,
  conversionDurationUnit
) => {
  const query = {
    cl: QUERY_TYPE_FUNNEL,
    ty: TYPE_UNIQUE_USERS,
    ec: eventsCondition,
    grpa: groupAnalysis,
    fr: dateRange.from
      ? MomentTz(dateRange.from).startOf('day').utc().unix()
      : MomentTz().startOf('week').utc().unix(),
    to: dateRange.to
      ? MomentTz(dateRange.to).endOf('day').utc().unix()
      : MomentTz().format('dddd') !== 'Sunday'
      ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
      : MomentTz().utc().unix(),
    ewp: getEventsWithProperties(queries),
    gbt: dateRange.frequency,
    cnvtm:
      conversionDurationNumber != null && conversionDurationUnit != null
        ? conversionDurationNumber + conversionDurationUnit
        : undefined,
    gbp: formatBreakdownsForQuery([...groupBy.event, ...groupBy.global]),
    gup: formatFiltersForQuery(globalFilters, 'global'),
    // if (session_analytics_seq.start && session_analytics_seq.end) {
    //   query.sse = session_analytics_seq.start;
    //   query.see = session_analytics_seq.end;
    // }
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata'
  };

  return query;
};

const getProfileWithProperties = (queries) => {
  return queries.map((ev) => {
    const filterProps = formatFiltersForQuery(ev.filters);
    return {
      an: ev.alias,
      ty: ev.label,
      pr: filterProps
    };
  });
};

export const getProfileQuery = (
  queries,
  groupBy,
  globalFilters = [],
  dateRange,
  groupAnalysis
) => {
  const query = {
    cl: QUERY_TYPE_PROFILE,
    grpa: groupAnalysis,
    queries: getProfileWithProperties(queries),
    gup: formatFiltersForQuery(globalFilters, 'global'),
    from: MomentTz(dateRange.from).utc().unix(),
    to: MomentTz().utc().unix(),
    gbp: formatBreakdownsForQuery([...groupBy.event, ...groupBy.global]),
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata'
  };

  return query;
};

export const getEventsWithPropertiesCustomKPI = (filters, category) => {
  const filterProps = [];
  // adding fil?.extra ? fil?.extra[*] check as a hotfix for timestamp filters
  const filtersGroupedByRef = Object.values(groupFilters(filters, 'ref'));
  filtersGroupedByRef.forEach((filtersGr) => {
    if (filtersGr.length === 1) {
      const fil = filtersGr[0];
      if (fil.props.length === 4) {
        fil.props.shift();
      }
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : fil?.props?.[2],
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : fil?.props?.[2]
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : fil?.props?.[2],
          objTy:
            category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : fil?.props?.[2]
              : '',
          va:
            fil.props[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values
        });
      }
    } else {
      let fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : fil?.props?.[2],
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : fil?.props?.[2]
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : fil?.props?.[2],
          objTy:
            category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : fil?.props?.[2]
              : '',
          va:
            fil.props[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values
        });
      }
      fil = filtersGr[1];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val) => {
          filterProps.push({
            prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3]
                : fil?.props?.[2],
            objTy:
              category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : fil?.props?.[2]
                : '',
            va: fil.props[1] === 'datetime' ? formatFilterDate(val) : val
          });
        });
      } else {
        filterProps.push({
          prNa: fil?.extra ? fil?.extra[1] : fil?.props[0],
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'OR',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : fil?.props?.[2],
          objTy:
            category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : fil?.props?.[2]
              : '',
          va:
            fil.props[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values
        });
      }
    }
  });
  return filterProps;
};

export const getEventsWithPropertiesKPI = (filters, category) => {
  const filterProps = [];
  // adding fil?.extra ? fil?.extra[*] check as a hotfix for timestamp filters
  const filtersGroupedByRef = Object.values(groupFilters(filters, 'ref'));
  filtersGroupedByRef.forEach((filtersGr) => {
    if (filtersGr.length === 1) {
      const fil = filtersGr[0];
      if (fil.props.length === 4) {
        fil.props.shift();
      }
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            extra: fil?.extra ? fil?.extra : null,
            prNa: fil?.extra ? fil?.extra[1] : `$${lowerCase(fil?.props[0])}`,
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3] === 'propMap'
                  ? ''
                  : fil?.extra[3]
                : 'event',
            objTy:
              fil?.extra?.[3] == 'propMap'
                ? ''
                : category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props?.[1] === 'datetime' ? formatFilterDate(val) : val,
            isPrMa: fil?.extra?.[3] == 'propMap' ? true : false
          });
        });
      } else {
        filterProps.push({
          extra: fil?.extra ? fil?.extra : null,
          prNa: fil?.extra ? fil?.extra[1] : `$${lowerCase(fil?.props[0])}`,
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3] === 'propMap'
                ? ''
                : fil?.extra[3]
              : 'event',
          objTy:
            fil?.extra?.[3] == 'propMap'
              ? ''
              : category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : 'event'
              : '',
          va:
            fil.props?.[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values,
          isPrMa: fil?.extra?.[3] == 'propMap' ? true : false
        });
      }
    } else {
      let fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          filterProps.push({
            extra: fil?.extra ? fil?.extra : null,
            prNa: fil?.extra ? fil?.extra[1] : `$${lowerCase(fil?.props[0])}`,
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: !index ? 'AND' : 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3] === 'propMap'
                  ? ''
                  : fil?.extra[3]
                : 'event',
            objTy:
              fil?.extra?.[3] == 'propMap'
                ? ''
                : category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props?.[1] === 'datetime' ? formatFilterDate(val) : val,
            isPrMa: fil?.extra?.[3] == 'propMap' ? true : false
          });
        });
      } else {
        filterProps.push({
          extra: fil?.extra ? fil?.extra : null,
          prNa: fil?.extra ? fil?.extra[1] : `$${lowerCase(fil?.props[0])}`,
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'AND',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3]
              : 'event',
          objTy:
            fil?.extra?.[3] == 'propMap'
              ? ''
              : category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3] === 'propMap'
                  ? ''
                  : fil?.extra[3]
                : 'event'
              : '',
          va:
            fil.props?.[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values,
          isPrMa: fil?.extra?.[3] == 'propMap' ? true : false
        });
      }
      fil = filtersGr[1];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val) => {
          filterProps.push({
            extra: fil?.extra ? fil?.extra : null,
            prNa: fil?.extra ? fil?.extra[1] : `$${lowerCase(fil?.props[0])}`,
            prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
            co: operatorMap[fil.operator],
            lOp: 'OR',
            en:
              category === 'channels' || category === 'custom_channels'
                ? ''
                : fil?.extra
                ? fil?.extra[3] === 'propMap'
                  ? ''
                  : fil?.extra[3]
                : 'event',
            objTy:
              fil?.extra?.[3] == 'propMap'
                ? ''
                : category === 'channels' || category === 'custom_channels'
                ? fil?.extra
                  ? fil?.extra[3]
                  : 'event'
                : '',
            va: fil.props?.[1] === 'datetime' ? formatFilterDate(val) : val,
            isPrMa: fil?.extra?.[3] == 'propMap' ? true : false
          });
        });
      } else {
        filterProps.push({
          extra: fil?.extra ? fil?.extra : null,
          prNa: fil?.extra ? fil?.extra[1] : `$${lowerCase(fil?.props[0])}`,
          prDaTy: fil?.extra ? fil?.extra[2] : fil?.props[1],
          co: operatorMap[fil.operator],
          lOp: 'OR',
          en:
            category === 'channels' || category === 'custom_channels'
              ? ''
              : fil?.extra
              ? fil?.extra[3] === 'propMap'
                ? ''
                : fil?.extra[3]
              : 'event',
          objTy:
            fil?.extra?.[3] == 'propMap'
              ? ''
              : category === 'channels' || category === 'custom_channels'
              ? fil?.extra
                ? fil?.extra[3]
                : 'event'
              : '',
          va:
            fil.props?.[1] === 'datetime'
              ? formatFilterDate(fil.values)
              : fil.values,
          isPrMa: fil?.extra?.[3] == 'propMap' ? true : false
        });
      }
    }
  });
  return filterProps;
};

const getGroupByWithPropertiesKPI = (appliedGroupBy, index, category) =>
  appliedGroupBy.map((opt) => {
    let appGbp = {};
    if (opt.eventIndex === index) {
      appGbp = {
        gr: '',
        prNa: opt.property,
        prDaTy: opt.prop_type,
        eni: opt.eventIndex,
        en:
          category === 'channels' || category === 'custom_channels'
            ? ''
            : opt.prop_category === 'propMap'
            ? ''
            : opt.prop_category,
        objTy:
          category === 'channels' || category === 'custom_channels'
            ? opt.prop_category === 'propMap'
              ? ''
              : opt.prop_category
            : '',
        dpNa: opt?.display_name ? opt?.display_name : '',
        isPrMa: opt.prop_category === 'propMap' ? true : false
      };
    } else {
      appGbp = {
        gr: '',
        prNa: opt.property,
        prDaTy: opt.prop_type,
        en:
          category === 'channels' || category === 'custom_channels'
            ? ''
            : opt.prop_category === 'propMap'
            ? ''
            : opt.prop_category,
        objTy:
          category === 'channels' || category === 'custom_channels'
            ? opt.prop_category === 'propMap'
              ? ''
              : opt.prop_category
            : '',
        dpNa: opt?.display_name ? opt?.display_name : '',
        isPrMa: opt.prop_category === 'propMap' ? true : false
      };
    }
    if (opt.prop_type === 'datetime') {
      opt.grn ? (appGbp.grn = opt.grn) : (appGbp.grn = 'day');
    }
    if (opt.prop_type === 'numerical') {
      opt.gbty ? (appGbp.gbty = opt.gbty) : (appGbp.gbty = '');
    }
    return appGbp;
  });

const getKPIqueryGroup = (queries, eventGrpBy, period) => {
  const queryArr = [];
  queries.forEach((item, index) => {
    const GrpByItem = eventGrpBy.filter(
      (item) => item.eventIndex === index + 1
    );
    queryArr.push({
      ca: item?.category,
      pgUrl: item?.pageViewVal ? item?.pageViewVal : '',
      dc: item.group,
      me: [item.metric],
      fil: getEventsWithPropertiesKPI(item.filters, item?.category),
      gBy: getGroupByWithPropertiesKPI(GrpByItem, index, item?.category),
      fr: period.from,
      to: period.to,
      tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
      qt: item.qt,
      an: item?.alias
    });
    queryArr.push({
      ca: item?.category,
      pgUrl: item?.pageViewVal ? item?.pageViewVal : '',
      dc: item.group,
      me: [item.metric],
      fil: getEventsWithPropertiesKPI(item.filters, item?.category),
      gBy: getGroupByWithPropertiesKPI(GrpByItem, index, item?.category),
      gbt: period.frequency,
      fr: period.from,
      to: period.to,
      tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
      qt: item.qt,
      an: item?.alias
    });
  });
  return queryArr;
};

export const getKPIQuery = (
  queries,
  dateRange,
  groupBy,
  queryOptions
  // globalFilters = []
) => {
  const query = {};
  query.cl = QUERY_TYPE_KPI?.toLocaleLowerCase();
  const period = {};
  if (dateRange?.from && dateRange?.to) {
    period.from = MomentTz(dateRange.from).startOf('day').utc().unix();
    period.to = MomentTz(dateRange.to).endOf('day').utc().unix();
    period.frequency = dateRange.frequency;
  } else {
    period.from = MomentTz().startOf('week').utc().unix();
    period.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
    period.frequency = dateRange.frequency;
  }

  const eventGrpBy = [...groupBy.event];
  query.qG = getKPIqueryGroup(queries, eventGrpBy, period);

  const GlobalGrpBy = [...groupBy.global];
  query.gGBy = getGroupByWithPropertiesKPI(
    GlobalGrpBy,
    null,
    queries[0]?.category
  );

  query.gFil = getEventsWithPropertiesKPI(
    queryOptions?.globalFilters,
    queries[0]?.category
  );

  return query;
};
const getGroupByWithPropertiesCustomKPI = (appliedGroupBy, index, category) => {
  return appliedGroupBy.map((opt) => {
    let appGbp = {};
    if (opt.eventIndex === index) {
      appGbp = {
        gr: '',
        prNa: opt.property,
        prDaTy: opt.prop_type,
        eni: opt.eventIndex,
        en:
          category === 'channels' || category === 'custom_channels'
            ? ''
            : opt.prop_category,
        objTy:
          category === 'channels' || category === 'custom_channels'
            ? opt.prop_category
            : '',
        dpNa: opt?.display_name ? opt?.display_name : ''
      };
    } else {
      appGbp = {
        gr: '',
        prNa: opt.property,
        prDaTy: opt.prop_type,
        en:
          category === 'channels' || category === 'custom_channels'
            ? ''
            : opt.prop_category,
        objTy:
          category === 'channels' || category === 'custom_channels'
            ? opt.prop_category
            : '',
        dpNa: opt?.display_name ? opt?.display_name : ''
      };
    }
    if (opt.prop_type === 'datetime') {
      opt.grn ? (appGbp.grn = opt.grn) : (appGbp.grn = 'day');
    }
    if (opt.prop_type === 'numerical') {
      opt.gbty ? (appGbp.gbty = opt.gbty) : (appGbp.gbty = '');
    }
    return appGbp;
  });
};
const getCustomKPIqueryGroup = (queries, eventGrpBy, period) => {
  const alphabetIndex = 'abcdefghijk';
  const queryArr = [];
  queries.forEach((item, index) => {
    const GrpByItem = eventGrpBy.filter(
      (item) => item.eventIndex === index + 1
    );
    queryArr.push({
      ca: item?.category,
      pgUrl: item?.pageViewVal ? item?.pageViewVal : '',
      dc: item.group,
      me: [item.metric],
      fil: getEventsWithPropertiesCustomKPI(item.filters, item?.category),
      gBy: getGroupByWithPropertiesCustomKPI(GrpByItem, index, item?.category),
      fr: period.from,
      to: period.to,
      tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
      na: alphabetIndex[index],
      qt: item.qt
    });
  });
  return queryArr;
};

export const getCustomKPIQuery = (
  queries,
  date_range,
  groupBy,
  queryOptions,
  formula
  // globalFilters = []
) => {
  const query = {};
  query.cl = QUERY_TYPE_KPI?.toLocaleLowerCase();
  const period = {};
  if (date_range?.from && date_range?.to) {
    period.from = MomentTz(date_range.from).startOf('day').utc().unix();
    period.to = MomentTz(date_range.to).endOf('day').utc().unix();
    period.frequency = date_range.frequency;
  } else {
    period.from = MomentTz().startOf('week').utc().unix();
    period.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
    period.frequency = date_range.frequency;
  }

  const eventGrpBy = [...groupBy.event];
  query.qG = getCustomKPIqueryGroup(queries, eventGrpBy, period);
  query.for = formula;
  return query;
};

const mapQueriesByGroup = (queries) => {
  const group = {};
  queries.forEach((query) => {
    group[query.group]
      ? group[query.group].push(query)
      : (group[query.group] = [query]);
  });
  return group;
};

const getGroupByByGroup = (grp) => {
  if (grp !== 'hubspot_deals' && grp !== 'salesforce_opportunities') {
    return [
      {
        gr: '',
        prNa: '$user_id',
        prDaTy: 'numerical',
        en: 'user',
        objTy: '',
        gbty: 'raw_values'
      }
    ];
  } else if (grp === 'hubspot_deals') {
    return [
      {
        gr: '',
        prNa: '$hubspot_deal_hs_object_id',
        prDaTy: 'numerical',
        en: 'user',
        objTy: '',
        gbty: 'raw_values'
      }
    ];
  } else if (grp === 'salesforce_opportunities') {
    return [
      {
        gr: '',
        prNa: '$salesforce_opportunity_id',
        prDaTy: 'numerical',
        en: 'user',
        objTy: '',
        gbty: 'raw_values'
      }
    ];
  }
};

export const getKPIQueryAttributionV1 = (
  queries,
  dateRange,
  groupBy,
  queryOptions
  // globalFilters = []
) => {
  const kpiQueriesByGroup = mapQueriesByGroup(queries);
  const kpiQueries = [];

  const period = {};
  if (dateRange?.from && dateRange?.to) {
    period.from = MomentTz(dateRange.from).startOf('day').utc().unix();
    period.to = MomentTz(dateRange.to).endOf('day').utc().unix();
    period.frequency = 'second';
  } else {
    period.from = MomentTz().startOf('week').utc().unix();
    period.to =
      MomentTz().format('dddd') !== 'Sunday'
        ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
        : MomentTz().utc().unix();
    period.frequency = 'second';
  }

  Object.keys(kpiQueriesByGroup).forEach((groupKey) => {
    const kpiQuery = {
      kpi_query_group: [],
      analyze_type: ''
    };
    const eventGrpBy = [...groupBy.event];

    kpiQuery.kpi_query_group = {
      cl: QUERY_TYPE_KPI?.toLocaleLowerCase(),
      qG: getKPIqueryGroup(kpiQueriesByGroup[groupKey], eventGrpBy, period),
      gGBy: getGroupByByGroup(groupKey),
      gFil: getEventsWithPropertiesKPI(
        queryOptions?.globalFilters,
        kpiQueriesByGroup[groupKey][0]?.category
      )
    };
    kpiQuery.analyze_type = [
      'hubspot_deals',
      'salesforce_opportunities'
    ].includes(groupKey)
      ? groupKey
      : 'user_kpi';
    if (kpiQueriesByGroup[groupKey].length) {
      kpiQueries.push(kpiQuery);
    }
  });

  return kpiQueries;
};

export const getSessionsQuery = ({ period }) => {
  return {
    cl: QUERY_TYPE_EVENT,
    ty: TYPE_UNIQUE_USERS,
    fr: period.from,
    to: period.to,
    ewp: [
      {
        na: '$session',
        pr: []
      }
    ],
    gup: [],
    gbt: '',
    ec: EVENT_QUERY_USER_TYPE.each,
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata'
  };
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
      rows: metrics
    }
  };
  return result;
};

const getEventIdx = (eventData, userObj) => {
  const str = userObj.slice(0, userObj.length - 1).join(', ');
  const eventIdx = eventData.findIndex(
    (elem) => elem.slice(0, elem.length - 1).join(', ') === str
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
      parseFloat(eVal.toFixed(2))
    ];
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
      rows: metrics
    }
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
      rows: metrics
    }
  };
  return result;
};

export const calculateFrequencyData = (
  eventData,
  userData,
  appliedBreakdown
) => {
  if (appliedBreakdown.length) {
    return calculateFrequencyDataForBreakdown(eventData, userData);
  }
  return calculateFrequencyDataForNoBreakdown(eventData, userData);
};

export const calculateActiveUsersData = (
  userData,
  sessionData,
  appliedBreakdown
) => {
  if (appliedBreakdown.length) {
    return calculateActiveUsersDataForBreakdown(userData, sessionData);
  }
  return calculateActiveUsersDataForNoBreakdown(userData, sessionData);
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

export const numberWithCommas = (x) =>
  x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');

export const formatApiData = (data, metrics) => ({ ...data, metrics });

const createFilterStruct = (pr, ref) => ({
  operator:
    pr.ty === 'datetime'
      ? reverseDateOperatorMap[pr.op]
      : reverseOperatorMap[pr.op],
  props: [pr.grpn, pr.pr, pr.ty, pr.en],
  values: pr.ty === 'categorical' ? [pr.va] : pr.va,
  ref
});

export const processFiltersFromQuery = (prArray) => {
  const filtersArray = [];
  let ref = -1;
  let lastProp = '';
  let lastOp = '';
  prArray.forEach((pr) => {
    if (pr.lop === 'AND') {
      ref += 1;
      filtersArray.push(createFilterStruct(pr, ref));
      lastProp = pr.pr;
      lastOp = pr.op;
    } else if (lastProp === pr.pr && lastOp === pr.op) {
      filtersArray[filtersArray.length - 1].values.push(pr.va);
    } else {
      filtersArray.push(createFilterStruct(pr, ref));
      lastProp = pr.pr;
      lastOp = pr.op;
    }
  });
  return filtersArray;
};

export const processBreakdownsFromQuery = (breakdownArray) => {
  return breakdownArray.map((opt, index) => ({
    groupName: opt.grpn,
    property: opt.pr,
    prop_category: opt.en,
    prop_type: opt.pty,
    eventName: opt.ena,
    eventIndex: opt.eni || 0,
    grn: opt.grn,
    gbty: opt.gbty,
    overAllIndex: index
  }));
};

export const getStateQueryFromRequestQuery = (requestQuery) => {
  const events = (requestQuery?.ewp || []).map((e) => {
    const eventFilters = processFiltersFromQuery(e.pr);
    return {
      alias: e.an,
      label: e.na,
      group: e.grpa,
      filters: eventFilters,
      key: generateRandomKey()
    };
  });

  const globalFilters =
    requestQuery?.gup && Array.isArray(requestQuery.gup)
      ? processFiltersFromQuery(requestQuery.gup)
      : null;

  const queryType = requestQuery.cl;
  const eventsCondition = requestQuery.ec;
  const groupAnalysis =
    requestQuery.grpa || QUERY_OPTIONS_DEFAULT_VALUE.group_analysis;
  const sessionAnalyticsSeq = INITIAL_SESSION_ANALYTICS_SEQ;

  const event = processBreakdownsFromQuery(requestQuery?.gbp || []).filter(
    (b) => b.eventIndex
  );
  const global = processBreakdownsFromQuery(requestQuery?.gbp || []).filter(
    (b) => !b.eventIndex
  );

  const dateRange = {
    from: requestQuery.fr * 1000,
    to: requestQuery.to * 1000,
    frequency: requestQuery.gbt
  };

  const result = {
    events,
    queryType,
    eventsCondition,
    groupAnalysis,
    session_analytics_seq: sessionAnalyticsSeq,
    globalFilters,
    breakdown: { event, global },
    dateRange,
    ...(queryType === QUERY_TYPE_FUNNEL && {
      funnelConversionDurationNumber:
        requestQuery.cnvtm != null
          ? parseInt(requestQuery.cnvtm.slice(0, -1))
          : CORE_QUERY_INITIAL_STATE.funnelConversionDurationNumber,
      funnelConversionDurationUnit:
        requestQuery.cnvtm != null
          ? requestQuery.cnvtm.slice(-1)
          : CORE_QUERY_INITIAL_STATE.funnelConversionDurationUnit
    })
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
      : PREDEFINED_DATES.THIS_WEEK
};

export const DashboardDefaultDateRangeFormat = {
  from: MomentTz().subtract(7, 'days').startOf('week'),
  to: MomentTz().subtract(7, 'days').endOf('week'),
  frequency: 'date',
  dateType: PREDEFINED_DATES.LAST_WEEK
};

export const getStateFromKPIFilters = (rawFilters) => {
  const eventFilters = [];

  let ref = -1,
    lastProp = '',
    lastOp = '';
  rawFilters.forEach((pr) => {
    if (pr.lOp === 'AND') {
      ref += 1;
      const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
      const DNa = _.startCase(pr.prNa);
      const isCamp =
        pr?.ca === 'channels' || pr?.ca === 'custom_channels'
          ? pr.objTy
          : pr.en;
      eventFilters.push({
        operator:
          pr.prDaTy === 'datetime'
            ? reverseDateOperatorMap[pr.co]
            : reverseOperatorMap[pr.co],
        props: [DNa, pr.prDaTy, isCamp],
        values:
          pr.prDaTy === FILTER_TYPES.DATETIME
            ? convertDateTimeObjectValuesToMilliSeconds(val)
            : val,
        extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
        ref
      });
      lastProp = pr.prNa;
      lastOp = pr.co;
    } else if (lastProp === pr.prNa && lastOp === pr.co) {
      eventFilters[eventFilters.length - 1].values.push(pr.va);
    } else {
      const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
      const DNa = _.lowerCase(pr.prNa);
      const isCamp =
        pr?.ca === 'channels' || pr?.ca === 'custom_channels'
          ? pr.objTy
          : pr.en;
      eventFilters.push({
        operator:
          pr.prDaTy === 'datetime'
            ? reverseDateOperatorMap[pr.co]
            : reverseOperatorMap[pr.co],
        props: [DNa, pr.prDaTy, isCamp],
        values:
          pr.prDaTy === FILTER_TYPES.DATETIME
            ? convertDateTimeObjectValuesToMilliSeconds(val)
            : val,
        extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
        ref
      });
      lastProp = pr.prNa;
      lastOp = pr.co;
    }
  });
  return eventFilters;
};

const getFiltersTouchpoints = (filters, touchpoint) => {
  const result = [];
  const filtersGroupedByRef = Object.values(groupFilters(filters, 'ref'));
  filtersGroupedByRef.forEach((filtersGr) => {
    if (filtersGr.length === 1) {
      const fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          result.push({
            attribution_key: touchpoint,
            lop: !index ? 'AND' : 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: val
          });
        });
      } else {
        result.push({
          attribution_key: touchpoint,
          lop: 'AND',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? fil.values : fil.values
        });
      }
    } else {
      let fil = filtersGr[0];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val, index) => {
          result.push({
            attribution_key: touchpoint,
            lop: !index ? 'AND' : 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: val
          });
        });
      } else {
        result.push({
          attribution_key: touchpoint,
          lop: 'AND',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? fil.values : fil.values
        });
      }
      fil = filtersGr[1];
      if (Array.isArray(fil.values)) {
        fil.values.forEach((val) => {
          result.push({
            attribution_key: touchpoint,
            lop: 'OR',
            op: operatorMap[fil.operator],
            pr: fil.props[0],
            ty: fil.props[1],
            va: val
          });
        });
      } else {
        result.push({
          attribution_key: touchpoint,
          lop: 'OR',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? fil.values : fil.values
        });
      }
    }
  });
  return result;
};

export const getAttributionQuery = (
  eventGoal = { filters: [] },
  touchpoint,
  attrDimensions,
  contentGroups,
  touchpointFilters,
  queryType,
  models,
  window,
  linkedEvents,
  dateRange = {},
  tacticOfferType,
  v1 = false
) => {
  const eventFilters = formatFiltersForQuery(eventGoal.filters);
  let touchPointFiltersQuery = [];
  if (touchpointFilters.length) {
    touchPointFiltersQuery = getFiltersTouchpoints(
      touchpointFilters,
      touchpoint
    );
  }

  let attrQueryV1 = {};

  if (v1) {
    attrQueryV1 = new AttributionQueryV1();
  }

  attrQueryV1.cm = ['Impressions', 'Clicks', 'Spend'];
  attrQueryV1.ce = {
    na: eventGoal.label,
    pr: eventFilters
  };
  attrQueryV1.attribution_key = touchpoint;
  attrQueryV1.attribution_key_f = touchPointFiltersQuery;
  attrQueryV1.query_type = queryType;
  attrQueryV1.attribution_methodology = models[0];
  attrQueryV1.lbw = window;
  attrQueryV1.tactic_offer_type = tacticOfferType;

  const query = {
    cl: QUERY_TYPE_ATTRIBUTION,
    meta: {
      metrics_breakdown: true
    },
    query: attrQueryV1
  };
  if (!eventGoal || !eventGoal.label) {
    query.query.ce = {};
  }
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
      const linkedEventFilters = formatFiltersForQuery(le.filters);
      return {
        na: le.label,
        pr: linkedEventFilters
      };
    });
  }
  const listDimensions =
    touchpoint === 'LandingPage'
      ? contentGroups.slice()
      : attrDimensions.slice();

  const attributionKeyDimensions = listDimensions
    .filter((d) => d.touchPoint === touchpoint && d.enabled && d.type === 'key')
    .map((d) => d.header);
  const attributionKeyCustomDimensions = listDimensions
    .filter(
      (d) => d.touchPoint === touchpoint && d.enabled && d.type === 'custom'
    )
    .map((d) => d.header);
  const attributionContentGroups = listDimensions
    .filter(
      (d) =>
        d.touchPoint === touchpoint && d.enabled && d.type === 'content_group'
    )
    .map((d) => d.header);

  if (touchpoint !== MARKETING_TOUCHPOINTS.SOURCE) {
    query.query.attribution_key_dimensions = attributionKeyDimensions;
    query.query.attribution_key_custom_dimensions =
      attributionKeyCustomDimensions;
    query.query.attribution_content_groups = attributionContentGroups;
  }

  return query;
};

export const getAttributionStateFromRequestQuery = (
  requestQuery,
  initial_attr_dimensions,
  initial_content_groups,
  kpiConfig
) => {
  let attrQueries = [];
  if (requestQuery.kpi_queries && requestQuery.kpi_queries.length) {
    requestQuery.kpi_queries.map((query) => {
      const kpiQuery = getKPIStateFromRequestQuery(
        query?.kpi_query_group,
        kpiConfig
      );
      attrQueries.push(...kpiQuery.events);
    });
  }

  const filters = processFiltersFromQuery(get(requestQuery, 'ce.pr', []));
  const touchPointFilters = requestQuery.attribution_key_f
    ? processFiltersFromQuery(requestQuery.attribution_key_f)
    : null;

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
            ) > -1
      };
    }
    return dimension;
  });

  const content_groups = initial_content_groups.map((dimension) => {
    if (dimension.touchPoint === touchpoint) {
      return {
        ...dimension,
        enabled: !requestQuery.attribution_key_dimensions
          ? dimension.defaultValue
          : requestQuery.attribution_key_dimensions?.indexOf(dimension.header) >
              -1 ||
            requestQuery.attribution_content_groups?.indexOf(dimension.header) >
              -1
      };
    }
    return dimension;
  });

  const result = {
    queryType: QUERY_TYPE_ATTRIBUTION,
    eventGoal: {
      label: requestQuery.ce.na,
      filters
    },
    attrQueries,
    touchpoint_filters: touchPointFilters,
    attr_query_type: requestQuery.query_type,
    touchpoint,
    attr_dimensions,
    content_groups,
    models: [requestQuery.attribution_methodology],
    window: requestQuery.lbw,
    tacticOfferType: requestQuery.tactic_offer_type,
    analyze_type: requestQuery.analyze_type || 'all'
  };

  if (requestQuery.attribution_methodology_c) {
    result.models.push(requestQuery.attribution_methodology_c);
  }

  if (requestQuery.lfe && requestQuery.lfe.length) {
    result.linkedEvents = requestQuery.lfe.map((le) => {
      const linkedFilters = processFiltersFromQuery(le.pr);
      return {
        label: le.na,
        filters: linkedFilters
      };
    });
  } else {
    result.linkedEvents = [];
  }
  return result;
};

export const isComparisonEnabled = (queryType, events, groupBy, models) => {
  if (queryType === QUERY_TYPE_FUNNEL) {
    return true;
  }

  if (queryType === QUERY_TYPE_EVENT) {
    const newAppliedBreakdown = [...groupBy.event, ...groupBy.global];
    if (newAppliedBreakdown.length === 0) {
      return true;
    }
    if (events.length === 1 && newAppliedBreakdown.length === 1) {
      return true;
    }
    return false;
  }

  if (queryType === QUERY_TYPE_ATTRIBUTION) {
    if (models.length === 1) {
      return true;
    }
  }
  if (queryType === QUERY_TYPE_KPI) {
    return true;
  }
  return false;
};

export const getProfileQueryFromRequestQuery = (requestQuery) => {
  const queryType = requestQuery.cl;
  const groupAnalysis = requestQuery.grpa;

  const queries = requestQuery.queries.map((e) => {
    const evfilters = processFiltersFromQuery(e.pr);
    return {
      alias: e.an,
      label: e.ty,
      filters: evfilters
    };
  });

  const filters =
    requestQuery?.gup && Array.isArray(requestQuery.gup)
      ? processFiltersFromQuery(requestQuery.gup)
      : null;

  const globalBreakdown = processBreakdownsFromQuery(
    requestQuery?.gbp || []
  ).filter((b) => !b.eventIndex);

  const groupBy = {
    global: globalBreakdown,
    event: []
  };
  const dateRange = {
    from: requestQuery.from * 1000,
    to: requestQuery.to * 1000
  };
  const result = {
    queryType,
    groupAnalysis,
    events: queries,
    globalFilters: filters,
    breakdown: groupBy,
    dateRange
  };
  return result;
};

export const convertDateTimeObjectValuesToMilliSeconds = (obj) => {
  const parsedObj = JSON.parse(obj);
  parsedObj.fr = isDateInMilliSeconds(parsedObj.fr)
    ? parsedObj.fr
    : parsedObj.fr * 1000;
  parsedObj.to = isDateInMilliSeconds(parsedObj.to)
    ? parsedObj.to
    : parsedObj.to * 1000;
  return JSON.stringify(parsedObj);
};

export const getKPIStateFromRequestQuery = (requestQuery, kpiConfig = []) => {
  const queryType = requestQuery.cl;
  const queries = [];
  for (let i = 0; i < requestQuery.qG.length; i += 2) {
    const q = requestQuery.qG[i];
    const config = kpiConfig.find((elem) => elem.display_category === q.dc);
    const metric = config
      ? config.metrics.find((m) => m.name === q.me[0])
      : null;

    const eventFilters = [];
    const fil = get(q, 'fil', EMPTY_ARRAY)
      ? get(q, 'fil', EMPTY_ARRAY)
      : EMPTY_ARRAY;
    let ref = -1;
    let lastProp = '';
    let lastOp = '';
    fil.forEach((pr, index) => {
      if (pr.lOp === 'AND') {
        ref += 1;
        const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
        const DNa = pr.extra ? pr.extra[0] : startCase(pr.prNa);
        const isCamp =
          requestQuery?.qG[i]?.ca === 'channels' ||
          requestQuery?.qG[i]?.ca === 'custom_channels'
            ? pr.objTy
            : pr.en;
        eventFilters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy, isCamp],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
          ref
        });
        lastProp = pr.prNa;
        lastOp = pr.co;
      } else if (lastProp === pr.prNa && lastOp === pr.co) {
        eventFilters[eventFilters.length - 1].values.push(pr.va);
      } else {
        const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
        const DNa = pr.extra ? pr.extra[0] : startCase(pr.prNa);
        const isCamp =
          requestQuery?.qG[i]?.ca === 'channels' ||
          requestQuery?.qG[i]?.ca === 'custom_channels'
            ? pr.objTy
            : pr.en;
        eventFilters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy, isCamp],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
          ref
        });
        lastProp = pr.prNa;
        lastOp = pr.co;
      }
    });

    queries.push({
      category: q.ca,
      group: q.dc,
      pageViewVal: q.pgUrl,
      metric: q.me[0],
      label: metric ? metric.display_name : q.me[0],
      filters: eventFilters,
      alias: q?.an,
      metricType: get(metric, 'type', null),
      qt: q.qt
    });
  }
  // const globalFilters = [];

  const filters = [];
  let ref = -1;
  let lastProp = '';
  let lastOp = '';
  requestQuery.gFil.forEach((pr) => {
    if (pr.lOp === 'AND') {
      ref += 1;
      const val = pr.prDaTy === FILTER_TYPES.CATEGORICAL ? [pr.va] : pr.va;
      const DNa = pr.extra ? pr.extra[0] : startCase(pr.prNa);
      const isCamp =
        requestQuery?.qG[0]?.ca === 'channels' ||
        requestQuery?.qG[0]?.ca === 'custom_channels'
          ? pr.objTy
          : pr.en;
      filters.push({
        operator:
          pr.prDaTy === 'datetime'
            ? reverseDateOperatorMap[pr.co]
            : reverseOperatorMap[pr.co],
        props: [DNa, pr.prDaTy, isCamp],
        values:
          pr.prDaTy === FILTER_TYPES.DATETIME
            ? convertDateTimeObjectValuesToMilliSeconds(val)
            : val,
        extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
        ref
      });
      lastProp = pr.prNa;
      lastOp = pr.co;
    } else if (lastProp === pr.prNa && lastOp === pr.co) {
      filters[filters.length - 1].values.push(pr.va);
    } else {
      const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
      const DNa = pr.extra ? pr.extra[0] : startCase(pr.prNa);
      const isCamp =
        requestQuery?.qG[0]?.ca === 'channels' ||
        requestQuery?.qG[0]?.ca === 'custom_channels'
          ? pr.objTy
          : pr.en;
      filters.push({
        operator:
          pr.prDaTy === 'datetime'
            ? reverseDateOperatorMap[pr.co]
            : reverseOperatorMap[pr.co],
        props: [DNa, pr.prDaTy, isCamp],
        values:
          pr.prDaTy === FILTER_TYPES.DATETIME
            ? convertDateTimeObjectValuesToMilliSeconds(val)
            : val,
        extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
        ref
      });
      lastProp = pr.prNa;
      lastOp = pr.co;
    }
  });

  const globalBreakdown = requestQuery.gGBy.map((opt, index) => {
    let appGbp = {};
    appGbp = {
      property: opt.prNa,
      prop_type: opt.prDaTy,
      overAllIndex: index,
      prop_category: opt?.isPrMa ? 'propMap' : opt.en || opt.objTy,
      display_name: opt?.dpNa ? opt?.dpNa : ''
    };
    if (opt.prDaTy === 'datetime') {
      opt.grn ? (appGbp.grn = opt.grn) : (appGbp.grn = 'day');
    }
    if (opt.prDaTy === 'numerical') {
      opt.gbty ? (appGbp.gbty = opt.gbty) : (appGbp.gbty = '');
    }
    return appGbp;
  });

  const groupBy = {
    global: globalBreakdown,
    event: [] // will be added later
  };
  const dateRange = {
    ...DefaultDateRangeFormat,
    from: requestQuery.qG[1].fr * 1000,
    to: requestQuery.qG[1].to * 1000,
    frequency: requestQuery.qG[1].gbt ? requestQuery.qG[1].gbt : 'date' // fix on .gbt for saved channel queries migrated to kpi queries
  };
  const result = {
    events: queries,
    queryType,
    globalFilters: filters,
    breakdown: groupBy,
    dateRange
  };
  return result;
};

export const getStateFromCustomKPIqueryGroup = (
  requestQuery,
  kpiConfig = []
) => {
  const queries = [];
  for (let i = 0; i < requestQuery.qG.length; i += 1) {
    const q = requestQuery.qG[i];
    const config = kpiConfig.find((elem) => elem.display_category === q.dc);
    const metric = config
      ? config.metrics.find((m) => m.name === q.me[0])
      : null;

    const eventFilters = [];
    const fil = get(q, 'fil', EMPTY_ARRAY)
      ? get(q, 'fil', EMPTY_ARRAY)
      : EMPTY_ARRAY;
    let ref = -1;
    let lastProp = '';
    let lastOp = '';
    fil.forEach((pr, index) => {
      if (pr.lOp === 'AND') {
        ref += 1;
        const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
        const DNa = pr.prNa;
        const isCamp =
          requestQuery?.qG[i]?.ca === 'channels' ||
          requestQuery?.qG[i]?.ca === 'custom_channels'
            ? pr.objTy
            : pr.en;
        eventFilters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy, isCamp],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
          ref
        });
        lastProp = pr.prNa;
        lastOp = pr.co;
      } else if (lastProp === pr.prNa && lastOp === pr.co) {
        eventFilters[eventFilters.length - 1].values.push(pr.va);
      } else {
        const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
        const DNa = pr.prNa;
        const isCamp =
          requestQuery?.qG[i]?.ca === 'channels' ||
          requestQuery?.qG[i]?.ca === 'custom_channels'
            ? pr.objTy
            : pr.en;
        eventFilters.push({
          operator:
            pr.prDaTy === 'datetime'
              ? reverseDateOperatorMap[pr.co]
              : reverseOperatorMap[pr.co],
          props: [DNa, pr.prDaTy, isCamp],
          values:
            pr.prDaTy === FILTER_TYPES.DATETIME
              ? convertDateTimeObjectValuesToMilliSeconds(val)
              : val,
          extra: [DNa, pr.prNa, pr.prDaTy, isCamp],
          ref
        });
        lastProp = pr.prNa;
        lastOp = pr.co;
      }
    });

    queries.push({
      category: q.ca,
      group: q.dc,
      pageViewVal: q.pgUrl,
      metric: q.me[0],
      label: metric ? metric.display_name : q.me[0],
      filters: eventFilters,
      alias: q?.an,
      metricType: get(metric, 'type', null),
      qt: q.qt
    });
  }
  return queries;
};
