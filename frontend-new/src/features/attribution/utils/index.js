import get from 'lodash/get';
import startCase from 'lodash/startCase';
import {
  reverseOperatorMap,
  reverseDateOperatorMap
} from 'Utils/operatorMapping';
import { EMPTY_ARRAY } from 'Utils/global';
import { QUERY_TYPE_ATTRIBUTION } from 'Utils/constants';
import {
  FILTER_TYPES,
} from 'Views/CoreQuery/constants';
import {
  convertDateTimeObjectValuesToMilliSeconds,
  DefaultDateRangeFormat
} from 'Views/CoreQuery/utils';

export const convertToEventOptions = (eventNames, display_names = []) => {
  // Based on groups Currently clubbing all into one, untill we have backend support
  const options = [];

  Object.keys(eventNames).forEach((key) => {
    const icon = key.toLowerCase().split(' ').join('_');
    options.push({
      label: key,
      icon: icon,
      values: eventNames[key].map((v) => {
        const name = display_names[v];
        return [name ? name : v, v];
      })
    });
  });
  return options;
};

const setFiltersFromRequestQuery = (
  requestQuery,
  filters = [],
  ref = -1,
  lastProp = '',
  lastOp = ''
) => {
  get(requestQuery, 'ce.pr', []).forEach((pr) => {
    if (pr.lop === 'AND') {
      ref += 1;
      const val = pr.ty === 'categorical' ? [pr.va] : pr.va;
      filters.push({
        operator:
          pr.ty === 'datetime'
            ? reverseDateOperatorMap[pr.op]
            : reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: val,
        ref
      });
      lastProp = pr.pr;
      lastOp = pr.op;
    } else if (lastProp === pr.pr && lastOp === pr.op) {
      filters[filters.length - 1].values.push(pr.va);
    } else {
      const val = pr.ty === 'categorical' ? [pr.va] : pr.va;
      filters.push({
        operator:
          pr.ty === 'datetime'
            ? reverseDateOperatorMap[pr.op]
            : reverseOperatorMap[pr.op],
        props: [pr.pr, pr.ty, pr.en],
        values: val,
        ref
      });
      lastProp = pr.pr;
      lastOp = pr.op;
    }
  });
};

export const getAttributionStateFromRequestQuery = (
  requestQuery,
  initial_attr_dimensions,
  initial_content_groups,
  kpiConfig
) => {
  let attrQueries = [];
  const filters = [];
  let ref = -1;
  let lastProp = '';
  let lastOp = '';
  requestQuery.kpi_queries?.forEach((kpiQ) => {
    const qState = getKPIStateFromRequestQuery(kpiQ.kpi_query_group, kpiConfig);
    attrQueries.push(...qState.events);
    setFiltersFromRequestQuery(kpiQ, filters, ref, lastProp, lastOp);
    return qState;
  });

  const touchPointFilters = [];
  if (requestQuery.attribution_key_f) {
    let ref = -1;
    let lastProp = '';
    let lastOp = '';
    requestQuery.attribution_key_f.forEach((pr) => {
      if (pr.lop === 'AND') {
        ref += 1;
        const val = pr.ty === 'categorical' ? [pr.va] : pr.va;
        touchPointFilters.push({
          operator:
            pr.ty === 'datetime'
              ? reverseDateOperatorMap[pr.op]
              : reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.attribution_key],
          values: val,
          ref
        });
        lastProp = pr.pr;
        lastOp = pr.op;
      } else if (lastProp === pr.pr && lastOp === pr.op) {
        touchPointFilters[touchPointFilters.length - 1].values.push(pr.va);
      } else {
        const val = pr.ty === 'categorical' ? [pr.va] : pr.va;
        touchPointFilters.push({
          operator:
            pr.ty === 'datetime'
              ? reverseDateOperatorMap[pr.op]
              : reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.attribution_key],
          values: val,
          ref
        });
        lastProp = pr.pr;
        lastOp = pr.op;
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
    analyze_type: 'all'
  };

  if (requestQuery.attribution_methodology_c) {
    result.models.push(requestQuery.attribution_methodology_c);
  }

  return result;
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
    fil.forEach((pr) => {
      if (pr.lOp === 'AND') {
        ref += 1;
        const val = pr.prDaTy === 'categorical' ? [pr.va] : pr.va;
        const DNa = pr.extra ? pr.extra[0] : startCase(pr.prNa);
        const isCamp =
          requestQuery?.qG[0]?.ca === 'channels' ||
            requestQuery?.qG[0]?.ca === 'custom_channels'
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
          requestQuery?.qG[0]?.ca === 'channels' ||
            requestQuery?.qG[0]?.ca === 'custom_channels'
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
      alias: '',
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

  const globalBreakdown = requestQuery.gGBy?.map((opt, index) => {
    let appGbp = {};
    appGbp = {
      property: opt.prNa,
      prop_type: opt.prDaTy,
      overAllIndex: index,
      prop_category: opt.en || opt.objTy,
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
