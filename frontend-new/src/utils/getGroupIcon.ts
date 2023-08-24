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
  | 'Salesforce_ads'
  | 'Others'
  | 'ArrowProgress'
  | 'UserMagnify'
  | 'G2crowd'
  | 'marketo'
  | 'leadSquared'
  | 'profile'
  | 'LaptopMobile'
  | 'PeopleRoof'
  | 'FaceBook'
  | 'BullsEyePointer';

const getGroupIcon = (groupName: any): GroupIcon => {
  const checkIcon = groupName?.toLowerCase().split(' ').join('_');
  switch (checkIcon) {
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
    case 'capterra':
      return 'Capterra';
    case 'source1':
      return 'Source1';
    case 'quora':
      return 'Quora';
    case 'twitter':
      return 'Twitter';
    case 'event_based':
      return 'EventBased';
    case 'others':
      return 'Others';
    case 'salesforce_users':
      return 'Salesforce_ads';
    case 'salesforce_accounts':
      return 'Salesforce_ads';
    case 'salesforce_opportunities':
      return 'Salesforce_ads';
    case 'traffic_source':
      return 'ArrowProgress';
    case 'session_properties':
      return 'mouseevent';
    case 'user_identification':
      return 'UserMagnify';
    case 'company_identification':
      return 'PeopleRoof';
    case 'platform/device':
      return 'LaptopMobile';
  }

  //Mapping Icons With Similar Name.
  if (checkIcon?.includes('salesforce')) {
    return 'Salesforce_ads';
  }
  if (checkIcon?.includes('hubspot')) {
    return 'hubspot_ads';
  }
  if (checkIcon?.includes('marketo')) {
    return 'marketo';
  }
  if (checkIcon?.includes('leadsquared')) {
    return 'leadSquared';
  }
  if (checkIcon?.includes('group')) {
    return 'profile';
  }
  if (checkIcon?.includes('linkedin')) {
    return 'linkedin_ads';
  }
  if (checkIcon?.includes('g2')) {
    return 'G2crowd';
  }
  if (checkIcon?.includes('page')) {
    return 'PageViews';
  }
  if (checkIcon?.includes('facebook')) {
    return 'FaceBook';
  }
  return 'BullsEyePointer';
};

export default getGroupIcon;
