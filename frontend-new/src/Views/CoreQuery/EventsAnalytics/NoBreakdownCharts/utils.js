import moment from 'moment';
import { SortData, getTitleWithSorter } from '../../../../utils/dataFormatter';

export const getNoGroupingTableData = (data, currentSorter, searchText, reverseEventsMapper) => {
  const clonedData = data.map(elem => {
    const element = { ...elem };
    for (const key in element) {
      if (key !== 'date') {
        element[reverseEventsMapper[key]] = element[key];
        delete element[key];
      }
    }
    return element;
  });

  const result = clonedData.map((elem, index) => {
    return {
      index,
      ...elem,
      date: moment(elem.date).format('MMM D, YYYY')
    };
  });

  return SortData(result, currentSorter.key, currentSorter.order);
};

export const getColumns = (events, currentSorter, handleSorting) => {
  const result = [
    {
      title: '',
      dataIndex: '',
      width: 37
    },
    {
      title: 'Date',
      dataIndex: 'date'
    }];

  const eventColumns = events.map(e => {
    return {
      title: getTitleWithSorter(e, e, currentSorter, handleSorting),
      dataIndex: e
    };
  });
  return [...result, ...eventColumns];
};

export const formatSingleEventAnalyticsData = (response, event, eventsMapper) => {
  const result = response.rows.map(row => {
    return {
      date: new Date(row[0]),
      [eventsMapper[event]]: row[1]
    };
  });
  return result;
};

export const formatMultiEventsAnalyticsData = (response, queries, eventsMapper) => {
  const result = [];
  response.rows.forEach(r => {
    const eventsData = {};
    response.headers.slice(1).forEach((h, index) => {
      eventsData[eventsMapper[h]] = r[index + 1];
    });
    result.push({
      date: new Date(r[0]),
      ...eventsData
    });
  });
  return result;
};

export const getDataInLineChartFormat = (data, queries, eventsMapper, hiddenEvents = [], frequency) => {
  data.sort((a, b) => {
    return moment(a.date).utc().unix() > moment(b.date).utc().unix() ? 1 : -1;
  });
  const result = [];
  const hashedData = {};
  const format = 'YYYY-MM-DD HH-mm';
  hashedData.x = data.map(elem => {
    return moment(elem.date).format(format);
  });
  queries.forEach(q => {
    if (hiddenEvents.indexOf(q) === -1) {
      hashedData[eventsMapper[q]] = data.map(elem => {
        return elem[eventsMapper[q]];
      });
    }
  });
  for (const obj in hashedData) {
    result.push([obj, ...hashedData[obj]]);
  }
  return result;
};

export const getDateBasedColumns = (data, currentSorter, handleSorting, frequency) => {
  const result = [
    {
      title: 'Events',
      dataIndex: 'event',
      fixed: 'left',
      width: 200
    }];
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D'
  }

  const dateColumns = data.map(elem => {
    return {
      title: getTitleWithSorter(moment(elem.date).format(format), moment(elem.date).format(format), currentSorter, handleSorting),
      width: 100,
      dataIndex: moment(elem.date).format(format)
    };
  });
  return [...result, ...dateColumns];
};

export const getNoGroupingTablularDatesBasedData = (data, currentSorter, searchText, reverseEventsMapper, frequency) => {
  const events = Object.keys(reverseEventsMapper);
  let format = 'MMM D';
  if (frequency === 'hour') {
    format = 'h A, MMM D'
  }
  const dates = data.map(elem => moment(elem.date).format(format));
  const filteredEvents = events.filter(event => reverseEventsMapper[event].includes(searchText));
  const result = filteredEvents.map((elem, index) => {
    const eventsData = {};
    dates.forEach(date => {
      eventsData[date] = data.find(elem => moment(elem.date).format(format) === date)[elem];
    });
    return {
      index,
      event: reverseEventsMapper[elem],
      ...eventsData
    };
  });

  return SortData(result, currentSorter.key, currentSorter.order);
};
