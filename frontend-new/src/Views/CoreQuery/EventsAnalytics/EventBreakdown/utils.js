import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';
import { parseForDateTimeLabel } from '../SingleEventSingleBreakdown/utils';
import { getBreakDownGranularities } from '../SingleEventMultipleBreakdown/utils';

export const formatData = (data) => {
  const headerSlice = data.headers.slice(0, data.headers.length - 1);
  const breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  const grns = getBreakDownGranularities(headerSlice, breakdowns);

  const result = data.rows.map((d, index) => {
    const str = d.slice(0, d.length - 1).map((x, ind) => 
      parseForDateTimeLabel(grns[ind], x)).join(',');
    return {
      index,
      label: str,
      value: d[d.length - 1]
    };
  });
  return SortData(result, 'value', 'descend');
};

export const getTableColumns = (breakdown, currentSorter, handleSorting) => {
  const result = breakdown.map(b => {
    return {
      title: b.property,
      dataIndex: b.property
    };
  });

  const countCol = {
    title: getTitleWithSorter('User Count', 'User Count', currentSorter, handleSorting),
    dataIndex: 'User Count'
  };
  return [...result, countCol];
};

export const getTableData = (data, breakdown, searchText, currentSorter) => {
  const filteredData = data.filter(elem => elem.label.toLowerCase().includes(searchText.toLowerCase()));
  const result = filteredData.map(d => {
    const breakdownValues = {};
    breakdown.forEach((b, index) => {
      breakdownValues[b.property] = d.label.split(',')[index];
    });
    return { ...breakdownValues, 'User Count': d.value, index: d.index };
  });
  return SortData(result, currentSorter.key, currentSorter.order);
};
