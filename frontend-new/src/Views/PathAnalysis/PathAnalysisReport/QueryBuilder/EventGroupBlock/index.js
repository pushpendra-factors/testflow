import React, { useState, useEffect } from 'react';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { SVG, Text } from 'factorsComponents';
import FaSelect from 'Components/FaSelect';
import {
  convertAndAddPropertiesToGroupSelectOptions,
  PropTextFormat
} from 'Utils/dataFormatter';
import getGroupIcon from 'Utils/getGroupIcon';
import { CustomGroupDisplayNames } from 'Components/GlobalFilter/FilterWrapper/utils';
import GroupSelect2 from '../GroupSelect2';
import styles from './index.module.scss';

function EventGroupBlock({
  eventGroup,
  index,
  eventIndex,
  grpIndex,
  groupByEvent,
  event,
  eventUserPropertiesV2,
  userPropNames,
  eventPropertiesV2,
  eventPropNames,
  groupProperties,
  groupPropNames,
  setGroupState,
  delGroupState,
  closeDropDown,
  groups
}) {
  const [filterOptions, setFilterOptions] = useState();

  const [propSelVis, setSelVis] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

  useEffect(() => {
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
        : groups?.all_groups?.[eventGroup]
          ? groups?.all_groups?.[eventGroup]
          : PropTextFormat(eventGroup);
      const groupValues =
        groupProperties[eventGroup]?.map((op) => ({
          value: op?.[1],
          label: op?.[0],
          extraProps: {
            valueType: op?.[2],
            propertyType: 'group'
          }
        })) || [];
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
    } else if (eventUserPropertiesV2) {
      convertAndAddPropertiesToGroupSelectOptions(
        eventUserPropertiesV2,
        filterOptsObj,
        'user'
      );
    }
    setFilterOptions(Object.values(filterOptsObj));
  }, [eventUserPropertiesV2, eventPropertiesV2, groupProperties]);

  const onChange = (group, val, ind) => {
    const newGroupByState = { ...groupByEvent };
    newGroupByState.prop_category = option?.extraProps?.propertyType;
    newGroupByState.eventName = event.label;
    newGroupByState.property = val[1];
    newGroupByState.prop_type = val[2];
    newGroupByState.eventIndex = eventIndex;

    if (newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = 'raw_values';
    }
    if (newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = 'day';
    }

    setGroupState(newGroupByState, ind);
    setGroupByDDVisible(false);
    closeDropDown();
  };

  const onGrpPropChange = (val) => {
    const newGroupByState = { ...groupByEvent };
    if (newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = val;
    }
    if (newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = val;
    }
    setGroupState(newGroupByState, index);
    setSelVis(false);
  };

  const renderGroupPropertyOptions = (opt) => {
    if (!opt || opt.prop_type === 'categorical') return null;

    const propOpts = {
      numerical: [
        ['original values', null, 'raw_values'],
        ['bucketed values', null, 'with_buckets']
      ],
      datetime: [
        ['hour', null, 'hour'],
        ['date', null, 'day'],
        ['week', null, 'week'],
        ['month', null, 'month']
      ]
    };

    const getProp = (op) => {
      if (op.prop_type === 'numerical') {
        const propSel = propOpts.numerical.filter((v) => v[2] === op.gbty);
        return propSel[0] ? propSel[0][0] : 'Select options';
      }
      if (op.prop_type === 'datetime') {
        const propSel = propOpts.datetime.filter((v) => v[2] === op.grn);
        return propSel[0] ? propSel[0][0] : 'Select options';
      }
      return null;
    };

    const setProp = (op) => {
      onGrpPropChange(op[2]);
      setSelVis(false);
    };

    return (
      <div className='flex items-center m-0 mx-2'>
        show as
        <div
          className={`flex relative m-0 mx-2 ${styles.grpProps__select__opt}`}
          onClick={() => setSelVis(!propSelVis)}
        >
          {getProp(opt)}
          {propSelVis && (
            <FaSelect
              options={propOpts[opt.prop_type]}
              optionClick={setProp}
              onClickOutside={() => setSelVis(false)}
            />
          )}
        </div>
      </div>
    );
  };

  const renderGroupDisplayName = (state) => {
    if (!state.property) {
      return 'Select Property';
    }
    switch (state.prop_category) {
      case 'event':
        return eventPropNames[state.property] || PropTextFormat(state.property);
      case 'user' || 'user_g':
        return userPropNames[state.property] || PropTextFormat(state.property);
      case 'group':
        return (
          groupPropNames[eventGroup][state.property] ||
          PropTextFormat(state.property)
        );
      default:
        return PropTextFormat(state.property);
    }
  };

  const renderGroupContent = () => {
    const title = renderGroupDisplayName(groupByEvent);
    return isGroupByDDVisible ? (
      <div className='relative'>
        <Tooltip title={title}>
          <Button
            icon={
              <SVG name={groupByEvent.prop_category} size={16} color='purple' />
            }
            type='link'
            className='fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin'
          >
            {title}
          </Button>
        </Tooltip>
        <div className={styles.group_block__event_selector}>
          <GroupSelect2
            groupedProperties={filterOptions}
            placeholder='Select Property'
            optionClick={(group, val) => onChange(group, val, index)}
            onClickOutside={() => setGroupByDDVisible(false)}
          />
        </div>
      </div>
    ) : (
      <>
        <Tooltip title={title}>
          <Button
            icon={
              <SVG name={groupByEvent.prop_category} size={16} color='purple' />
            }
            type='link'
            className='fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin'
            onClick={() => setGroupByDDVisible(true)}
          >
            {title}
          </Button>
        </Tooltip>
        {renderGroupPropertyOptions(groupByEvent)}
      </>
    );
  };

  const renderGroupBySelect = () => (
    <div className={styles.group_block__event_selector}>
      <GroupSelect2
        groupedProperties={filterOptions}
        placeholder='Select Property'
        optionClick={(group, val) => onChange(group, val)}
        onClickOutside={() => closeDropDown()}
      />
    </div>
  );

  return (
    <div className='flex items-center relative ml-10'>
      {grpIndex >= 1 ? (
        <Text level={8} type='title' extraClass='m-0 mr-16' weight='thin'>
          and
        </Text>
      ) : (
        <Text
          level={8}
          type='title'
          extraClass='m-0 breakdown-margin'
          weight='thin'
        >
          Breakdown
        </Text>
      )}
      {groupByEvent && groupByEvent.property ? (
        renderGroupContent()
      ) : (
        <>{renderGroupBySelect()}</>
      )}
      <Button
        type='text'
        onClick={() => delGroupState(groupByEvent)}
        size='small'
        className='fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button'
      >
        <SVG name='remove' />
      </Button>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  groupProperties: state.coreQuery.groupProperties,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  userPropNames: state.coreQuery.userPropNames,
  eventPropNames: state.coreQuery.eventPropNames,
  groupPropNames: state.coreQuery.groupPropNames,
  groups: state.coreQuery.groups
});

export default connect(mapStateToProps)(EventGroupBlock);
