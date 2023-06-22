import { formatDuration } from 'Utils/dataFormatter';
import { KEY_LABELS, PAGE_COUNT_KEY, SESSION_SPENT_TIME } from '../../../const';
export const formatCellData = (
  title: string | number,
  columnName: keyof typeof KEY_LABELS
) => {
  let formattedValue = title;
  if (columnName === SESSION_SPENT_TIME) {
    if (isNaN(Number(title))) {
      formattedValue = 'NA';
    } else if (Number(title) < 1800) {
      formattedValue = formatDuration(title);
    } else {
      formattedValue = '> 30mins';
    }
  } else if (columnName === PAGE_COUNT_KEY) {
    formattedValue = `${title} ${Number(title) > 1 ? 'Pages' : 'Page'}`;
  }
  return formattedValue;
};
