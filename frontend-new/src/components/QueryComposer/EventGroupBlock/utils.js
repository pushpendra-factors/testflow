import {
  PropTextFormat,
  convertAndAddPropertiesToGroupSelectOptions,
  processProperties
} from 'Utils/dataFormatter';
import { CustomGroupDisplayNames } from 'Components/GlobalFilter/FilterWrapper/utils';
import getGroupIcon from 'Utils/getGroupIcon';
import _ from 'lodash';

export const defaultPropertyList = (
  eventPropertiesV2,
  eventUserPropertiesV2,
  groupProperties,
  eventGroup,
  groups,
  event
) => {
  const filterOptsObj = {};
  const eventGroups = eventPropertiesV2[event?.label] || {};
  convertAndAddPropertiesToGroupSelectOptions(
    eventGroups,
    filterOptsObj,
    'event'
  );
  if (eventGroup) {
    const groupLabel = CustomGroupDisplayNames[eventGroup]
      ? CustomGroupDisplayNames[eventGroup]
      : groups[eventGroup]
        ? groups[eventGroup]
        : PropTextFormat(eventGroup);
    const groupValues =
      processProperties(groupProperties[eventGroup], 'user', eventGroup) || [];
    const groupPropIconName = getGroupIcon(groupLabel);
    if (!filterOptsObj[groupLabel]) {
      filterOptsObj[groupLabel] = {
        iconName: groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
        label: groupLabel,
        values: groupValues
      };
    } else {
      filterOptsObj[groupLabel].values.push(...groupValues);
    }
  } else if (eventUserPropertiesV2) {
    convertAndAddPropertiesToGroupSelectOptions(
      eventUserPropertiesV2,
      filterOptsObj,
      'user'
    );
  }

  return filterOptsObj;
};

function removeDuplicateAndEmptyKeys(obj) {
  const uniqueKeys = {};
  // blacklisted groups
  const removeGroupList = ['Company identification'];
  Object.entries(obj).forEach(([key, value]) => {
    // remove duplicate keys
    if (!(key in uniqueKeys)) {
      // remove blacklisted keys
      if (!key.includes(removeGroupList)) {
        // remove empty keys
        if (!_.isEmpty(value?.values)) {
          uniqueKeys[key] = value;
        }
      }
    }
  });

  return uniqueKeys;
}

export const alertsGroupPropertyList = (
  eventPropertiesV2,
  userPropertiesV2,
  groupProperties,
  eventGroup = '',
  groups,
  event
) => {
  const filterOptsObj = {};

  const eventGroups = eventPropertiesV2[event?.label] || {};
  convertAndAddPropertiesToGroupSelectOptions(
    eventGroups,
    filterOptsObj,
    'event'
  );

  if (groupProperties) {
    Object.entries(groupProperties || {}).forEach(([group, properties]) => {
      if (Object.keys(groups).includes(group)) {
        const groupLabel =
          CustomGroupDisplayNames[group] ||
          (groups[group] ? groups[group] : PropTextFormat(group));

        const groupValues = processProperties(properties, 'user', group);
        const groupPropIconName = getGroupIcon(groupLabel);

        filterOptsObj[groupLabel] = {
          iconName:
            groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
          label: groupLabel,
          values: groupValues
        };
      }
    });
  }
  if (!eventGroup) {
    if (userPropertiesV2) {
      convertAndAddPropertiesToGroupSelectOptions(
        userPropertiesV2,
        filterOptsObj,
        'user'
      );
    }

    const groupLabel = CustomGroupDisplayNames[eventGroup]
      ? CustomGroupDisplayNames[eventGroup]
      : groups[eventGroup]
        ? groups[eventGroup]
        : PropTextFormat(eventGroup);
    const groupValues =
      processProperties(groupProperties[eventGroup], 'user', eventGroup) || [];
    const groupPropIconName = getGroupIcon(groupLabel);
    if (!filterOptsObj[groupLabel]) {
      filterOptsObj[groupLabel] = {
        iconName: groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
        label: groupLabel,
        values: groupValues
      };
    } else {
      filterOptsObj[groupLabel].values.push(...groupValues);
    }
  }

  // remove duplicate, blacklisted and empty keys/group
  const finalOptsObj = removeDuplicateAndEmptyKeys(filterOptsObj);

  return finalOptsObj;
};
