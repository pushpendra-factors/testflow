import React from 'react';
import { visualizationColors } from '../../../utils/dataFormatter';
import SparkChart from '../Chart';

export default {
  title: 'Components/SparkLineChart',
  component: SparkChart
};

const SPARK_CHART_SAMPLE_KPI_DATA = [
  {
    date: new Date('Sat Jul 16 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 866
  },
  {
    date: new Date('Sun Jul 17 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 766
  },
  {
    date: new Date('Mon Jul 18 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 803
  },
  {
    date: new Date('Tue Jul 19 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 826
  },
  {
    date: new Date('Wed Jul 20 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 793
  },
  {
    date: new Date('Thu Jul 21 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 829
  },
  {
    date: new Date('Fri Jul 22 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 879
  }
];

const SPARK_CHART_SAMPLE_KPI_HOURLY_DATA = [
  {
    date: new Date('Thu Jul 28 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 26
  },
  {
    date: new Date('Thu Jul 28 2022 20:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 42
  },

  {
    date: new Date('Thu Jul 28 2022 21:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 36
  },

  {
    date: new Date('Thu Jul 28 2022 22:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 26
  },

  {
    date: new Date('Thu Jul 28 2022 23:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 45
  },

  {
    date: new Date('Fri Jul 29 2022 00:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 29
  },

  {
    date: new Date('Fri Jul 29 2022 01:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 27
  },

  {
    date: new Date('Fri Jul 29 2022 02:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 40
  },

  {
    date: new Date('Fri Jul 29 2022 03:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 23
  },

  {
    date: new Date('Fri Jul 29 2022 04:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 22
  },

  {
    date: new Date('Fri Jul 29 2022 05:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 32
  },

  {
    date: new Date('Fri Jul 29 2022 06:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 25
  },

  {
    date: new Date('Fri Jul 29 2022 07:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 36
  },

  {
    date: new Date('Fri Jul 29 2022 08:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 31
  },

  {
    date: new Date('Fri Jul 29 2022 09:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 42
  },

  {
    date: new Date('Fri Jul 29 2022 10:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 26
  },

  {
    date: new Date('Fri Jul 29 2022 11:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 25
  },

  {
    date: new Date('Fri Jul 29 2022 12:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 23
  },

  {
    date: new Date('Fri Jul 29 2022 13:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 29
  },

  {
    date: new Date('Fri Jul 29 2022 14:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 34
  },

  {
    date: new Date('Fri Jul 29 2022 15:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 30
  },

  {
    date: new Date('Fri Jul 29 2022 16:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 15
  },

  {
    date: new Date('Fri Jul 29 2022 17:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 37
  },

  {
    date: new Date('Fri Jul 29 2022 18:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 34
  }
];

const SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA = [
  {
    date: new Date('Sat Jul 16 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Sat Jul 16 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    'Engaged Users who tested': 866,
    compareValue: 800
  },
  {
    date: new Date('Sun Jul 17 2022 19:30:00 GMT+0530 (India Standard Time)'),
    'Engaged Users who tested': 766,
    compareDate: new Date(
      'Sun Jul 17 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    compareValue: 800
  },
  {
    date: new Date('Mon Jul 18 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Mon Jul 18 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    'Engaged Users who tested': 803,
    compareValue: 800
  },
  {
    date: new Date('Tue Jul 19 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Tue Jul 19 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    'Engaged Users who tested': 826,
    compareValue: 800
  },
  {
    date: new Date('Wed Jul 20 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Wed Jul 20 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    'Engaged Users who tested': 793,
    compareValue: 800
  },
  {
    date: new Date('Thu Jul 21 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Thu Jul 21 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    'Engaged Users who tested': 829,
    compareValue: 800
  },
  {
    date: new Date('Fri Jul 22 2022 19:30:00 GMT+0530 (India Standard Time)'),
    compareDate: new Date(
      'Fri Jul 22 2022 19:30:00 GMT+0530 (India Standard Time)'
    ),
    'Engaged Users who tested': 879,
    compareValue: 1000
  }
];

export const DefaultChart = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is Engaged Users who tested
    <SparkChart chartData={SPARK_CHART_SAMPLE_KPI_DATA} event="Engaged Users who tested" />
  );
};

export const ChartWithDifferentColor = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is Engaged Users who tested
    <SparkChart
      chartColor={visualizationColors[5]}
      chartData={SPARK_CHART_SAMPLE_KPI_DATA}
      event="Engaged Users who tested"
    />
  );
};

export const ChartWithHourlyFrequeny = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is Engaged Users who tested
    <SparkChart
      frequency="hour"
      chartColor={visualizationColors[8]}
      chartData={SPARK_CHART_SAMPLE_KPI_HOURLY_DATA}
      event="Engaged Users who tested"
    />
  );
};

export const ChartWithComparisonData = () => {
  return (
    // event prop should match the key present in the chartData array. In this case, event is Engaged Users who tested
    <SparkChart
      chartColor={visualizationColors[9]}
      chartData={SPARK_CHART_SAMPLE_KPI_DATA_WITH_COMPARISON_DATA}
      event="Engaged Users who tested"
      comparisonEnabled={true}
    />
  );
};
