import {
  SortResults,
  getClickableTitleSorter
} from '../../../../utils/dataFormatter';
import {
  getBreakdownDisplayName,
  parseForDateTimeLabel
} from '../eventsAnalytics.helpers';
import { getBreakDownGranularities } from '../SingleEventMultipleBreakdown/utils';
import {
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  DISPLAY_PROP
} from '../../../../utils/constants';

export const getDefaultSortProp = () => [
  {
    key: 'User Count',
    type: 'numerical',
    subtype: null,
    order: 'descend'
  }
];

export const getVisibleData = (aggregateData, sorter) => {
  const result = SortResults(aggregateData, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

export const formatData = (data) => {
  const headerSlice = data.headers.slice(0, data.headers.length - 1);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);

  const result = data.rows.map((d, index) => {
    const breakdownVals = d
      .slice(0, d.length - 1)
      .map((vl) => (DISPLAY_PROP[vl] ? DISPLAY_PROP[vl] : vl));
    const breakdownData = {};
    for (let i = 0; i < breakdowns.length; i++) {
      const bkd = breakdowns[i];
      breakdownData[`${bkd.pr} - ${i}`] = parseForDateTimeLabel(
        grns[i],
        breakdownVals[i]
      );
    }
    const grpLabel = Object.values(breakdownData).join(', ');
    return {
      index,
      label: grpLabel,
      value: d[d.length - 1],
      'User Count': d[d.length - 1],
      ...breakdownData
    };
  });
  return result;
};

export const getTableColumns = (
  breakdown,
  currentSorter,
  handleSorting,
  userPropNames,
  eventPropertiesDisplayNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle = getBreakdownDisplayName({
      breakdown: e,
      userPropNames,
      eventPropertiesDisplayNames
    });

    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200
    };
  });

  const countCol = {
    title: getClickableTitleSorter(
      'User Count',
      { key: 'User Count', type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: 'User Count',
    width: 200
  };
  return [...breakdownColumns, countCol];
};

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  return SortResults(filteredData, currentSorter);
};
