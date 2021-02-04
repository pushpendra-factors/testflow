import React from "react";
import { Text } from "../factorsComponents";
import { linechart_legend_length } from "../../utils/constants";

function DashboardWidgetLegends({
  colors,
  legends,
  cardSize,
  arrayMapper,
  parentClassName = "flex justify-center mb-4 py-3",
}) {
  return (
    <div className={parentClassName}>
      {legends.map((legend, index) => {
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
                {label.length > linechart_legend_length[cardSize]
                  ? label.substr(0, linechart_legend_length[cardSize]) + "..."
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
