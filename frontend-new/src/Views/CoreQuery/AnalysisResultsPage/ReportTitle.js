import React from "react";
import moment from "moment";
import { Button } from "antd";
import { SVG } from "../../../components/factorsComponents";

function ReportTitle({ title, setDrawerVisible, queryDetail }) {
  return (
    <div style={{ borderBottom: "1px solid #E7E9ED" }} className="pb-4">
      <div
        style={{ fontSize: "32px", letterSpacing: "-0.02em", color: title ? "#3E516C" : "#8692A3" }}
        className="leading-9 font-semibold"
      >
        {title || `Untitled Analysis ${moment().format("DD/MM/YYYY")}`}
      </div>
      <div className="flex items-center mt-3">
        <div
          style={{ color: "#3E516C" }}
          className="mr-2 text-base leading-6 font-medium"
        >
          {queryDetail}
        </div>
        <div>
          <Button
            style={{
              display: "flex",
              padding: "4px",
              color: "#0E2647",
              opacity: 0.56,
            }}
            className="items-center"
            size={"large"}
            type="text"
            onClick={setDrawerVisible.bind(this, true)}
          >
            <SVG
              extraClass="mr-1"
              name="edit_query"
              size="24"
              color={"#0E2647"}
            />
            Edit
          </Button>
        </div>
      </div>
    </div>
  );
}

export default ReportTitle;
