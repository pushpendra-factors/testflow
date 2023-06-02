import { OptionType } from '../FaSelect/types';

// extraProps are those props which are sent from parent and never used in componenet and returned back to parent.
export type GroupSelectOptionType = {
  iconName?: string;
  value?: string;
  label: string;
  extraProps?: any;
  values: OptionType[];
};

export type GroupSelectOptionClickCallbackType = (
  value: OptionType,
  group: GroupSelectOptionType
) => void;
