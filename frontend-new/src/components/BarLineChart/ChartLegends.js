import React from "react";

function ChartLegends() {
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
        <div
          style={{ color: "#08172B", fontSize: "0.875rem", lineHeight: "1.25rem" }}
        >
          Conversions as Unique users (Last Touch)
        </div>
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
        <div
          style={{ color: "#08172B", fontSize: "0.875rem", lineHeight: "1.25rem" }}
        >
          Cost per conversion
        </div>
      </div>
    </div>
  );
}

export default ChartLegends;
