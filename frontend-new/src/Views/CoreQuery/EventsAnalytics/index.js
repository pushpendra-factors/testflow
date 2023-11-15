import React from 'react';
import NoBreakdownCharts from './NoBreakdownCharts';
import SingleEventSingleBreakdown from './SingleEventSingleBreakdown';
import SingleEventMultipleBreakdown from './SingleEventMultipleBreakdown';
import MultipleEventsWithBreakdown from './MultipleEventsWIthBreakdown';
import {
  EACH_USER_TYPE,
  ANY_USER_TYPE,
  ALL_USER_TYPE,
} from '../../../utils/constants';
import EventBreakdownCharts from './EventBreakdown/EventBreakdownCharts';
import { getErrorMessage } from 'Utils/global';

function EventsAnalytics({
  queries,
  breakdown,
  resultState,
  page,
  arrayMapper,
  title = 'chart',
  chartType,
  durationObj,
  breakdownType,
  section,
  renderedCompRef,
}) {
  let content = null;

  const [errMsg,setErrMsg]=useState('');
  useEffect(() => {
    const errorMessage = getErrorMessage(resultState);
    setErrMsg(errorMessage);
  }, [resultState]);

  
  if (breakdownType === EACH_USER_TYPE) {
    if (resultState.data && !resultState.data.metrics.rows.length) {
      content = (
        <div className='flex justify-center items-center w-full h-full pt-4 pb-4'>
        <NoDataInTimeRange message={errMsg}/>
      </div>
      );
    }

    if (resultState.data && resultState.data.metrics.rows.length) {
      if (!breakdown.length) {
        content = (
          <NoBreakdownCharts
            queries={queries}
            resultState={resultState}
            page={page}
            chartType={chartType}
            arrayMapper={arrayMapper}
            durationObj={durationObj}
            section={section}
            renderedCompRef={renderedCompRef}
          />
        );
      }

      if (queries.length === 1 && breakdown.length === 1) {
        content = (
          <SingleEventSingleBreakdown
            queries={queries}
            breakdown={breakdown}
            resultState={resultState}
            page={page}
            chartType={chartType}
            title={title}
            durationObj={durationObj}
            section={section}
            renderedCompRef={renderedCompRef}
          />
        );
      }

      if (queries.length > 1 && breakdown.length) {
        content = (
          <MultipleEventsWithBreakdown
            queries={queries}
            breakdown={breakdown}
            resultState={resultState}
            page={page}
            chartType={chartType}
            title={title}
            durationObj={durationObj}
            section={section}
            ref={renderedCompRef}
          />
        );
      }

      if (queries.length === 1 && breakdown.length > 1) {
        content = (
          <SingleEventMultipleBreakdown
            queries={queries}
            breakdown={breakdown}
            resultState={resultState}
            page={page}
            chartType={chartType}
            title={title}
            durationObj={durationObj}
            section={section}
            ref={renderedCompRef}
          />
        );
      }
    }
  }

  if (breakdownType === ANY_USER_TYPE || breakdownType === ALL_USER_TYPE) {
    content = (
      <EventBreakdownCharts
        section={section}
        data={resultState.data}
        breakdown={breakdown}
        ref={renderedCompRef}
      />
    );
  }

  return content;
}

export default EventsAnalytics;
