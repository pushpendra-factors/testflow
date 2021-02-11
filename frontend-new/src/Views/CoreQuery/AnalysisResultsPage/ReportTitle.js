import React, { useCallback } from "react";
import moment from "moment";
import { Button } from "antd";
import { SVG } from "../../../components/factorsComponents";
import {
  REPORT_SECTION,
  DASHBOARD_MODAL,
  QUERY_TYPE_WEB,
} from "../../../utils/constants";

function ReportTitle({
  title,
  setDrawerVisible,
  queryDetail,
  section,
  onReportClose,
  queryType,
}) {
  const handleClick = useCallback(() => {
    if (section === REPORT_SECTION) {
      setDrawerVisible(true);
    }
    if (section === DASHBOARD_MODAL) {
      console.log("adaddad");
      setDrawerVisible();
    }
  }, [section, setDrawerVisible]);

  return (
    <div style={{ borderBottom: "1px solid #E7E9ED" }} className="pb-4">
      <div className="flex justify-between items-center">
        <div
          style={{
            fontSize: "32px",
            letterSpacing: "-0.02em",
            color: title ? "#3E516C" : "#8692A3",
          }}
          className="leading-9 font-semibold"
        >
          {title || `Untitled Analysis ${moment().format("DD/MM/YYYY")}`}
        </div>
        {section === DASHBOARD_MODAL ? (
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
              onClick={onReportClose.bind(this, false)}
            >
              <SVG extraClass="mr-1" name="close" size="32" color={"#0E2647"} />
            </Button>
          </div>
        ) : null}
      </div>
      <div className="flex items-center mt-3">
        <div
          style={{ color: "#3E516C" }}
          className="mr-2 text-base leading-6 font-medium"
        >
          {queryDetail}
        </div>
        {queryType !== QUERY_TYPE_WEB ? (
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
              onClick={handleClick}
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
        ) : null}
      </div>
    </div>
  );
}

export default ReportTitle;
