import { DISPLAY_PROP } from 'Utils/constants';
import { cloneDeep } from 'lodash';
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

export const moveSelectedOptionsToTop = (options: OptionType[]) => {
  const _options = cloneDeep(options) as OptionType[];
  const selectedOptions = _options.filter((option) => option?.isSelected);
  const unselectedOptions = _options.filter((option) => !option?.isSelected);
  return [...selectedOptions, ...unselectedOptions];
};

export const filterSearchFunction = (op: OptionType, searchTerm: string) => {
  if (!searchTerm) return true;
  let searchTermLowerCase = searchTerm.toLowerCase();
  // Regex to detect https/http is there or not as a protocol
  let testURLRegex =
    /^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_\+.~#?&\/=]*)$/;
  if (testURLRegex.test(searchTermLowerCase)) {
    searchTermLowerCase = searchTermLowerCase.split('://')[1];
  }
  searchTermLowerCase = searchTermLowerCase.replace(/\/$/, '');
  return (
    op.label.toLowerCase().includes(searchTermLowerCase) ||
    (op.label === '$none' &&
      DISPLAY_PROP[op.label].toLowerCase().includes(searchTermLowerCase))
  );
};
