import React from "react";
import { Text } from "../factorsComponents";
import { linechart_legend_length } from "../../utils/constants";

function TopLegends({
  colors,
  parentClassName = "flex justify-center mb-4 py-3",
  cardSize,
}) {
  const legends = Object.keys(colors);
  return (
    <div className={parentClassName}>
      {legends.map((legend, index) => {
        return (
          <div key={legend + index} className="flex items-center">
            <div
              style={{
                backgroundColor: Object.values(colors)[index],
                width: "16px",
                height: "16px",
                borderRadius: "8px",
              }}
            ></div>
            <div className="px-2">
              <Text mini type="paragraph">
                {legend.length > linechart_legend_length
                  ? legend.substr(0, linechart_legend_length) + "..."
                  : legend}
              </Text>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default TopLegends;
