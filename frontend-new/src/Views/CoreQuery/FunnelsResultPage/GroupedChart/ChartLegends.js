import React from "react";
import { Text } from "../../../../components/factorsComponents";
import {
  legend_counts,
  charts_legend_length,
  DASHBOARD_WIDGET_SECTION,
} from "../../../../utils/constants";

function ChartLegends({
  colors,
  legends,
  arrayMapper,
  cardSize = 1,
  section,
  parentClassName = "flex justify-center py-3",
}) {
  let itemsCount = legend_counts[cardSize];
  return (
    <div className={parentClassName}>
      {legends.slice(0, itemsCount).map((legend, index) => {
        let label = arrayMapper.find((elem) => elem.mapper === legend)
          .eventName;
        if (
          label.length > charts_legend_length[cardSize] &&
          section === DASHBOARD_WIDGET_SECTION
        ) {
          label = label.substr(0, charts_legend_length[cardSize]) + "...";
        }
        return (
          <div key={legend} className="flex items-center">
            <div
              style={{
                backgroundColor: colors[index],
                width: "16px",
                height: "16px",
                borderRadius: "8px",
              }}
            ></div>
            <div className="px-2">
              <Text mini type="paragraph">
                {label}
              </Text>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default ChartLegends;
