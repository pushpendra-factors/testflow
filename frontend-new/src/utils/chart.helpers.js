import { QUERY_TYPE_KPI } from "./constants";

export const isPivotSupported = ({ queryType }) => {
  return queryType === QUERY_TYPE_KPI;
};
