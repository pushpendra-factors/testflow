import React from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import isEqual from 'lodash/isEqual';
import { PropTextFormat } from 'Utils/dataFormatter';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import truncateURL from 'Utils/truncateURL';
import { Button, Popover, Tag, Tooltip } from 'antd';
import { ACCOUNTS_TABLE_COLUMN_TYPES, COLUMN_TYPE_PROPS } from 'Utils/table';
import TextWithOverflowTooltip from 'Components/GenericComponents/TextWithOverflowTooltip';
import { EngagementTag, headerClassStr, placeholderIcon } from '../constants';
import {
  flattenObjects,
  getHost,
  getPropType,
  propValueFormat,
  sortNumericalColumn,
  sortStringColumn
} from '../utils';
import styles from './index.module.scss';

export const defaultSegmentsList = [
  'In Hubspot',
  'In Salesforce',
  'Visited Website',
  'Engaged on LinkedIn',
  'Visited G2'
];

export const reorderDefaultDomainSegmentsToTop = (segments = []) => {
  const defaultSegments = segments
    .filter((segment) => defaultSegmentsList.includes(segment.name))
    .sort((a, b) => a.name.localeCompare(b.name));
  const createdSegments = segments
    .filter((segment) => !defaultSegmentsList.includes(segment.name))
    .sort((a, b) => a.name.localeCompare(b.name));
  return defaultSegments.concat(createdSegments);
};

export const getGroupList = (groupOptions) => {
  const groups = Object.entries(groupOptions || {}).map(
    ([groupName, displayName]) => [displayName, groupName]
  );
  groups.unshift(['All Accounts', GROUP_NAME_DOMAINS]);
  return groups;
};

const getTitleText = ({ title, extraClass = '' }) => (
  <Text
    type='title'
    level={7}
    color='grey-2'
    weight='bold'
    extraClass={`m-0 truncate ${extraClass}`}
  >
    {title}
  </Text>
);

export const renderValue = (
  value,
  propType,
  prop,
  domainsList,
  isText = false
) => {
  const formattedValue = propValueFormat(prop, value, propType) || '-';
  const urlTruncatedValue = truncateURL(formattedValue, domainsList);
  if (isText) return `"${formattedValue}"`;
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
};

const EngagementSignalTag = ({ eventName, score, displayNames }) => (
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
      {displayNames[eventName] || PropTextFormat(eventName)}
    </Text>
    <span className={styles['tag-seperator']}>|</span>
    {parseInt(score)}
  </Tag>
);

const getTablePropColumn = ({
  prop,
  groupPropNames,
  eventNames,
  listProperties,
  projectDomainsList
}) => {
  const mergedGroupPropNames = flattenObjects(groupPropNames);
  const propDisplayName = mergedGroupPropNames[prop]
    ? mergedGroupPropNames[prop]
    : PropTextFormat(prop);
  const propType = getPropType(listProperties, prop);

  if (prop === '$engagement_level') {
    return {
      title: getTitleText({ title: propDisplayName }),
      width:
        COLUMN_TYPE_PROPS[ACCOUNTS_TABLE_COLUMN_TYPES[prop]?.Type || 'string']
          ?.min || 264,
      dataIndex: prop,
      key: prop,
      showSorterTooltip: null,
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
  }

  if (prop === '$top_enagagement_signals') {
    return {
      title: getTitleText({ title: propDisplayName, extraClass: 'p-4' }),
      width: COLUMN_TYPE_PROPS.string.max,
      type: 'actions',
      dataIndex: prop,
      key: prop,
      render: (value) => {
        const eventsArr = value?.split(' , ');
        const renderTags = (events) =>
          events.map((item) => {
            const splitItem = item.trim().split(' ');
            const eventName = splitItem.slice(0, -1).join(' ');
            const score = splitItem.slice(-1);
            return (
              <EngagementSignalTag
                displayNames={eventNames}
                eventName={eventName}
                score={score}
              />
            );
          });

        return (
          <div className={styles.top_eng_names}>
            {value && renderTags(eventsArr.slice(0, 2))}
            {value && eventsArr.length > 2 && (
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
                    {renderTags(eventsArr.slice(2))}
                  </div>
                }
              >
                <Tag color='default' className={styles['tag-enagagementrule']}>
                  <span>and </span> +{eventsArr.length - 2}
                </Tag>
              </Popover>
            )}
          </div>
        );
      }
    };
  }

  return {
    title: getTitleText({ title: propDisplayName, extraClass: 'capitalize' }),
    dataIndex: prop,
    key: prop,
    width:
      COLUMN_TYPE_PROPS[ACCOUNTS_TABLE_COLUMN_TYPES[prop]?.Type || 'string']
        ?.min || 264,
    align: ACCOUNTS_TABLE_COLUMN_TYPES[prop]?.Align || 'left',
    showSorterTooltip: null,
    sorter: (a, b) =>
      propType === 'numerical'
        ? sortNumericalColumn(a[prop], b[prop])
        : sortStringColumn(a[prop], b[prop]),
    render: (value) => renderValue(value, propType, prop, projectDomainsList)
  };
};

