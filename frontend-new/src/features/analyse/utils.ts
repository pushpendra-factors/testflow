import MomentTz from "Components/MomentTz";
import { QueryOptions } from "./types";

export const getQueryOptionsFromEquivalentQuery = (currOpts: QueryOptions, equivalentQuery: any) => ({
    ...currOpts,
    date_range: {
      from: MomentTz(equivalentQuery.dateRange.from),
      to: MomentTz(equivalentQuery.dateRange.to),
      frequency: equivalentQuery.dateRange.frequency
    },
    session_analytics_seq: equivalentQuery.session_analytics_seq,
    groupBy: {
      global: [...equivalentQuery.breakdown.global],
      event: [...equivalentQuery.breakdown.event]
    },
    globalFilters: equivalentQuery.globalFilters
});