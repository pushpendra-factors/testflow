import {
    PropTextFormat,
    convertAndAddPropertiesToGroupSelectOptions,
    processProperties
  } from 'Utils/dataFormatter';
  import { CustomGroupDisplayNames } from 'Components/GlobalFilter/FilterWrapper/utils';
  import getGroupIcon from 'Utils/getGroupIcon';
  import { GroupDisplayNames, IsDomainGroup } from 'Components/Profile/utils';


  export const defaultPropertyList = (eventPropertiesV2, eventUserPropertiesV2, groupProperties, eventGroup, groupOpts) =>{

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
        : groupOpts[eventGroup]
        ? groupOpts[eventGroup]
        : PropTextFormat(eventGroup);
      const groupValues =
        processProperties(groupProperties[eventGroup], 'group', eventGroup) ||
        [];
      const groupPropIconName = getGroupIcon(groupLabel);
      if (!filterOptsObj[groupLabel]) {
        filterOptsObj[groupLabel] = {
          iconName:
            groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
          label: groupLabel,
          values: groupValues
        };
      } else {
        filterOptsObj[groupLabel].values.push(...groupValues);
      }
    } else {
      if (eventUserPropertiesV2) {
        convertAndAddPropertiesToGroupSelectOptions(
          eventUserPropertiesV2,
          filterOptsObj,
          'user'
        );
      }
    }

    return filterOptsObj

}


export const alertsGroupPropertyList = (eventPropertiesV2, userPropertiesV2, groupProperties, eventGroup, groupOpts, groupName="") => {
    const filterOptsObj = {};

      if (userPropertiesV2) {
        convertAndAddPropertiesToGroupSelectOptions(
          userPropertiesV2,
          filterOptsObj,
          'user'
        );
      } 
      if(groupProperties){
          for (const [group, properties] of Object.entries(groupProperties || {})) {
            if (Object.keys(GroupDisplayNames).includes(group)) {
              const groupLabel = CustomGroupDisplayNames[group]
                ? CustomGroupDisplayNames[group]
                : groupOpts[group]
                ? groupOpts[group]
                : PropTextFormat(group);
              const groupValues = processProperties(properties, 'group', group);
              const groupPropIconName = getGroupIcon(groupLabel);
              filterOptsObj[groupLabel] = {
                iconName:
                  groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
                label: groupLabel,
                values: groupValues
              };
            }
          } 
      } 
      if (!eventGroup) {
        const groupLabel = CustomGroupDisplayNames[eventGroup]
          ? CustomGroupDisplayNames[eventGroup]
          : groupOpts[eventGroup]
          ? groupOpts[eventGroup]
          : PropTextFormat(eventGroup);
        const groupValues =
          processProperties(groupProperties[eventGroup], 'group', eventGroup) ||
          [];
        const groupPropIconName = getGroupIcon(groupLabel);
        if (!filterOptsObj[groupLabel]) {
          filterOptsObj[groupLabel] = {
            iconName:
              groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
            label: groupLabel,
            values: groupValues
          };
        } else {
          filterOptsObj[groupLabel].values.push(...groupValues);
        }
      }

    return filterOptsObj
  }