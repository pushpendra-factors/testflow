import React from "react";
import tableStyles from "./FunnelsResultTable/index.module.scss";
import {
  calculatePercentage,
  SortData,
  getTitleWithSorter,
  formatDuration,
} from "../../../utils/dataFormatter";
import { SVG } from "../../../components/factorsComponents";

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
  const result = groups
    .filter((g) => g.is_visible)
    .map((g) => {
      return { name: g.name };
    });
  const firstEventIdx = response.headers.findIndex((elem) => elem === "step_0");
  response.rows.forEach((row) => {
    const breakdownName = row.slice(0, firstEventIdx).join(",");
    const obj = result.find((r) => r.name === breakdownName);
    if (obj) {
      const netCounts = row.filter((val) => typeof val === "number");
      queries.forEach((_, idx) => {
        const eventIdx = response.headers.findIndex(
          (elem) => elem === `step_${idx}`
        );
        obj[arrayMapper[idx].mapper] = calculatePercentage(
          row[eventIdx],
          netCounts[0]
        );
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
    const row = elem.map((item) => {
      return item;
    });
    const netCounts = row.filter((row) => typeof row === "number");
    const name = row.slice(0, firstEventIdx).join(",");
    return {
      index,
      name,
      value:
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
      dataIndex: "Grouping",
      className: tableStyles.groupColumn,
      fixed: "left",
      width: 100,
      render: (d) => {
        if (d.includes("$no_group")) {
          return "Overall";
        } else {
          return d;
        }
      },
    },
    {
      title: "Total Conversion",
      dataIndex: "Conversion",
      className: tableStyles.conversionColumn,
      fixed: "left",
      width: 100,
    },
    {
      title: "Conversion Time",
      dataIndex: "Converstion Time",
      width: 100,
    },
  ];
  const eventColumns = [];
  queries.forEach((elem, index) => {
    eventColumns.push({
      title: breakdown.length
        ? getTitleWithSorter(elem, elem, currentSorter, handleSorting)
        : elem,
      dataIndex: breakdown.length
        ? `${arrayMapper[index].mapper}-${index}`
        : `${elem}-${index}`,
      width: 150,
      className: index === queries.length - 1 ? tableStyles.lastColumn : "",
    });
    if (index < queries.length - 1) {
      eventColumns.push({
        title: (
          <div className="flex items-center justify-between">
            <div className="text-base" style={{ color: "#8692A3" }}>
              &mdash;
            </div>
            <SVG name="clock" />
            <div className="text-base" style={{ color: "#8692A3" }}>
              &rarr;
            </div>
          </div>
        ),
        dataIndex: `time[${index}-${index + 1}]`,
        width: 50,
      });
    }
  });

  const blankCol = {
    title: "",
    dataIndex: "",
    width: 37,
    fixed: "left",
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
  searchText,
  durations
) => {
  if (!breakdown.length) {
    const queryData = {};
    const overallDuration = getOverAllDuration(durations);
    queries.forEach((q, index) => {
      queryData[
        `${q}-${index}`
      ] = `${data[index].netCount} (${data[index].value}%)`;
      if (index < queries.length - 1) {
        queryData[`time[${index}-${index + 1}]`] = getStepDuration(
          durations,
          index,
          index + 1
        );
      }
    });
    return [
      {
        index: 0,
        Grouping: "All",
        Conversion: data[data.length - 1].value + "%",
        "Converstion Time": overallDuration,
        ...queryData,
      },
    ];
  } else {
    const appliedGroups = groups
      .map((elem) => elem.name)
      .filter(
        (elem) => elem.toLowerCase().indexOf(searchText.toLowerCase()) > -1
      );
    const durationMetric = durations.metrics.find(
      (elem) => elem.title === "MetaStepTimeInfo"
    );
    const firstEventIdx = durationMetric.headers.findIndex(
      (elem) => elem === "step_0_1_time"
    );
    const result = appliedGroups.map((grp, index) => {
      const group = grp;
      const durationGrp = durationMetric.rows.find(
        (elem) => elem.slice(0, firstEventIdx).join(",") === grp
      );
      const eventsData = {};
      let totalDuration = 0;
      data.forEach((d, idx) => {
        eventsData[`${d.name}-${idx}`] =
          d.data[group] +
          " (" +
          calculatePercentage(d.data[group], data[0].data[group]) +
          "%)";
        if (idx < data.length - 1) {
          const durationIdx = durationMetric.headers.findIndex(
            (elem) => elem === `step_${idx}_${idx + 1}_time`
          );
          eventsData[`time[${idx}-${idx + 1}]`] = durationGrp
            ? formatDuration(durationGrp[durationIdx])
            : "NA";
          totalDuration += durationGrp ? Number(durationGrp[durationIdx]) : 0;
        }
      });
      return {
        index,
        Grouping: grp,
        "Converstion Time": formatDuration(totalDuration),
        Conversion:
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

export const getOverAllDuration = (durationsObj) => {
  if (durationsObj && durationsObj.metrics) {
    const durationMetric = durationsObj.metrics.find(
      (d) => d.title === "MetaStepTimeInfo"
    );
    if (durationMetric && durationMetric.rows && durationMetric.rows.length) {
      try {
        let total = 0;
        durationMetric.rows[0].forEach((r) => {
          total += Number(r);
        });
        return formatDuration(total);
      } catch (err) {
        return "NA";
      }
    }
  }
  return "NA";
};

export const getStepDuration = (durationsObj, index1, index2) => {
  let durationVal = "NA";
  if (durationsObj && durationsObj.metrics) {
    const durationMetric = durationsObj.metrics.find(
      (d) => d.title === "MetaStepTimeInfo"
    );
    if (
      durationMetric &&
      durationMetric.headers &&
      durationMetric.headers.length
    ) {
      try {
        const stepIndex = durationMetric.headers.findIndex(
          (elem) => elem === `step_${index1}_${index2}_time`
        );
        if (stepIndex > -1) {
          durationVal = formatDuration(durationMetric.rows[0][stepIndex]);
        }
      } catch (err) {
        console.log(err);
      }
    }
  }
  return durationVal;
};
