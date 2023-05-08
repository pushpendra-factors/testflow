import { OptionType } from './types';
export const selectedOptionsMapper = (
  options: OptionType[],
  selectedOptions: string[]
) => {
  const optionsValues = options.map((op) => op.value);
  return selectedOptions.map((opValue: string) => {
    const index = optionsValues.indexOf(opValue);
    if (index > -1) {
      return options[index];
    }
    //Custom Selected Option By User.
    return { value: opValue, label: opValue };
  });
};
