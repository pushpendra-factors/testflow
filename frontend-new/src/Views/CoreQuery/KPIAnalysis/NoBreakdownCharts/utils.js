import React from 'react';
import moment from 'moment';

import { DATE_FORMATS } from '../../../../utils/constants';
import { Number as NumFormat } from '../../../../components/factorsComponents';
import {
  addQforQuarter,
  getClickableTitleSorter,
  SortResults,
} from '../../../../utils/dataFormatter';

import { getKpiLabel } from '../kpiAnalysis.helpers';

export const getDefaultSortProp = (kpis) => {
  if (Array.isArray(kpis) && kpis.length) {
    return [
      {
        key: `${getKpiLabel(kpis[0])} - 0`,
        type: 'numerical',
        subtype: null,
        order: 'descend',
      },
    ];
  }
  return [];
};

export const getDefaultDateSortProp = () => {
  return [
    {
      key: 'Overall',
      type: 'numerical',
      subtype: null,
      order: 'descend',
    },
  ];
};

export const formatData = (data, kpis) => {
  try {
    const result = kpis.map((kpi, index) => {
      const kpiLabel = getKpiLabel(kpi);
      const totalIndex = 1;
      const dateSplitIndex = 0;
      const obj = {
        index,
        name: kpiLabel,
      };
      if (data[totalIndex] && data[dateSplitIndex]) {
        const dateIndex = data[dateSplitIndex].headers.findIndex(
          (h) => h === 'datetime'
        );
        const kpiIndex = index + 1;
        return {
          ...obj,
          total: data[totalIndex].rows.length
            ? data[totalIndex].rows[0][index]
            : 0,
          dataOverTime: data[dateSplitIndex].rows.map((row) => {
            return {
              date: new Date(row[dateIndex]),
              [kpiLabel]: row[kpiIndex],
            };
          }),
        };
      } else {
        return {
          ...obj,
          total: 0,
        };
      }
    });
    return result;
  } catch (err) {
    console.log('formatData -> err', err);
    return [];
  }
};

export const formatDataInSeriesFormat = (aggData) => {
  try {
    const differentDates = new Set();
    aggData.forEach((d) => {
      d.dataOverTime.forEach((elem) => {
        differentDates.add(new Date(elem.date).getTime());
      });
    });
    const categories = Array.from(differentDates);
    const initializedDatesData = categories.map(() => {
      return 0;
    });
    const data = aggData.map((m) => {
      return {
        index: m.index,
        name: m.name,
        data: [...initializedDatesData],
        marker: {
          enabled: false,
        },
        total: m.total,
      };
    });
    aggData.forEach((m, index) => {
      categories.forEach((cat, catIndex) => {
        const dateIndex = m.dataOverTime.findIndex(
          (elem) => new Date(elem.date).getTime() === cat
        );
        if (dateIndex > -1) {
          data[index].data[catIndex] = m.dataOverTime[dateIndex][m.name];
        }
      });
    });
    return {
      categories,
      data,
    };
  } catch (err) {
    return {
      categories: [],
      data: [],
    };
  }
};

export const getTableColumns = (
  kpis,
  currentSorter,
  handleSorting,
  eventNames,
  frequency
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const result = [
    {
      title: getClickableTitleSorter(
        'Date',
        { key: 'date', type: 'datetime', subtype: 'date' },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'date',
      render: (d) => {
        return addQforQuarter(frequency) + moment(d).format(format);
      },
    },
  ];
  const eventColumns = kpis.map((e, idx) => {
    const kpiLabel = getKpiLabel(e);
    return {
      title: getClickableTitleSorter(
        eventNames[kpiLabel] || kpiLabel,
        {
          key: `${kpiLabel} - ${idx}`,
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: 'text-right',
      dataIndex: `${kpiLabel} - ${idx}`,
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...eventColumns];
};

export const getDataInTableFormat = (
  data,
  categories,
  queries,
  currentSorter
) => {
  const result = categories.map((cat, catIndex) => {
    const obj = {
      index: catIndex,
      date: cat,
    };
    queries.forEach((q, qIndex) => {
      const kpiLabel = getKpiLabel(q);
      obj[`${kpiLabel} - ${qIndex}`] = data[qIndex].data[catIndex];
    });
    return obj;
  });
  return SortResults(result, currentSorter);
};

export const getDateBasedColumns = (
  categories,
  currentSorter,
  handleSorting,
  eventNames,
  frequency
) => {
  const OverallColumn = {
    title: getClickableTitleSorter(
      'Overall',
      { key: `Overall`, type: 'numerical', subtype: null },
      currentSorter,
      handleSorting,
      'right'
    ),
    className: 'text-right',
    dataIndex: `Overall`,
    width: 150,
    render: (d) => {
      return <NumFormat number={d} />;
    },
  };
  const result = [
    {
      title: getClickableTitleSorter(
        'Event',
        {
          key: 'event',
          type: 'categorical',
          subtype: null,
        },
        currentSorter,
        handleSorting
      ),
      dataIndex: 'event',
      fixed: 'left',
      width: 200,
      render: (d) => {
        return eventNames[d] || d;
      },
    },
  ];
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const dateColumns = categories.map((cat) => {
    return {
      title: getClickableTitleSorter(
        addQforQuarter(frequency) + moment(cat).format(format),
        {
          key: addQforQuarter(frequency) + moment(cat).format(format),
          type: 'numerical',
          subtype: null,
        },
        currentSorter,
        handleSorting,
        'right'
      ),
      className: 'text-right',
      width: frequency === 'hour' ? 200 : 150,
      dataIndex: addQforQuarter(frequency) + moment(cat).format(format),
      render: (d) => {
        return <NumFormat number={d} />;
      },
    };
  });
  return [...result, ...dateColumns, OverallColumn];
};

export const getDateBasedTableData = (
  seriesData,
  categories,
  searchText,
  currentSorter,
  frequency
) => {
  const format = DATE_FORMATS[frequency] || DATE_FORMATS['date'];
  const result = seriesData.map((sd, index) => {
    const obj = {
      index,
      event: sd.name,
      Overall: sd.total,
    };
    const dateData = {};
    categories.forEach((cat, catIndex) => {
      dateData[addQforQuarter(frequency) + moment(cat).format(format)] =
        sd.data[catIndex];
    });
    return {
      ...obj,
      ...dateData,
    };
  });
  const filteredResult = result.filter(
    (elem) => elem.event.toLowerCase().indexOf(searchText.toLowerCase()) > -1
  );
  return SortResults(filteredResult, currentSorter);
};
