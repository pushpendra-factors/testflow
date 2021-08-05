import {
  SortData,
  SortResults,
  getClickableTitleSorter,
} from '../../../../utils/dataFormatter';
import { parseForDateTimeLabel } from '../SingleEventSingleBreakdown/utils';
import { getBreakDownGranularities } from '../SingleEventMultipleBreakdown/utils';

export const formatData = (data) => {
  const headerSlice = data.headers.slice(0, data.headers.length - 1);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);

  const result = data.rows.map((d, index) => {
    const str = d
      .slice(0, d.length - 1)
      .map((x, ind) => parseForDateTimeLabel(grns[ind], x))
      .join(',');
    return {
      index,
      label: str,
      value: d[d.length - 1],
    };
  });
  return SortData(result, 'value', 'descend');
};

export const getTableColumns = (
  breakdown,
  currentSorter,
  handleSorting,
  userPropNames,
  eventPropNames
) => {
  const breakdownColumns = breakdown.map((e, index) => {
    const displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || e.property
        : e.property;

    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: `${e.property} - ${index}`, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
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
  };
  return [...breakdownColumns, countCol];
};

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  const result = filteredData.map((d) => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[`${b.property} - ${index}`] = d.label.split(',')[index];
    });
    return { ...breakdownValues, 'User Count': d.value, index: d.index };
  });
  return SortResults(result, currentSorter);
};
