import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';

import { connect } from 'react-redux';
import GroupSelect from '../GroupSelect';

const EventGroupBlock = ({
  eventIndex, groupByEvent, event, userProperties, eventProperties, setGroupState, closeDropDown
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

  const renderGroupContent = () => {
    return (
          <div className={`${styles.group_block__group_content} ml-4`}>
            {groupByEvent.property}
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
