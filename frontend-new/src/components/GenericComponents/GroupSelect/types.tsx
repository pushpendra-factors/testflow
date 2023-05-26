import { OptionType } from '../FaSelect/types';

export type GroupSelectOptionType = {
  iconName?: string;
  label: string;
  values: OptionType[];
};

export type GroupSelectOptionClickCallbackType = (
  value: OptionType,
  group: GroupSelectOptionType
) => void;
