import React from 'react';
import moment from 'moment';
import { labelsObj } from '../../utils';
import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';

export const getTableColumns = (
  events,
  breakdown,
  currentSorter,
  handleSorting,
  page,
  eventNames,
  userPropNames,
  eventPropNames
) => {
  const breakdownColumns = breakdown.map((e) => {
    let displayTitle = e;
    if (userPropNames[e]) {
      displayTitle = userPropNames[e] ? userPropNames[e] : e;
    }
    if (eventPropNames[e]) {
      displayTitle = eventPropNames[e] ? eventPropNames[e] : e;
    }

    return {
      title: displayTitle,
      dataIndex: e,
      width: '50%',
      fixed: 'left',
    };
  });

  const e = events[0];

  const title = eventNames[e] || e;

  const countColumn = {
    title: getTitleWithSorter(
      `${title}: ${labelsObj[page]}`,
      'Event Count',
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Event Count',
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };
  return [...breakdownColumns, countColumn];
};

export const getDataInTableFormat = (
  data,
  events,
  breakdown,
  searchText,
  currentSorter
) => {
  if (breakdown.length === 1 && events.length === 1) {
    const filteredData = data.filter(
      (d) => d.label.toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );
    const result = filteredData.map((d, index) => {
      return {
        index: d.index,
        [breakdown[0]]: d.label,
        'Event Count': d.value,
      };
    });
    return SortData(result, currentSorter.key, currentSorter.order);
  }
  return [];
};

const getWeekFormat = (m) => {
  const startDate = m.format('D-MMM-YYYY');
  const endDate = m.endOf('week').format('D-MMM-YYYY');
  return startDate + ' to ' + endDate;
};

export const parseForDateTimeLabel = (grn, label) => {
  let labelValue = label;
  if (grn && moment(label).isValid()) {
    let dateLabel;
    try {
      const newDatr = new Date(label);
      dateLabel = moment(newDatr);
    } catch (e) {
      return label;
    }

    if (grn === 'date') {
      labelValue = dateLabel.format('D-MMM-YYYY');
    } else if (grn === 'day') {
      labelValue = dateLabel.format('D-MMM-YYYY');
    } else if (grn === 'week') {
      labelValue = getWeekFormat(dateLabel);
    } else if (grn === 'hour') {
      labelValue = dateLabel.format('D-MMM-YYYY H') + 'h';
    } else if (grn === 'month') {
      labelValue = dateLabel.format('MMM-YYYY');
    }
  }

  return labelValue;
};

export const formatData = (data) => {
  if (
    !data ||
    !data.metrics ||
    !data.metrics.rows ||
    !data.metrics.rows.length
  ) {
    return [];
  }
  const result = data.metrics.rows.map((elem, index) => {
    const labelVal = parseForDateTimeLabel(
      data.meta?.query?.gbp[0]?.grn,
      elem[2]
    );
    // console.log(labelVal);
    return {
      label: labelVal,
      value: elem[3],
      index,
    };
  });
  return SortData(result, 'value', 'descend');
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting,
  frequency
) => {
  const result = [
    {
      title: breakdown[0],
      dataIndex: breakdown[0],
      fixed: 'left',
      width: 200,
    },
  ];

  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }

  const dateColumns = categories.map((cat) => {
    return {
      title: getTitleWithSorter(
        moment(cat).format(format),
        moment(cat).format(format),
        currentSorter,
        handleSorting
      ),
      width: 100,
      dataIndex: moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...dateColumns];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  breakdown,
  searchText,
  currentSorter,
  frequency
) => {
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D';
  }
  const result = seriesData
    .filter((sd) => sd.name.toLowerCase().includes(searchText.toLowerCase()))
    .map((sd) => {
      const dateWiseData = {};
      categories.forEach((cat, index) => {
        dateWiseData[moment(cat).format(format)] = sd.data[index];
      });
      return {
        index: sd.index,
        [breakdown[0]]: sd.name,
        ...dateWiseData,
      };
    });
  return SortData(result, currentSorter.key, currentSorter.order);
};

export const formatDataInStackedAreaFormat = (data, aggregateData) => {
  if (
    !data.headers ||
    !data.headers.length ||
    !data.rows ||
    !data.rows.length ||
    !aggregateData.length
  ) {
    return {
      categories: [],
      data: [],
    };
  }
  const dateIndex = data.headers.findIndex((h) => h === 'datetime');
  const countIndex = data.headers.findIndex((h) => h === 'count');
  const eventIndex = data.headers.findIndex((h) => h === 'event_name');
  const breakdownIndex = eventIndex + 1;
  let differentDates = new Set();
  data.rows.forEach((row) => {
    differentDates.add(row[dateIndex]);
  });
  differentDates = Array.from(differentDates);
  const initializedDatesData = differentDates.map(() => {
    return 0;
  });
  const labelsMapper = {};
  const resultantData = aggregateData.map((d, index) => {
    labelsMapper[d.label] = index;
    return {
      name: d.label,
      data: [...initializedDatesData],
      index: d.index,
      marker: {
        enabled: false,
      },
    };
  });

  data.rows.forEach((row) => {
    let breakdownJoin = row
      .slice(breakdownIndex, countIndex)
      .map((x) => parseForDateTimeLabel(data.meta?.query?.gbp[0]?.grn, x))
      .join(',');
    console.log(breakdownJoin);
    const bIdx = labelsMapper[breakdownJoin];
    const idx = differentDates.indexOf(row[dateIndex]);
    if (resultantData[bIdx]) {
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });

  return {
    categories: differentDates,
    data: resultantData,
  };
};
