import React, { useState, useCallback, useContext } from 'react';
import { defaultSortProp } from '../../CoreQuery/ProfilesResultPage/NoBreakdownCharts/utils';
import NoBreakdownTable from '../../CoreQuery/ProfilesResultPage/NoBreakdownCharts/NoBreakdownTable';
import { getNewSorterState } from '../../../utils/dataFormatter';
import HorizontalBarChartTable from '../../CoreQuery/ProfilesResultPage/NoBreakdownCharts/HorizontalBarChartTable';
import {
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  CHART_TYPE_TABLE
} from '../../../utils/constants';
import { DashboardContext } from '../../../contexts/DashboardContext';

const NoBreakdownCharts = ({ chartType, data, unit, section, queries }) => {
  const { handleEditQuery } = useContext(DashboardContext);
  const [sorter, setSorter] = useState(defaultSortProp());

  const handleSorting = useCallback((prop) => {
    setSorter((currentSorter) => {
      return getNewSorterState(currentSorter, prop);
    });
  }, []);

  let chartContent = null;
  let tableContent = null;

  // if (chartType === CHART_TYPE_TABLE) {
  //   tableContent = (
  //     <div
  //       onClick={handleEditQuery}
  //       style={{ color: '#5949BC' }}
  //       className='mt-3 font-medium text-base cursor-pointer flex justify-end item-center'
  //     >
  //       Show More &rarr;
  //     </div>
  //   );
  // }

  if (chartType === CHART_TYPE_TABLE) {
    chartContent = (
      <NoBreakdownTable
        data={data}
        sorter={sorter}
        handleSorting={handleSorting}
        isWidgetModal={false}
        section={section}
        queries={queries}
      />
    );
  }

  if (chartType === CHART_TYPE_HORIZONTAL_BAR_CHART) {
    chartContent = (
      <HorizontalBarChartTable
        data={data}
        queries={queries}
        cardSize={unit.cardSize}
        isDashboardWidget={true}
      />
    );
  }

  return (
    <div className={`w-full flex-1`}>
      {chartContent}
      {tableContent}
    </div>
  );
};

export default NoBreakdownCharts;
