import { formatDurationIntoString, PropTextFormat } from 'Utils/dataFormatter';
import {
  ANY_USER_TYPE,
  EVENT_QUERY_USER_TYPE,
  PREDEFINED_DATES,
  QUERY_TYPE_EVENT,
  REVERSE_USER_TYPES,
  TYPE_UNIQUE_USERS
} from 'Utils/constants';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { operatorMap, reverseOperatorMap } from 'Utils/operatorMapping';
import {
  getEventsWithProperties,
  getStateQueryFromRequestQuery
} from '../../Views/CoreQuery/utils';
import MomentTz from '../MomentTz';
import { INITIAL_USER_PROFILES_FILTERS_STATE } from './AccountProfiles/accountProfiles.constants';

export const groups = {
  Timestamp: (item) =>
    MomentTz(item.timestamp * 1000).format('DD MMM YYYY, hh:mm:ss A'),
  Hourly: (item) =>
    `${MomentTz(item.timestamp * 1000)
      .startOf('hour')
      .format('hh A')} - ${MomentTz(item.timestamp * 1000)
      .add(1, 'hour')
      .startOf('hour')
      .format('hh A')} ${MomentTz(item.timestamp * 1000)
      .startOf('hour')
      .format('DD MMM YYYY')}`,
  Daily: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('day')
      .format('DD MMM YYYY'),
  Weekly: (item) =>
    `${MomentTz(item.timestamp * 1000)
      .startOf('week')
      .format('DD MMM YYYY')} - ${MomentTz(item.timestamp * 1000)
      .endOf('week')
      .format('DD MMM YYYY')}`,
  Monthly: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('month')
      .format('MMM YYYY'),
  Timeline: (item) =>
    MomentTz(item.timestamp * 1000)
      .startOf('day')
      .format(' DD MMM YYYY ddd')
};

const getEntityName = (caller, grpn) => {
  if (caller === 'account_profiles') {
    return grpn === 'user' ? 'user_group' : 'user_g';
  }
  return 'user_g';
};

export const formatFiltersForPayload = (
  filters = [],
  caller = 'user_profiles'
) => {
  const filterProps = [];
  filters.forEach((filter) => {
    const { values, props, operator } = filter;
    const vals = Array.isArray(values) ? filter.values : [filter.values];

    vals.forEach((val, index) => {
      filterProps.push({
        en: getEntityName(caller, props[0]),
        lop: index === 0 ? 'AND' : 'OR',
        op: operatorMap[operator],
        grpn: props[0],
        pr: props[1],
        ty: props[2],
        va: val
      });
    });
  });

  return filterProps;
};

