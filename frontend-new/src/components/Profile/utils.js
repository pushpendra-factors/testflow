import MomentTz from '../MomentTz';
import { formatDurationIntoString, PropTextFormat } from 'Utils/dataFormatter';
import {
  EVENT_QUERY_USER_TYPE,
  PREDEFINED_DATES,
  QUERY_TYPE_EVENT,
  RevAvailableGroups,
  ReverseProfileMapper,
  TYPE_UNIQUE_USERS
} from 'Utils/constants';
import { getGlobalFilters } from 'Views/PathAnalysis/PathAnalysisReport/QueryBuilder/utils';
import { getEventsWithProperties } from '../../Views/CoreQuery/utils';
import { generateRandomKey } from 'Utils/global';
import {
  operatorMap,
  reverseDateOperatorMap,
  reverseOperatorMap
} from 'Utils/operatorMapping';

export const granularityOptions = [
  'Timestamp',
  'Hourly',
  'Daily',
  'Weekly',
  'Monthly'
];

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
      .format('MMM YYYY')
};

export const hoverEvents = [
  '$session',
  '$form_submitted',
  '$offline_touch_point',
  '$sf_campaign_member_created',
  '$sf_campaign_member_updated',
  '$hubspot_form_submission',
  '$hubspot_engagement_email',
  '$hubspot_engagement_meeting_created',
  '$hubspot_engagement_meeting_updated',
  '$hubspot_engagement_call_created',
  '$hubspot_engagement_call_updated'
];

export const TimelineHoverPropDisplayNames = {
  $timestamp: 'Date and Time',
  '$hubspot_form_submission_form-type': 'Form Type',
  $hubspot_form_submission_title: 'Form Title',
  '$hubspot_form_submission_form-id': 'Form ID',
  '$hubspot_form_submission_conversion-id': 'Conversion ID',
  $hubspot_form_submission_email: 'Email',
  '$hubspot_form_submission_page-url-no-qp': 'Page URL',
  '$hubspot_form_submission_page-title': 'Page Title',
  $hubspot_form_submission_timestamp: 'Form Submit Timestamp'
};

export const displayFilterOpts = {
  All: 'All Accounts',
  $hubspot_company: 'Hubspot Companies',
  $salesforce_account: 'Salesforce Accounts'
};

export const formatFiltersForPayload = (filters = []) => {
  const filterProps = [];
  filters.forEach((fil) => {
    if (Array.isArray(fil.values)) {
      fil.values.forEach((val, index) => {
        filterProps.push({
          en: 'user_g',
          lop: !index ? 'AND' : 'OR',
          op: operatorMap[fil.operator],
          pr: fil.props[0],
          ty: fil.props[1],
          va: fil.props[1] === 'datetime' ? val : val
        });
      });
    } else {
      filterProps.push({
        en: 'user_g',
        lop: 'AND',
        op: operatorMap[fil.operator],
        pr: fil.props[0],
        ty: fil.props[1],
        va: fil.props[1] === 'datetime' ? fil.values : fil.values
      });
    }
  });
  return filterProps;
};

export const formatEventsFromSegment = (ewp) => {
  const events = ewp?.map((e) => {
    const filters = [];
    let ref = -1;
    let lastProp = '';
    let lastOp = '';
    e.pr.forEach((pr) => {
      if (pr.lop === 'AND') {
        ref += 1;
        filters.push({
          operator:
            pr.ty === 'datetime'
              ? reverseDateOperatorMap[pr.op]
              : reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.en],
          values: [pr.va],
          ref
        });
        lastProp = pr.pr;
        lastOp = pr.op;
      } else if (lastProp === pr.pr && lastOp === pr.op) {
        filters[filters.length - 1].values.push(pr.va);
      } else {
        filters.push({
          operator:
            pr.ty === 'datetime'
              ? reverseDateOperatorMap[pr.op]
              : reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.en],
          values: [pr.va],
          ref
        });
        lastProp = pr.pr;
        lastOp = pr.op;
      }
    });
    return {
      alias: e.an,
      label: e.na,
      group: e.grpa,
      filters,
      key: generateRandomKey()
    };
  });
  return events;
};

