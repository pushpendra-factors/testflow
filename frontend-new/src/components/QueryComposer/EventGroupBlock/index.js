import React, { useState, useEffect } from 'react';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { SVG, Text } from '../../factorsComponents';
import styles from './index.module.scss';
import GroupSelect2 from '../GroupSelect2';
import FaSelect from '../../FaSelect';

function EventGroupBlock({
  eventGroup,
  index,
  eventIndex,
  grpIndex,
  groupByEvent,
  event,
  userProperties,
  userPropNames,
  eventProperties,
  eventPropNames,
  groupProperties,
  groupPropNames,
  setGroupState,
  delGroupState,
  closeDropDown,
  hideText = false, // added to hide the text from UI (Used in event based alerts)
  posTop = false // used to open the drop down at the top( Event based alerts)
}) {
  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'Event Properties',
      icon: 'event',
      values: []
    },
    {
      label: 'User Properties',
      icon: 'user',
      values: []
    },
    {
      label: 'Group Properties',
      icon: 'group',
      values: []
    }
  ]);

  const [propSelVis, setSelVis] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[0].values = eventProperties[event.label];
    if (eventGroup?.length) {
      filterOpts[2].values = groupProperties[eventGroup[1]];
      filterOpts[1].values = [];
    } else {
      filterOpts[1].values = userProperties;
      filterOpts[2].values = [];
    }
    setFilterOptions(filterOpts);
  }, [userProperties, eventProperties, groupProperties]);

  const onChange = (group, val, ind) => {
    const newGroupByState = { ...groupByEvent };
    if (group === 'User Properties') {
      newGroupByState.prop_category = 'user';
    } else if (group === 'Group Properties') {
      newGroupByState.prop_category = 'group';
    } else {
      newGroupByState.prop_category = 'event';
    }
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

  const renderGroupContent = () => {
    let propName = '';
    if (groupByEvent.property && groupByEvent.prop_category === 'user') {
      propName = userPropNames[groupByEvent.property]
        ? userPropNames[groupByEvent.property]
        : groupByEvent.property;
    }

    if (groupByEvent.property && groupByEvent.prop_category === 'event') {
      propName = eventPropNames[groupByEvent.property]
        ? eventPropNames[groupByEvent.property]
        : groupByEvent.property;
    }

    if (groupByEvent.property && groupByEvent.prop_category === 'group') {
      propName = groupPropNames[groupByEvent.property]
        ? groupPropNames[groupByEvent.property]
        : groupByEvent.property;
    }

    return isGroupByDDVisible ? (
      <div className='relative'>
        <Tooltip title={propName}>
          <Button
            icon={
              <SVG name={groupByEvent.prop_category} size={16} color='purple' />
            }
            type='link'
            className='fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin'
          >
            {propName}
          </Button>
        </Tooltip>
        <div
          className={`${styles.group_block__event_selector} ${
            posTop && styles.group_block__select_ct
          }`}
        >
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
        <Tooltip title={propName}>
          <Button
            icon={
              <SVG name={groupByEvent.prop_category} size={16} color='purple' />
            }
            type='link'
            className='fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin'
            onClick={() => setGroupByDDVisible(true)}
          >
            {propName}
          </Button>
        </Tooltip>
        {renderGroupPropertyOptions(groupByEvent)}
      </>
    );
  };

  const renderGroupBySelect = () => (
    <div
      className={`${styles.group_block__event_selector} ${
        posTop && styles.group_block__select_ct
      }`}
    >
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
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  userPropNames: state.coreQuery.userPropNames,
  eventPropNames: state.coreQuery.eventPropNames,
  groupPropNames: state.coreQuery.groupPropNames
});

export default connect(mapStateToProps)(EventGroupBlock);
