import React, { useState } from 'react';
import styles from './index.module.scss';
import { SVG, Text } from 'factorsComponents';

import { Select, Button } from 'antd';
import { group } from 'd3';
import { queries } from '@testing-library/react';

const { OptGroup, Option } = Select;

export default function GroupBlock({ events, groupBy }) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isValueDDVisible, setValueDDVisible] = useState(false);
  const [groupByState, setGroupByState] = useState(groupBy);

  const filterOptions = [
    {
      label: 'User Properties',
      icon: 'fav',
      values: [
        'Country',
        'City',
        'Age'
      ]
    }, {
      label: 'Event Properties',
      icon: 'virtual',
      values: [
        'Add to WishList',
        'Applied Coupon',
        'Cart Updated'
      ]
    }
  ];

  const onChange = (value) => {
    const newGroupByState = Object.assign({}, groupByState);
    newGroupByState.property = value;
    setDDVisible(false);
    setGroupByState(newGroupByState);
  };

  const onEventValueChange = (value) => {
    const newGroupByState = Object.assign({}, groupByState);
    newGroupByState.eventValue = value;
    setValueDDVisible(false);
    setGroupByState(newGroupByState);
  };

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const triggerValueDropDown = () => {
    setValueDDVisible(true);
  };

  const renderGroupedEventOptions = () => {
    return filterOptions.map(group => (
      <OptGroup label={(
        <div className={styles.group_block__selector_group}>
          <SVG name={group.icon}></SVG>
          <span >{group.label}</span>
        </div>
      )}>
        {group.values.map((option) => (
          <Option value={option}></Option>
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

      <div className={'flex justify-start items-center ml-10 mt-2'} >
        {!isDDVisible && <Button type="link" onClick={triggerDropDown}>{!groupByState.property && <SVG name="plus" />} {groupByState.property ? groupByState.property : 'Select user property'}</Button> }
        {isDDVisible
          ? (<Select
            placeholder="Select Property"
            onChange={onChange} defaultOpen={true}
            dropdownRender={menu => (
              <div className={styles.group_block__selector_body}>
                {menu}
              </div>
            )} style={{ width: 200 }} showArrow={false} showSearch>
            {renderGroupedEventOptions()}
          </Select>)

          : null
        }

        <Text type={'title'} level={7} weight={'thin'} extraClass={'mx-2 m-0'}>with values</Text>

        {!isValueDDVisible && <Button type="link" onClick={triggerValueDropDown}>{groupByState.eventValue ? groupByState.eventValue : events[0].label }</Button> }
        {isValueDDVisible
          ? <Select style={{ width: 200 }} showArrow={false}
            defaultOpen={true}
            onChange={onEventValueChange} >
            {events.map(event => (
              <Option value={event.label}></Option>
            ))}
          </Select> : null }
      </div>

    </div>
  );
}
