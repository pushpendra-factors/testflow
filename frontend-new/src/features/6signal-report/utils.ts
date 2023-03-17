import React from 'react';
import MomentTz from 'Components/MomentTz';
import { getClickableTitleSorter, SortResults } from 'Utils/dataFormatter';
import {
  SESSION_SPENT_TIME,
  KEY_LABELS,
  PAGE_COUNT_KEY,
  CHANNEL_KEY,
  CHANNEL_QUICK_FILTERS,
  COMPANY_KEY,
  CAMPAIGN_KEY,
  SHARE_QUERY_PARAMS,
  DEFAULT_COLUMNS,
  EMP_RANGE_KEY,
  REVENUE_RANGE_KEY,
  INDUSTRY_KEY
} from './const';
import { ResultGroup, StringObject, WeekStartEnd, ShareApiData } from './types';
import momentTz from 'moment-timezone';
import { intersection } from 'lodash';
import TableCell from './ui/components/ReportTable/TableCell';

export const generateFirstAndLastDayOfLastWeeks = (
  n: number = 5
): WeekStartEnd[] => {
  const lastWeek = MomentTz().subtract(7, 'd');
  let dateArray: WeekStartEnd[] = [];
  for (let i = 0; i < n; i++) {
    const day = MomentTz(lastWeek).subtract(7 * i, 'd');
    const weekStart = day.clone().startOf('week');
    const weekEnd = day.clone().endOf('week');
    const formattedRangeOption = `${weekStart.format(
      'MMM Do'
    )} to ${weekEnd.format('MMM Do')}`;
    const formattedRange = `${weekStart.format('MMM D, Y')} - ${weekEnd.format(
      'MMM D, Y'
    )}`;
    dateArray.push({
      from: weekStart.unix(),
      to: weekEnd.unix(),
      formattedRange,
      formattedRangeOption
    });
  }
  return dateArray;
};

export const getFormattedRange = (
  from: number,
  to: number,
  timezone: string = 'Asia/Kolkata'
) => {
  const fromDay = momentTz.unix(from).tz(timezone);
  const toDay = momentTz.unix(to).tz(timezone);

  return `${fromDay.format('MMM D, Y')} - ${toDay.format('MMM D, Y')}`;
};

export const parseResultGroupResponse = ({
  headers,
  rows
}: ResultGroup): {
  campaigns: string[];
  channels: string[];
  channelIndex: number;
  campaignIndex: number;
} => {
  const returnValue = {
    campaigns: [],
    channels: [],
    channelIndex: 0,
    campaignIndex: 0
  };

  const campaignKeyIndex = headers.findIndex(
    (header) => header === CAMPAIGN_KEY
  );
  const channelKeyIndex = headers.findIndex((header) => header === CHANNEL_KEY);
  if (!rows || !rows.length) return returnValue;
  let uniqueCampains = new Set<string>();
  let uniqueChannels = new Set<string>();
  rows.forEach((row) => {
    if (row) {
      if (row?.[campaignKeyIndex]) uniqueCampains.add(row[campaignKeyIndex]);
      if (row?.[channelKeyIndex]) uniqueChannels.add(row[channelKeyIndex]);
    }
  });
  return {
    campaigns: Array.from(uniqueCampains),
    channels: Array.from(uniqueChannels),
    campaignIndex: campaignKeyIndex,
    channelIndex: channelKeyIndex
  };
};

export const getSortType = (header: string) => {
  if (header === SESSION_SPENT_TIME || header === PAGE_COUNT_KEY) {
    return 'numerical';
  } else if (header === EMP_RANGE_KEY || header === REVENUE_RANGE_KEY) {
    return 'rangeNumeric';
  }
  return 'categorical';
};

const getColumnWidth = (header: string) => {
  if ([COMPANY_KEY, INDUSTRY_KEY].includes(header)) {
    const windowSize = window.innerWidth;
    if (header === INDUSTRY_KEY)
      return {
        width: 200
      };
    if (windowSize <= 1440) {
      return {
        width: 260
      };
    } else {
      return {
        width: 350
      };
    }
  }
  return {};
};

export const getTableColumuns = (
  data: ResultGroup,
  sorter: any,
  handleSorting: (sorter: string) => void
) => {
  const { headers } = data;
  const tColumns = intersection(DEFAULT_COLUMNS, headers)
    .map((header: string, i: number) => {
      let returnObj = {
        key: i,
        dataIndex: header,
        title: getClickableTitleSorter(
          // @ts-ignore
          KEY_LABELS?.[header] || header,
          {
            key: header,
            type: getSortType(header),
            subtype: null
          },
          sorter,
          handleSorting,
          'left',
          'center',
          'p-0 m-0 text-ellipsis'
        ),
        render: (text: string, record: StringObject) =>
          // @ts-ignore
          React.createElement(TableCell, {
            text: text,
            record: record,
            header: header
          }),
        ...getColumnWidth(header)
      };

      return returnObj;
    })
    .filter((d) => !!d);

  return tColumns;
};

export const getTableData = (
  data: ResultGroup,
  searchText: string,
  selectedChannel: string,
  selectedCampaigns: string[],
  sorter: any
) => {
  const { rows, headers } = data;
  let dataSource: StringObject[] = rows?.map((row, i) => {
    let rowObj: StringObject = {
      key: String(i)
    };
    headers.forEach((header, j) => {
      rowObj[header] = row[j];
    });

    return rowObj;
  });
  //filtering table data with selected Channel filter
  if (selectedChannel && selectedChannel !== CHANNEL_QUICK_FILTERS[0].id) {
    dataSource = dataSource?.filter((data: StringObject) =>
      data?.[CHANNEL_KEY] === selectedChannel ? true : false
    );
  }
  //filtering using selected campaigns
  if (
    selectedCampaigns &&
    Array.isArray(selectedCampaigns) &&
    selectedCampaigns?.length > 0
  ) {
    dataSource = dataSource?.filter((data: StringObject) =>
      selectedCampaigns.includes(data?.[CAMPAIGN_KEY])
    );
  }
  //filtering using search key
  if (searchText) {
    dataSource = dataSource?.filter((data: StringObject) =>
      data?.[COMPANY_KEY]?.toLowerCase()?.includes(searchText.toLowerCase())
    );
  }
  return SortResults(dataSource, sorter);
};

export const getDefaultTableColumns = () => {
  return DEFAULT_COLUMNS.map((key, index) => {
    return {
      index,
      dataIndex: key,
      title: getClickableTitleSorter(
        // @ts-ignore
        KEY_LABELS?.[key] || key,
        { key: key, type: 'categorical', subtype: null },
        {},
        () => {
          console.log('handle sorting');
        },
        'left',
        'center',
        'px-6 py-3'
      )
    };
  });
};

export const checkStringEquality = (
  str1: string,
  str2: string,
  caseSensitive: boolean = false
): boolean => {
  if (caseSensitive) return str1 === str2;
  return str1?.toLowerCase() === str2?.toLowerCase();
};

export const getPublicUrl = (obj: ShareApiData): string => {
  return (
    window.location.protocol +
    '//' +
    window.location.host +
    `/reports/6_signal?${SHARE_QUERY_PARAMS.queryId}=${obj.query_id}&${SHARE_QUERY_PARAMS.projectId}=${obj.project_id}&${SHARE_QUERY_PARAMS.routeVersion}=${obj.route_version}`
  );
};
