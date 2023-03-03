import React from 'react';
import { Text } from 'Components/factorsComponents';
import MomentTz from 'Components/MomentTz';
import {
  formatDuration,
  getClickableTitleSorter,
  SortResults
} from 'Utils/dataFormatter';
import {
  SESSION_SPENT_TIME,
  KEY_LABELS,
  PAGE_COUNT_KEY,
  CHANNEL_KEY,
  CHANNEL_QUICK_FILTERS,
  COMPANY_KEY,
  CAMPAIGN_KEY,
  SHARE_QUERY_PARAMS
} from './const';
import { ResultGroup, StringObject, WeekStartEnd, ShareApiData } from './types';
import moment from 'moment';

export const generateFirstAndLastDayOfLastWeeks = (
  n: number = 5
): WeekStartEnd[] => {
  const today = MomentTz();
  let dateArray: WeekStartEnd[] = [];
  for (let i = 0; i < n; i++) {
    const day = MomentTz(today).subtract(7 * i, 'd');
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

export const getFormattedRange = (from: number, to: number) => {
  const fromDay = moment(from);
  const toDay = moment(to);
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

export const getTableColumuns = (
  data: ResultGroup,
  sorter: any,
  handleSorting: (sorter: string) => void
) => {
  const { headers } = data;
  const tColumns = headers
    .map((header, i) => {
      let returnObj = {
        key: i,
        dataIndex: header,
        title: getClickableTitleSorter(
          // @ts-ignore
          KEY_LABELS?.[header] || header,
          {
            key: header,
            type:
              checkStringEquality(header, SESSION_SPENT_TIME) ||
              checkStringEquality(header, PAGE_COUNT_KEY)
                ? 'numerical'
                : 'categorical',
            subtype: null
          },
          sorter,
          handleSorting,
          'left',
          'center',
          'px-6 py-3'
        ),
        render: (d) => React.createElement(Text, { type: 'title', level: 7 }, d)
      };

      if (header === SESSION_SPENT_TIME) {
        returnObj.render = (d) =>
          React.createElement(
            Text,
            { type: 'title', level: 7 },
            formatDuration(d)
          );
      }
      if (header === PAGE_COUNT_KEY) {
        returnObj.render = (d) =>
          React.createElement(
            Text,
            { type: 'title', level: 7 },
            `${d} ${Number(d) > 1 ? 'Pages' : 'Page'}`
          );
      }
      if (header === CHANNEL_KEY) {
        return null;
      }
      // if (i === headers?.length - 1) {
      //   // @ts-ignore
      //   returnObj.render = (
      //     d // @ts-ignore
      //   ) => React.createElement(LastCell, { text: d }, null);
      // }

      return returnObj;
    })
    .filter((d) => !!d);

  // if (tColumns.length > 1) {
  //   // @ts-ignore
  //   tColumns[tColumns.length - 2].colSpan = 1;
  // }

  // tColumns.push({
  //   title: getClickableTitleSorter(
  //     // @ts-ignore
  //     '',
  //     {
  //       key: '',
  //       type: 'numerical',
  //       subtype: null
  //     },
  //     sorter,
  //     handleSorting,
  //     'left',
  //     'center',
  //     'px-6 py-3'
  //   ),
  //   dataIndex: 'action',
  //   render: (
  //     d // @ts-ignore
  //   ) => React.createElement(LastCell, { text: '' }, null),
  //   colSpan: 0
  // });

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
  return Object.keys(KEY_LABELS).map((key, index) => {
    return {
      index,
      dataIndex: KEY_LABELS[key],
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
