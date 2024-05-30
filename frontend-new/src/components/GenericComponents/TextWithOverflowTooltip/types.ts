export interface TextWithTooltipProps {
  text: string;
  tooltipText?: string;
  extraClass?: string;
  disabled?: boolean;
  maxLines?: number;
  color?: string;
  hasLink?: boolean;
  linkTo?: string;
  linkState?: object;
  onClick: () => void;
  active?: boolean;
  activeClass?: string;
  alwaysShowTooltip?: boolean;
}