export const getSegmentQuery = (queries, queryOptions, userType, caller) => {
  const query = {};
  query.grpa = queryOptions?.group_analysis;
  query.source = queryOptions?.source;
  query.caller = queryOptions?.caller;
  query.table_props = queryOptions?.table_props;
  query.cl = QUERY_TYPE_EVENT;
  query.ty = TYPE_UNIQUE_USERS;

  const period = {};
  if (queryOptions.date_range.from && queryOptions.date_range.to) {
    period.from = MomentTz(queryOptions.date_range.from).utc().unix();
    period.to = MomentTz(queryOptions.date_range.to).utc().unix();
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
  query.gup = formatFiltersForPayload(queryOptions?.globalFilters, caller);

  query.ec = EVENT_QUERY_USER_TYPE[userType];
  query.tz = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  return query;
};

export const IsDomainGroup = (source) =>
  source === GROUP_NAME_DOMAINS || source === 'All';

export const getFiltersRequestPayload = ({
  selectedFilters,
  tableProps,
  caller = 'account_profiles'
}) => {
  const { eventsList, eventProp, filters, account, secondaryFilters } =
    selectedFilters;

  const queryOptions = {
    group_analysis: account[1],
    source: caller === 'account_profiles' ? account[1] : 'All',
    caller,
    table_props: tableProps,
    globalFilters: [...filters, ...secondaryFilters],
    date_range: {}
  };

  return {
    query: getSegmentQuery(eventsList, queryOptions, eventProp, caller)
  };
};

export const formatReqPayload = (payload) => {
  let req = { query: { source: payload.source } };

  if (payload.segment) {
    const { query = {} } = payload.segment;
    req.query = {
      grpa: query.grpa || '',
      source: payload.source,
      ty: query.ty || '',
      ec: query.ec || '',
      ewp: query.ewp || [],
      gup: query.gup || [],
      table_props: query.table_props || []
    };
  }

  if (payload?.search_filter?.length) {
    req = { ...req, search_filter: [...payload.search_filter] };
  }

  return req;
};

export const eventsFormattedForGranularity = (
  events,
  granularity,
  collapse = true
) => {
  const output = events.reduce((result, item) => {
    const byTimestamp = (result[groups[granularity](item)] =
      result[groups[granularity](item)] || {});
    const byUser = (byTimestamp[item.user] = byTimestamp[item.user] || {
      events: [],
      collapsed: collapse
    });
    byUser.events.push(item);
    return result;
  }, {});
  return output;
};

export const eventsGroupedByGranularity = (events, granularity) => {
  const groupedEvents = events.reduce((result, item) => {
    const timestampKey = groups[granularity](item);

    if (!result[timestampKey]) {
      result[timestampKey] = [];
    }

    result[timestampKey].push(item);

    return result;
  }, {});

  return groupedEvents;
};

export const toggleCellCollapse = (
  formattedData,
  timestamp,
  username,
  collapseState
) => {
  const data = { ...formattedData };
  data[timestamp][username].collapsed = collapseState;
  return data;
};

const isValidHttpUrl = (string) => {
  let url;
  try {
    url = new URL(string);
  } catch (_) {
    return false;
  }
  return url.protocol === 'http:' || url.protocol === 'https:';
};

export const getHost = (urlstr) => {
  const uri = isValidHttpUrl(urlstr) ? new URL(urlstr).hostname : urlstr;
  return uri;
};

export const getUniqueItemsByKeyAndSearchTerm = (
  activities,
  searchTerm = ''
) => {
  if (!activities) {
    return [];
  }

  const isNotMilestone = (event) =>
    event && event.user !== 'milestone' && event.event_type !== 'milestone';

  const isUnique = (value, index, self) =>
    index === self.findIndex((t) => t && t.display_name === value.display_name);

  const matchesSearchTerm = (value) =>
    value &&
    value.display_name &&
    value.display_name.toLowerCase().includes(searchTerm.toLowerCase());

  return activities
    .filter(isNotMilestone)
    .filter(isUnique)
    .filter(matchesSearchTerm);
};

export const getPropType = (propsList, searchProp) => {
  let propType = 'categorical';
  propsList?.forEach((propArr) => {
    if (propArr[1] === searchProp) {
      propType = propArr[2];
    }
  });
  return propType;
};

export const propValueFormat = (searchKey, value, type) => {
  if (!value) return '-';

  const isDate = searchKey?.toLowerCase()?.includes('date');
  const isNumDuration = searchKey?.toLowerCase()?.includes('time');
  const isCatDuration = searchKey?.endsWith('time');
  const isDurationMilliseconds = searchKey?.includes('durationmilliseconds');
  const isTimestamp = searchKey?.includes('timestamp');

  const formatDatetime = (value) => {
    const dateFormat = isDate ? 'DD MMM YYYY' : 'DD MMM YYYY, hh:mm A zz';
    return MomentTz(value * 1000).format(dateFormat);
  };

  const formatNumerical = (value) => {
    let localeParam;
    if (searchKey === '$6Signal_annual_revenue') {
      localeParam = {
        style: 'currency',
        currency: 'USD',
        minimumFractionDigits: 0,
        maximumFractionDigits: 0
      };
    }
    return isNumDuration
      ? formatDurationIntoString(parseFloat(value))
      : isDurationMilliseconds
        ? formatDurationIntoString(parseFloat(value / 1000))
        : parseInt(value).toLocaleString('en-US', localeParam);
  };

  const formatCategorical = (value) =>
    isTimestamp
      ? MomentTz(value * 1000).format('DD MMM YYYY, hh:mm A zz')
      : isCatDuration
        ? formatDurationIntoString(parseFloat(value))
        : value;

  switch (type) {
    case 'datetime':
      return !isNaN(parseInt(value)) ? formatDatetime(value) : value;

    case 'numerical':
      return !isNaN(parseInt(value)) ? formatNumerical(value) : value;

    case 'categorical':
      return formatCategorical(value);

    default:
      return value;
  }
};

export const getEventCategory = (event, eventNamesMap) => {
  let category = 'others';
  Object.entries(eventNamesMap).forEach(([groupName, events]) => {
    if (events.includes(event.event_name)) {
      category = groupName;
    }
  });
  if (event.display_name === 'Page View') {
    category = 'website';
  }
  return category;
};

export const getIconForCategory = (category) => {
  const source = category.toLowerCase();

  if (source.includes('hubspot')) {
    return 'hubspot';
  }
  if (source.includes('salesforce')) {
    return 'salesforce';
  }
  if (source.includes('leadsquared')) {
    return 'leadsquared';
  }
  if (source.includes('marketo')) {
    return 'marketo';
  }
  if (source === 'website') {
    return 'globe';
  }
  return 'mouseclick';
};

export const DefaultDateRangeForSegments = {
  from: MomentTz().subtract(28, 'days').startOf('day'),
  to: MomentTz().subtract(1, 'days').endOf('day'),
  frequency: MomentTz().format('dddd') === 'Monday' ? 'hour' : 'date',
  dateType:
    MomentTz().format('dddd') === 'Sunday'
      ? PREDEFINED_DATES.LAST_MONTH
      : PREDEFINED_DATES.THIS_MONTH
};

export const timestampToString = {
  Timestamp: (item) => MomentTz(item * 1000).format('DD MMM YYYY, hh:mm:ss A'),
  Hourly: (item) =>
    `${MomentTz(item * 1000)
      .startOf('hour')
      .format('hh A')} - ${MomentTz(item * 1000)
      .add(1, 'hour')
      .startOf('hour')
      .format('hh A')} ${MomentTz(item * 1000)
      .startOf('hour')
      .format('DD MMM YYYY')}`,
  Daily: (item) =>
    MomentTz(item * 1000)
      .startOf('day')
      .format('DD MMM YYYY'),

  Weekly: (item) =>
    `${MomentTz(item * 1000)
      .startOf('week')
      .format('DD MMM YYYY')} - ${MomentTz(item * 1000)
      .endOf('week')
      .format('DD MMM YYYY')}`,
  Monthly: (item) =>
    MomentTz(item * 1000)
      .startOf('month')
      .format('MMM YYYY')
};

export const sortStringColumn = (a = '', b = '') => {
  const compareA = typeof a === 'string' ? a.toLowerCase() : a;
  const compareB = typeof b === 'string' ? b.toLowerCase() : b;
  return compareA > compareB ? 1 : compareB > compareA ? -1 : 0;
};

export const sortNumericalColumn = (a = 0, b = 0) => a - b;

export const transformPayloadForWeightConfig = (payload) => {
  const { key: wid, label: event_name, weight, vr, filters, fname } = payload;
  const output = {
    wid,
    event_name,
    weight,
    is_deleted: false,
    rule: [],
    vr: vr === 0 ? 0 : 1,
    fname
  };

  if (filters?.length) {
    filters.forEach((filter) => {
      const { props, operator, values } = filter;
      const [property_type, key, value_type] = props;
      const rule = {
        key,
        value: value_type === 'categorical' ? values : [],
        operator: operatorMap[operator] || operator,
        property_type,
        value_type,
        lower_bound: value_type === 'numerical' ? parseInt(values) : 0
      };
      output.rule.push(rule);
    });
  } else {
    output.rule = null;
  }

  return output;
};

export const transformWeightConfigForQuery = (config) => {
  const { wid: key, event_name: label, weight, vr, rule, fname } = config;
  const output = {
    key,
    label,
    weight,
    filters: [],
    vr,
    fname
  };

  if (rule) {
    const rules = Array.isArray(rule) ? rule : [rule];

    rules.forEach((rule) => {
      const { value, value_type, lower_bound, property_type, key, operator } =
        rule;
      const ruleValues =
        Array.isArray(value) && value.length > 0
          ? value
          : value_type === 'categorical'
            ? [value]
            : value_type === 'numerical'
              ? lower_bound
              : value;
      const filter = {
        props: [property_type, key, value_type, property_type],
        operator: reverseOperatorMap[operator] || operator,
        values: ruleValues,
        ref: 1
      };
      output.filters.push(filter);
    });
  }

  return output;
};

export const getSelectedFiltersFromQuery = ({
  query,
  groupsList,
  caller = 'account_profiles'
}) => {
  const eventProp =
    REVERSE_USER_TYPES[query.ec] != null
      ? REVERSE_USER_TYPES[query.ec]
      : ANY_USER_TYPE;
  const grpa = Boolean(query.grpa) === true ? query.grpa : GROUP_NAME_DOMAINS;
  const filters = getStateQueryFromRequestQuery(query);
  const result = {
    eventProp,
    filters:
      caller === 'account_profiles'
        ? filters.globalFilters.filter((elem) => elem.props[0] !== 'user')
        : filters.globalFilters,
    eventsList: filters.events,
    account:
      caller === 'account_profiles'
        ? groupsList.find((g) => g[1] === grpa)
        : INITIAL_USER_PROFILES_FILTERS_STATE.account,
    eventTimeline: '7',
    secondaryFilters:
      caller === 'account_profiles'
        ? filters.globalFilters.filter((elem) => elem.props[0] === 'user')
        : []
  };
  return result;
};

export const findKeyByValue = (data, targetValue) => {
  let foundKey = null;

  Object.entries(data).forEach(([key, values]) => {
    if (values.includes(targetValue)) {
      foundKey = key;
    }
  });

  return foundKey;
};
