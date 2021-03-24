import React from "react";
import { Text } from "../factorsComponents";

function ChartLegends({ legends }) {
  return (
    <div className="flex justify-center items-center">
      <div className="flex items-center mr-2">
        <div
          className="mr-1"
          style={{
            backgroundColor: "#4d7db4",
            width: "16px",
            height: "16px",
            borderRadius: "8px",
          }}
        ></div>
        <Text extraClass="mb-0 text-sm" type="title" color="grey-8">{legends[0]}</Text>
      </div>
      <div className="flex items-center mr-2">
        <div
          className="mr-1"
          style={{
            backgroundColor: "#D4787D",
            width: "16px",
            height: "4px",
          }}
        ></div>
        <Text extraClass="mb-0 text-sm" type="title" color="grey-8">{legends[1]}</Text>
      </div>
    </div>
  );
}

export default ChartLegends;
