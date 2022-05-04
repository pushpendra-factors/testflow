export const getBreakdownDisplayName = ({
  breakdown,
  userPropNames,
  eventPropNames,
}) => {
  const property = breakdown.pr || breakdown.property;
  const prop_category = breakdown.en || breakdown.prop_category;
  const displayTitle =
    prop_category === 'user'
      ? _.get(userPropNames, property, property)
      : prop_category === 'event'
      ? _.get(eventPropNames, property, property)
      : property;

  if (breakdown.eventIndex) {
    return displayTitle + ' (event)';
  }
  return displayTitle;
};
