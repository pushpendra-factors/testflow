import React from 'react';
import {
  EngagementTag,
  formatSegmentsObjToGroupSelectObj,
  getHost,
  getPropType,
  IsDomainGroup,
  propValueFormat,
  sortNumericalColumn,
  sortStringColumn
} from '../utils';
import { Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import isEqual from 'lodash/isEqual';
import { PropTextFormat } from 'Utils/dataFormatter';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';

const placeholderIcon = '/assets/avatar/company-placeholder.png';

export const defaultSegmentsList = [
  'In Hubspot',
  'In Salesforce',
  'Visited Website',
  'Engaged on LinkedIn',
  'Visited G2'
];

const reorderDefaultSegmentsToTop = (segments) => {
  segments?.[0]?.values.sort((a, b) => {
    const aIsMatch = defaultSegmentsList.includes(a?.[0]);
    const bIsMatch = defaultSegmentsList.includes(b?.[0]);

    if (aIsMatch && !bIsMatch) {
      return -1;
    } else if (!aIsMatch && bIsMatch) {
      return 1;
    }

    return 0;
  });

  return segments;
};

export const getGroupList = (groupOptions) => {
  const groups = Object.entries(groupOptions || {}).map(
    ([group_name, display_name]) => [display_name, group_name]
  );
  groups.unshift(['All Accounts', GROUP_NAME_DOMAINS]);
  return groups;
};

export const generateSegmentsList = ({ segments }) => {
  const segmentsList = [];

  Object.entries(segments)
    .filter((segment) => segment[0] === GROUP_NAME_DOMAINS)
    .map(([group, vals]) => formatSegmentsObjToGroupSelectObj(group, vals))
    .forEach((obj) => segmentsList.push(obj));
  return reorderDefaultSegmentsToTop(segmentsList);
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
      <Text type='title' level={7} extraClass='m-0' truncate shouldTruncateURL>
        {value ? propValueFormat(prop, value, propType) : '-'}
      </Text>
    )
  };
};

export const getColumns = ({
  isScoringLocked,
  displayTableProps,
  groupPropNames,
  listProperties,
  defaultSorterInfo
}) => {
  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';
  const columns = [
    {
      // Company Name Column
      title: <div className={headerClassStr}>Account Domain</div>,
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
              src={`https://logo.clearbit.com/${getHost(item.host)}`}
              onError={(e) => {
                if (e.target.src !== placeholderIcon) {
                  e.target.src = placeholderIcon;
                }
              }}
              alt=''
              width='24'
              height='24'
              loading='lazy'
            />
            <span className='ml-2'>{item.name}</span>
          </div>
        ) || '-'
    }
  ];
  // Engagement Column

  if (!isScoringLocked) {
    columns.push(
      {
        title: <div className={headerClassStr}>Engagement</div>,
        width: 150,
        dataIndex: 'engagement',
        key: 'engagement',
        fixed: 'left',
        defaultSortOrder: 'descend',
        sorter: (a, b) => sortNumericalColumn(a.score, b.score),
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
      },
      {
        title: <div className={headerClassStr}>Score</div>,
        width: 150,
        dataIndex: 'score',
        key: 'score',
        defaultSortOrder: 'descend',
        sorter: (a, b) => sortNumericalColumn(a.score, b.score),
        render: (value) => (
          <Text type='title' level={7} extraClass='m-0'>
            {value ? value.toFixed() : '-'}
          </Text>
        )
      }
    );
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
    sorter: (a, b) => sortStringColumn(a.lastActivity, b.lastActivity),
    render: (item) => MomentTz(item).fromNow()
  });

  columns.forEach((column) => {
    if (column.key === defaultSorterInfo?.key) {
      column.sortOrder = defaultSorterInfo?.order;
    } else {
      delete column.sortOrder;
    }
  });
  const hasSorter = columns.find((item) =>
    ['ascend', 'descend'].includes(item.sortOrder)
  );
  if (!hasSorter) {
    columns.forEach((column) => {
      if (['engagement', 'lastActivity'].includes(column.key)) {
        column.defaultSortOrder = 'descend';
        return;
      }
    });
  }
  return columns;
};

export const checkFiltersEquality = ({
  appliedFilters,
  filtersList,
  newSegmentMode,
  eventsList,
  eventProp,
  areFiltersDirty,
  isActiveSegment,
  secondaryFiltersList
}) => {
  if (newSegmentMode === true && filtersList.length > 0) {
    return {
      saveButtonDisabled: false,
      applyButtonDisabled: false
    };
  }
  const areFiltersEqual = isEqual(filtersList, appliedFilters.filters);
  const areSecondaryFiltersEqual = isEqual(
    secondaryFiltersList,
    appliedFilters.secondaryFilters
  );
  const areEventsEqual = isEqual(eventsList, appliedFilters.eventsList);
  const isEventPropEqual = eventProp === appliedFilters.eventProp;
  const applyButtonDisabled =
    areSecondaryFiltersEqual &&
    areFiltersEqual === true &&
    areEventsEqual === true &&
    isEventPropEqual === true;
  const saveButtonDisabled =
    isActiveSegment === true
      ? (filtersList.length === 0 && eventsList.length === 0) ||
        areFiltersDirty === false
      : applyButtonDisabled === false ||
        (filtersList.length === 0 && eventsList.length === 0);
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
    if (IsDomainGroup(source)) {
      props[GROUP_NAME_DOMAINS] = groupProperties[GROUP_NAME_DOMAINS];
      Object.keys(availableGroups || {}).forEach((group) => {
        props[group] = groupProperties[group];
      });
    } else props[source] = groupProperties[source];
    // props.user = userProperties;
  } else if (profileType === 'user') {
    props.user = userProperties;
  }
  return props;
};
