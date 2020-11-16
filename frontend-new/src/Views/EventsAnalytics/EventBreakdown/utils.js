import { getTitleWithSorter } from '../../CoreQuery/FunnelsResultPage/utils';
import { SortData } from '../../CoreQuery/utils';

export const formatData = (data) => {
  const resultInObjFormat = {};
  data.rows.forEach(d => {
    const date = d[0];
    const str = d.slice(1, d.length - 1).join(',');
    if (resultInObjFormat[str]) {
      resultInObjFormat[str].datewise.push({
        date,
        value: d[d.length - 1]
      });
      resultInObjFormat[str].value += d[d.length - 1];
    } else {
      resultInObjFormat[str] = {
        value: d[d.length - 1],
        datewise: [{
          date,
          value: d[d.length - 1]
        }]
      };
    }
  });
  const result = [];
  let idx = 0;
  for (const key in resultInObjFormat) {
    result.push({
      ...resultInObjFormat[key],
      label: key,
      index: idx
    });
    idx++;
  }
  result.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  return result;
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
