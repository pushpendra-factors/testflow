import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';
import { bindActionCreators } from 'redux';

import { Button } from 'antd';
import GroupSelect from '../GroupSelect';

import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';

function GroupBlock({
  groupByState,
  setGroupBy,
  delGroupBy,
  userProperties
}) {
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
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

  const onChange = (value, index) => {
    const newGroupByState = Object.assign({}, groupByState.global[index]);
    newGroupByState.prop_category = 'user';
    newGroupByState.eventName = '$present';
    newGroupByState.property = value[1][0];
    newGroupByState.prop_type = value[1][1];
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
    return (<div key={0} className={`${styles.group_block__select} flex justify-start items-center ml-10 mt-2`} >
      {!isDDVisible[index] &&
        <Button size={'large'} type="link" onClick={() => triggerDropDown(index)}><SVG name="plus" /> Select user property</Button> }
      {isDDVisible[index]
        ? (<GroupSelect groupedProperties={filterOptions}
          placeholder="Select Property"
          optionClick={(group, val) => onChange([group, val], index)}
          onClickOutside={() => triggerDropDown(index, true)}

          >
            </GroupSelect>

        )

        : null
      }
    </div>);
  };

  const renderExistingBreakdowns = () => {
    if (groupByState.global.length < 1) return;
    return (groupByState.global.map((opt, index) => (
      <div key={index} className={`${styles.group_block__select} flex justify-start items-center ml-10 mt-2`} >
        {!isDDVisible[index] && <>
        <Button size={'small'}
        type="text"
        onClick={() => delOption(index)}
        className={`${styles.group_block__remove} mr-2ß`}>
          <SVG name="remove"></SVG></Button>

        <Button type="link" onClick={() => triggerDropDown(index)}>{!opt.property && <SVG name="plus" extraClass={`mr-2`} />} {opt.property ? opt.property : 'Select user property'}</Button>
        </>
        }
        {isDDVisible[index]
          ? (<GroupSelect groupedProperties={filterOptions}
            placeholder="Add new"
            optionClick={(group, val) => onChange([group, val], index)}
            onClickOutside={() => triggerDropDown(index, true)}

            >
              </GroupSelect>

          )

          : null
        }
      </div>
    )));
  };

  return (
    <div className={'flex flex-col justify-start'}>

      <div className={`${styles.group_block__event} flex justify-start items-center`}>
        <div className={'fa--query_block--add-event inactive flex justify-center items-center mr-2'}><SVG name={'groupby'} size={24} color={'purple'}/></div>
        <Text type={'title'} level={6} weight={'thin'} extraClass={'m-0'}>Group By</Text>
      </div>

      {renderExistingBreakdowns()}
      {renderInitGroupSelect(groupByState.global.length)}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  groupByState: state.coreQuery.groupBy
});

const mapDispatchToProps = dispatch => bindActionCreators({
  setGroupBy,
  delGroupBy

}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(GroupBlock);
