import { visualizationColors } from '../utils/dataFormatter';

export const SPARK_LINE_CHART_TITLE_CHAR_COUNT = 40;

export const COLOR_CLASSNAMES = visualizationColors.reduce(
  (prev, curr, currIndex) => {
    return {
      ...prev,
      [curr]: `charts-color-class-${currIndex}`
    };
  },
  {}
);
