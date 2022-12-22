import React from 'react';
import MetricChart from '../MetricChart';

export default {
  title: 'Components/MetricChart',
  component: MetricChart
};

export const DefaultChart = () => {
  return <MetricChart value={100} />;
};

export const WithComparison = () => {
  return <MetricChart value={100} compareValue={80} showComparison={true} />;
};
