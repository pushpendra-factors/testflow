import React from 'react';
import MultipleEventsWithBreakdown from '../EventsAnalytics/MultipleEventsWIthBreakdown';
// import { MultipleEventsMultipleBreakdown as data } from '../EventsAnalytics/SampleResponse';
import { MultipleEventsMultipleBreakdownUserData as data } from '../EventsAnalytics/SampleResponse';

function DummyCharts() {
  const queries = ['www.acme.com/product/collaboration', 'www.acme.com/product/analytics'];

  const appliedBreakdown = [{
    prop_category: 'event',
    property: 'Browser',
    prop_type: 'categorical'
  }];

  return (
        <MultipleEventsWithBreakdown
            breakdown={appliedBreakdown}
            queries={queries}
            // page="totalEvents"
            page="totalUsers"
            resultState={{ data: data.result_group[0] }}
        />
  );
}

export default DummyCharts;
