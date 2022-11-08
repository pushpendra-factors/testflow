import React from 'react';
import { CHART_COLOR_1 } from '../../../constants/color.constants';
import ColumnChart from '../ColumnChart';

export default {
  title: 'Components/ColumnChart',
  component: ColumnChart
};

export function DefaultChart() {
  return (
    <ColumnChart
      categories={[
        '(Not Set)',
        'Brand_awareness',
        'Brand_launch',
        'Context_marketing',
        'Email_marketing',
        'Product_launch'
      ]}
      series={[
        {
          data: [8550, 585, 81, 966, 632, 240],
          color: CHART_COLOR_1
        }
      ]}
    />
  );
}

export function WithComparison() {
  return (
    <ColumnChart
      categories={['(Not Set)', 'Brand_awareness', 'Brand_launch']}
      comparisonApplied
      series={[
        {
          data: [8550, 585, 81]
        },
        {
          data: [7550, 685, 91]
        }
      ]}
    />
  );
}
