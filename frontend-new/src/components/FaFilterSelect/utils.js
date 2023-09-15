import { OPERATORS } from 'Utils/constants';
import { processProperties } from 'Utils/dataFormatter';
import { DEFAULT_OP_PROPS } from 'Utils/operatorMapping';

export const DEFAULT_OPERATOR_PROPS = {
  categorical: DEFAULT_OP_PROPS['categorical'],
  numerical: DEFAULT_OP_PROPS['numerical'],
  datetime: DEFAULT_OP_PROPS['datetime']
};
export const dateTimeSelect = new Map([
  ['Days', 'days'],
  ['Weeks', 'week'],
  ['Months', 'month'],
  ['Quarters', 'quarter'],
  ['days', 'Days'],
  ['week', 'Weeks'],
  ['month', 'Months'],
  ['quarter', 'Quarters']
]);

const getGroupLabel = (grp) => {
  if (grp === 'event') return 'Event Properties';
  if (grp === 'user') return 'User Properties';
  if (!grp || !grp.length) return 'Properties';
  return grp;
};

export const convertOptionsToGroupSelectFormat = (options) => {
  // To remove Duplicate Groups.(Groups With Same Names)
  const optionsObj = {};

  options?.forEach((groupOpt) => {
    const label = getGroupLabel(groupOpt?.label);
    if (!optionsObj[label]) {
      optionsObj[label] = {
        iconName: groupOpt?.icon,
        label: label,
        values: processProperties(
          groupOpt?.values,
          groupOpt?.propertyType,
          groupOpt?.key
        )
      };
    } else {
      optionsObj[label].values.push(
        ...(processProperties(
          groupOpt.values,
          groupOpt?.propertyType,
          groupOpt?.key
        ) || [])
      );
    }
  });
  return Object.values(optionsObj);
};

export const checkIfValueSelectorCanRender = (operatorState) => {
  return (
    operatorState &&
    operatorState !== OPERATORS['isKnown'] &&
    operatorState !== OPERATORS['isUnknown'] &&
    operatorState !== OPERATORS['inList'] &&
    operatorState !== OPERATORS['notInList']
  );
};
