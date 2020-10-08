import React, { useState } from 'react';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';

import { Select, Button } from 'antd';

const { OptGroup, Option } = Select;

function GroupBlock({
  queryType, events, groupByState, setGroupByState, filterOptions
}) {
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  // const [groupByState, setGroupByState] = useState(groupBy);

  const onChange = (value, index) => {
    const newGroupByState = Object.assign({}, groupByState[index]);
    value[0] === 'Event Properties'
      ? newGroupByState.prop_category = 'event'
      : newGroupByState.prop_category = 'user';

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

  const onEventValueChange = (value, index) => {
    const newGroupByState = Object.assign({}, groupByState[index]);
    newGroupByState.eventValue = value;
    setGroupByState(newGroupByState, index);
    const ddVis = [...isDDVisible];
    ddVis[index] = false;
    setValueDDVisible(ddVis);
  };

  const triggerDropDown = (index) => {
    const ddVis = [...isDDVisible];
    ddVis[index] = true;
    setDDVisible(ddVis);
  };

  const triggerValueDropDown = (index) => {
    const ddVis = [...isValueDDVisible];
    ddVis[index] = true;
    setValueDDVisible(ddVis);
  };

  const renderGroupedEventOptions = () => {
    return filterOptions.map((group, index) => (
      <OptGroup key={index} label={(
        <div className={styles.group_block__selector_group}>
          <SVG name={group.icon}></SVG>
          <span >{group.label}</span>
        </div>
      )}>
        {group.values.map((option, index) => (
          <Option key={index} value={[group.label, option]}>{option[0]}</Option>
        ))}
      </OptGroup>
    ));
  };

  return (
    <div className={'flex flex-col justify-start'}>

      <div className={`${styles.query_block__event} flex justify-start items-center`}>
        <div className={'fa--query_block--add-event inactive flex justify-center items-center mr-2'}><SVG name={'groupby'} size={36} color={'purple'}/></div>
        <Text type={'title'} level={6} weight={'thin'} extraClass={'m-0'}>Group By</Text>
      </div>

      {
        groupByState.map((opt, index) => (
          <div key={0} className={'flex justify-start items-center ml-10 mt-2'} >
          {!isDDVisible[index] && <Button type="link" onClick={() => triggerDropDown(index)}>{!opt.property && <SVG name="plus" />} {opt.property ? opt.property : 'Select user property'}</Button> }
          {isDDVisible[index]
            ? (<Select
              placeholder="Select Property"
              onChange={(val) => onChange(val, index)} defaultOpen={true}
              dropdownRender={menu => (
                <div className={styles.group_block__selector_body}>
                  {menu}
                </div>
              )} style={{ width: 200 }} showArrow={false} showSearch>
              {renderGroupedEventOptions()}
            </Select>)

            : null
          }

        {opt.property && queryType === 'funnel' &&
        <>
          <Text type={'title'} level={7} weight={'thin'} extraClass={'mx-2 m-0'}>with values</Text>

          {!isValueDDVisible[index] && <Button type="link" onClick={triggerValueDropDown}>{opt.eventValue ? opt.eventValue : events[0].label }</Button> }
          {isValueDDVisible[index] &&
          <Select style={{ width: 200 }} showArrow={false}
              defaultOpen={true}
              onChange={(val) => onEventValueChange(val, index)} >
              {events.map((event, index) => (
                <Option key={index} value={event.label}></Option>
              ))}
            </Select>
          }
        </>
          }
        </div>
        ))

      }
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

export default connect(mapStateToProps)(GroupBlock);
