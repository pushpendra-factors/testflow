type eventOptionsType = {
  icon?: string;
  label: string;
  values: string[][];
};
export const findGroupNameUsingOptionValue = (
  eventOptions: eventOptionsType[],
  optionValue: string
) => {
  const eventOptionGroup = eventOptions.find((groupOption) => {
    let valueFound = false;
    for (const valueOption of groupOption.values) {
      if (valueOption[1] === optionValue) {
        valueFound = true;
        break;
      }
    }
    return valueFound;
  });
  return eventOptionGroup?.label;
};
