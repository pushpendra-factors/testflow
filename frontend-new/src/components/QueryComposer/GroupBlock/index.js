import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { bindActionCreators } from 'redux';

import { Button, Tooltip } from 'antd';
import GroupSelect2 from '../GroupSelect2';

import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';

function GroupBlock({
  groupByState,
  setGroupBy,
  delGroupBy,
  userProperties,
  userPropNames,
  eventPropNames
}) {
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  const [propSelVis, setSelVis] = useState([false]);
  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'User Properties',
      icon: 'userplus',
      values: []
    }
  ]);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[0].values = userProperties;
    setFilterOptions(filterOpts);
  }, [userProperties]);

  const delOption = (index) => {
    delGroupBy('global', groupByState.global[index], index);
  };

  const onGrpPropChange = (val, index) => {
    const newGroupByState = Object.assign({}, groupByState.global[index]);
    if(newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = val;
    }
    if(newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = val;
    }
    setGroupBy('global', newGroupByState, index);
    const ddVis = [...propSelVis];
    ddVis[index] = false;
    setSelVis(ddVis);
  }

  const onChange = (value, index) => {
    const newGroupByState = Object.assign({}, groupByState.global[index]);
    newGroupByState.prop_category = 'user';
    newGroupByState.eventName = '$present';
    newGroupByState.property = value[1][1];
    newGroupByState.prop_type = value[1][2];
    if(newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = 'raw_values';
    }
    if(newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = 'day';
    }
    
    setGroupBy('global', newGroupByState, index);
    const ddVis = [...isDDVisible];
    ddVis[index] = false;
    const valDD = [isValueDDVisible];
    valDD[index] = true;
    setDDVisible(ddVis);
    setValueDDVisible(valDD);
  };

  const triggerDropDown = (index, close = false) => {
    const ddVis = [...isDDVisible];
    ddVis[index] = !close;
    setDDVisible(ddVis);
  };

  const renderInitGroupSelect = (index) => {
    return (<div key={0} className={`${styles.group_block__select} flex justify-start items-center mt-2`} >
      {
        <Button className={`fa-button--truncate`} type="text" onClick={() => triggerDropDown(index)} icon={<SVG name="plus" />}> Add new </Button>}
      
      <div className={styles.group_block__event_selector}>
      {isDDVisible[index]
        ? (
          <div className={styles.group_block__event_selector__btn}>
            <GroupSelect2 groupedProperties={filterOptions}
          placeholder="Select Property"
          optionClick={(group, val) => onChange([group, val], index)}
          onClickOutside={() => triggerDropDown(index, true)}
        ></GroupSelect2>
        </div>
        ) : null}
      </div>
    </div>);
  };

  const renderGroupPropertyOptions = (opt, index) => {
    if(!opt || opt.prop_type === 'categorical') return;

    const propOpts = {
      'numerical': [
          ['original values', null, 'raw_values'], 
          ['bucketed values', null, 'with_buckets']],
      'datetime': [
        ['hour', null, 'hour'],
        ['date', null, 'day'],
        ['week', null, 'week'],
        ['month', null, 'month']
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

    const setProp = (opt, i = index) => {
      onGrpPropChange(opt[2], i);
      selectVisToggle()
    }

    const selectVisToggle = (i = index) => {
      const visState = [...propSelVis];
      visState[i] = !visState[i];
      setSelVis(visState);
    }

    return (<div className={styles.grpProps}>
      show as <div className={styles.grpProps__select}>
          <span className={styles.grpProps__select__opt} 
            onClick={() => selectVisToggle()}> 
            { getProp(opt)}  
          </span>
          {propSelVis[index] && 
            <FaSelect options={propOpts[opt.prop_type]}
              optionClick={setProp}
              onClickOutside={() => selectVisToggle()}
            
            ></FaSelect> 
          }
        </div>
    </div>);
  }

  const renderGroupDisplayName = (opt, index) => {
    let propertyName = '';
    if(opt.property && opt.prop_category === 'user') {
      propertyName = userPropNames[opt.property]?  userPropNames[opt.property] : opt.property;
    }
    if(opt.property && opt.prop_category === 'event') {
      propertyName = eventPropNames[opt.property]?  eventPropNames[opt.property] : opt.property;
    }
    if(!opt.property) {
      propertyName = 'Select user property';
    }
    return (
      <Tooltip title={propertyName}>
        <Button className={`fa-button--truncate`} type="link" onClick={() => triggerDropDown(index)}>{!opt.property && <SVG name="plus" extraClass={`mr-2`} />} {propertyName}</Button>
      </Tooltip>
    )
  }

  const renderExistingBreakdowns = () => {
    if (groupByState.global.length < 1) return;
    return (groupByState.global.map((opt, index) => (
      <div key={index} className={`${styles.group_block__select} flex justify-start items-center mt-2`} >
        {
          <>
        <Button
        type="text"
        className={`fa-button--truncate`}
        onClick={() => delOption(index)}
        className={`${styles.group_block__remove} mr-2`}
        icon={<SVG name="delete" />}
        /> 
        {renderGroupDisplayName(opt, index)}
        {renderGroupPropertyOptions(opt, index)}
        </>
        }
        {isDDVisible[index]
          ? (<GroupSelect2 groupedProperties={filterOptions}
            placeholder="Add new"
            optionClick={(group, val) => onChange([group, val], index)}
            onClickOutside={() => triggerDropDown(index, true)}

            >
              </GroupSelect2>

          )

          : null
        }
      </div>
    )));
  };

  return (
    <div className={'flex flex-col justify-start'}>

      {/* <div className={`${styles.group_block__event} flex justify-start items-center`}>
        <div className={'fa--query_block--add-event inactive flex justify-center items-center mr-2'}><SVG name={'groupby'} size={24} color={'purple'}/></div>
        <Text type={'title'} level={6} weight={'thin'} extraClass={'m-0'}>Group By</Text>
      </div> */}

      {renderExistingBreakdowns()}
      {renderInitGroupSelect(groupByState.global.length)}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  userPropNames: state.coreQuery.userPropNames,
  eventPropNames: state.coreQuery.eventPropNames,
  groupByState: state.coreQuery.groupBy
});

const mapDispatchToProps = dispatch => bindActionCreators({
  setGroupBy,
  delGroupBy

}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(GroupBlock);
