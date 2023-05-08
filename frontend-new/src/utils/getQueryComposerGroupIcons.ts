type Icon = String;
type IconName =
  | 'salesforce_ads'
  | 'hubspot_ads'
  | 'marketo'
  | 'leadSquared'
  | 'profile'
  | String;
export const getQueryComposerGroupIcon = (icon: Icon): IconName => {
  const checkIcon = icon?.toLowerCase().split(' ').join('_');
  if (checkIcon?.includes('salesforce')) {
    return 'salesforce_ads';
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
  return icon;
};
