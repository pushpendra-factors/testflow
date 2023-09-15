import { visualizationColors } from '../utils/dataFormatter';

export const SPARK_LINE_CHART_TITLE_CHAR_COUNT = 40;

export const METRIC_CHART_TITLE_CHAR_COUNT = 40;

export const COLOR_CLASSNAMES = visualizationColors.reduce(
  (prev, curr, currIndex) => ({
    ...prev,
    [curr]: `charts-color-class-${currIndex}`
  }),
  {}
);

export const cardSizeToMetricCount = {
  0: 2,
  1: 3,
  2: 1
};
