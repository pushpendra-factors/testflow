import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';

import { Button } from 'antd';

import { SVG, Text } from 'factorsComponents';

import { connect } from 'react-redux';
import GroupSelect2 from '../GroupSelect2';
import FaSelect from '../../FaSelect';

const EventGroupBlock = ({
  index,
  eventIndex,
  grpIndex,
  groupByEvent,
  event,
  userProperties,
  userPropNames,
  eventProperties,
  eventPropNames,
  setGroupState,
  delGroupState,
  closeDropDown,
}) => {
  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'Event Properties',
      icon: 'event',
      values: [],
    },
    {
      label: 'User Properties',
      icon: 'user',
      values: [],
    },
  ]);

  const [propSelVis, setSelVis] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[1].values = userProperties;
    setFilterOptions(filterOpts);
  }, [userProperties]);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[0].values = eventProperties[event.label];
    setFilterOptions(filterOpts);
  }, [eventProperties]);

  const onChange = (group, val, ind) => {
    const newGroupByState = Object.assign({}, groupByEvent);
    if (group === 'User Properties') {
      newGroupByState.prop_category = 'user';
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
    const newGroupByState = Object.assign({}, groupByEvent);
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
    if (!opt || opt.prop_type === 'categorical') return;

    const propOpts = {
      numerical: [
        ['original values', null, 'raw_values'],
        ['bucketed values', null, 'with_buckets'],
      ],
      datetime: [
        ['hour', null, 'hour'],
        ['date', null, 'day'],
        ['week', null, 'week'],
        ['month', null, 'month'],
      ],
    };

    const getProp = (opt) => {
      if (opt.prop_type === 'numerical') {
        const propSel = propOpts['numerical'].filter((v) => v[2] === opt.gbty);
        return propSel[0] ? propSel[0][0] : 'Select options';
      }
      if (opt.prop_type === 'datetime') {
        const propSel = propOpts['datetime'].filter((v) => v[2] === opt.grn);
        return propSel[0] ? propSel[0][0] : 'Select options';
      }
    };

    const setProp = (opt) => {
      onGrpPropChange(opt[2]);
      setSelVis(false);
    };

    return (
      <div className={styles.grpProps}>
        show as{' '}
        <div className={styles.grpProps__select}>
          <span
            className={styles.grpProps__select__opt}
            onClick={() => setSelVis(true)}
          >
            {getProp(opt)}
          </span>
          {propSelVis && (
            <FaSelect
              options={propOpts[opt.prop_type]}
              optionClick={setProp}
              onClickOutside={() => setSelVis(false)}
            ></FaSelect>
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

    return isGroupByDDVisible ? (
      <>
        <Button
          icon={
            <SVG name={groupByEvent.prop_category} size={16} color={'purple'} />
          }
          type={'link'}
          className={'ml-2'}
        >
          {propName}
        </Button>
        <div className={styles.group_block__event_selector}>
          <GroupSelect2
            groupedProperties={filterOptions}
            placeholder='Select Property'
            optionClick={(group, val) => onChange(group, val, index)}
            onClickOutside={() => setGroupByDDVisible(false)}
          ></GroupSelect2>
        </div>
      </>
    ) : (
      <>
        <Button
          icon={
            <SVG name={groupByEvent.prop_category} size={16} color={'purple'} />
          }
          type={'link'}
          className={'ml-2'}
          onClick={() => setGroupByDDVisible(true)}
        >
          {propName}
        </Button>
        {renderGroupPropertyOptions(groupByEvent)}
      </>
    );
  };

  const renderGroupBySelect = () => {
    return (
      <div className={styles.group_block__event_selector}>
        <GroupSelect2
          groupedProperties={filterOptions}
          placeholder='Select Property'
          optionClick={(group, val) => onChange(group, val)}
          onClickOutside={() => closeDropDown()}
        ></GroupSelect2>
      </div>
    );
  };

  return (
    <div className={`flex items-center relative w-full`}>
      <Button
        type='text'
        onClick={() => delGroupState(groupByEvent)}
        size={'small'}
        className={`mr-1 fa-btn--custom`}
      >
        {' '}
        <SVG name={'remove'} />{' '}
      </Button>
      <Text level={8} type={'title'} extraClass={'m-0'} weight={'thin'}>
        {grpIndex < 1 ? 'Breakdown' : '...and'}{' '}
      </Text>
      {groupByEvent && groupByEvent.property ? (
        renderGroupContent()
      ) : (
        <>{renderGroupBySelect()}</>
      )}
    </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  userPropNames: state.coreQuery.userPropNames,
  eventPropNames: state.coreQuery.eventPropNames,
});

export default connect(mapStateToProps)(EventGroupBlock);
