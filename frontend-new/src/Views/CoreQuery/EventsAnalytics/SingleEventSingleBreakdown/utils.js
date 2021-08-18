import React from 'react';
import moment from 'moment';
import { labelsObj } from '../../utils';
import {
  getClickableTitleSorter,
  SortResults,
} from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  DATE_FORMATS,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
} from '../../../../utils/constants';

export const defaultSortProp = () => {
  return {
    order: 'descend',
    key: 'Event Count',
    type: 'numerical',
    subtype: null,
  };
};

export const getVisibleData = (aggregateData, sorter) => {
  const result = SortResults(aggregateData, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

export const getVisibleSeriesData = (data, sorter) => {
  const result = SortResults(data, sorter).slice(
    0,
    MAX_ALLOWED_VISIBLE_PROPERTIES
  );
  return result;
};

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
  console.log('sesb getTableColumns');
  const breakdownColumns = breakdown.map((e) => {
    const displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || e.property
        : e.property;
    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: e.property, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: e.property,
      width: '50%',
      fixed: 'left',
    };
  });

  const e = events[0];

  const title = eventNames[e] || e;

  const countColumn = {
    title: getClickableTitleSorter(
      `${title}: ${labelsObj[page]}`,
      { key: 'Event Count', type: 'numerical', subtype: null },
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

export const getDataInTableFormat = (data, searchText, currentSorter) => {
  console.log('sesb getDataInTableFormat');
  const filteredData = data.filter(
    (d) => d.label.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  return SortResults(filteredData, currentSorter);
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

    if (grn === 'date' || grn === 'day' || grn === 'month' || grn === 'hour') {
      labelValue = dateLabel.format(DATE_FORMATS[grn]);
    } else if (grn === 'week') {
      labelValue = getWeekFormat(dateLabel);
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
  console.log('sesb Format Data');
  const result = data.metrics.rows.map((elem, index) => {
    const labelVal = parseForDateTimeLabel(
      data.meta?.query?.gbp[0]?.grn,
      elem[2]
    );
    const breakdowns = data.meta.query.gbp;
    return {
      label: labelVal,
      value: elem[3],
      [breakdowns[0].pr]: labelVal,
      'Event Count': elem[3], //used for sorting, value key will be removed soon
      index,
    };
  });
  return result;
};

export const getDateBasedColumns = (
  categories,
  breakdown,
  currentSorter,
  handleSorting,
  frequency,
  userPropNames,
  eventPropNames
) => {
  console.log('sesb getDateBasedColumns');
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: `Event Count`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting
    ),
    dataIndex: `Event Count`,
    width: 150,
  };
  const breakdownColumns = breakdown.map((e) => {
    const displayTitle =
      e.prop_category === 'user'
        ? userPropNames[e.property] || e.property
        : e.prop_category === 'event'
        ? eventPropNames[e.property] || e.property
        : e.property;

    return {
      title: getClickableTitleSorter(
        displayTitle,
        { key: e.property, type: e.prop_type, subtype: e.grn },
        currentSorter,
        handleSorting
      ),
      dataIndex: e.property,
      width: 200,
      fixed: 'left',
    };
  });

  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  const dateColumns = categories.map((cat) => {
    return {
      title: getClickableTitleSorter(
        moment(cat).format(format),
        { key: moment(cat).format(format), type: 'numerical', subtype: null },
        currentSorter,
        handleSorting
      ),
      width: 150,
      dataIndex: moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...breakdownColumns, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  searchText,
  currentSorter,
) => {
  console.log('sesb getDateBasedTableData');
  const result = seriesData.filter((sd) =>
    sd.name.toLowerCase().includes(searchText.toLowerCase())
  );

  return SortResults(result, currentSorter);
};

export const formatDataInStackedAreaFormat = (
  data,
  aggregateData,
  frequency
) => {
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
  console.log('sesb formatDataInStackedAreaFormat');
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
      ...d,
    };
  });

  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];

  data.rows.forEach((row) => {
    let breakdownJoin = row
      .slice(breakdownIndex, countIndex)
      .map((x) => parseForDateTimeLabel(data.meta?.query?.gbp[0]?.grn, x))
      .join(',');
    const bIdx = labelsMapper[breakdownJoin];
    const category = row[dateIndex];
    const idx = differentDates.indexOf(category);
    if (resultantData[bIdx]) {
      resultantData[bIdx][moment(category).format(format)] = row[countIndex];
      resultantData[bIdx].data[idx] = row[countIndex];
    }
  });

  return {
    categories: differentDates,
    data: resultantData,
  };
};
