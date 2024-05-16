import React, { useState, useEffect } from 'react';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { PropTextFormat } from 'Utils/dataFormatter';
import FaSelect from '../../FaSelect';
import styles from './index.module.scss';
import { SVG, Text } from '../../factorsComponents';
import { defaultPropertyList, alertsGroupPropertyList } from './utils';

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
  hideText = false, // added to hide the text from UI (Used in event based alerts)
  noMargin = false,
  groups,
  userPropertiesV2,
  groupAnalysis = false
}) {
  const [filterOptions, setFilterOptions] = useState([]);
  const [propSelVis, setSelVis] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

  useEffect(() => {
    let filterOptsObj = {};

    if (!groupAnalysis || groupAnalysis === 'users') {
      filterOptsObj = defaultPropertyList(
        eventPropertiesV2,
        eventUserPropertiesV2,
        groupProperties,
        eventGroup,
        groups?.all_groups,
        event
      );
    } else {
      filterOptsObj = alertsGroupPropertyList(
        eventPropertiesV2,
        userPropertiesV2,
        groupProperties,
        eventGroup,
        groups?.all_groups,
        event
      );
    }

    setFilterOptions(Object.values(filterOptsObj));
  }, [
    userPropertiesV2,
    eventPropertiesV2,
    eventUserPropertiesV2,
    groups,
    groupProperties,
    groupAnalysis,
    eventGroup,
    event
  ]);

  const onChange = (option, group, ind) => {
    const newGroupByState = { ...groupByEvent };
    newGroupByState.prop_category = option?.extraProps?.propertyType;
    newGroupByState.eventName = event?.label;
    newGroupByState.property = option?.value;
    newGroupByState.prop_type = option?.extraProps?.valueType;
    newGroupByState.eventIndex = eventIndex;

    if (groupAnalysis) {
      newGroupByState.groupName = option?.extraProps?.groupName;
    }
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

  const getIcon = (state) => {
    const { property, prop_category } = state || {};
    if (!property) return null;
    const iconName = prop_category === 'group' ? 'user' : prop_category;
    return <SVG name={iconName} size={16} color='purple' />;
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
            icon={getIcon(groupByEvent)}
            type='link'
            className='fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin'
          >
            {title}
          </Button>
        </Tooltip>
        <div className={styles.group_block__event_selector}>
          <GroupSelect
            options={filterOptions}
            searchPlaceHolder='Select Property'
            optionClickCallback={(option, group) =>
              onChange(option, group, index)
            }
            onClickOutside={() => setGroupByDDVisible(false)}
            allowSearch
            allowSearchTextSelection={false}
            extraClass={styles.group_block__event_selector__select}
          />
        </div>
      </div>
    ) : (
      <>
        <Tooltip title={title}>
          <Button
            icon={getIcon(groupByEvent)}
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
    <div className={`${styles.group_block__event_selector}`}>
      <GroupSelect
        options={filterOptions}
        searchPlaceHolder='Select Property'
        optionClickCallback={onChange}
        onClickOutside={() => closeDropDown()}
        allowSearch
        allowSearchTextSelection={false}
        extraClass={`${styles.group_block__event_selector__select}`}
      />
    </div>
  );
  return (
    <div
      className={`flex items-center relative ${styles.draghandleparent} ${
        noMargin ? '' : 'ml-6'
      }`}
    >
      <div className={`p-1 flex ${styles.draghandle}`}>
        <div style={{ margin: 'auto 0', cursor: 'pointer' }}>
          <SVG name='drag' />
        </div>
      </div>
      {!hideText &&
        (grpIndex >= 1 ? (
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
        ))}
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
  userPropertiesV2: state.coreQuery.userPropertiesV2,
  userPropNames: state.coreQuery.userPropNames,
  eventPropNames: state.coreQuery.eventPropNames,
  groupPropNames: state.coreQuery.groupPropNames,
  eventNames: state.coreQuery.eventNames,
  groups: state.coreQuery.groups
});

export default connect(mapStateToProps)(EventGroupBlock);
