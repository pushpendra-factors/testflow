type GroupName =
  | 'website_session'
  | 'form_submission'
  | 'hubspot_contacts'
  | 'hubspot_companies'
  | 'hubspot_deals'
  | 'google_ads_metrics'
  | 'google_organic_metrics'
  | 'linkedin_metrics'
  | 'linkedin_company_engagements'
  | 'all_channels_metrics'
  | 'bingads_metrics'
  | 'marketo_leads'
  | 'page_views'
  | 'Capterra'
  | 'source1'
  | 'Quora'
  | 'Twitter'
  | 'event_based'
  | 'others';

type GroupIcon =
  | 'brand'
  | 'List_Checkv2'
  | 'hubspot_ads'
  | 'google_ads'
  | 'google'
  | 'linkedin_ads'
  | 'bingads_metrics'
  | 'AllChannelsMetrics'
  | 'bing'
  | 'Marketo'
  | 'PageViews'
  | 'Capterra'
  | 'Source1'
  | 'Quora'
  | 'Twitter'
  | 'EventBased'
  | 'mouseevent'
  | '';

const getGroupIcon = (groupName: GroupName): GroupIcon => {
  switch (groupName) {
    case 'website_session':
      return 'brand';
    case 'form_submission':
      return 'List_Checkv2';
    case 'hubspot_contacts':
      return 'hubspot_ads';
    case 'hubspot_companies':
      return 'hubspot_ads';
    case 'hubspot_deals':
      return 'hubspot_ads';
    case 'google_ads_metrics':
      return 'google_ads';
    case 'google_organic_metrics':
      return 'google';
    case 'linkedin_metrics':
      return 'linkedin_ads';
    case 'linkedin_company_engagements':
      return 'linkedin_ads';
    case 'all_channels_metrics':
      return 'AllChannelsMetrics';
    case 'bingads_metrics':
      return 'bing';
    case 'marketo_leads':
      return 'Marketo';
    case 'page_views':
      return 'PageViews';
    case 'Capterra':
      return 'Capterra';
    case 'source1':
      return 'Source1';
    case 'Quora':
      return 'Quora';
    case 'Twitter':
      return 'Twitter';
    case 'event_based':
      return 'EventBased';
    case 'others':
      return 'mouseevent';
    default:
      return '';
  }
};

export default getGroupIcon;
