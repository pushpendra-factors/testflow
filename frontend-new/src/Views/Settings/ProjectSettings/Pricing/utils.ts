export const showUpgradeNudge = (
  utilisedAmount: number,
  totalAmount: number,
  flag: boolean = false
) => {
  if (!flag) return false;
  let percentage = Number(((utilisedAmount / totalAmount) * 100).toFixed(2));
  return percentage >= 75;
};

export const PRICING_PAGE_TABS = {
  BILLING: 'billing',
  ENRICHMENT_RULES: 'enrichment_rules'
};
