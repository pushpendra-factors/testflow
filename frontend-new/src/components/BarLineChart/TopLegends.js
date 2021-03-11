import React from "react";
import { Text } from "../factorsComponents";

function TopLegends({
  parentClassName = "flex justify-center mb-4 py-3",
  cardSize,
}) {
  return (
    <div className={parentClassName}>
      <div className="flex items-center">
        <div
          style={{
            backgroundColor: "rgb(77, 125, 180)",
            width: "16px",
            height: "16px",
            borderRadius: "8px",
          }}
        ></div>
        <div className="px-2">
          <Text mini type="paragraph">
            Opportunities
          </Text>
        </div>
      </div>
      {cardSize !== 2 ? (
        <div className="flex items-center">
          <div
            style={{
              backgroundColor: "rgb(212, 120, 125)",
              width: "16px",
              height: "16px",
              borderRadius: "8px",
            }}
          ></div>
          <div className="px-2">
            <Text mini type="paragraph">
              Cost Per Conversion
            </Text>
          </div>
        </div>
      ) : null}
    </div>
  );
}

export default TopLegends;
