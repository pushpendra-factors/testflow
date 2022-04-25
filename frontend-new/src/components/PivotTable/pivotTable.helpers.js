import { EMPTY_ARRAY } from 'Utils/global';

export const formatPivotData = ({ data, breakdown, kpis }) => {
  try {
    const breakdownAttributes = breakdown.map((b) => b.property);
    const kpiAttributes = kpis.map((kpi) => kpi);
    const attributesRow = breakdownAttributes.concat(kpiAttributes);
    const values = data.map((d) => {
      const breakdownVals = breakdown.map((b, index) => {
        return d[`${b.property} - ${index}`];
      });
      const kpiVals = kpis.map((kpi, index) => {
        return d[`${kpi} - ${index}`];
      });
      const current = breakdownVals.concat(kpiVals);
      return current;
    });
    return [breakdownAttributes, attributesRow, values];
  } catch (err) {
    console.log('formatPivotData -> err', err);
    return EMPTY_ARRAY;
  }
};

export const getValueOptions = ({ kpis }) => {
  return _.map(kpis, (kpi) => kpi);
};

export const getColumnOptions = ({ breakdown }) => {
  return _.map(breakdown, (b) => b.property);
};

export const getRowOptions = ({ selectedRows, kpis, breakdown }) => {
  const valueOptions = getValueOptions({ kpis });
  const columnOptions = getColumnOptions({ breakdown });
  const allOptions = _.concat(valueOptions, columnOptions);
  return _.filter(allOptions, (option) => !selectedRows.includes(option));
};

export const SortRowOptions = ({ data, kpis, breakdown }) => {
  const breakdownOptions = getColumnOptions({ breakdown });
  const kpiOptions = getValueOptions({ kpis });
  const breakdownsSelected = data.filter((d) => breakdownOptions.includes(d));
  const kpisSelected = data.filter((d) => kpiOptions.includes(d));
  return [...breakdownsSelected, ...kpisSelected];
};

export const getFunctionOptions = () => {
  return [
    'Integer Sum',
    'Sum',
    'Count',
    'Average',
    'Median',
    'Sum as Fraction of Rows',
    'Sum as Fraction of Columns',
    'Sum as Fraction of Total',
  ];
};
