/* eslint-disable */
import tableStyles from "./FunnelsResultTable/index.module.scss";
import {
  calculatePercentage,
  SortData,
  getTitleWithSorter,
} from "../../../utils/dataFormatter";
import { valueMapper } from "../../../utils/constants";

const windowSize = {
  w: window.outerWidth,
  h: window.outerHeight,
  iw: window.innerWidth,
  ih: window.innerHeight,
};

export const generateGroupedChartsData = (
  response,
  queries,
  groups,
  arrayMapper
) => {
  if (!response) {
    return [];
  }
  const result = queries.map((_, idx) => {
    return [arrayMapper[idx].mapper];
  });
  const firstEventIdx = response.headers.findIndex((elem) => elem === "step_0");
  response.rows.forEach((elem) => {
    const row = elem.map(item=>{
      return item;
    });
    const breakdownName = row.slice(0, firstEventIdx).join(",");
    const isVisible = groups.filter(
      (g) => g.is_visible && g.name === breakdownName
    ).length;
    if (isVisible) {
      const netCounts = row.filter((row) => typeof row === "number");
      netCounts.forEach((n, idx) => {
        result[idx].push(calculatePercentage(n, netCounts[0]));
      });
    }
  });
  return result;
};

export const generateGroups = (response, maxAllowedVisibleProperties) => {
  if (!response) {
    return [];
  }
  const firstEventIdx = response.headers.findIndex((elem) => elem === "step_0");
  const result = response.rows.map((elem, index) => {
    const row = elem.map(item=>{
      return item;
    })
    const netCounts = row.filter((row) => typeof row === "number");
    const name = row.slice(0, firstEventIdx).join(",");
    return {
      index,
      name,
      conversion_rate:
        calculatePercentage(netCounts[netCounts.length - 1], netCounts[0]) +
        "%",
      is_visible: index < maxAllowedVisibleProperties ? true : false,
    };
  });
  return result;
};

export const generateTableColumns = (
  breakdown,
  queries,
  currentSorter,
  handleSorting,
  arrayMapper
) => {
  const result = [
    {
      title: breakdown.length ? "Grouping" : "Users",
      dataIndex: "name",
      className: tableStyles.groupColumn,
    },
    {
      title: "Conversion",
      dataIndex: "conversion",
      className: tableStyles.conversionColumn,
    },
  ];
  const eventColumns = queries.map((elem, index) => {
    return {
      title: getTitleWithSorter(elem, elem, currentSorter, handleSorting),
      dataIndex: breakdown.length
        ? `${arrayMapper[index].mapper}-${index}`
        : `${elem}-${index}`,
      className: index === queries.length - 1 ? tableStyles.lastColumn : "",
    };
  });

  const blankCol = {
    title: "",
    dataIndex: "",
    width: 37,
  };
  if (breakdown.length) {
    return [...result, ...eventColumns];
  } else {
    return [blankCol, ...result, ...eventColumns];
  }
};

export const generateTableData = (
  data,
  breakdown,
  queries,
  groups,
  arrayMapper,
  currentSorter,
  searchText
) => {
  if (!breakdown.length) {
    const queryData = {};
    queries.forEach((q, index) => {
      queryData[
        `${q}-${index}`
      ] = `${data[index].netCount} (${data[index].value}%)`;
    });
    return [
      {
        index: 0,
        ...queryData,
        name: "All",
        conversion: data[data.length - 1].value + "%",
      },
    ];
  } else {
    const appliedGroups = groups
      .map((elem) => elem.name)
      .filter(
        (elem) => elem.toLowerCase().indexOf(searchText.toLowerCase()) > -1
      );
    const result = appliedGroups.map((grp, index) => {
      const group = grp;
      const eventsData = {};
      data.forEach((d, idx) => {
        eventsData[`${d.name}-${idx}`] =
          d.data[group] +
          " (" +
          calculatePercentage(d.data[group], data[0].data[group]) +
          "%)";
      });
      return {
        index,
        name: grp,
        conversion:
          calculatePercentage(
            data[data.length - 1].data[group],
            data[0].data[group]
          ) + "%",
        ...eventsData,
      };
    });

    if (currentSorter.key) {
      const sortKey = arrayMapper.find(
        (elem) => elem.eventName === currentSorter.key
      );
      return SortData(
        result,
        sortKey.mapper + "-" + sortKey.index,
        currentSorter.order
      );
    }
    return result;
  }
};

export const generateUngroupedChartsData = (response, arrayMapper) => {
  if (!response) {
    return [];
  }

  const netCounts = response.rows[0].filter((elem) => typeof elem === "number");
  const result = [];
  let index = 0;

  while (index < arrayMapper.length) {
    if (index === 0) {
      result.push({
        event: arrayMapper[index].mapper,
        netCount: netCounts[index],
        value: 100,
      });
    } else {
      result.push({
        event: arrayMapper[index].mapper,
        netCount: netCounts[index],
        value: calculatePercentage(netCounts[index], netCounts[0]),
      });
    }
    index++;
  }
  return result;
};

export const checkForWindowSizeChange = (callback) => {
  if (
    window.outerWidth !== windowSize.w ||
    window.outerHeight !== windowSize.h
  ) {
    setTimeout(() => {
      windowSize.w = window.outerWidth; // update object with current window properties
      windowSize.h = window.outerHeight;
      windowSize.iw = window.innerWidth;
      windowSize.ih = window.innerHeight;
    }, 0);
    callback();
  }

  // if the window doesn't resize but the content inside does by + or - 5%
  else if (
    window.innerWidth + window.innerWidth * 0.05 < windowSize.iw ||
    window.innerWidth - window.innerWidth * 0.05 > windowSize.iw
  ) {
    setTimeout(() => {
      windowSize.iw = window.innerWidth;
    }, 0);
    callback();
  }
};

export const generateEventsData = (response, queries, arrayMapper) => {
  if (!response) {
    return [];
  }
  const firstEventIdx = response.headers.findIndex((elem) => elem === "step_0");
  const result = queries.map((q, idx) => {
    const data = {};
    response.rows.forEach((r) => {
      const name = r.slice(0, firstEventIdx).join(",");
      const netCounts = r.filter((elem) => typeof elem === "number");
      data[name] = netCounts[idx];
    });
    return {
      index: idx + 1,
      data,
      name: arrayMapper[idx].mapper,
    };
  });
  return result;
};
