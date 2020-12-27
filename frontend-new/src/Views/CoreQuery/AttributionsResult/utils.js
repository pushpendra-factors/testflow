import React from "react";
import { SortData, getTitleWithSorter, formatCount } from "../../../utils/dataFormatter";
import { ATTRIBUTION_METHODOLOGY } from "../../../utils/constants";

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
  const { headers } = data.result;
  const touchpointIdx = headers.indexOf(touchpoint);
  const costIdx = headers.indexOf("Cost Per Conversion");
  const userIdx = headers.indexOf(`${event} - Users`);
  const rows = data.result.rows.filter(
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
  const { headers } = data.result;
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  // const campaignIdx = headers.indexOf("Campaign");
  // const costIdx = headers.indexOf("Cost Per Conversion");
  // const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  let rows = data.result.rows.filter(
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

export const getTableColumns = (
  currentSorter,
  handleSorting,
  attribution_method,
  attribution_method_compare,
  touchpoint,
  linkedEvents
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
      title: "Opportunities (as unique users)",
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
        title: `${le.label} - Users`,
        dataIndex: le.label,
      };
    });
  }
  return [...result, ...linkedEventsColumns];
};

export const getTableData = (
  data,
  event,
  searchText,
  currentSorter,
  attribution_method_compare,
  touchpoint,
  linkedEvents
) => {
  const { headers } = data.result;
  const touchpointIdx = headers.indexOf(touchpoint);
  const impressionsIdx = headers.indexOf("Impressions");
  const clicksIdx = headers.indexOf("Clicks");
  const spendIdx = headers.indexOf("Spend");
  const visitorsIdx = headers.indexOf("Website Visitors");
  const costIdx = headers.indexOf("Cost Per Conversion");
  const userIdx = headers.indexOf(`${event} - Users`);
  const compareUsersIdx = headers.indexOf(`Compare - Users`);
  const compareCostIdx = headers.indexOf(`Compare Cost Per Conversion`);
  const result = data.result.rows
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
          const eventIdx = headers.indexOf(`${le.label} - Users`);
          resultantRow[le.label] = row[eventIdx];
        });
      }
      if (attribution_method_compare) {
        resultantRow["conversion_compare"] = row[compareUsersIdx];
        resultantRow["cost_compare"] = row[compareCostIdx];
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
