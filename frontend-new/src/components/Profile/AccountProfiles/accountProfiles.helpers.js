import React from 'react';
import { ReverseProfileMapper } from 'Utils/constants';
import {
  EngagementTag,
  formatSegmentsObjToGroupSelectObj,
  getHost,
  getPropType,
  propValueFormat,
  sortNumericalColumn,
  sortStringColumn
} from '../utils';
import { Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import isEqual from 'lodash/isEqual';
import { PropTextFormat } from 'Utils/dataFormatter';

export const getGroupList = (groupOptions) => {
  const groups = Object.entries(groupOptions || {}).map(
    ([group_name, display_name]) => [display_name, group_name]
  );
  groups.unshift(['All Accounts', 'All']);
  return groups;
};

export const generateSegmentsList = ({ accountPayload, segments }) => {
  const segmentsList = [];

  Object.entries(segments)
    .filter(
      (segment) => !Object.keys(ReverseProfileMapper).includes(segment[0])
    )
    .map(([group, vals]) => formatSegmentsObjToGroupSelectObj(group, vals))
    .forEach((obj) => segmentsList.push(obj));
  return segmentsList;
};

const getTablePropColumn = ({ prop, groupPropNames, listProperties }) => {
  const propDisplayName = groupPropNames[prop]
    ? groupPropNames[prop]
    : PropTextFormat(prop);
  const propType = getPropType(listProperties, prop);
  return {
    title: (
      <Text
        type='title'
        level={7}
        color='grey-2'
        weight='bold'
        extraClass='m-0'
        truncate
        charLimit={25}
      >
        {propDisplayName}
      </Text>
    ),
    dataIndex: prop,
    key: prop,
    width: 280,
    sorter: (a, b) =>
      propType === 'numerical'
        ? sortNumericalColumn(a[prop], b[prop])
        : sortStringColumn(a[prop], b[prop]),
    render: (value) => (
      <Text type='title' level={7} extraClass='m-0' truncate>
        {value ? propValueFormat(prop, value, propType) : '-'}
      </Text>
    )
  };
};

export const getColumns = ({
  accounts,
  source,
  isEngagementLocked,
  displayTableProps,
  groupPropNames,
  listProperties
}) => {
  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';
  const columns = [
    {
      // Company Name Column
      title: (
        <div className={headerClassStr}>
          {source === 'All' ? 'Account Domain' : 'Company Name'}
        </div>
      ),
      dataIndex: 'account',
      key: 'account',
      width: 300,
      fixed: 'left',
      ellipsis: true,
      sorter: (a, b) => sortStringColumn(a.account.name, b.account.name),
      render: (item) =>
        (
          <div className='flex items-center'>
            <img
              src={`https://logo.uplead.com/${getHost(item.host)}`}
              onError={(e) => {
                if (
                  e.target.src !==
                  'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg'
                ) {
                  e.target.src =
                    'https://s3.amazonaws.com/www.factors.ai/assets/img/buildings.svg';
                }
              }}
              alt=''
              width='20'
              height='20'
            />
            <span className='ml-2'>{item.name}</span>
          </div>
        ) || '-'
    }
  ];
  // Engagement Column
  const engagementExists = accounts.data?.find(
    (item) =>
      item.engagement &&
      (item.engagement !== undefined || item.engagement !== '')
  );
  if (engagementExists && !isEngagementLocked) {
    columns.push({
      title: <div className={headerClassStr}>Engagement</div>,
      width: 150,
      dataIndex: 'engagement',
      key: 'engagement',
      fixed: 'left',
      defaultSortOrder: 'descend',
      sorter: {
        compare: (a, b) => sortNumericalColumn(a.score, b.score),
        multiple: 1
      },
      render: (status) =>
        status ? (
          <div
            className='engagement-tag'
            style={{ '--bg-color': EngagementTag[status]?.bgColor }}
          >
            <img
              src={`../../../assets/icons/${EngagementTag[status]?.icon}.svg`}
              alt=''
            />
            <Text type='title' level={7} extraClass='m-0'>
              {status}
            </Text>
          </div>
        ) : (
          '-'
        )
    });
  }
  // Table Prop Columns
  displayTableProps?.forEach((prop) => {
    columns.push(getTablePropColumn({ prop, groupPropNames, listProperties }));
  });
  // Last Activity Column
  columns.push({
    title: <div className={headerClassStr}>Last Activity</div>,
    dataIndex: 'lastActivity',
    key: 'lastActivity',
    width: 200,
    align: 'right',
    sorter: {
      compare: (a, b) => sortStringColumn(a.lastActivity, b.lastActivity),
      multiple: 2
    },
    render: (item) => MomentTz(item).fromNow()
  });
  return columns;
};

export const checkFiltersEquality = ({
  appliedFilters,
  filtersList,
  newSegmentMode,
  eventsList,
  eventProp
}) => {
  if (newSegmentMode === true && filtersList.length > 0) {
    return {
      saveButtonDisabled: false,
      applyButtonDisabled: false
    };
  }
  const areFiltersEqual = isEqual(filtersList, appliedFilters.filters);
  const areEventsEqual = isEqual(eventsList, appliedFilters.eventsList);
  const isEventPropEqual = eventProp === appliedFilters.eventProp;
  const saveButtonDisabled =
    areFiltersEqual === false || filtersList.length === 0;
  const applyButtonDisabled =
    areFiltersEqual === true && areEventsEqual && isEventPropEqual;
  return { saveButtonDisabled, applyButtonDisabled };
};

export const computeFilterProperties = ({
  userProperties,
  groupProperties,
  availableGroups,
  profileType,
  source
}) => {
  const props = {};
  if (profileType === 'account') {
    if (source === 'All') {
      Object.keys(availableGroups).forEach((group) => {
        props[group] = groupProperties[group];
      });
    } else props[source] = groupProperties[source];
    props.user = userProperties;
  } else if (profileType === 'user') {
    props.user = userProperties;
  }
  return props;
};
