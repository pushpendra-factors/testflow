import React from 'react';
import MultipleEventsWithBreakdown from '../EventsAnalytics/MultipleEventsWIthBreakdown';
import { MultipleEventsMultipleBreakdown as data } from '../EventsAnalytics/SampleResponse';
// import { MultipleEventsMultipleBreakdownUserData as data } from '../EventsAnalytics/SampleResponse';

function DummyCharts() {
  const queries = ['www.acme.com/pricing', 'www.acme.com/solutions', 'www.acme.com/product'];

  const appliedBreakdown = [
    {
      prop_category: 'event',
      property: 'Browser',
      prop_type: 'categorical',
      eventValue: 'www.acme.com/pricing'
    },
    {
      prop_category: 'event',
      property: 'Page Load Time',
      prop_type: 'numerical',
      eventValue: 'www.acme.com/solutions'
    },
    {
      prop_category: 'event',
      property: 'Device Type',
      prop_type: 'categorical',
      eventValue: 'www.acme.com/product'
    },
    {
      prop_category: 'event',
      property: 'Page Spent Time',
      prop_type: 'numerical',
      eventValue: 'www.acme.com/product'
    }
  ];

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
