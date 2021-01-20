import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';

import { Button } from 'antd';

import { SVG } from 'factorsComponents';

import { connect } from 'react-redux';
import GroupSelect from '../GroupSelect';
import FaSelect from '../../FaSelect';

const EventGroupBlock = ({
  index,
  eventIndex, groupByEvent, event, userProperties, eventProperties,
  setGroupState,
  delGroupState, closeDropDown
}) => {
  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'User Properties',
      icon: 'userplus',
      values: []
    },
    {
      label: 'Event Properties',
      icon: 'mouseclick',
      values: []
    }
  ]);

  const [propSelVis, setSelVis] = useState(false);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[0].values = userProperties;
    setFilterOptions(filterOpts);
  }, [userProperties]);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[1].values = eventProperties[event.label];
    setFilterOptions(filterOpts);
  }, [eventProperties]);

  const onChange = (group, val) => {
    const newGroupByState = Object.assign({}, groupByEvent);
    if (group === 'User Properties') {
      newGroupByState.prop_category = 'user';
    } else {
      newGroupByState.prop_category = 'event';
    }

    newGroupByState.eventName = event.label;
    newGroupByState.property = val[0];
    newGroupByState.prop_type = val[1];
    newGroupByState.eventIndex = eventIndex;

    setGroupState(newGroupByState);
    closeDropDown();
  };

  const onGrpPropChange = (val) => {
    const newGroupByState = Object.assign({}, groupByEvent);
    if(newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = val;
    }
    if(newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = val;
    }
    setGroupState(newGroupByState, index );
    setSelVis(false);
  }

  const renderGroupPropertyOptions = (opt) => {
    if(!opt || opt.prop_type === 'categorical') return;

    const propOpts = {
      'numerical': [
          ['original values', null, 'raw_values'], 
          ['bucketed values', null, 'with_buckets']],
      'datetime': [
        ['hourly', null, 'hour'],
        ['daily', null, 'day'],
        ['week', null, 'week'],
      ]
    }

    const getProp = (opt) => {
      if(opt.prop_type === 'numerical') {
        const propSel = propOpts['numerical'].filter((v) => v[2] === opt.gbty);
        return propSel[0]? propSel[0][0] : 'Select options';
      }
      if(opt.prop_type === 'datetime') {
        const propSel = propOpts['datetime'].filter((v) => v[2] === opt.grn);
        return propSel[0]? propSel[0][0] : 'Select options';
      } 
    }

    const setProp = (opt) => {
      onGrpPropChange(opt[2]);
      setSelVis(false);
    }

    return (<div className={styles.grpProps}>
      show as <div className={styles.grpProps__select}>
          <span className={styles.grpProps__select__opt} 
            onClick={() => setSelVis(true)}> 
            { getProp(opt)}  
          </span>
          {propSelVis && 
            <FaSelect options={propOpts[opt.prop_type]}
              optionClick={setProp}
              onClickOutside={() => setSelVis(false)}
            
            ></FaSelect> 
          }
        </div>
    </div>);
  }

  const renderGroupContent = () => {
    return (
          <div className={`${styles.group_block__group_content} ml-4`}>

            {groupByEvent.property}

            {renderGroupPropertyOptions(groupByEvent)}
          </div>
    );
  };

  const renderGroupBySelect = () => {
    return (<GroupSelect groupedProperties={filterOptions}
            placeholder="Select Property"
            optionClick={(group, val) => onChange(group, val)}
            onClickOutside={() => closeDropDown()}
            >
              </GroupSelect>);
  };

  return (
        <div className={styles.group_block}>
        <Button size={'small'} type="text" onClick={() => delGroupState(groupByEvent)} className={`${styles.group_block__remove} mr-1`}><SVG name="remove"></SVG></Button>
        <span className={`${styles.group_block__prefix} ml-10`}>group by</span>
        {groupByEvent && groupByEvent.property
          ? renderGroupContent()
          : <>
            {renderGroupBySelect()}
          </>
        }

        </div>
  );
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties
});

export default connect(mapStateToProps)(EventGroupBlock);
