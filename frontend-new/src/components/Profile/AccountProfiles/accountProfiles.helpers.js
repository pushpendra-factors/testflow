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
  <TextWithOverflowTooltip
    text={title}
    extraClass={`font-bold ${extraClass}`}
  />
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
    <TextWithOverflowTooltip
      alwaysShowTooltip
      text={urlTruncatedValue}
      tooltipText={formattedValue}
    />
  );
};

const EngagementSignalTag = ({ eventName, score, displayNames }) => (
  <div className='inline-flex gap-x-1 h-6'>
    <SVG name='event' size={16} />
    <TextWithOverflowTooltip
      text={displayNames[eventName] || PropTextFormat(eventName)}
      extraClass='font-normal'
    />
    <span className='font-normal'> {`(${parseInt(score)})`}</span>
  </div>
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
      render: (engagement) => {
        if (!engagement) return '-';
        const engagementLevel = engagement.$engagement_level;
        const engagementScore = engagement.$engagement_score;
        const signalsArr = engagement.$top_enagagement_signals.split(' , ');

        const renderTag = (item) => {
          const [eventName, score] = item.trim().split(/ (.+)$/);

          return (
            <EngagementSignalTag
              displayNames={eventNames}
              eventName={eventName}
              score={score}
            />
          );
        };

        const popoverContent = (
          <div className='flex flex-col gap-y-2'>
            <div className='text-xs font-medium'>Engagement Score</div>
            <div className='flex items-center gap-x-2'>
              <img
                src={`../../../assets/icons/${EngagementTag[engagementLevel]?.icon}.svg`}
                alt=''
                className='w-7 h-7'
              />
              <span className='font-semibold text-base'>{engagementScore}</span>
            </div>
            <div className='text-xs font-medium'>Engagement Signals</div>
            <div className='flex flex-col'>{signalsArr.map(renderTag)}</div>
          </div>
        );

        return (
          <Popover content={popoverContent} placement='bottom'>
            <div
              className='engagement-tag'
              style={{
                background: EngagementTag[engagementLevel]?.bgColor,
                color: EngagementTag[engagementLevel]?.textColor
              }}
            >
              <img
                src={`../../../assets/icons/${EngagementTag[engagementLevel]?.icon}.svg`}
                alt=''
              />
              {engagementLevel}
            </div>
          </Popover>
        );
      }
    };
  }

  if (prop === '$top_enagagement_signals') {
    return {
      title: getTitleText({ title: propDisplayName, extraClass: 'px-4' }),
      width: COLUMN_TYPE_PROPS.string.max,
      type: 'actions',
      dataIndex: prop,
      key: prop,
      render: (value) => {
        if (!value) return '';

        const eventsArr = value.split(' , ');
        const topEvent = eventsArr[0];
        const otherEvents = eventsArr.slice(1);

        const renderTag = (item) => {
          const splitItem = item.trim().split(' ');
          const eventName = splitItem.slice(0, -1).join(' ');
          const score = splitItem.slice(-1)[0];

          return (
            <EngagementSignalTag
              displayNames={eventNames}
              eventName={eventName}
              score={score}
            />
          );
        };

        const renderPopoverContent = (events) => (
          <div className='flex flex-col'>
            <Text type='title' level={8} color='grey'>
              Engagement Signals
            </Text>
            {events.map(renderTag)}
          </div>
        );

        return (
          <div className={`${styles.top_eng_names} flex items-center gap-x-2`}>
            {renderTag(topEvent)}
            {otherEvents.length > 0 && (
              <Popover
                content={renderPopoverContent(otherEvents)}
                placement='bottom'
              >
                <Tag
                  onClick={(e) => e.stopPropagation()}
                  color='default'
                  className={styles['tag-enagagementrule']}
                >
                  +{otherEvents.length}
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
        <div className='flex items-center gap-x-2'>
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
            linkTo={`/profiles/accounts/${btoa(domain.identity || domain.id)}`}
            onClick={(e) => e.preventDefault() && e.stopPropagation()}
            active={
              previewState?.drawerVisible &&
              previewState?.domain?.name === domain.name
            }
            activeClass='active-link'
          />
          <div className='preview-btns'>
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
