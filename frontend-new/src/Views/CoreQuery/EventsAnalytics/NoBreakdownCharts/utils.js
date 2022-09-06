import React from 'react';
import cx from 'classnames';
import get from 'lodash/get';
import MomentTz from 'Components/MomentTz';
import {
  addQforQuarter,
  formatCount,
  getClickableTitleSorter,
  SortResults
} from '../../../../utils/dataFormatter';
import {
  Number as NumFormat,
  SVG,
  Text
} from '../../../../components/factorsComponents';
import { DATE_FORMATS } from '../../../../utils/constants';

export const getTableData = ({ data, currentSorter }) => {
  const clonedData = data.map((elem) => {
    const element = { ...elem };
    return element;
  });

  const result = clonedData.map((elem, index) => {
    return {
      ...elem,
      index
    };
  });

  return SortResults(result, currentSorter);
};

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
      key: 'Overall',
      type: 'numerical',
      subtype: null,
      order: 'descend'
    }
  ];
};

export const getColumns = (
  events,
  arrayMapper,
  frequency,
  currentSorter,
  handleSorting,
  eventNames,
  comparisonApplied
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  const dateColumn = {
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
  };

  const eventColumns = events.map((e, idx) => {
    const mapperKey = arrayMapper.find((elem) => elem.index === idx).mapper;
    return {
      title: getClickableTitleSorter(
        eventNames[e] || e,
        {
          key: mapperKey,
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: 'text-right',
      dataIndex: mapperKey,
      render: (d, row) => {
        return (
          <div className="flex flex-col">
            <Text type="title" level={7} color="grey-6">
              <NumFormat number={d} />
            </Text>
            {comparisonApplied && (
              <>
                <Text type="title" level={7} color="grey">
                  <NumFormat number={row[`${mapperKey} - compareValue`]} />
                </Text>
                <div className="flex col-gap-1 items-center justify-end">
                  <SVG
                    color={
                      row[`${mapperKey} - change`] > 0 ? '#5ACA89' : '#FF0000'
                    }
                    name={
                      row[`${mapperKey} - change`] > 0
                        ? 'arrowLift'
                        : 'arrowDown'
                    }
                    size={16}
                  />
                  <Text
                    level={7}
                    type="title"
                    color={row[`${mapperKey} - change`] < 0 ? 'red' : 'green'}
                  >
                    <NumFormat
                      number={Math.abs(row[`${mapperKey} - change`])}
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
  return [dateColumn, ...eventColumns];
};

export const formatData = (response, arrayMapper, comparisonData) => {
  if (
    !response.headers ||
    !response.headers.length ||
    !response.rows ||
    !response.rows.length
  ) {
    return [];
  }
  const result = [];
  const { headers, rows } = response;
  const dateIndex = headers.findIndex((header) => header === 'datetime');
  rows.forEach((row, rowIdx) => {
    const eventsData = {};
    headers.slice(dateIndex + 1).forEach((_, index) => {
      const key = arrayMapper.find((m) => m.index === index).mapper;
      eventsData[key] = row[dateIndex + index + 1];

      if (comparisonData != null) {
        const comparisonKey = `${key} - compareValue`;
        eventsData[comparisonKey] = get(
          comparisonData,
          `rows.${rowIdx}.${dateIndex + index + 1}`,
          0
        );
        const changeKey = `${key} - change`;
        if (eventsData[comparisonKey]) {
          eventsData[changeKey] =
            ((eventsData[key] - eventsData[comparisonKey]) /
              eventsData[comparisonKey]) *
            100;
        } else {
          eventsData[changeKey] = 0;
        }
      }
    });
    result.push({
      date: new Date(row[dateIndex]),
      compareDate: new Date(
        get(comparisonData, `rows.${rowIdx}.${dateIndex}`, new Date())
      ),
      ...eventsData
    });
  });
  return result;
};

export const getDataInLineChartFormat = (
  data,
  arrayMapper,
  eventNames,
  comparisonData
) => {
  if (
    !data.headers ||
    !data.headers.length ||
    !data.rows ||
    !data.rows.length
  ) {
    return {
      categories: [],
      data: []
    };
  }
  const { headers, rows } = data;
  const dateIndex = headers.findIndex((h) => h === 'datetime');
  const differentDates = new Set();
  const differentComparisonDates = new Set();
  rows.forEach((row, rowIndex) => {
    differentDates.add(row[dateIndex]);
    if (comparisonData && comparisonData.rows.length > 0) {
      const compareDate = get(comparisonData, `rows.${rowIndex}.${dateIndex}`);
      differentComparisonDates.add(compareDate);
    }
  });

  const categories = Array.from(differentDates);
  const compareCategories = Array.from(differentComparisonDates);

  const initializedDatesData = differentDates.map(() => {
    return 0;
  });

  const resultantData = arrayMapper.map((m) => {
    return {
      name: m.displayName || eventNames[m.eventName] || m.eventName,
      data: [...initializedDatesData],
      index: m.index,
      marker: {
        enabled: false
      }
    };
  });

  const comparisonResultantData = arrayMapper.map((m) => {
    return {
      name: m.displayName || eventNames[m.eventName] || m.eventName,
      data: [...initializedDatesData],
      index: m.index,
      marker: {
        enabled: false
      },
      dashStyle: 'dash',
      compareIndex: m.index
    };
  });

  data.rows.forEach((row, rowIndex) => {
    const idx = categories.indexOf(row[dateIndex]);
    arrayMapper.forEach((_, index) => {
      resultantData[index].data[idx] = row[dateIndex + index + 1];
      if (comparisonData != null) {
        const compareValue = get(
          comparisonData,
          `rows.${rowIndex}.${dateIndex + index + 1}`,
          0
        );
        comparisonResultantData[index].data[idx] = compareValue;
      }
    });
  });

  return {
    categories,
    compareCategories,
    data:
      comparisonData != null
        ? [...resultantData, ...comparisonResultantData]
        : resultantData
  };
};

export const getDateBasedColumns = (
  data,
  currentSorter,
  handleSorting,
  frequency,
  eventNames,
  comparisonApplied
) => {
  // const OverallColumn = {
  //   title: getClickableTitleSorter(
  //     'Overall',
  //     { key: 'Overall', type: 'numerical', subtype: null },
  //     currentSorter,
  //     handleSorting,
  //     'right'
  //   ),
  //   dataIndex: 'Overall',
  //   className: 'text-right',
  //   width: 150
  // };

  const eventColumn = {
    title: getClickableTitleSorter(
      'Event',
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
      return eventNames[d] || d;
    }
  };
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;

  const dateColumns = [];
  data.forEach((elem) => {
    dateColumns.push({
      title: getClickableTitleSorter(
        addQforQuarter(frequency) + MomentTz(elem.date).format(format),
        {
          key: addQforQuarter(frequency) + MomentTz(elem.date).format(format),
          type: 'numerical',
          subtype: null
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      width: frequency === 'hour' ? 200 : 150,
      dataIndex: addQforQuarter(frequency) + MomentTz(elem.date).format(format),
      className: cx('text-right', { 'border-none': comparisonApplied }),
      render: (d) => {
        return <NumFormat number={d} />;
      }
    });
    if (comparisonApplied) {
      dateColumns.push({
        title: getClickableTitleSorter(
          addQforQuarter(frequency) + MomentTz(elem.compareDate).format(format),
          {
            key:
              addQforQuarter(frequency) +
              MomentTz(elem.compareDate).format(format),
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
          addQforQuarter(frequency) + MomentTz(elem.compareDate).format(format),
        render: (d, row) => {
          return <NumFormat number={d} />;
        }
      });
      dateColumns.push({
        title: getClickableTitleSorter(
          'Change',
          {
            key:
              addQforQuarter(frequency) +
              MomentTz(elem.compareDate).format(format) +
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
          MomentTz(elem.compareDate).format(format) +
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
  return [eventColumn, ...dateColumns];
};

export const getDateBasedTableData = (
  data,
  currentSorter,
  searchText,
  arrayMapper,
  frequency,
  metrics,
  comparisonApplied
) => {
  const filteredMapper = arrayMapper.filter((elem) =>
    elem.eventName.toLowerCase().includes(searchText.toLowerCase())
  );
  const format = DATE_FORMATS[frequency] || DATE_FORMATS.date;
  const dates = data.map(
    (elem) => addQforQuarter(frequency) + MomentTz(elem.date).format(format)
  );
  const comparisonDates = comparisonApplied
    ? data.map(
        (elem) =>
          addQforQuarter(frequency) + MomentTz(elem.compareDate).format(format)
      )
    : [];
  const result = filteredMapper.map((elem, index) => {
    let total = 0;
    if (
      metrics &&
      metrics.headers &&
      Array.isArray(metrics.headers) &&
      metrics.headers.length &&
      metrics.rows &&
      Array.isArray(metrics.rows) &&
      metrics.rows.length
    ) {
      const countIdx = metrics.headers.findIndex(
        (h) => h === 'count' || h === 'aggregate'
      );
      const eventIndexIdx = metrics.headers.findIndex(
        (h) => h === 'event_index'
      );
      const metricRow = metrics.rows.find(
        (mr) => mr[eventIndexIdx] === elem.index
      );
      total = metricRow ? metricRow[countIdx] : 0;
    }
    const eventsData = {};
    dates.forEach((date, dateIndex) => {
      const val1 = data.find(
        (d) =>
          addQforQuarter(frequency) + MomentTz(d.date).format(format) === date
      )[elem.mapper];
      eventsData[date] = val1;
      if (comparisonApplied) {
        const val2 = data.find(
          (d) =>
            addQforQuarter(frequency) +
              MomentTz(d.compareDate).format(format) ===
            comparisonDates[dateIndex]
        )[`${elem.mapper} - compareValue`];
        eventsData[comparisonDates[dateIndex]] = val2;
        eventsData[`${comparisonDates[dateIndex]} - Change`] = val2
          ? formatCount(((val1 - val2) / val2) * 100)
          : 0;
      }
    });
    return {
      index,
      event: elem.eventName,
      Overall: total,
      ...eventsData
    };
  });

  return SortResults(result, currentSorter, currentSorter.order);
};
