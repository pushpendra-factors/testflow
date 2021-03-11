import React from "react";
import { SortData, getTitleWithSorter, formatCount } from "../../../utils/dataFormatter";
import { ATTRIBUTION_METHODOLOGY } from "../../../utils/constants";
import styles from './index.module.scss';

import { SVG } from '../../../components/factorsComponents';

export const getDifferentCampaingns = (data) => {
  const { headers } = data.result;
  const campaignIdx = headers.indexOf("Campaign");
  let differentCampaigns = new Set();
  data.result.rows.forEach((row) => {
    differentCampaigns.add(row[campaignIdx]);
  });
  return Array.from(differentCampaigns);
};

export const formatData = (data, event, visibleIndices, touchpoint) => {
  const { headers } = data;
  const touchpointIdx = headers.indexOf(touchpoint);
  const costIdx = headers.indexOf("Cost Per Conversion");
  const userIdx = headers.indexOf(`${event} - Users`);
  const rows = data.rows.filter(
    (_, index) => visibleIndices.indexOf(index) > -1
  );
  const result = rows.map((row) => {
    return [row[touchpointIdx], row[costIdx], row[userIdx]];
  });
  return SortData(result, 2, "descend");
};

export const formatGroupedData = (
  data,
  event,
  visibleIndices,
  attribution_method,
  attribution_method_compare
) => {
  const { headers } = data;
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  // const campaignIdx = headers.indexOf("Campaign");
  // const costIdx = headers.indexOf("Cost Per Conversion");
  // const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  let rows = data.rows.filter(
    (_, index) => visibleIndices.indexOf(index) > -1
  );
  rows = SortData(rows, userIdx, "descend");
  const chartData = [
    [`Conversions as Unique users (${attribution_method})`],
    [`Conversions as Unique users (${attribution_method_compare})`],
  ];
  rows.forEach((row) => {
    chartData[0].push(row[userIdx]);
    chartData[1].push(row[compareUsersIdx]);
  });
  return chartData;
};

