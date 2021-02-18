import React, { useCallback } from "react";
import moment from "moment";
import { Button } from "antd";
import { SVG, Text } from "../../../components/factorsComponents";
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
      setDrawerVisible();
    }
  }, [section, setDrawerVisible]);

  return (
    <div className="pb-2 border-bottom--thin-2">
      <div className="flex justify-between items-center"> 
        <Text type={"title"} level={3} weight={"bold"} extraClass={'m-0'}> {title || `Untitled Analysis ${moment().format("DD/MM/YYYY")}`} </Text>
        {section === DASHBOARD_MODAL ? ( 
            <Button  
            type={'text'}
              onClick={onReportClose.bind(this, false)}
              icon={<SVG name="Remove" />}
            />  
        ) : null}
      </div>
      <div className="flex items-center"> 
      <div className={'fa-title--editable flex items-center cursor-pointer '} onClick={queryType !== QUERY_TYPE_WEB ? handleClick : null}>
        <Text type={"title"} level={6} color={'grey'} extraClass={'m-0 mr-2'}> {queryDetail} </Text>  
        <SVG 
                name="edit" 
                color={'grey'}
              /> 
      </div> 

      </div>
    </div>
  );
}

export default ReportTitle;
