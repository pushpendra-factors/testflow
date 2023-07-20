export const getIcon = (icon: string) => {
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
  if (checkIcon?.includes('linkedin')) {
    return 'linkedin_ads';
  }
  if (checkIcon?.includes('g2')) {
    return 'G2crowd';
  }
  return icon;
};
