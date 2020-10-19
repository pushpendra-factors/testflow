import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from '../../factorsComponents';

import { Button } from 'antd';
import GroupSelect from '../GroupSelect';

function GroupBlock({
  groupByState,
  setGroupByState,
  userProperties
}) {
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'User Properties',
      icon: 'fav',
      values: []
    }
  ]);

  useEffect(() => {
    const filterOpts = [...filterOptions];
    filterOpts[0].values = userProperties;
    setFilterOptions(filterOpts);
  }, [userProperties]);

  const onChange = (value, index) => {
    const newGroupByState = Object.assign({}, groupByState[index]);
    newGroupByState.prop_category = 'user';
    newGroupByState.eventName = '$present';
    newGroupByState.property = value[1][0];
    newGroupByState.prop_type = value[1][1];
    setGroupByState(newGroupByState, index);
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

  return (
    <div className={'flex flex-col justify-start'}>

      <div className={`${styles.group_block__event} flex justify-start items-center`}>
        <div className={'fa--query_block--add-event inactive flex justify-center items-center mr-2'}><SVG name={'groupby'} size={36} color={'purple'}/></div>
        <Text type={'title'} level={6} weight={'thin'} extraClass={'m-0'}>Group By</Text>
      </div>

      {
        groupByState.map((opt, index) => (
          <div key={index} className={`${styles.group_block__select} flex justify-start items-center ml-10 mt-2`} >
            {!isDDVisible[index] && <Button type="link" onClick={() => triggerDropDown(index)}>{!opt.property && <SVG name="plus" />} {opt.property ? opt.property : 'Select user property'}</Button> }
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
        </div>
        ))

      }
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties
});

export default connect(mapStateToProps)(GroupBlock);
