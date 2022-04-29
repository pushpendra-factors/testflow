export const getKpiLabel = (kpi) => {
  const label = kpi.alias || kpi.label;
  if (kpi.category !== 'channels') {
    return label;
  }
  return _.startCase(kpi.group) + ' ' + label;
};
