import React from 'react';
import { Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import isEqual from 'lodash/isEqual';
import { PropTextFormat } from 'Utils/dataFormatter';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import truncateURL from 'Utils/truncateURL';
import { Popover, Tag } from 'antd';
import { ACCOUNTS_TABLE_COLUMN_TYPES, COLUMN_TYPE_PROPS } from 'Utils/table';
import { AdminLock } from 'Routes/feature';
import { EngagementTag } from '../constants';
import {
  getHost,
  getPropType,
  propValueFormat,
  sortNumericalColumn,
  sortStringColumn
} from '../utils';
import styles from './index.module.scss';

const placeholderIcon = '/assets/avatar/company-placeholder.png';

export const defaultSegmentsList = [
  'In Hubspot',
  'In Salesforce',
  'Visited Website',
  'Engaged on LinkedIn',
  'Visited G2'
];

export const reorderDefaultDomainSegmentsToTop = (segments = []) => {
  segments?.sort((a, b) => {
    const aIsMatch = defaultSegmentsList.includes(a?.name);
    const bIsMatch = defaultSegmentsList.includes(b?.name);

    if (aIsMatch && !bIsMatch) {
      return -1;
    }
    if (!aIsMatch && bIsMatch) {
      return 1;
    }

    return 0;
  });

  return segments;
};

export const getGroupList = (groupOptions) => {
  const groups = Object.entries(groupOptions || {}).map(
    ([groupName, displayName]) => [displayName, groupName]
  );
  groups.unshift(['All Accounts', GROUP_NAME_DOMAINS]);
  return groups;
};

const getTablePropColumn = ({
  prop,
  groupPropNames,
  listProperties,
  projectDomainsList
}) => {
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
        extraClass='m-0 truncate'
        truncate
        charLimit={25}
      >
        {propDisplayName}
      </Text>
    ),
    dataIndex: prop,
    key: prop,
    width:
      COLUMN_TYPE_PROPS[ACCOUNTS_TABLE_COLUMN_TYPES[prop]?.Type || 'string']
        ?.min || 264,
    showSorterTooltip: null,
    sorter: (a, b) =>
      propType === 'numerical'
        ? sortNumericalColumn(a[prop], b[prop])
        : sortStringColumn(a[prop], b[prop]),
    render: (value) => {
      const formattedValue = propValueFormat(prop, value, propType) || '-';
      const urlTruncatedValue = truncateURL(formattedValue, projectDomainsList);
      return (
        <Text
          type='title'
          level={7}
          extraClass='m-0'
          truncate
          toolTipTitle={formattedValue}
        >
          {urlTruncatedValue}
        </Text>
      );
    }
  };
};