const renderComparCell = (obj, xcl) => {
  let changeMetric = null;
  if(obj.change) {
    if(obj.change > 0 || obj.change < 0) {
      const change = Math.abs(obj.change);
      changeMetric = (
        <div className={`${styles.cmprCell__change} ${xcl}`}>
          <SVG name={obj.change > 0 ? `arrowLift` : `arrowDown`} size={16}></SVG>
          <span>
            {obj.change === 'Infinity' ? <>&#8734;</> : <>{change} &#37;</>} 
          </span>
        </div>  
      )
    } 
  }
  
  return (<div className={styles.cmprCell}>
    <span className={styles.cmprCell__first}>{obj.first}</span>
    <span className={styles.cmprCell__second}>{obj.second}</span>
    {changeMetric}
  </div>)
}

export const getCompareTableColumns = (
  currentSorter,
  handleSorting,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  event
) => {
  const result = [
    {
      title: touchpoint,
      dataIndex: touchpoint,
    },
    {
      title: getTitleWithSorter(
        "Impressions",
        "impressions",
        currentSorter,
        handleSorting
      ),
      dataIndex: "impressions",
      render: renderComparCell
    },
    {
      title: getTitleWithSorter(
        "Clicks",
        "clicks",
        currentSorter,
        handleSorting
      ),
      dataIndex: "clicks",
      render: renderComparCell
    },
    {
      title: getTitleWithSorter("Spend", "spend", currentSorter, handleSorting),
      dataIndex: "spend",
      render: renderComparCell
    },
    {
      title: getTitleWithSorter(
        "Visitors",
        "visitors",
        currentSorter,
        handleSorting
      ),
      dataIndex: "visitors",
      render: renderComparCell
    },
    {
      title: event,
      className: "tableParentHeader",
      children: [
        {
          title: (
            <div className="flex flex-col items-center justify-ceneter">
              <div>Conversion</div>
              <div style={{ fontSize: "10px", color: "#8692A3" }}>
                {
                  ATTRIBUTION_METHODOLOGY.find(
                    (m) => m.value === attribution_method
                  ).text
                }
              </div>
            </div>
          ),
          dataIndex: "conversion",
          className: "text-center",
          render: (obj) => renderComparCell(obj, "justify-center")
        },
        {
          title: (
            <div className="flex flex-col items-center justify-ceneter">
              <div>Cost per Conversion</div>
              <div style={{ fontSize: "10px", color: "#8692A3" }}>
                {
                  ATTRIBUTION_METHODOLOGY.find(
                    (m) => m.value === attribution_method
                  ).text
                }
              </div>
            </div>
          ),
          dataIndex: "cost",
          className: "text-center",
          render: (obj) => renderComparCell(obj, "justify-center")
        },
      ],
    },
  ];
  if (attribution_method_compare) {
    result[result.length - 1].children.push({
      title: (
        <div className="flex flex-col items-center justify-ceneter">
          <div>Conversion</div>
          <div style={{ fontSize: "10px", color: "#8692A3" }}>
            {
              ATTRIBUTION_METHODOLOGY.find(
                (m) => m.value === attribution_method_compare
              ).text
            }
          </div>
        </div>
      ),
      dataIndex: "conversion_compare",
      className: "text-center",
      render: (obj) => renderComparCell(obj, "justify-center")
    });
    result[result.length - 1].children.push({
      title: (
        <div className="flex flex-col items-center justify-ceneter">
          <div>Cost per Conversion</div>
          <div style={{ fontSize: "10px", color: "#8692A3" }}>
            {
              ATTRIBUTION_METHODOLOGY.find(
                (m) => m.value === attribution_method_compare
              ).text
            }
          </div>
        </div>
      ),
      dataIndex: "cost_compare",
      className: "text-center",
      render: (obj) => renderComparCell(obj, "justify-center")
    });
  }
  let linkedEventsColumns = [];
  if (linkedEvents.length) {
    linkedEventsColumns = linkedEvents.map((le) => {
      return {
        title: le.label,
        className: "tableParentHeader",
        children: [
          {
            title: (
              <div className="flex flex-col items-center justify-ceneter">
                <div>Users</div>
              </div>
            ),
            dataIndex: le.label + " - Users",
            className: "text-center",
            render: (obj) => renderComparCell(obj, "justify-center")
          },
          {
            title: (
              <div className="flex flex-col items-center justify-ceneter">
                <div>Cost per Conversion</div>
              </div>
            ),
            dataIndex: le.label + " - CPC",
            className: "text-center",
            render: (obj) => renderComparCell(obj, "justify-center")
          },
        ],
      }
    });
  }
  return [...result, ...linkedEventsColumns];
};

export const getTableColumns = (
  currentSorter,
  handleSorting,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents,
  event
) => {
  const result = [
    {
      title: touchpoint,
      dataIndex: touchpoint,
    },
    {
      title: getTitleWithSorter(
        "Impressions",
        "impressions",
        currentSorter,
        handleSorting
      ),
      dataIndex: "impressions",
    },
    {
      title: getTitleWithSorter(
        "Clicks",
        "clicks",
        currentSorter,
        handleSorting
      ),
      dataIndex: "clicks",
    },
    {
      title: getTitleWithSorter("Spend", "spend", currentSorter, handleSorting),
      dataIndex: "spend",
    },
    {
      title: getTitleWithSorter(
        "Visitors",
        "visitors",
        currentSorter,
        handleSorting
      ),
      dataIndex: "visitors",
    },
    {
      title: event,
      className: "tableParentHeader",
      children: [
        {
          title: (
            <div className="flex flex-col items-center justify-ceneter">
              <div>Conversion</div>
              <div style={{ fontSize: "10px", color: "#8692A3" }}>
                {
                  ATTRIBUTION_METHODOLOGY.find(
                    (m) => m.value === attribution_method
                  ).text
                }
              </div>
            </div>
          ),
          dataIndex: "conversion",
          className: "text-center",
        },
        {
          title: (
            <div className="flex flex-col items-center justify-ceneter">
              <div>Cost per Conversion</div>
              <div style={{ fontSize: "10px", color: "#8692A3" }}>
                {
                  ATTRIBUTION_METHODOLOGY.find(
                    (m) => m.value === attribution_method
                  ).text
                }
              </div>
            </div>
          ),
          dataIndex: "cost",
          className: "text-center",
        },
      ],
    },
  ];
  if (attribution_method_compare) {
    result[result.length - 1].children.push({
      title: (
        <div className="flex flex-col items-center justify-ceneter">
          <div>Conversion</div>
          <div style={{ fontSize: "10px", color: "#8692A3" }}>
            {
              ATTRIBUTION_METHODOLOGY.find(
                (m) => m.value === attribution_method_compare
              ).text
            }
          </div>
        </div>
      ),
      dataIndex: "conversion_compare",
      className: "text-center",
    });
    result[result.length - 1].children.push({
      title: (
        <div className="flex flex-col items-center justify-ceneter">
          <div>Cost per Conversion</div>
          <div style={{ fontSize: "10px", color: "#8692A3" }}>
            {
              ATTRIBUTION_METHODOLOGY.find(
                (m) => m.value === attribution_method_compare
              ).text
            }
          </div>
        </div>
      ),
      dataIndex: "cost_compare",
      className: "text-center",
    });
  }
  let linkedEventsColumns = [];
  if (linkedEvents.length) {
    linkedEventsColumns = linkedEvents.map((le) => {
      return {
        title: le.label,
        className: "tableParentHeader",
        children: [
          {
            title: (
              <div className="flex flex-col items-center justify-ceneter">
                <div>Users</div>
              </div>
            ),
            dataIndex: le.label + " - Users",
            className: "text-center",
          },
          {
            title: (
              <div className="flex flex-col items-center justify-ceneter">
                <div>Cost per Conversion</div>
              </div>
            ),
            dataIndex: le.label + " - CPC",
            className: "text-center",
          },
        ],
      }
    });
  }
  return [...result, ...linkedEventsColumns];
};

const constrComparisionCellData = (row, row2, index) => {
  return {
    first: formatCount(row[index], 1),
    second: row2? formatCount(row2[index], 1) : NaN,
    change: row2? calcChangePerc(row[index], row2[index]) : NaN
  }
}

export const getCompareTableData = (
  data,
  data2,
  event,
  searchText,
  currentSorter,
  attribution_method_compare,
  touchpoint,
  linkedEvents
) => {
  const { headers } = data;
  const touchpointIdx = headers.indexOf(touchpoint);
  const impressionsIdx = headers.indexOf("Impressions");
  const clicksIdx = headers.indexOf("Clicks");
  const spendIdx = headers.indexOf("Spend");
  const visitorsIdx = headers.indexOf("Website Visitors");
  const costIdx = headers.indexOf("Cost Per Conversion");
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  const data2Rows = data2.rows;
  const result = data.rows
    .map((row, index) => {
      const row2 = data2Rows.filter((r) => r[touchpointIdx] === row[touchpointIdx])[0];
      let resultantRow = {
        index,
        [touchpoint]: row[touchpointIdx],
        impressions: constrComparisionCellData(row, row2, impressionsIdx),
        clicks: constrComparisionCellData(row, row2, clicksIdx),
        spend: constrComparisionCellData(row, row2, spendIdx),
        visitors: constrComparisionCellData(row, row2, visitorsIdx),
        conversion: constrComparisionCellData(row, row2, userIdx),
        cost: constrComparisionCellData(row, row2, costIdx),
      };
      if (linkedEvents.length) {
        linkedEvents.forEach((le) => {
          const eventUsersIdx = headers.indexOf(`${le.label} - Users`);
          const eventCPCIdx = headers.indexOf(`${le.label} - CPC`);
          resultantRow[`${le.label} - Users`] = constrComparisionCellData(row, row2, eventUsersIdx);
          resultantRow[`${le.label} - CPC`] = constrComparisionCellData(row, row2, eventCPCIdx);
        });
      }
      if (attribution_method_compare) {
        resultantRow["conversion_compare"] = constrComparisionCellData(row, row2, [compareUsersIdx]);
        resultantRow["cost_compare"] = constrComparisionCellData(row, row2, [compareCostIdx]);
      }
      return resultantRow;
    })
    .filter(
      (row) =>
        row[touchpoint].toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );

  if (!currentSorter) {
    return SortData(result, "conversion", "descend");
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};

const calcChangePerc = (val1, val2) => {
  return formatCount(((val1 - val2) / val2 * 100), 1);
}

export const getTableData = (
  data,
  event,
  searchText,
  currentSorter,
  attribution_method_compare,
  touchpoint,
  linkedEvents
) => {
  const { headers } = data;
  const touchpointIdx = headers.indexOf(touchpoint);
  const impressionsIdx = headers.indexOf("Impressions");
  const clicksIdx = headers.indexOf("Clicks");
  const spendIdx = headers.indexOf("Spend");
  const visitorsIdx = headers.indexOf("Website Visitors");
  const costIdx = headers.indexOf("Cost Per Conversion");
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  const result = data.rows
    .map((row, index) => {
      let resultantRow = {
        index,
        [touchpoint]: row[touchpointIdx],
        impressions: formatCount(row[impressionsIdx], 1),
        clicks: formatCount(row[clicksIdx], 1),
        spend: formatCount(row[spendIdx], 1),
        visitors: formatCount(row[visitorsIdx], 1),
        conversion: formatCount(row[userIdx], 1),
        cost: formatCount(row[costIdx], 1),
      };
      if (linkedEvents.length) {
        linkedEvents.forEach((le) => {
          const eventUsersIdx = headers.indexOf(`${le.label} - Users`);
          const eventCPCIdx = headers.indexOf(`${le.label} - CPC`);
          resultantRow[`${le.label} - Users`] = formatCount(row[eventUsersIdx], 0);
          resultantRow[`${le.label} - CPC`] = formatCount(row[eventCPCIdx], 0);
        });
      }
      if (attribution_method_compare) {
        resultantRow["conversion_compare"] = row[compareUsersIdx];
        resultantRow["cost_compare"] = formatCount(row[compareCostIdx], 0);
      }
      return resultantRow;
    })
    .filter(
      (row) =>
        row[touchpoint].toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );

  if (!currentSorter) {
    return SortData(result, "conversion", "descend");
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};
