import React from 'react';
import moment from 'moment';
import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {parseForDateTimeLabel} from '../SingleEventSingleBreakdown/utils';

export const formatData = (data) => {
  if (
    !data ||
    !data.metrics ||
    !data.metrics.headers ||
    !data.metrics.headers.length ||
    !data.metrics.rows ||
    !data.metrics.rows.length
  ) {
    return [];
  }
  const { headers, rows } = data.metrics;
  const eventNameIndex = headers.findIndex((header) => header === 'event_name');
  const countIndex = headers.findIndex((header) => header === 'count');
  
  const headerSlice = headers.slice(eventNameIndex + 1, countIndex);
  let breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  let grns = getBreakDownGranularities(headerSlice, breakdowns)
  
  const result = rows.map((d, index) => {
    const str = d.slice(eventNameIndex + 1, countIndex).join(',');
    const grpLabel = str.split(',').map((x, ind) => parseForDateTimeLabel(grns[ind], x)).join(',')
    return {
      label: grpLabel,
      value: d[countIndex],
      index,
    };
  });
  return SortData(result, 'value', 'descend');
};

export const getBreakDownGranularities = (breakDownSlice, breakdowns) => {
  const grns = [];
  let brks = breakdowns;
  breakDownSlice.forEach((h) => {
    const brkIndex = brks.findIndex((x) => h===x.pr);
    grns.push(brks[brkIndex]?.grn);
    brks.splice(brkIndex, 1)
  });
  return grns;
}

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
  const breakdownColumns = breakdown.map((e, index) => {
    let displayTitle = e.property;
    if (e.prop_category === 'user') {
      displayTitle = userPropNames[e.property]
        ? userPropNames[e.property]
        : e.property;
    }

    if (e.prop_category === 'event') {
      displayTitle = eventPropNames[e.property]
        ? eventPropNames[e.property]
        : e.property;
    }

    return {
      title: displayTitle,
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
    };
  });

  const e = events[0];
  const title = eventNames[e] || e;

  const countColumn = {
    title: getTitleWithSorter(
      title,
      'Event Count',
      currentSorter,
      handleSorting
    ),
    dataIndex: 'Event Count',
    width: 200,
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
  const filteredData = data.filter((elem) =>
    elem.label.toLowerCase().includes(searchText.toLowerCase())
  );
  const result = filteredData.map((d) => {
    const splittedLabel = d.label.split(',');
    const breakdownData = {};
    breakdown.forEach((b, index) => {
      let brkLabel = splittedLabel[index];
      if(b.grn) {
        brkLabel = parseForDateTimeLabel(b.grn, brkLabel)
      }
      breakdownData[`${b.property} - ${index}`] = brkLabel;
    });
    return {
      index: d.index,
      'Event Count': d.value,
      ...breakdownData,
    };
  });
  return SortData(result, currentSorter.key, currentSorter.order);
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
  const breakdownColumns = breakdown.map((e, index) => {
    let displayTitle = e.property;
    if (e.prop_category === 'user') {
      displayTitle = userPropNames[e.property]
        ? userPropNames[e.property]
        : e.property;
    }

    if (e.prop_category === 'event') {
      displayTitle = eventPropNames[e.property]
        ? eventPropNames[e.property]
        : e.property;
    }

    return {
      title: displayTitle,
      dataIndex: `${e.property} - ${index}`,
      fixed: !index ? 'left' : '',
      width: 200,
    };
  });
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
  return [...breakdownColumns, ...dateColumns];
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
      const splittedLabel = sd.name.split(',');
      const breakdownData = {};
      breakdown.forEach((b, index) => {
        let brkLabel = splittedLabel[index];
        if(b.grn) {
          brkLabel = parseForDateTimeLabel(b.grn, brkLabel)
        }
        breakdownData[`${b.property} - ${index}`] = brkLabel;
      });
      return {
        index: sd.index,
        ...breakdownData,
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

  const headerSlice = data.headers.slice(breakdownIndex, countIndex);
  let breakdowns = data.meta.query.gbp ? [...data.meta.query.gbp] : [];
  let grns = getBreakDownGranularities(headerSlice, breakdowns)

  data.rows.forEach((row) => {
    const breakdownJoin = row.slice(breakdownIndex, countIndex).map((x, ind) => parseForDateTimeLabel(grns[ind], x)).join(',');
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
