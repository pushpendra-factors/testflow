import React, { useEffect, useState } from "react";
import { generateUngroupedChartsData } from "../utils";

import Chart from "./Chart";
import FunnelsResultTable from "../FunnelsResultTable";
import { DASHBOARD_MODAL } from "../../../../utils/constants";

function UngroupedChart({ resultState, queries, section, arrayMapper }) {
  const [chartData, setChartData] = useState([]);

  useEffect(() => {
    const formattedData = generateUngroupedChartsData(
      resultState.data,
      arrayMapper
    );
    setChartData(formattedData);
  }, [arrayMapper, resultState.data]);

  if (!chartData.length) {
    return null;
  }

  return (
    <div className="flex items-center justify-center flex-col">
      <Chart chartData={chartData} arrayMapper={arrayMapper} />

      <div className="mt-12 w-full">
        <FunnelsResultTable
          isWidgetModal={section === DASHBOARD_MODAL}
          chartData={chartData}
          breakdown={[]}
          queries={queries}
          groups={[]}
          arrayMapper={arrayMapper}
        />
      </div>
    </div>
  );
}

export default UngroupedChart;
