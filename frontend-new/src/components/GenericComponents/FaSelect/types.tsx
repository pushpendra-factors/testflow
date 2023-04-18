import { ReactNode } from 'react';

export type OptionType = {
  value: string;
  label: string;
  labelNode?: ReactNode;
};
export type Variant = 'Single' | 'Multi';
export type handleOptionFunctionType = (
  options: OptionType | OptionType[]
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