export const formatPayloadForFilters = (gp) => {
  const globalFilters = [];

  if (gp && Array.isArray(gp)) {
    let ref = -1;
    let lastProp = '';
    let lastOp = '';
    gp.forEach((pr) => {
      if (pr.lop === 'AND') {
        ref += 1;
        globalFilters.push({
          operator:
            pr.ty === 'datetime'
              ? reverseDateOperatorMap[pr.op]
              : reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.en],
          values: [pr.va],
          ref
        });
        lastProp = pr.pr;
        lastOp = pr.op;
      } else if (lastProp === pr.pr && lastOp === pr.op) {
        globalFilters[globalFilters.length - 1].values.push(pr.va);
      } else {
        globalFilters.push({
          operator:
            pr.ty === 'datetime'
              ? reverseDateOperatorMap[pr.op]
              : reverseOperatorMap[pr.op],
          props: [pr.pr, pr.ty, pr.en],
          values: [pr.va],
          ref
        });
        lastProp = pr.pr;
        lastOp = pr.op;
      }
    });
  }
  return globalFilters;
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

export const getUniqueItemsByKeyAndSearchTerm = (activities, searchTerm) => {
  return activities
    ?.filter((event) => event.user !== 'milestone')
    ?.filter((event) => event.event_type !== 'milestone')
    ?.filter(
      (value, index, self) =>
        index ===
          self.findIndex((t) => t.display_name === value.display_name) &&
        value.display_name.toLowerCase().includes(searchTerm.toLowerCase())
    );
};

export const getPropType = (propsList, searchProp) => {
  let propType = 'categorical';
  propsList.forEach((propArr) => {
    if (propArr[1] === searchProp) {
      propType = propArr[2];
    }
  });
  return propType;
};

export const propValueFormat = (searchKey, value, type) => {
  if (typeof value === 'string' && value.startsWith('$')) {
    return PropTextFormat(value);
  }
  switch (type) {
    case 'datetime':
      if (searchKey.toLowerCase().includes('date'))
        return MomentTz(value * 1000).format('DD MMM YYYY');
      else return MomentTz(value * 1000).format('DD MMM YYYY, hh:mm A zz');
    case 'numerical':
      if (searchKey.toLowerCase().includes('time'))
        return formatDurationIntoString(parseInt(value));
      else if (searchKey.includes('durationmilliseconds'))
        return formatDurationIntoString(parseInt(value / 1000));
      else return parseInt(value);
    case 'categorical':
      if (searchKey.includes('timestamp'))
        return MomentTz(value * 1000).format('DD MMM YYYY, hh:mm A zz');
      else if (searchKey.endsWith('time'))
        return formatDurationIntoString(parseInt(value));
      else return value;
    default:
      return value;
  }
};

export const formatSegmentsObjToGroupSelectObj = (group, vals) => {
  const obj = {
    label:
      ReverseProfileMapper[group]?.users ||
      RevAvailableGroups[group] ||
      'Others',
    icon: '',
    values: []
  };
  obj.values = vals?.map((val) => [
    val?.name,
    val?.id,
    {
      name: val?.name,
      description: val?.description,
      type: val?.type,
      query: val?.query
    }
  ]);
  return obj;
};
export const getEventCategory = (event, eventNamesMap) => {
  let category = 'others';
  Object.entries(eventNamesMap).forEach(([groupName, events]) => {
    if (events.includes(event.event_name)) {
      category = groupName;
      return;
    }
  });
  if (event.display_name === 'Page View') {
    category = 'website';
  }
  return category;
};

export const getIconForCategory = (category) => {
  if (category.toLowerCase().includes('hubspot')) {
    return 'hubspot';
  }
  if (category.toLowerCase().includes('salesforce')) {
    return 'salesforce';
  }
  if (category.toLowerCase().includes('leadsquared')) {
    return 'leadsquared';
  }
  if (category.toLowerCase().includes('marketo')) {
    return 'marketo';
  }
  if (category === 'website') {
    return 'globe';
  }
  return 'events_blue';
};

