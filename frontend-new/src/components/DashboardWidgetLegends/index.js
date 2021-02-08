import React from "react";
import { Text } from "../factorsComponents";
import { linechart_legend_length, legend_counts } from "../../utils/constants";

function DashboardWidgetLegends({
  colors,
  legends,
  cardSize,
  arrayMapper,
  parentClassName = "flex justify-center py-3",
}) {
  let itemsCount = legend_counts[cardSize];
  return (
    <div className={parentClassName}>
      {legends.slice(0, itemsCount).map((legend, index) => {
        const label = legend
          .split(",")
          .filter((elem) => elem)
          .join(",");
        const key = arrayMapper.find((m) => m.index === index).mapper;
        return (
          <div key={legend + index} className="flex items-center">
            <div
              style={{
                backgroundColor: colors[key],
                width: "16px",
                height: "16px",
                borderRadius: "8px",
              }}
            ></div>
            <div className="px-2">
              <Text mini type="paragraph">
                {label.length > linechart_legend_length
                  ? label.substr(0, linechart_legend_length) + "..."
                  : label}
              </Text>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default DashboardWidgetLegends;
