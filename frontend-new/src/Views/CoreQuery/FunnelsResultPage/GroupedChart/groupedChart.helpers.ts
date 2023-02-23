import {
  CHART_TYPE_BARCHART,
  CHART_TYPE_HORIZONTAL_BAR_CHART,
  MAX_ALLOWED_VISIBLE_PROPERTIES,
  METRIC_TYPES
} from 'Utils/constants';
import { generateColors } from 'Utils/dataFormatter';
import { EMPTY_ARRAY } from 'Utils/global';

const colors = generateColors(MAX_ALLOWED_VISIBLE_PROPERTIES);

interface FunnelDataPoint {
  name: string;
  value: string;
}

interface ColumnChartPoint {
  name: string;
  data: number[];
}

interface HorizontalBarChartDataPoint {
  y: number;
  color: string;
  metricType?: string;
}

interface HorizontalBarChartPoint {
  name: string;
  data: HorizontalBarChartDataPoint[];
}

export const getValueFromPercentString = (str: string) => {
  return Number(str.split('%')[0]);
};

export const getCompareGroupsByName = ({
  compareGroups
}: {
  compareGroups: FunnelDataPoint[];
}): Record<string, FunnelDataPoint> => {
  if (compareGroups == null) return {};
  return compareGroups.reduce((prev, curr) => {
    return {
      ...prev,
      [curr.name]: curr
    };
  }, {});
};

interface GetSeriesProps {
  visibleProperties: FunnelDataPoint[];
  chartType: string;
  compareGroups: FunnelDataPoint[];
}

export const getColumChartSeries = ({
  visibleProperties,
  compareGroups,
  chartType
}: GetSeriesProps): ColumnChartPoint[] => {
  if (visibleProperties.length === 0 || chartType !== CHART_TYPE_BARCHART) {
    return EMPTY_ARRAY;
  }
  const s: ColumnChartPoint[] = [
    {
      name: 'OG',
      data: visibleProperties.map((v) => getValueFromPercentString(v.value))
    }
  ];
  if (compareGroups != null) {
    const compareGroupsByName = getCompareGroupsByName({ compareGroups });
    const compareSeriesData = visibleProperties.map((vp) => {
      const d = compareGroupsByName[vp.name];
      if (d != null) {
        return getValueFromPercentString(d.value);
      }
      return 0;
    });
    s.push({
      name: 'compare',
      data: compareSeriesData
    });
  }
  return s;
};

export const getHorizontalBarChartSeries = ({
  visibleProperties,
  compareGroups,
  chartType
}: GetSeriesProps): HorizontalBarChartPoint[] => {
  if (
    visibleProperties.length === 0 ||
    chartType !== CHART_TYPE_HORIZONTAL_BAR_CHART
  ) {
    return EMPTY_ARRAY;
  }
  const s: HorizontalBarChartPoint[] = [
    {
      name: 'OG',
      data: visibleProperties.map((v, index) => {
        return {
          y: getValueFromPercentString(v.value),
          color: colors[index],
          metricType: METRIC_TYPES.percentType
        };
      })
    }
  ];
  if (compareGroups != null) {
    const compareGroupsByName = getCompareGroupsByName({ compareGroups });
    const compareSeriesData = visibleProperties.map((vp, index) => {
      const d = compareGroupsByName[vp.name];
      const value = d != null ? getValueFromPercentString(d.value) : 0;
      return {
        y: value,
        color: colors[index],
        metricType: METRIC_TYPES.percentType
      };
    });
    s.push({
      name: 'compare',
      data: compareSeriesData
    });
  }
  return s;
};
