export const showUpgradeNudge = (
  utilisedAmount: number,
  totalAmount: number,
  currentProjectSettings: any
) => {
  if (
    (currentProjectSettings?.int_client_six_signal_key &&
      currentProjectSettings?.client6_signal_key) ||
    (currentProjectSettings?.int_clear_bit &&
      currentProjectSettings?.clearbit_key)
  )
    return false;
  let percentage = Number(((utilisedAmount / totalAmount) * 100).toFixed(2));
  return percentage >= 75;
};

export const PRICING_PAGE_TABS = {
  BILLING: 'billing',
  ENRICHMENT_RULES: 'enrichment_rules'
};
