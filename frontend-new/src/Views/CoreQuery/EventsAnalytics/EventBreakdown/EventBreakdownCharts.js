import React, {
  useState,
  useEffect,
  forwardRef,
  useContext,
  useImperativeHandle
} from 'react';
import { useSelector } from 'react-redux';
import { formatData, getDefaultSortProp, getVisibleData } from './utils';
import BarChart from '../../../../components/BarChart';
import EventBreakdownTable from './EventBreakdownTable';
import ChartHeader from '../../../../components/SparkLineChart/ChartHeader';
import { CoreQueryContext } from '../../../../contexts/CoreQueryContext';
import { CHART_COLOR_1 } from '../../../../constants/color.constants';

const EventBreakdownCharts = forwardRef(
  ({ data, breakdown, section, durationObj, resultState }, ref) => {
    const {
      coreQueryState: { savedQuerySettings }
    } = useContext(CoreQueryContext);

    const [chartsData, setChartsData] = useState([]);
    const [visibleProperties, setVisibleProperties] = useState([]);
    const [sorter, setSorter] = useState(
      savedQuerySettings.sorter || getDefaultSortProp()
    );
    const { eventNames } = useSelector((state) => state.coreQuery);

    useEffect(() => {
      const formattedData = formatData(data);
      setChartsData(formattedData);
    }, [data]);

    useEffect(() => {
      setVisibleProperties(getVisibleData(chartsData, sorter));
    }, [chartsData, sorter]);

    useImperativeHandle(ref, () => ({
      currentSorter: { sorter }
    }));

    if (!chartsData.length) {
      return (
        <div className='h-64 flex items-center justify-center w-full'>
          No Data Found!
        </div>
      );
    }

    let chart = null;

    const table = (
      <div className='mt-12 w-full'>
        <EventBreakdownTable
          data={chartsData}
          breakdown={breakdown}
          setVisibleProperties={setVisibleProperties}
          visibleProperties={visibleProperties}
          sorter={sorter}
          setSorter={setSorter}
          durationObj={durationObj}
          resultState={resultState}
        />
      </div>
    );

    if (breakdown.length) {
      chart = <BarChart section={section} chartData={visibleProperties} />;
    } else {
      chart = (
        <ChartHeader
          eventNames={eventNames}
          total={data.rows[0]}
          query='Count'
          bgColor={CHART_COLOR_1}
        />
      );
    }

    return (
      <div className='flex items-center justify-center flex-col'>
        {chart}
        {table}
      </div>
    );
  }
);

export default EventBreakdownCharts;
