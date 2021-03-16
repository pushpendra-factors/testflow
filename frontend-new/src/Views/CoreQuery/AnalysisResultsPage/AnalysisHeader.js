import React, { useCallback, useEffect, useState } from "react";
import styles from "./index.module.scss";
import { SVG, Text } from "../../../components/factorsComponents";
import { Button } from "antd";
import moment from "moment";
import { EVENT_BREADCRUMB } from "../../../utils/constants";
import SaveQuery from "../../../components/SaveQuery";

function AnalysisHeader({
  queryType,
  onBreadCrumbClick,
  requestQuery,
  queryTitle,
  setQuerySaved,
  breakdownType
}) {
  const [showSaveModal, setShowSaveModal] = useState(false);

  const addShadowToHeader = useCallback(() => {
    const scrollTop =
      window.pageYOffset !== undefined
        ? window.pageYOffset
        : (
            document.documentElement ||
            document.body.parentNode ||
            document.body
          ).scrollTop;
    if (scrollTop > 0) {
      document.getElementById("app-header").style.filter =
        "drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))";
    } else {
      document.getElementById("app-header").style.filter = "none";
    }
  }, []);

  useEffect(() => {
    document.addEventListener("scroll", addShadowToHeader);
    return () => {
      document.removeEventListener("scroll", addShadowToHeader);
    };
  }, [addShadowToHeader]);

  return (
    <div
      id="app-header"
      className={`bg-white z-50	flex fixed items-center justify-between py-3 px-8 ${styles.topHeader}`}
    >
      <div
        onClick={onBreadCrumbClick}
        className="flex items-center cursor-pointer"
      > 
          <Button  
            size={"large"}
            type="text"
            icon={<SVG size={32} name="Brand"/>}
            className={'mr-2'}
          />
        <div>
        <Text type={'title'} level={7} extraClass={'m-0 mt-1'} color={'grey'} lineHeight={'small'} >
            {queryTitle
              ? `Reports / ${EVENT_BREADCRUMB[queryType]} / ${queryTitle}`
              : `Reports / ${EVENT_BREADCRUMB[queryType]} / Untitled Analysis${" "}
            ${moment().format("DD/MM/YYYY")}`} 
        </Text>
        </div>
      </div>
      <div className="flex items-center">
        {/* <Button
          style={{ display: "flex", padding: "4px" }}
          className="items-center mr-4"
          size={"small"}
          type="text"
        >
          <SVG name={"annotation"} />
        </Button> */}

        {/* <Tooltip placement="bottom" title={"Created by Jitesh Kriplani"}>
          <div className="mr-4 cursor-pointer">
            <SVG name={"report_user"} />
          </div>
        </Tooltip> */}

        {/* <Button
          // onClick={setVisible.bind(this, true)}
          style={{ display: "flex" }}
          className="items-center"
          type="primary"
          icon={
            <SVG extraClass="mr-1" name={"save"} size={24} color="#FFFFFF" />
          }
        >
          Save
        </Button> */}
        <SaveQuery
          requestQuery={requestQuery}
          visible={showSaveModal}
          setVisible={setShowSaveModal}
          queryType={queryType}
          setQuerySaved={setQuerySaved}
          breakdownType={breakdownType}
        />

        <Button 
          size={"large"}
          type="text"
          icon={<SVG size={20} name={"threedot"} />}
          className={'ml-2'}
        >
          
        </Button>
      </div>
    </div>
  );
}

export default AnalysisHeader;
