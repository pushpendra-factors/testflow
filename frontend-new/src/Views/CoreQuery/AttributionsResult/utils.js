import React from "react";
import { SortData, getTitleWithSorter } from "../../../utils/dataFormatter";

export const getDifferentCampaingns = (data) => {
  const { headers } = data.result;
  const campaignIdx = headers.indexOf("Campaign");
  let differentCampaigns = new Set();
  data.result.rows.forEach((row) => {
    differentCampaigns.add(row[campaignIdx]);
  });
  return Array.from(differentCampaigns);
};

export const formatData = (data, event, visibleIndices) => {
  const { headers } = data.result;
  const campaignIdx = headers.indexOf("Campaign");
  const costIdx = headers.indexOf("Cost Per Conversion");
  const userIdx = headers.indexOf(`${event} - Users`);
  const rows = data.result.rows.filter(
    (_, index) => visibleIndices.indexOf(index) > -1
  );
  const result = rows.map((row) => {
    return [row[campaignIdx], row[costIdx], row[userIdx]];
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
  attribution_method_compare
) => {
  const result = [
    {
      title: "Marketing Touchpoint",
      dataIndex: "campaign",
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
                {attribution_method}
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
                {attribution_method}
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
            {attribution_method_compare}
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
            {attribution_method_compare}
          </div>
        </div>
      ),
      dataIndex: "cost_compare",
      className: "text-center",
    });
  }
  return result;
};

export const getTableData = (
  data,
  event,
  searchText,
  currentSorter,
  attribution_method_compare
) => {
  const { headers } = data.result;
  const campaignIdx = headers.indexOf("Campaign");
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
      const resultantRow = {
        index,
        campaign: row[campaignIdx],
        impressions: row[impressionsIdx],
        clicks: row[clicksIdx],
        spend: row[spendIdx],
        visitors: row[visitorsIdx],
        conversion: row[userIdx],
        cost: row[costIdx],
      };
      if (attribution_method_compare) {
        return {
          ...resultantRow,
          conversion_compare: row[compareUsersIdx],
          cost_compare: row[compareCostIdx],
        };
      }
      return resultantRow;
    })
    .filter(
      (row) => row.campaign.toLowerCase().indexOf(searchText.toLowerCase()) > -1
    );

  if (!currentSorter) {
    return SortData(result, "conversion", "descend");
  }
  return SortData(result, currentSorter.key, currentSorter.order);
};
