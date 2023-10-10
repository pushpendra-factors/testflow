import MomentTz from "Components/MomentTz";
import { QUERY_TYPE_KPI } from "Utils/constants";

const getPredefinedqueryGroupForWidget1 = (widget, filter, groupBy, period) => {
  const queryArr = [];
  widget?.me?.forEach((item, index) => {
    queryArr.push({
      me: [{na: item?.na, d_na: item?.d_na}],
      inter_e_type: item?.inter_e_type,
      // fil: getEventsWithPropertiesKPI(item.filters, item?.category),
      // gBy: getGroupByWithPropertiesKPI(GrpByItem, index, item?.category),
      fil: filter,
      g_by: groupBy,
      gbt: period.frequency,
      fr: period.from,
      to: period.to,
      tz: "UTC",
      "inter_id": widget.inter_id
    });
    queryArr.push({
      me: [{na: item?.na, d_na: item?.d_na}],
      inter_e_type: item?.inter_e_type,
      // fil: getEventsWithPropertiesKPI(item.filters, item?.category),
      // gBy: getGroupByWithPropertiesKPI(GrpByItem, index, item?.category),
      fil: filter,
      g_by: groupBy,
      gbt: "",
      fr: period.from,
      to: period.to,
      tz: "UTC",
      "inter_id": widget.inter_id
    });
  });
  return queryArr;
};

const getPredefinedqueryGroup = (widget, filter, groupBy, period) => {
    const queryArr = [];
    // Object.keys(widget).forEach((item, index) => {
      queryArr.push({
        me: widget.me.map(obj => {
            const { inter_e_type, ...rest } = obj; // Use destructuring to exclude "inter_e_type"
            return rest; // Return the object without "inter_e_type"
          }),
          inter_e_type: widget.me[0]?.inter_e_type,
        // fil: getEventsWithPropertiesKPI(item.filters, item?.category),
        // gBy: getGroupByWithPropertiesKPI(GrpByItem, index, item?.category),
        fil: filter,
        g_by: groupBy,
        gbt: period.frequency,
        fr: period.from,
        to: period.to,
        tz: "UTC",
        "inter_id": widget.inter_id
      });
      queryArr.push({
        me: widget.me.map(obj => {
            const { inter_e_type, ...rest } = obj; // Use destructuring to exclude "inter_e_type"
            return rest; // Return the object without "inter_e_type"
          }),
          inter_e_type: widget.me[0]?.inter_e_type,
        // fil: getEventsWithPropertiesKPI(item.filters, item?.category),
        // gBy: getGroupByWithPropertiesKPI(GrpByItem, index, item?.category),
        fil: filter,
        g_by: groupBy,
        gbt: "",
        fr: period.from,
        to: period.to,
        tz: "UTC",
        "inter_id": widget.inter_id
      });
    // });
    return queryArr;
  };
  
  export const getPredefinedQuery = (
    widget,
    dateRange,
    filter = [],
    groupBy = {}
  ) => {
    const query = {};
    query.cl = 'predef_web_analytics_query';
    const period = {};
    if (dateRange?.from && dateRange?.to) {
      period.from = MomentTz(dateRange.from).startOf('day').utc().unix();
      period.to = MomentTz(dateRange.to).endOf('day').utc().unix();
      period.frequency = dateRange.frequency;
    } else {
      period.from = MomentTz().startOf('week').utc().unix();
      period.to =
        MomentTz().format('dddd') !== 'Sunday'
          ? MomentTz().subtract(1, 'day').endOf('day').utc().unix()
          : MomentTz().utc().unix();
      period.frequency = dateRange.frequency;
    }
  
    if(widget?.inter_id === 1) {
      query.q_g = getPredefinedqueryGroupForWidget1(widget, filter, groupBy, period);
    } else {
      query.q_g = getPredefinedqueryGroup(widget, filter, groupBy, period);
    }
  
    return query;
  };