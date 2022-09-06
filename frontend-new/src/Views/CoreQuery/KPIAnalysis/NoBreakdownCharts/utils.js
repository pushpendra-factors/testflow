import React from 'react';
import get from 'lodash/get';
import { has } from 'lodash';
import cx from 'classnames';
import MomentTz from 'Components/MomentTz';

import { DATE_FORMATS } from '../../../../utils/constants';
import {
  Number as NumFormat,
  SVG,
  Text
} from '../../../../components/factorsComponents';
import {
  addQforQuarter,
  formatCount,
  getClickableTitleSorter,
  SortResults
} from '../../../../utils/dataFormatter';

import { getKpiLabel, getFormattedKpiValue } from '../kpiAnalysis.helpers';

export const getDefaultSortProp = () => {
  return [
    {
      key: 'date',
      type: 'datetime',
      subtype: 'date',
      order: 'descend'
    }
  ];
};

export const getDefaultDateSortProp = () => {
  return [
    {
      key: 'event',
      type: 'categorical',
      subtype: null,
      order: 'descend'
    }
  ];
};

export const formatData = (data, kpis, comparisonData) => {
  try {
    const result = kpis.map((kpi, index) => {
      const kpiLabel = getKpiLabel(kpi);
      const totalIndex = 1;
      const dateSplitIndex = 0;
      const obj = {
        index,
        name: kpiLabel,
        metricType: kpi.metricType
      };
      if (data[totalIndex] && data[dateSplitIndex]) {
        const dateIndex = data[dateSplitIndex].headers.findIndex(
          (h) => h === 'datetime'
        );
        const kpiIndex = index + 1;
        const kpiData = {
          ...obj,
          total: data[totalIndex].rows.length
            ? data[totalIndex].rows[0][index]
            : 0,
          dataOverTime: data[dateSplitIndex].rows.map((row, rowIdx) => {
            const dataObj = {
              date: new Date(row[dateIndex]),
              [kpiLabel]: row[kpiIndex]
            };
            if (comparisonData != null) {
              dataObj.compareDate = new Date(
                comparisonData[dateSplitIndex]?.rows[rowIdx]?.[dateIndex]
              );
              dataObj.compareValue =
                comparisonData[dateSplitIndex]?.rows[rowIdx]?.[kpiIndex];
            }
            return dataObj;
          })
        };
        if (comparisonData != null) {
          kpiData.compareTotal = comparisonData[totalIndex].rows.length
            ? comparisonData[totalIndex].rows[0][index]
            : 0;
        }
        return kpiData;
      } else {
        return {
          ...obj,
          total: 0
        };
      }
    });
    return result;
  } catch (err) {
    console.log('formatData -> err', err);
    return [];
  }
};

export const formatDataInSeriesFormat = (aggData, comparisonApplied) => {
  try {
    const differentDates = new Set();
    const differentComparisonDates = new Set();
    aggData.forEach((d) => {
      d.dataOverTime.forEach((elem) => {
        differentDates.add(new Date(elem.date).getTime());
        if (comparisonApplied) {
          differentComparisonDates.add(new Date(elem.compareDate).getTime());
        }
      });
    });

    const categories = Array.from(differentDates);
    const compareCategories = Array.from(differentComparisonDates);

    const initializedDatesData = categories.map(() => {
      return 0;
    });
    const data = aggData.map((m) => {
      return {
        index: m.index,
        name: m.name,
        data: [...initializedDatesData],
        marker: {
          enabled: false
        },
        metricType: get(m, 'metricType', null),
        total: m.total
      };
    });

    const compareData = comparisonApplied
      ? aggData.map((m) => {
          return {
            index: m.index,
            name: m.name,
            data: [...initializedDatesData],
            marker: {
              enabled: false
            },
            dashStyle: 'dash',
            metricType: get(m, 'metricType', null),
            total: m.compareTotal,
            compareIndex: m.index
          };
        })
      : [];

    aggData.forEach((m, index) => {
      categories.forEach((cat, catIndex) => {
        const dateIndex = m.dataOverTime.findIndex(
          (elem) => new Date(elem.date).getTime() === cat
        );
        if (dateIndex > -1) {
          data[index].data[catIndex] = m.dataOverTime[dateIndex][m.name];
          if (comparisonApplied) {
            compareData[index].data[catIndex] =
              m.dataOverTime[dateIndex].compareValue;
          }
        }
      });
    });
    return {
      categories,
      compareCategories,
      data: comparisonApplied ? [...data, ...compareData] : data
    };
  } catch (err) {
    console.error(err);
    return {
      categories: [],
      data: [],
      compareCategories: []
    };
  }
};

