import React, { useState, useEffect, useMemo } from 'react';
import cx from 'classnames';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import { SVG, Text } from 'factorsComponents';
import styles from './index.module.scss';
import {
  setGroupBy,
  delGroupBy,
  getGroupProperties
} from 'Reducers/coreQuery/middleware';
import getGroupIcon from 'Utils/getGroupIcon';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
function QueryBlock({
  availableGroups,
  index,
  event,
  eventChange,
  queries,
  queryType,
  eventOptions,
  eventNames,
  activeProject,
  groupBy,
  setGroupBy,
  delGroupBy,
  groupProperties,
  getGroupProperties,
  groupAnalysis
}) {
  const [isDDVisible, setDDVisible] = useState(false);

  const eventGroup = useMemo(() => {
    const group =
      availableGroups?.find((group) => group[0] === event?.group) || [];
    return group[1];
  }, [availableGroups, event]);

  const showGroups = useMemo(() => {
    let showOpts = [];
    if (['users', 'events'].includes(groupAnalysis)) {
      showOpts = [...eventOptions];
    } else {
      const groupOpts = eventOptions?.filter((item) => {
        const [groupDisplayName] =
          availableGroups?.find((group) => group[1] === groupAnalysis) || [];
        return item.label === groupDisplayName;
      });
      const groupNamesList = availableGroups.map((item) => item[0]);
      const userOpts = eventOptions?.filter(
        (item) => !groupNamesList.includes(item?.label)
      );
      showOpts = groupOpts.concat(userOpts);
    }
    showOpts = showOpts?.map((opt) => {
      return {
        iconName: getGroupIcon(opt?.icon),
        label: opt?.label,
        values: opt?.values?.map((op) => {
          return { value: op[1], label: op[0] };
        })
      };
    });
    // Moving MostRecent as first Option.
    const mostRecentGroupindex = showOpts
      ?.map((opt) => opt.label)
      ?.indexOf('Most Recent');
    if (mostRecentGroupindex > 0) {
      showOpts = [
        showOpts[mostRecentGroupindex],
        ...showOpts.slice(0, mostRecentGroupindex),
        ...showOpts.slice(mostRecentGroupindex + 1)
      ];
    }
    return showOpts;
  }, [eventOptions, groupAnalysis, availableGroups]);

  useEffect(() => {
    if (!event) return;
    if (eventGroup?.length && !groupProperties[eventGroup]) {
      getGroupProperties(activeProject.id, eventGroup);
    }
  }, [event, activeProject.id, eventGroup]);

  const onChange = (option, group) => {
    const newEvent = { alias: '', label: '', filters: [], group: '', icon: '' };
    newEvent.icon = group.icon;
    newEvent.label = option.value;
    newEvent.group = group.label;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const selectEvents = () =>
    isDDVisible ? (
      <div className={styles.query_block__event_selector}>
        <GroupSelect
          options={showGroups}
          optionClickCallback={onChange}
          allowSearch={true}
          onClickOutside={() => setDDVisible(false)}
          extraClass={`${styles.query_block__event_selector__select}`}
        />
      </div>
    ) : null;

  const ifQueries = queries.length > 0;
  if (!event) {
    return (
      <div
        className={`${styles.query_block} fa--query_block my-2 ${
          ifQueries ? 'borderless no-padding' : 'borderless no-padding'
        }`}
      >
        <div
          className={`${styles.query_block__event} flex justify-start items-center`}
        >
          <Button
            type='text'
            onClick={triggerDropDown}
            icon={<SVG name='plus' color='grey' />}
          >
            {ifQueries ? 'Add another event' : 'Select Event'}
          </Button>
          {selectEvents()}
        </div>
      </div>
    );
  }

  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding mt-2`}
    >
      <div
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${
          styles.query_block__event
        } block_section items-center`}
      >
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-2'}`}>
          <div className='relative'>
            <Tooltip
              title={
                eventNames[event.label] ? eventNames[event.label] : event.label
              }
            >
              <Button
                icon={
                  <SVG
                    name={
                      showGroups.find((group) => group.label === event.group)
                        ?.iconName
                    }
                    size={20}
                  />
                }
                className='fa-button--truncate fa-button--truncate-lg btn-total-round'
                type='link'
                onClick={triggerDropDown}
              >
                {eventNames[event.label]
                  ? eventNames[event.label]
                  : event.label}
              </Button>
              {selectEvents()}
            </Tooltip>
          </div>
        </div>
      </div>
    </div>
  );
}

const mapStateToProps = (state) => ({
  eventOptions: state.coreQuery.eventOptions,
  activeProject: state.global.active_project,
  groupProperties: state.coreQuery.groupProperties,
  groupBy: state.coreQuery.groupBy.event,
  groupByMagic: state.coreQuery.groupBy,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy,
      getGroupProperties
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
