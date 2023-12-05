import MomentTz from 'Components/MomentTz';
import { QUERY_TYPE_KPI } from 'Utils/constants';
import { formatFilterDate } from 'Utils/dataFormatter';
import { groupFilters } from 'Utils/global';
import { operatorMap } from 'Utils/operatorMapping';

export const formatFiltersForPayload = (filters = []) => {
  const filterProps = [];
  filters.forEach((filter) => {
    const { values, props, operator } = filter;
    const vals = Array.isArray(values) ? filter.values : [filter.values];
    if (props.length === 4) {
      props.shift();
    }

    vals.forEach((val, index) => {
      filterProps.push({
        pr_na: props[0],
        pr_da_ty: props[1],
        co: operatorMap[operator],
        l_op: index === 0 ? 'AND' : 'OR',
        va: val
      });
    });
  });

  return filterProps;
};

const getPredefinedqueryGroupForWidget1 = (widget, filter, groupBy, period) => {
  const queryArr = [];
  widget?.me?.forEach((item, index) => {
    queryArr.push({
      me: [{ na: item?.na, d_na: item?.d_na }],
      inter_e_type: item?.inter_e_type,
      fil: formatFiltersForPayload(filter),
      g_by: groupBy,
      gbt: period.frequency,
      fr: period.from,
      to: period.to,
      tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
      inter_id: widget.inter_id
    });
    queryArr.push({
      me: [{ na: item?.na, d_na: item?.d_na }],
      inter_e_type: item?.inter_e_type,
      fil: formatFiltersForPayload(filter),
      g_by: groupBy,
      gbt: '',
      fr: period.from,
      to: period.to,
      tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
      inter_id: widget.inter_id
    });
  });
  return queryArr;
};

const getPredefinedqueryGroup = (widget, filter, groupBy, period) => {
  const queryArr = [];
  queryArr.push({
    me: widget.me.map((obj) => {
      const { inter_e_type, ...rest } = obj; // Use destructuring to exclude "inter_e_type"
      return rest; // Return the object without "inter_e_type"
    }),
    inter_e_type: widget.me[0]?.inter_e_type,
    fil: formatFiltersForPayload(filter),
    g_by: groupBy,
    gbt: period.frequency,
    fr: period.from,
    to: period.to,
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
    inter_id: widget.inter_id
  });
  queryArr.push({
    me: widget.me.map((obj) => {
      const { inter_e_type, ...rest } = obj; // Use destructuring to exclude "inter_e_type"
      return rest; // Return the object without "inter_e_type"
    }),
    inter_e_type: widget.me[0]?.inter_e_type,
    fil: formatFiltersForPayload(filter),
    g_by: groupBy,
    gbt: '',
    fr: period.from,
    to: period.to,
    tz: localStorage.getItem('project_timeZone') || 'Asia/Kolkata',
    inter_id: widget.inter_id
  });
  return queryArr;
};

export const getPredefinedQuery = (
  widget,
  dateRange,
  filter = [],
  groupBy = {}
) => {
  const query = {};
  query.cl = 'predef_web_analytics_query';
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

  if (widget?.inter_id === 1) {
    query.q_g = getPredefinedqueryGroupForWidget1(
      widget,
      filter,
      groupBy,
      period
    );
  } else {
    query.q_g = getPredefinedqueryGroup(widget, filter, groupBy, period);
  }

  return query;
};

export const transformWidgetResponse = (response) => {
  // Initialize empty arrays for odd and even elements
  const originalData = response;
  let oddHeaders = [];
  let evenHeaders = [];
  let oddRows = [];
  let evenRows = [];

  // Separate odd and even elements
  for (let i = 0; i < response.length; i++) {
    if (i % 2 === 0) {
      // Even element
      evenHeaders.push(...response[i].headers);
      evenRows.push(response[i].rows);
    } else {
      // Odd element
      oddHeaders.push(...response[i].headers);
      oddRows.push(...response[i].rows);
    }
  }

  evenHeaders = evenHeaders.filter((obj) => obj != 'datetime');

  let evenRowTransform = evenRows[0];
  for(let i = 1; i < evenRows.length; i++) {
    for(let j = 0; j < evenRowTransform.length; j++) {
      evenRowTransform[j].push(evenRows[i][j][1])
    }
  }
  
  const formateData = {
    result : [
      {
        "cache_meta": originalData.cache_meta,
        "headers": ['datetime', ...evenHeaders],
        "meta": originalData.meta,
        "query": originalData.query,
        "rows": evenRowTransform,
      },
      {
        "cache_meta": originalData.cache_meta,
        "headers": oddHeaders,
        "meta": originalData.meta,
        "query": originalData.query,
        "rows": [oddRows.flat()],
      }
    ]
  }

  return formateData;
};
