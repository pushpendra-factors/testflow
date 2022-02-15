import { QUERY_TYPE_EVENT, QUERY_TYPE_FUNNEL, QUERY_TYPE_KPI, QUERY_TYPE_CAMPAIGN, QUERY_TYPE_PROFILE, QUERY_TYPE_ATTRIBUTION } from "../../utils/constants";

export const IconAndTextSwitchQueryType = (queryType) => {
  switch (queryType) {
    case QUERY_TYPE_EVENT:
      return {
        text: 'Analyse Events',
        icon: 'events_cq',
      };
    case QUERY_TYPE_FUNNEL:
      return {
        text: 'Find event funnel for',
        icon: 'funnels_cq',
      };
    case QUERY_TYPE_CAMPAIGN:
      return {
        text: 'Campaign Analytics',
        icon: 'campaigns_cq',
      };
    case QUERY_TYPE_ATTRIBUTION:
      return {
        text: 'Attributions',
        icon: 'attributions_cq',
      };
    case QUERY_TYPE_KPI:
      return {
        text: 'KPI',
        icon: 'attributions_cq',
      };
    case QUERY_TYPE_PROFILE:
      return {
        text: 'Profile Analysis',
        icon: 'profiles_cq',
      };
    default:
      return {
        text: 'Templates',
        icon: 'templates_cq',
      };
  }
};