export const convertSVGtoURL = (svg = '') => {
  // svg needs to be passed with backticks
  const escapeRegExp = (str) => {
    return str.replace(/([.*+?^=!:${}()|\[\]\/\\])/g, '\\$1');
  };

  const replaceAll = (str, find, replace) => {
    return str.replace(new RegExp(escapeRegExp(find), 'g'), replace);
  };

  var encoded = svg.replace(/\s+/g, ' ');
  encoded = replaceAll(encoded, '%', '%25');
  encoded = replaceAll(encoded, '> <', '><');
  encoded = replaceAll(encoded, '; }', ';}');
  encoded = replaceAll(encoded, '<', '%3c');
  encoded = replaceAll(encoded, '>', '%3e');
  encoded = replaceAll(encoded, '"', "'");
  encoded = replaceAll(encoded, '#', '%23');
  encoded = replaceAll(encoded, '{', '%7b');
  encoded = replaceAll(encoded, '}', '%7d');
  encoded = replaceAll(encoded, '|', '%7c');
  encoded = replaceAll(encoded, '^', '%5e');
  encoded = replaceAll(encoded, '`', '%60');
  encoded = replaceAll(encoded, '@', '%40');

  var uri = 'url("data:image/svg+xml;charset=UTF-8,' + encoded + '")';
  return uri;
};

export const DEFAULT_TIMELINE_CONFIG = {
  disabled_events: [],
  user_config: {
    table_props: [],
    leftpane_props: [],
    milestones: []
  },
  account_config: {
    table_props: [],
    leftpane_props: [],
    milestones: [],
    user_prop: ''
  }
};

export const eventIconsColorMap = {
  brand: {
    iconColor: '#EE3C3C',
    bgColor: '#FAFAFA',
    borderColor: '#EEEEEE'
  },
  envelope: {
    iconColor: '#FF7875',
    bgColor: '#FFF4F4',
    borderColor: '#FFDEDE'
  },
  handshake: {
    iconColor: '#85A5FF',
    bgColor: '#EFF3FF',
    borderColor: '#D3DEFF'
  },
  phone: {
    iconColor: '#95DE64',
    bgColor: '#F0FFE7',
    borderColor: '#D5F4C1'
  },
  listcheck: {
    iconColor: '#5CDBD3',
    bgColor: '#EBFFFE',
    borderColor: '#C6F6F4'
  },
  'hand-pointer': {
    iconColor: '#FAAD14',
    bgColor: '#FFF3DB',
    borderColor: '#FBE5BA'
  },
  hubspot: {
    iconColor: '#FF7A59',
    bgColor: '#FFE8E2',
    borderColor: '#FED0C5'
  },
  salesforce: {
    iconColor: '#00A1E0',
    bgColor: '#E8F8FF',
    borderColor: '#CDF0FF'
  },
  window: {
    iconColor: '#FF85C0',
    bgColor: '#FFF0F7',
    borderColor: '#FFD9EB'
  },
  'calendar-star': {
    iconColor: '#B37FEB',
    bgColor: '#F6EDFF',
    borderColor: '#E9D4FF'
  }
};

export const iconColors = [
  '#85A5FF',
  '#B37FEB',
  '#5CDBD3',
  '#FF9C6E',
  '#FF85C0',
  '#FFC069',
  '#A0D911',
  '#FAAD14'
];

export const ALPHANUMSTR = '0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ';

export const DefaultDateRangeForSegments = {
  from: MomentTz().subtract(28, 'days').startOf('day'),
  to: MomentTz().subtract(1, 'days').endOf('day'),
  frequency: MomentTz().format('dddd') === 'Monday' ? 'hour' : 'date',
  dateType:
    MomentTz().format('dddd') === 'Sunday'
      ? PREDEFINED_DATES.LAST_MONTH
      : PREDEFINED_DATES.THIS_MONTH
};

export const getSegmentQuery = (queries, queryOptions, userType) => {
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
  query.gup = formatFiltersForPayload(queryOptions?.globalFilters);

  query.ec = EVENT_QUERY_USER_TYPE[userType];
  query.tz = localStorage.getItem('project_timeZone') || 'Asia/Kolkata';
  return query;
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

export const getElemenetHeight = (elem) => {
  const divElement = document.getElementById(elem);
  return divElement.offsetHeight;
};