export const getColumns = ({
  isScoringLocked,
  displayTableProps,
  groupPropNames,
  listProperties,
  defaultSorterInfo,
  projectDomainsList,
  activeAgent
}) => {
  const headerClassStr =
    'fai-text fai-text__color--grey-2 fai-text__size--h7 fai-text__weight--bold';

  const accountColumn = {
    title: <div className={headerClassStr}>Account Domain</div>,
    dataIndex: 'account',
    key: 'account',
    width: 264,
    type: 'string',
    fixed: 'left',
    ellipsis: true,
    sorter: (a, b) => sortStringColumn(a.account.name, b.account.name),
    render: (item) =>
      (
        <div className='flex items-center' id={item.name}>
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
          <span
            style={{
              textOverflow: 'ellipsis',
              overflow: 'hidden',
              whiteSpace: 'nowrap'
            }}
            className='ml-2'
          >
            <Text
              type='title'
              level={7}
              extraClass='truncate'
              truncate
              charLimit={25}
            >
              {item.name}
            </Text>
          </span>
        </div>
      ) || '-'
  };

  const engagementColumn = {
    title: <div className={headerClassStr}>Engagement</div>,
    width: 152,
    type: 'string',
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
  };

  const scoreColumn = {
    title: <div className={headerClassStr}>Score</div>,
    width: 152,
    type: 'number',
    dataIndex: 'score',
    key: 'score',
    defaultSortOrder: 'descend',
    sorter: (a, b) => sortNumericalColumn(a.score, b.score),
    render: (value) => (
      <Text type='title' level={7} extraClass='m-0'>
        {value ? value.toFixed() : '-'}
      </Text>
    )
  };

  const topEngagementsColumn = {
    title: <div className={headerClassStr}>Engagement Signals</div>,
    width: COLUMN_TYPE_PROPS.string.max,
    dataIndex: 'top_engagements',
    type: 'actions',
    key: 'top_engagements',
    render: (value) => (
      <div className={styles.top_eng_names}>
        {value &&
          Object.keys(value)
            .slice(0, 2)
            .map((eachKey, eachIndex) => (
              <Tag color='default' className={styles['tag-enagagementrule']}>
                <Text
                  type='title'
                  level={7}
                  color='grey-2'
                  extraClass='m-0 truncate'
                  truncate
                  size='h2'
                  charLimit={20}
                >
                  {eachKey}
                </Text>
                <span className={styles['tag-seperator']}>|</span>
                {value[eachKey]}
              </Tag>
            ))}
        {value && Object.keys(value).length > 2 ? (
          <Popover
            content={
              <div
                className='flex flex-col'
                onClick={(e) => {
                  e.stopPropagation();
                }}
              >
                <Text type='title' level={7} color='grey'>
                  Engagement Signals
                </Text>
                {Object.keys(value)
                  .slice(2)
                  .map((eachKey, eachIndex) => (
                    <Tag
                      color='default'
                      className={styles['tag-enagagementrule']}
                    >
                      <Text
                        type='title'
                        level={7}
                        color='grey-2'
                        extraClass='m-0 truncate'
                        truncate
                        size='h2'
                        charLimit={20}
                      >
                        {eachKey}
                      </Text>
                      <span className={styles['tag-seperator']}>|</span>
                      {value[eachKey]}
                    </Tag>
                  ))}
              </div>
            }
          >
            <Tag color='default' className={styles['tag-enagagementrule']}>
              <span>and </span> +{Object.keys(value).length - 2}
            </Tag>
          </Popover>
        ) : null}
      </div>
    )
  };

  const tablePropColumns = displayTableProps?.map((prop) =>
    getTablePropColumn({
      prop,
      groupPropNames,
      listProperties,
      projectDomainsList
    })
  );

  const lastActivityColumn = {
    title: <div className={headerClassStr}>Last Activity</div>,
    dataIndex: 'lastActivity',
    key: 'lastActivity',
    width: 224,
    type: 'datetime',
    align: 'left',
    sorter: (a, b) => sortStringColumn(a.lastActivity, b.lastActivity),
    render: (item) => MomentTz(item).fromNow()
  };

  const scoringColumns = [engagementColumn, scoreColumn];

  if (AdminLock(activeAgent)) {
    scoringColumns.push(topEngagementsColumn);
  }

  const columns = [
    accountColumn,
    ...(isScoringLocked ? [] : scoringColumns),
    ...tablePropColumns,
    lastActivityColumn
  ];

  columns.forEach((column) => {
    if (column.key === defaultSorterInfo?.key) {
      column.sortOrder = defaultSorterInfo?.order;
    } else {
      delete column.sortOrder;
    }
  });

  const hasSorter = columns.some((item) =>
    ['ascend', 'descend'].includes(item.sortOrder)
  );

  if (!hasSorter) {
    columns.forEach((column) => {
      if (['engagement', 'lastActivity'].includes(column.key)) {
        column.defaultSortOrder = 'descend';
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
      ? (filtersList.length === 0 &&
          eventsList.length === 0 &&
          secondaryFiltersList.length === 0) ||
        areFiltersDirty === false
      : applyButtonDisabled === false ||
        (filtersList.length === 0 &&
          eventsList.length === 0 &&
          secondaryFiltersList.length === 0);
  return { saveButtonDisabled, applyButtonDisabled };
};

export const computeFilterProperties = ({
  userProperties,
  groupProperties,
  availableGroups,
  profileType
}) => {
  const props = {};
  if (profileType === 'account') {
    props[GROUP_NAME_DOMAINS] = groupProperties[GROUP_NAME_DOMAINS];
    Object.keys(availableGroups || {}).forEach((group) => {
      props[group] = groupProperties[group];
    });
  } else if (profileType === 'user') {
    props.user = userProperties;
  }
  return props;
};