export const getTableColumns = ({
  kpis,
  currentSorter,
  handleSorting,
  frequency,
  comparisonApplied
}) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const result = [
    {
      title: getClickableTitleSorter(
        'Date',
        { key: 'date', type: 'datetime', subtype: 'date' },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'date',
      render: (d, row) => {
        return (
          <div className="flex flex-col">
            <Text type="title" level={7} color="grey-6">
              {addQforQuarter(frequency) + MomentTz(d).format(format)}
            </Text>
            {comparisonApplied && (
              <Text type="title" level={7} color="grey">
                Vs{' '}
                {addQforQuarter(frequency) +
                  MomentTz(row.compareDate).format(format)}
              </Text>
            )}
          </div>
        );
      }
    }
  ];
  const eventColumns = kpis.map((e, idx) => {
    const kpiLabel = getKpiLabel(e);
    return {
      title: getClickableTitleSorter(
        kpiLabel,
        {
          key: `${kpiLabel} - ${idx}`,
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: 'text-right',
      dataIndex: `${kpiLabel} - ${idx}`,
      render: (d, row) => {
        return (
          <div className="flex flex-col">
            <Text type="title" level={7} color="grey-6">
              {e.metricType ? (
                getFormattedKpiValue({ value: d, metricType: e.metricType })
              ) : (
                <NumFormat number={d} />
              )}
            </Text>
            {comparisonApplied && (
              <>
                <Text type="title" level={7} color="grey">
                  {e.metricType ? (
                    getFormattedKpiValue({
                      value: row[`${kpiLabel} - ${idx} - compareValue`],
                      metricType: e.metricType
                    })
                  ) : (
                    <NumFormat
                      number={row[`${kpiLabel} - ${idx} - compareValue`]}
                    />
                  )}
                </Text>
                <div className="flex col-gap-1 items-center justify-end">
                  <SVG
                    color={
                      row[`${kpiLabel} - ${idx} - change`] > 0
                        ? '#5ACA89'
                        : '#FF0000'
                    }
                    name={
                      row[`${kpiLabel} - ${idx} - change`] > 0
                        ? 'arrowLift'
                        : 'arrowDown'
                    }
                    size={16}
                  />
                  <Text
                    level={7}
                    type="title"
                    color={
                      row[`${kpiLabel} - ${idx} - change`] < 0 ? 'red' : 'green'
                    }
                  >
                    <NumFormat
                      number={Math.abs(row[`${kpiLabel} - ${idx} - change`])}
                    />
                    %
                  </Text>
                </div>
              </>
            )}
          </div>
        );
      }
    };
  });
  return [...result, ...eventColumns];
};

export const getDataInTableFormat = (
  data,
  categories,
  kpis,
  currentSorter,
  comparisonApplied,
  compareCategories
) => {
  const compareRows = {};
  const result = categories.map((cat, catIndex) => {
    const obj = {
      index: catIndex,
      date: cat
    };
    if (comparisonApplied) {
      obj.compareDate = compareCategories[catIndex];
    }
    kpis.forEach((q, qIndex) => {
      const kpiLabel = getKpiLabel(q);
      obj[`${kpiLabel} - ${qIndex}`] = data[qIndex].data[catIndex];
      if (comparisonApplied) {
        if (compareRows[`${kpiLabel} - ${qIndex}`] == null) {
          compareRows[`${kpiLabel} - ${qIndex}`] = data.find(
            (elem) => elem.compareIndex === data[qIndex].index
          );
        }
        const val1 = data[qIndex].data[catIndex];
        const val2 = compareRows[`${kpiLabel} - ${qIndex}`]?.data[catIndex];
        obj[`${kpiLabel} - ${qIndex} - compareValue`] = val2;
        obj[`${kpiLabel} - ${qIndex} - change`] =
          val1 && val2 ? ((val1 - val2) / val2) * 100 : 0;
      }
    });
    return obj;
  });
  return SortResults(result, currentSorter);
};

export const getDateBasedColumns = ({
  kpis,
  categories,
  currentSorter,
  handleSorting,
  frequency,
  comparisonApplied,
  compareCategories
}) => {
  // const OverallColumn = {
  //   title: getClickableTitleSorter(
  //     'Overall',
  //     { key: 'Overall', type: 'numerical', subtype: null },
  //     currentSorter,
  //     handleSorting,
  //     'right'
  //   ),
  //   className: 'text-right',
  //   dataIndex: 'Overall',
  //   width: 150,
  //   render: (d, _, index) => {
  //     const metricType = get(kpis[index], 'metricType', null);
  //     return metricType ? (
  //       getFormattedKpiValue({ value: d, metricType })
  //     ) : (
  //       <NumFormat number={d} />
  //     );
  //   }
  // };
  const result = [
    {
      title: getClickableTitleSorter(
        'KPI',
        {
          key: 'event',
          type: 'categorical',
          subtype: null
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'event',
      fixed: 'left',
      width: 200,
      render: (d) => {
        return d;
      }
    }
  ];
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const dateColumns = [];
  const metricTypes = {};
  categories.forEach((cat, catIndex) => {
    dateColumns.push({
      title: getClickableTitleSorter(
        addQforQuarter(frequency) + MomentTz(cat).format(format),
        {
          key: addQforQuarter(frequency) + MomentTz(cat).format(format),
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: cx('text-right', { 'border-none': comparisonApplied }),
      width: frequency === 'hour' ? 200 : 150,
      dataIndex: addQforQuarter(frequency) + MomentTz(cat).format(format),
      render: (d, row) => {
        let metricType;
        if (metricTypes[row.event] != null) {
          metricType = metricTypes[row.event];
        } else {
          metricType = kpis.find((kpi) => kpi.label === row.event)?.metricType;
          metricTypes[row.event] = metricType;
        }
        return metricType ? (
          getFormattedKpiValue({ value: d, metricType })
        ) : (
          <NumFormat number={d} />
        );
      }
    });
    if (comparisonApplied) {
      dateColumns.push({
        title: getClickableTitleSorter(
          addQforQuarter(frequency) +
            MomentTz(compareCategories[catIndex]).format(format),
          {
            key:
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format),
            type: 'numerical',
            subtype: null
          },
          currentSorter,
          handleSorting,
          'right'
        ),
        className: 'text-right border-none',
        width: frequency === 'hour' ? 200 : 150,
        dataIndex:
          addQforQuarter(frequency) +
          MomentTz(compareCategories[catIndex]).format(format),
        render: (d, row) => {
          let metricType;
          if (metricTypes[row.event] != null) {
            metricType = metricTypes[row.event];
          } else {
            metricType = kpis.find(
              (kpi) => kpi.label === row.event
            )?.metricType;
            metricTypes[row.event] = metricType;
          }
          return metricType ? (
            getFormattedKpiValue({ value: d, metricType })
          ) : (
            <NumFormat number={d} />
          );
        }
      });
      dateColumns.push({
        title: getClickableTitleSorter(
          'Change',
          {
            key:
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format) +
              ' - Change',
            type: 'percent',
            subtype: null
          },
          currentSorter,
          handleSorting,
          'right'
        ),
        className: 'text-right',
        width: frequency === 'hour' ? 200 : 150,
        dataIndex:
          addQforQuarter(frequency) +
          MomentTz(compareCategories[catIndex]).format(format) +
          ' - Change',
        render: (d) => {
          const changeIcon = (
            <SVG
              color={d > 0 ? '#5ACA89' : '#FF0000'}
              name={d > 0 ? 'arrowLift' : 'arrowDown'}
              size={16}
            />
          );
          return (
            <div className="flex col-gap-1 items-center justify-end">
              {changeIcon}
              <Text level={7} type="title" color={d < 0 ? 'red' : 'green'}>
                <NumFormat number={Math.abs(d)} />%
              </Text>
            </div>
          );
        }
      });
    }
  });
  return [...result, ...dateColumns];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  searchText,
  currentSorter,
  frequency,
  comparisonApplied,
  compareCategories
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const result = seriesData
    .filter((s) => !has(s, 'compareIndex'))
    .map((sd, index) => {
      const obj = {
        index,
        event: sd.name,
        Overall: sd.total
      };
      const compareRow = seriesData.find(
        (elem) => elem.compareIndex === sd.index
      );
      const dateData = {};
      categories.forEach((cat, catIndex) => {
        dateData[addQforQuarter(frequency) + MomentTz(cat).format(format)] =
          sd.data[catIndex];
        if (comparisonApplied && compareRow != null) {
          const val1 = sd.data[catIndex];
          const val2 = compareRow.data[catIndex];
          dateData[
            addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format)
          ] = val2;
          dateData[
            `${
              addQforQuarter(frequency) +
              MomentTz(compareCategories[catIndex]).format(format)
            } - Change`
          ] = formatCount(((val1 - val2) / val2) * 100);
        }
      });
      return {
        ...obj,
        ...dateData
      };
    });
  const filteredResult = result.filter(
    (elem) => elem.event.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  return SortResults(filteredResult, currentSorter);
};
