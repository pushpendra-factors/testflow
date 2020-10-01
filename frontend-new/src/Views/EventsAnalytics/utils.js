import moment from 'moment';

import { getTitleWithSorter } from '../CoreQuery/FunnelsResultPage/utils';
import {
  singleEventAnyEventUserResponse, multiEventAnyEventUserResponse
} from './SampleResponse';

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

  result.sort((a, b) => {
    if (currentSorter.order === 'ascend') {
      return parseInt(a[currentSorter.key]) >= parseInt(b[currentSorter.key]) ? 1 : -1;
    }
    if (currentSorter.order === 'descend') {
      return parseInt(a[currentSorter.key]) <= parseInt(b[currentSorter.key]) ? 1 : -1;
    }
    return 0;
  });
  return result;
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
  const dates = [];
  const result = [];
  response.rows.forEach(r => {
    if (dates.indexOf(r[0]) === -1) {
      const currentDateData = response.rows.filter(elem => elem[0] === r[0]);
      const eventsData = {};
      currentDateData.forEach(d => {
        const query = queries.find(q => q === d[1]);
        eventsData[eventsMapper[query]] = d[2];
      });
      result.push({
        date: new Date(r[0]),
        ...eventsData
      });
      dates.push(r[0]);
    }
  });
  return result;
};

export const getSingleEventAnyEventUserData = (event, eventsMapper) => {
  const response = singleEventAnyEventUserResponse;
  const result = response.rows.map(row => {
    return {
      date: new Date(row[0]),
      [eventsMapper[event]]: row[1]
    };
  });
  return result;
};

export const getMultiEventsAnyEventUserData = (queries, eventsMapper) => {
  const response = multiEventAnyEventUserResponse;
  const dates = [];
  const result = [];
  response.rows.forEach(r => {
    if (dates.indexOf(r[0]) === -1) {
      const currentDateData = response.rows.filter(elem => elem[0] === r[0]);
      const eventsData = {};
      currentDateData.forEach(d => {
        const query = queries.find(q => q === d[1]);
        eventsData[eventsMapper[query]] = d[2];
      });
      result.push({
        date: new Date(r[0]),
        ...eventsData
      });
      dates.push(r[0]);
    }
  });
  return result;
};

export const getDataInLineChartFormat = (data, queries, eventsMapper, hiddenEvents = []) => {
  data.sort((a, b) => {
    return moment(a.date).utc().unix() > moment(b.date).utc().unix() ? 1 : -1;
  });
  const result = [];
  const hashedData = {};
  hashedData.x = data.map(elem => {
    return moment(elem.date).format('YYYY-MM-DD');
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

export const getDateBasedColumns = (data, currentSorter, handleSorting) => {
  const result = [
    {
      title: 'Events',
      dataIndex: 'event',
      fixed: 'left',
      width: 200
    }];

  const dateColumns = data.map(elem => {
    return {
      title: getTitleWithSorter(moment(elem.date).format('MMM D'), moment(elem.date).format('MMM D'), currentSorter, handleSorting),
      width: 100,
      dataIndex: moment(elem.date).format('MMM D')
    };
  });
  return [...result, ...dateColumns];
};

export const getNoGroupingTablularDatesBasedData = (data, currentSorter, searchText, reverseEventsMapper) => {
  const events = Object.keys(reverseEventsMapper);
  const dates = data.map(elem => moment(elem.date).format('MMM D'));
  const filteredEvents = events.filter(event => reverseEventsMapper[event].includes(searchText));
  const result = filteredEvents.map((elem, index) => {
    const eventsData = {};
    dates.forEach(date => {
      eventsData[date] = data.find(elem => moment(elem.date).format('MMM D') === date)[elem];
    });
    return {
      index,
      event: reverseEventsMapper[elem],
      ...eventsData
    };
  });
  result.sort((a, b) => {
    if (currentSorter.order === 'ascend') {
      return parseInt(a[currentSorter.key]) >= parseInt(b[currentSorter.key]) ? 1 : -1;
    }
    if (currentSorter.order === 'descend') {
      return parseInt(a[currentSorter.key]) <= parseInt(b[currentSorter.key]) ? 1 : -1;
    }
    return 0;
  });
  return result;
};

export const formatSingleEventSinglePropertyData = (data) => {
  const properties = {};
  const result = [];
  data.rows.forEach(elem => {
    if (elem[1] !== '$none') {
      if (properties.hasOwnProperty(elem[1])) {
        result[properties[elem[1]]].value += elem[2];
      } else {
        properties[elem[1]] = result.length;
        result.push({
          label: elem[1],
          value: elem[2]
        })
      }
    }
  });
  result.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  return result;
}