export const getColumns = ({
  displayTableProps,
  groupPropNames,
  eventNames,
  listProperties,
  defaultSorterInfo,
  projectDomainsList,
  onClickOpen,
  onClickOpenNewTab,
  previewState
}) => {
  const accountColumn = {
    title: <div className={headerClassStr}>Account Domain</div>,
    dataIndex: 'domain',
    key: 'domain',
    width: 264,
    type: 'string',
    fixed: 'left',
    showSorterTooltip: null,
    sorter: (a, b) => sortStringColumn(a.domain.name, b.domain.name),
    render: (domain) =>
      (
        <div className='flex items-center justify-between'>
          <div className='inline-flex gap--8' id={domain.id}>
            <img
              src={`https://logo.clearbit.com/${getHost(domain.name)}`}
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
            <TextWithOverflowTooltip
              alwaysShowTooltip
              text={domain.name}
              hasLink
              linkTo={`/profiles/accounts/${btoa(
                domain.identity || domain.id
              )}`}
              onClick={(e) => e.preventDefault() && e.stopPropagation()}
              active={
                previewState?.drawerVisible &&
                previewState?.domain?.name === domain.name
              }
              activeClass='active-link'
            />
          </div>
          <div className='inline-flex gap--4 preview-btns'>
            <Tooltip title='Open'>
              <Button
                size='small'
                onClick={(e) => {
                  e.stopPropagation();
                  onClickOpen(domain);
                }}
                className='flex items-center'
                style={{ padding: '0 4px' }} // temp styling
              >
                <SVG name='expand' />
              </Button>
            </Tooltip>
            <Tooltip title='Open in new tab'>
              <Button
                onClick={(e) => {
                  e.stopPropagation();
                  onClickOpenNewTab(domain);
                }}
                size='small'
                className='flex items-center'
                style={{ padding: '0 6px' }} // temp styling
              >
                <SVG name='ArrowUpRightSquare' size='12' />
              </Button>
            </Tooltip>
          </div>
        </div>
      ) || '-'
  };

  const tablePropColumns = displayTableProps?.map((prop) =>
    getTablePropColumn({
      headerClassStr,
      prop,
      groupPropNames,
      eventNames,
      listProperties,
      projectDomainsList
    })
  );

  const lastActivityColumn = {
    title: <div className={headerClassStr}>Last Activity</div>,
    dataIndex: 'last_activity',
    key: 'last_activity',
    width: 224,
    type: 'datetime',
    align: 'left',
    sorter: (a, b) => sortStringColumn(a.last_activity, b.last_activity),
    render: (item) => MomentTz(item).fromNow()
  };

  let columns = [accountColumn, ...tablePropColumns, lastActivityColumn];

  columns = columns.map((column) => {
    const updatedColumn = {
      ...column
    };
    if (updatedColumn.key === defaultSorterInfo?.key) {
      updatedColumn.sortOrder = defaultSorterInfo?.order;
    } else {
      delete updatedColumn.sortOrder;
    }
    return updatedColumn;
  });

  const hasSorter = columns.some((item) =>
    ['ascend', 'descend'].includes(item.sortOrder)
  );

  if (!hasSorter) {
    columns = columns.map((column) => {
      const updatedColumn = {
        ...column
      };
      if (['$engagement_level', 'last_activity'].includes(updatedColumn.key)) {
        updatedColumn.defaultSortOrder = 'descend';
      }
      return updatedColumn;
    });
  }

  return columns;
};

export const checkFiltersEquality = ({
  appliedFilters,
  selectedFilters,
  newSegmentMode,
  areFiltersDirty,
  isActiveSegment
}) => {
  if (newSegmentMode && selectedFilters.filters.length > 0) {
    return {
      saveButtonDisabled: false,
      applyButtonDisabled: false
    };
  }

  const areFiltersEqual = isEqual(
    selectedFilters.filters,
    appliedFilters.filters
  );
  const areSecondaryFiltersEqual = isEqual(
    selectedFilters.secondaryFilters,
    appliedFilters.secondaryFilters
  );
  const areEventsEqual = isEqual(
    selectedFilters.eventsList,
    appliedFilters.eventsList
  );
  const isEventPropEqual =
    selectedFilters.eventProp === appliedFilters.eventProp;

  const applyButtonDisabled =
    areFiltersEqual &&
    areSecondaryFiltersEqual &&
    areEventsEqual &&
    isEventPropEqual;

  const noSelectedFilters =
    selectedFilters.filters.length === 0 &&
    selectedFilters.eventsList.length === 0 &&
    selectedFilters.secondaryFilters.length === 0;

  const saveButtonDisabled = isActiveSegment
    ? noSelectedFilters || !areFiltersDirty
    : noSelectedFilters;

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
