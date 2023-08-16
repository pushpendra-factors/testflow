import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import styles from 'Components/QueryComposer/GroupBlock/index.module.scss';
import { SVG, Text } from 'Components/factorsComponents';
import { bindActionCreators } from 'redux';
import { Button, Tooltip } from 'antd';
import {
  setGroupBy,
  delGroupBy,
  getEventPropertiesV2
} from 'Reducers/coreQuery/middleware';
import FaSelect from 'Components/FaSelect';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import { convertGroupedPropertiesToUngrouped } from 'Utils/dataFormatter';

function GroupBlock({
  groupByState,
  setGroupBy,
  delGroupBy,
  userPropertiesV2,
  groupProperties,
  userPropNames,
  groupPropNames,
  groupName = 'users',
  isDDVisible,
  setDDVisible,
  setEventLevelExpandBy,
  eventLevelExpandBy,
  buttonClickPropNames,
  pageViewPropNames,
  eventPropertiesV2,
  getEventPropertiesV2,
  eventTypeName = '',
  eventItem,
  activeProject,
  eventPropNames
}) {
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  const [propSelVis, setSelVis] = useState([false]);
  const [filterOptions, setFilterOptions] = useState([
    {
      label: 'Event Properties',
      iconName: 'event',
      values: []
    }
  ]);

  const selectedEventName = eventItem?.value?.split(',')[0];
  useEffect(() => {
    getEventPropertiesV2(activeProject?.id, selectedEventName);
  }, [eventItem]);

  const modifyUserProperties = () => {
    const userPropertiesModified = [];
    if (userPropertiesV2) {
      convertGroupedPropertiesToUngrouped(
        userPropertiesV2,
        userPropertiesModified
      );
    }
    return userPropertiesModified;
  };
  const modifyEventProperties = (eventName) => {
    const eventProps = [];
    if (eventName && eventPropertiesV2?.[eventName]) {
      convertGroupedPropertiesToUngrouped(
        eventPropertiesV2?.[eventName],
        eventProps
      );
    }
    return eventProps;
  };
  useEffect(() => {
    const filterOpts = [...filterOptions];

    switch (selectedEventName) {
      case 'Button Clicks':
        filterOpts[0].values = buttonClickPropNames;
        break;
      case 'Sessions':
        filterOpts[0].values = modifyEventProperties('$session');
        break;
      case 'CRM Events':
        filterOpts[0].values = modifyUserProperties();
        break;
      case 'Page Views':
        filterOpts[0].values = pageViewPropNames;
        break;
      case selectedEventName:
        filterOpts[0].values = modifyEventProperties(selectedEventName);
        break;
      default:
        filterOpts[0].values = [];
    }

    const modifiedFilterOpts = filterOpts?.map((opt) => {
      return {
        iconName: opt?.iconName,
        label: opt?.label,
        values: opt?.values?.map((op) => {
          return {
            value: op[1],
            label: op[0],
            extraProps: {
              valueType: op[2]
            }
          };
        })
      };
    });
    setFilterOptions(modifiedFilterOpts);
  }, [
    userPropertiesV2,
    groupProperties,
    groupName,
    eventTypeName,
    eventPropertiesV2,
    buttonClickPropNames,
    pageViewPropNames,
    selectedEventName
  ]);

  const delOption = (index) => {
    // delGroupBy('global', groupByState.global[index], index);
    let newArr = eventLevelExpandBy?.filter((item, indx) => indx != index);
    setEventLevelExpandBy(newArr);
  };

  const onGrpPropChange = (val, index) => {
    const newGroupByState = Object.assign({}, eventLevelExpandBy[index]);
    if (newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = val;
    }
    if (newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = val;
    }
    // setGroupBy('global', newGroupByState, index);
    if (eventLevelExpandBy?.[index]) {
      let newArr = eventLevelExpandBy;
      newArr[index] = newGroupByState;
      setEventLevelExpandBy(newArr);
    } else {
      setEventLevelExpandBy([...eventLevelExpandBy, newGroupByState]);
    }
    const ddVis = [...propSelVis];
    ddVis[index] = false;
    setSelVis(ddVis);
  };

  const onChange = (option, group, index) => {
    const newGroupByState = Object.assign({}, eventLevelExpandBy[index]);
    if (group?.label === 'Group Properties') {
      newGroupByState.prop_category = 'group';
    } else {
      newGroupByState.prop_category = 'event';
    }
    newGroupByState.eventName = '$present';
    newGroupByState.property = option?.value;
    newGroupByState.prop_type = option?.extraProps?.valueType;
    if (newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = 'raw_values';
    }
    if (newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = 'day';
    }
    // setGroupBy('global', newGroupByState, index);
    if (eventLevelExpandBy?.[index]) {
      let newArr = eventLevelExpandBy;
      newArr[index] = newGroupByState;
      setEventLevelExpandBy(newArr);
    } else {
      setEventLevelExpandBy([...eventLevelExpandBy, newGroupByState]);
    }
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
    return (
      <div key={0} className={`m-0 mt-2`}>
        <div className={`flex relative`}>
          {/* {
            <Button
              className={`fa-button--truncate`}
              type='text'
              onClick={() => triggerDropDown(index)}
              icon={<SVG name='plus' />}
            >
              Add new
            </Button>
          } */}
          {isDDVisible[index] ? (
            <div className={`${styles.group_block__event_selector}`}>
              <GroupSelect
                options={filterOptions}
                searchPlaceHolder={'Select Property'}
                optionClickCallback={(option, group) =>
                  onChange(option, group, index)
                }
                onClickOutside={() => triggerDropDown(index, true)}
                allowSearch={true}
                extraClass={`${styles.group_block__event_selector__select}`}
              />
            </div>
          ) : null}
        </div>
      </div>
    );
  };

  const renderGroupPropertyOptions = (opt, index) => {
    if (!opt || opt.prop_type === 'categorical') return;

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

    const setProp = (opt, i = index) => {
      onGrpPropChange(opt[2], i);
      selectVisToggle();
    };

    const selectVisToggle = (i = index) => {
      const visState = [...propSelVis];
      visState[i] = !visState[i];
      setSelVis(visState);
    };

    return (
      <div className={`flex items-center m-0 mx-2`}>
        show as
        <div
          className={`flex relative m-0 mx-2 ${styles.grpProps__select__opt}`}
          onClick={() => selectVisToggle()}
        >
          {getProp(opt)}
          {propSelVis[index] && (
            <FaSelect
              options={propOpts[opt.prop_type]}
              optionClick={setProp}
              onClickOutside={() => selectVisToggle()}
            ></FaSelect>
          )}
        </div>
      </div>
    );
  };

  const getIcon = (groupByEvent) => {
    const { property, prop_category } = groupByEvent || {};
    if (!property) return null;
    const iconName = prop_category === 'group' ? 'user' : prop_category;
    return <SVG name={iconName} size={16} color={'purple'} />;
  };

  const renderGroupDisplayName = (opt, index) => {
    let propertyName = '';
    if (opt.property && opt.prop_category === 'event') {
      propertyName = eventPropNames[opt.property]
        ? eventPropNames[opt.property]
        : userPropNames[opt.property]
        ? userPropNames[opt.property]
        : opt.property;
    }
    if (opt.property && opt.prop_category === 'group') {
      propertyName = groupPropNames[opt.property]
        ? groupPropNames[opt.property]
        : opt.property;
    }
    if (!opt.property) {
      propertyName = 'Select user property';
    }
    return (
      <Tooltip title={propertyName} color={TOOLTIP_CONSTANTS.DARK}>
        <Button
          icon={getIcon(opt)}
          className={`fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin`}
          type='link'
          onClick={() => triggerDropDown(index)}
        >
          {!opt.property && <SVG name='plus' extraClass={`mr-2`} />}
          {propertyName}
        </Button>
      </Tooltip>
    );
  };

  const renderExistingBreakdowns = () => {
    if (eventLevelExpandBy < 1) return;
    return eventLevelExpandBy.map((opt, index) => (
      <div key={index} className={`flex relative items-center mt-2`}>
        <Text type={'title'} level={8} extraClass={`m-0 mt-2 mr-4`}>
          Expand by
        </Text>
        {
          <>
            <div className={`flex relative`}>
              {renderGroupDisplayName(opt, index)}
              {isDDVisible[index] ? (
                <div className={`${styles.group_block__event_selector}`}>
                  <GroupSelect
                    options={filterOptions}
                    searchPlaceHolder={'Select Property'}
                    optionClickCallback={(option, group) =>
                      onChange(option, group, index)
                    }
                    onClickOutside={() => triggerDropDown(index, true)}
                    allowSearch={true}
                    extraClass={`${styles.group_block__event_selector__select}`}
                  />
                </div>
              ) : null}
            </div>
            {renderGroupPropertyOptions(opt, index)}

            <Button
              type='text'
              onClick={() => delOption(index)}
              size={'small'}
              className={`fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button`}
            >
              <SVG name={'remove'} />
            </Button>
          </>
        }
      </div>
    ));
  };

  return (
    <div className={'flex flex-col relative justify-start items-start ml-20'}>
      {renderExistingBreakdowns()}
      {renderInitGroupSelect(eventLevelExpandBy?.length)}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userPropertiesV2: state.coreQuery.userPropertiesV2,
  groupProperties: state.coreQuery.groupProperties,
  userPropNames: state.coreQuery.userPropNames,
  groupPropNames: state.coreQuery.groupPropNames,
  groupByState: state.coreQuery.groupBy,
  buttonClickPropNames: state.coreQuery.buttonClickPropNames,
  pageViewPropNames: state.coreQuery.pageViewPropNames,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventPropNames: state.coreQuery.eventPropNames
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy,
      getEventPropertiesV2
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(GroupBlock);
