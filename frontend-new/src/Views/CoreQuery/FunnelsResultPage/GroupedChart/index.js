import React, {
  useEffect,
  useState,
  useContext,
  forwardRef,
  useImperativeHandle,
} from 'react';
import { formatData, getVisibleData } from '../utils';
import BarChart from './Chart';
import FunnelsResultTable from '../FunnelsResultTable';
import NoDataChart from '../../../../components/NoDataChart';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import FunnelsScatterPlot from './FunnelsScatterPlot';
import { CHART_TYPE_BARCHART } from '../../../../utils/constants';

const GroupedChart = forwardRef(
  (
    {
      resultState,
      queries,
      breakdown,
      isWidgetModal,
      arrayMapper,
      section,
      chartType,
    },
    ref
  ) => {
    const {
      coreQueryState: { savedQuerySettings },
    } = useContext(CoreQueryContext);
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter && Array.isArray(savedQuerySettings.sorter)
        ? savedQuerySettings.sorter
        : []
    );
    const [eventsData, setEventsData] = useState([]);
    const [groups, setGroups] = useState([]);

    useImperativeHandle(ref, () => {
      return {
        currentSorter: { sorter },
      };
    });

    useEffect(() => {
      const { groups: appliedGroups, events } = formatData(
        resultState.data,
        arrayMapper
      );
      setGroups(appliedGroups);
      setEventsData(events);
    }, [resultState.data, arrayMapper]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(groups, sorter));
    }, [groups, sorter]);

    if (!visibleProperties.length) {
      return (
        <div className='mt-4 flex justify-center items-center w-full h-full'>
          <NoDataChart />
        </div>
      );
    }

    let chart = null;

    if (chartType === CHART_TYPE_BARCHART) {
      chart = (
        <BarChart
          isWidgetModal={isWidgetModal}
          groups={visibleProperties}
          eventsData={eventsData}
          arrayMapper={arrayMapper}
          section={section}
          durations={resultState.data.meta}
        />
      );
    } else {
      chart = (
        <div className='w-full'>
          <FunnelsScatterPlot
            visibleProperties={visibleProperties}
            arrayMapper={arrayMapper}
            section={section}
          />
        </div>
      );
    }

    return (
      <div className='flex items-center justify-center flex-col'>
        {chart}
        <div className='mt-12 w-full'>
          <FunnelsResultTable
            breakdown={breakdown}
            queries={queries}
            groups={groups}
            visibleProperties={visibleProperties}
            setVisibleProperties={setVisibleProperties}
            chartData={eventsData}
            arrayMapper={arrayMapper}
            resultData={resultState.data}
            sorter={sorter}
            setSorter={setSorter}
          />
        </div>
      </div>
    );
  }
);

export default GroupedChart;
