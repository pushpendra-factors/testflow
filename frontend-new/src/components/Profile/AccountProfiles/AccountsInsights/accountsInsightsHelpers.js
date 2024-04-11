import MomentTz from 'Components/MomentTz';
import {
  currencyFormatter,
  formatCount,
  formatDuration
} from 'Utils/dataFormatter';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { reorderDefaultDomainSegmentsToTop } from '../accountProfiles.helpers';

export const getInsightsDataByKey = (insights) => {
  if (insights.completed === true) {
    return insights.data
      .filter((d) => d.headers)
      .reduce(
        (prev, curr) => ({
          ...prev,
          [curr.headers[0]]: curr.rows[0]
        }),
        {}
      );
  }
  return {};
};

export const getCompareDate = (dateRange) => {
  if (dateRange.dateType === 'this_month') {
    return {
      startDate: MomentTz().subtract(1, 'month').startOf('month'),
      endDate: MomentTz().subtract(1, 'month').endOf('month'),
      dateString: 'Last Month',
      dateType: 'last_month'
    };
  }
  if (dateRange.dateType === 'last_month') {
    return {
      startDate: MomentTz().subtract(2, 'month').startOf('month'),
      endDate: MomentTz().subtract(2, 'month').endOf('month'),
      dateString: 'Last to Last Month',
      dateType: 'last_to_last_month'
    };
  }
  if (dateRange.dateType === 'this_week') {
    return {
      startDate: MomentTz().subtract(1, 'week').startOf('week'),
      endDate: MomentTz().subtract(1, 'week').endOf('week'),
      dateString: 'Last Week',
      dateType: 'last_week'
    };
  }
  if (dateRange.dateType === 'last_week') {
    return {
      startDate: MomentTz().subtract(2, 'week').startOf('week'),
      endDate: MomentTz().subtract(2, 'week').endOf('week'),
      dateString: 'Last to Last Week',
      dateType: 'last_to_last_week'
    };
  }
  if (dateRange.dateType === 'this_quarter') {
    return {
      startDate: MomentTz().subtract(1, 'quarter').startOf('quarter'),
      endDate: MomentTz().subtract(1, 'quarter').endOf('quarter'),
      dateString: 'Last Quarter',
      dateType: 'last_quarter'
    };
  }
  if (dateRange.dateType === 'last_quarter') {
    return {
      startDate: MomentTz().subtract(2, 'quarter').startOf('quarter'),
      endDate: MomentTz().subtract(2, 'quarter').endOf('quarter'),
      dateString: 'Last to Last Quarter',
      dateType: 'last_to_last_quarter'
    };
  }
  return null;
};

export const getFormattedMetricValue = (value, valueType) => {
  if (valueType === 'currency') {
    return `$${currencyFormatter(value)}`;
  }
  if (valueType === 'percentage') {
    return `${formatCount(value, 1)}%`;
  }
  if (valueType === 'duration') {
    return formatDuration(value);
  }
  return value;
};

export const getSegmentName = (segments, segmentId) => {
  const segmentsList =
    reorderDefaultDomainSegmentsToTop(segments[GROUP_NAME_DOMAINS]) || [];
  return segmentsList.find((segment) => segment.id === segmentId)?.name;
};
