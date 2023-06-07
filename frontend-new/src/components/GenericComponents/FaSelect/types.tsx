import { ReactNode } from 'react';

export type OptionType = {
  value: string;
  label: string;
  labelNode?: ReactNode;
  isSelected?: boolean;
  isDisabled?: boolean;
  extraProps?: any;
};
export type Variant = 'Single' | 'Multi';

export type SingleSelectOptionClickCallbackType = (
  selectedOption: OptionType
) => void;

export type ApplyClickCallbackType = (
  updatedOptions: OptionType[],
  selectedOptions: string[]
) => void;

export type PlacementType =
  | 'Top'
  | 'Bottom'
  | 'Left'
  | 'Right'
  | 'TopLeft'
  | 'TopRight'
  | 'BottomLeft'
  | 'BottomRight';
