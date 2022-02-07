import React, {
  useState,
  useContext,
  useCallback,
  useImperativeHandle,
  forwardRef,
} from 'react';
import { defaultSortProp } from './utils';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import { getNewSorterState } from '../../../../utils/dataFormatter';
import HorizontalBarChartTable from './HorizontalBarChartTable';
import NoBreakdownTable from './NoBreakdownTable';

const NoBreakdownCharts = forwardRef(
  ({ data, title = 'Profile-chart', section, queries, groupAnalysis }, ref) => {
    const {
      coreQueryState: { savedQuerySettings },
    } = useContext(CoreQueryContext);

    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : defaultSortProp()
    );

    const handleSorting = useCallback((prop) => {
      setSorter((currentSorter) => {
        return getNewSorterState(currentSorter, prop);
      });
    }, []);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter },
      };
    });

    const chart = (
      <div className='w-full'>
        <HorizontalBarChartTable queries={queries} groupAnalysis={groupAnalysis} data={data} />
      </div>
    );

    const table = (
      <div className='mt-12 w-full'>
        <NoBreakdownTable
          data={data}
          sorter={sorter}
          handleSorting={handleSorting}
          isWidgetModal={false}
          section={section}
          queries={queries}
          reportTitle={title}
          groupAnalysis={groupAnalysis}
        />
      </div>
    );

    return (
      <div className='flex items-center justify-center flex-col'>
        {chart}
        {table}
      </div>
    );
  }
);

export default NoBreakdownCharts;
