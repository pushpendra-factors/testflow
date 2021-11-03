import React from "react";
import NoBreakdownCharts from "./NoBreakdownCharts";
import BreakdownCharts from "./BreakdownCharts";

function CampaignAnalytics({
  resultState,
  arrayMapper,
  campaignState,
  chartType,
  currMetricsValue,
  section,
  durationObj
}) {
  const { group_by: breakdown } = campaignState;
    

  const APIResponse = {"result":[{"headers":["aggregate"],"rows":[[3896]],"meta":{"query":{"cl":"","ty":"","ec":"","ewp":null,"gbp":null,"gup":null,"gbt":null,"tz":"","fr":0,"to":0,"ovp":false,"sse":0,"see":0,"agFn":"","agPr":"","agEn":""},"currency":"","metrics":null}},{"headers":["aggregate"],"rows":[[3893]],"meta":{"query":{"cl":"","ty":"","ec":"","ewp":null,"gbp":null,"gup":null,"gbt":null,"tz":"","fr":0,"to":0,"ovp":false,"sse":0,"see":0,"agFn":"","agPr":"","agEn":""},"currency":"","metrics":null}}]}
  const DuractionObjNew = {
  "from": "2021-10-09T18:30:00.000Z",
  "to": "2021-10-13T18:29:59.999Z",
  "frequency": "date",
  "dateType": "this_week"
} 
  
  let content = null;
  
  if (breakdown.length) {
    content = (
      <BreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        // data={resultState.data}
        data={APIResponse}
        breakdown={breakdown}
        currentEventIndex={currMetricsValue}
        section={section}
      />
    );
  } else {
    content = (
      <NoBreakdownCharts
        arrayMapper={arrayMapper}
        chartType={chartType}
        data={APIResponse}
        section={section}
        durationObj={DuractionObjNew}
      />
    );
  }

  return <>{content}</>;
}

export default CampaignAnalytics;
