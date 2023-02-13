import React from 'react';
import {
  CHART_COLOR_10,
  CHART_COLOR_3,
  CHART_COLOR_5,
  CHART_COLOR_7,
  CHART_COLOR_8,
  CHART_COLOR_9
} from '../../../constants/color.constants';
import ColumnChart from '../ColumnChart';
import { METRIC_TYPES } from '../../../utils/constants';

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
          data: [8550, 5850, 801, 9660, 6320, 2400]
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

export function WithMultipleColors() {
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
      multiColored
      series={[
        {
          data: [8550, 5850, 801, 9660, 6320, 2400]
        }
      ]}
    />
  );
}

export function WithMultipleColorsAndComparison() {
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
      // if colors wont be passed, 10 colors will be generated and used
      colors={[
        CHART_COLOR_10,
        CHART_COLOR_7,
        CHART_COLOR_3,
        CHART_COLOR_8,
        CHART_COLOR_5,
        CHART_COLOR_9
      ]}
      multiColored
      comparisonApplied
      valueMetricType={METRIC_TYPES.percentType}
      series={[
        {
          data: [85, 58, 8, 96, 63, 24]
        },
        {
          data: [75, 48, 9, 86, 73, 34]
        }
      ]}
    />
  );
}
