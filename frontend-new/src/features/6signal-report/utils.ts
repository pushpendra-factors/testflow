import React from 'react';
import MomentTz from 'Components/MomentTz';
import { getClickableTitleSorter, SortResults } from 'Utils/dataFormatter';
import momentTz from 'moment-timezone';
import { intersection } from 'lodash';
import {
  DATE_RANGE_TODAY_LABEL,
  DATE_RANGE_YESTERDAY_LABEL,
  DATE_RANGE_LABEL_LAST_7_DAYS,
  getRangeByLabel
} from 'Components/FaDatepicker/utils';
import { PathUrls } from 'Routes/pathUrls';
import TableCell from './ui/components/ReportTable/TableCell';
import { ResultGroup, StringObject, WeekStartEnd, ShareApiData } from './types';
import {
  SESSION_SPENT_TIME,
  KEY_LABELS,
  PAGE_COUNT_KEY,
  CHANNEL_KEY,
  COMPANY_KEY,
  CAMPAIGN_KEY,
  SHARE_QUERY_PARAMS,
  DEFAULT_COLUMNS,
  EMP_RANGE_KEY,
  REVENUE_RANGE_KEY,
  INDUSTRY_KEY,
  ALL_CHANNEL
} from './const';

export const generateFirstAndLastDayOfLastWeeks = (n = 5): WeekStartEnd[] => {
  const lastWeek = MomentTz().subtract(7, 'd');
  const dateArray: WeekStartEnd[] = [...generateUnsavedReportDateRanges()];

  // generating saved report dates
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
      formattedRangeOption,
      isSaved: true
    });
  }

  return dateArray;
};

export const generateUnsavedReportDateRanges = (): WeekStartEnd[] => {
  const dateArray: WeekStartEnd[] = [];
  const dateValues = [
    DATE_RANGE_TODAY_LABEL,
    DATE_RANGE_YESTERDAY_LABEL,
    DATE_RANGE_LABEL_LAST_7_DAYS
  ];

  dateValues.forEach((dateValue) => {
    const dateObj = getRangeByLabel(dateValue);
    const from = dateObj?.startDate ? momentTz(dateObj.startDate) : momentTz();
    const to = dateObj?.endDate ? momentTz(dateObj.endDate) : momentTz();
    dateArray.push({
      from: from.unix(),
      to: to.unix(),
      formattedRange:
        dateValue === DATE_RANGE_TODAY_LABEL
          ? momentTz().format('MMM DD, YYYY')
          : `${from.format('MMM D, Y')} - ${to.format('MMM D, Y')}`,
      formattedRangeOption: dateValue,
      isSaved: false
    });
  });
  return dateArray;
};

export const parseSavedReportDates = (dates: string[]): WeekStartEnd[] => {
  if (!dates || !dates?.length) return [];
  const dateArray: WeekStartEnd[] = [];
  dates.forEach((dateValue: string) => {
    if (typeof dateValue !== 'string') return;
    const dateValueArray = dateValue?.trim()?.split('-');
    if (dateValueArray.length < 2) return;
    const fromEpoch = Number(dateValueArray[0]);
    const fromDate = momentTz.unix(fromEpoch);
    const toEpoch = Number(dateValueArray[1]);
    const toDate = momentTz.unix(toEpoch);

    const formattedRangeOption = `${fromDate.format(
      'MMM Do'
    )} to ${toDate.format('MMM Do')}`;
    const formattedRange = `${fromDate.format('MMM D, Y')} - ${toDate.format(
      'MMM D, Y'
    )}`;
    dateArray.unshift({
      from: fromEpoch,
      to: toEpoch,
      formattedRange,
      formattedRangeOption,
      isSaved: true
    });
  });
  return dateArray;
};

export const getFormattedRange = (
  from: number,
  to: number,
  timezone = 'Asia/Kolkata'
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
  const uniqueCampains = new Set<string>();
  const uniqueChannels = new Set<string>();
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
  }
  if (header === EMP_RANGE_KEY || header === REVENUE_RANGE_KEY) {
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
    }
    return {
      width: 350
    };
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
      const returnObj = {
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
            text,
            record,
            header
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
    const rowObj: StringObject = {
      key: String(i)
    };
    headers.forEach((header, j) => {
      rowObj[header] = row[j];
    });

    return rowObj;
  });
  // filtering table data with selected Channel filter
  if (selectedChannel && selectedChannel !== ALL_CHANNEL) {
    dataSource = dataSource?.filter(
      (data: StringObject) => data?.[CHANNEL_KEY] === selectedChannel
    );
  }
  // filtering using selected campaigns
  if (
    selectedCampaigns &&
    Array.isArray(selectedCampaigns) &&
    selectedCampaigns?.length > 0
  ) {
    dataSource = dataSource?.filter((data: StringObject) =>
      selectedCampaigns.includes(data?.[CAMPAIGN_KEY])
    );
  }
  // filtering using search key
  if (searchText) {
    dataSource = dataSource?.filter(
      (data: StringObject) =>
        data?.[COMPANY_KEY]?.toLowerCase()?.includes(searchText.toLowerCase())
    );
  }
  return SortResults(dataSource, sorter);
};

export const getDefaultTableColumns = () =>
  DEFAULT_COLUMNS.map((key, index) => ({
    index,
    dataIndex: key,
    title: getClickableTitleSorter(
      // @ts-ignore
      KEY_LABELS?.[key] || key,
      { key, type: 'categorical', subtype: null },
      {},
      () => {
        console.log('handle sorting');
      },
      'left',
      'center',
      'px-6 py-3'
    )
  }));

export const checkStringEquality = (
  str1: string,
  str2: string,
  caseSensitive = false
): boolean => {
  if (caseSensitive) return str1 === str2;
  return str1?.toLowerCase() === str2?.toLowerCase();
};

export const getPublicUrl = (obj: ShareApiData, project_id: string): string =>
  `${window.location.protocol}//${window.location.host}${PathUrls.VisitorIdentificationReport}?${SHARE_QUERY_PARAMS.queryId}=${obj.query_id}&${SHARE_QUERY_PARAMS.projectId}=${project_id}&${SHARE_QUERY_PARAMS.routeVersion}=${obj.route_version}`;

export const generateEllipsisOption = (values: string[], charLimit = 35) => {
  let text = '';
  values.every((value, i) => {
    text += `${value}${i < values.length - 1 ? ', ' : ''}`;
    if (text.length > charLimit) {
      text = createEllipsis(text, charLimit, values.length - (i + 1));
      return false;
    }
    return true;
  });
  return text;
};

const createEllipsis = (text: string, charLimit: number, countLeft = 0) => {
  if (text.length < charLimit) return text;
  if (text.length > charLimit && countLeft) {
    return `${text.slice(0, charLimit)}...${countLeft}more`;
  }
  return `${text.slice(0, charLimit)}...`;
};
