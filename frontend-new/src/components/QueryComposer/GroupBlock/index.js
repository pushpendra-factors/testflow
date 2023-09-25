import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import styles from './index.module.scss';
import { SVG } from 'Components/factorsComponents';
import { bindActionCreators } from 'redux';

import { Button, Tooltip } from 'antd';

import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';
import FaSelect from '../../FaSelect';
import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';
import {
  PropTextFormat,
  convertAndAddPropertiesToGroupSelectOptions,
  processProperties
} from 'Utils/dataFormatter';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import { GroupDisplayNames } from 'Components/Profile/utils';
import { CustomGroupNames } from 'Components/GlobalFilter/FilterWrapper/utils';

function GroupBlock({
  groupByState,
  setGroupBy,
  delGroupBy,
  userPropertiesV2,
  groupProperties,
  userPropNames,
  groupPropNames,
  groupName,
  groupOpts
}) {
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  const [propSelVis, setSelVis] = useState([false]);
  const [filterOptions, setFilterOptions] = useState([]);

  useEffect(() => {
    const filterOptsObj = {};

    if (groupName === 'users' || groupName === 'events') {
      if (userPropertiesV2) {
        convertAndAddPropertiesToGroupSelectOptions(
          userPropertiesV2,
          filterOptsObj,
          'user'
        );
      }
    } else if (groupName !== '$domains') {
      const groupLabel = CustomGroupNames[groupName]
        ? CustomGroupNames[groupName]
        : groupOpts[groupName]
        ? groupOpts[groupName]
        : PropTextFormat(groupName);
      const groupValues = processProperties(
        groupProperties[groupName],
        'group',
        groupName
      );
      const groupPropIconName = getGroupIcon(groupLabel);
      filterOptsObj[groupLabel] = {
        iconName: groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
        label: groupLabel,
        values: groupValues
      };
    } else {
      for (const [group, properties] of Object.entries(groupProperties || {})) {
        if (Object.keys(GroupDisplayNames).includes(group)) {
          const groupLabel = CustomGroupNames[group]
            ? CustomGroupNames[group]
            : groupOpts[group]
            ? groupOpts[group]
            : PropTextFormat(group);
          const groupValues = processProperties(properties, 'group', group);
          const groupPropIconName = getGroupIcon(groupLabel);
          filterOptsObj[groupLabel] = {
            iconName:
              groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
            label: groupLabel,
            values: groupValues
          };
        }
      }
    }

    setFilterOptions(Object.values(filterOptsObj));
  }, [userPropertiesV2, groupProperties, groupName]);

  const delOption = (index) => {
    delGroupBy('global', groupByState.global[index], index);
  };

  const onGrpPropChange = (val, index) => {
    const newGroupByState = Object.assign({}, groupByState.global[index]);
    if (newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = val;
    }
    if (newGroupByState.prop_type === 'datetime') {
      newGroupByState.grn = val;
    }
    setGroupBy('global', newGroupByState, index);
    const ddVis = [...propSelVis];
    ddVis[index] = false;
    setSelVis(ddVis);
  };

  const onChange = (option, group, index) => {
    const newGroupByState = {
      prop_category: option?.extraProps?.propertyType,
      eventName: '$present',
      groupName: option?.extraProps?.groupName,
      property: option?.value,
      prop_type: option?.extraProps?.valueType,
      gbty: option?.extraProps?.valueType === 'numerical' ? 'raw_values' : '',
      grn: option?.extraProps?.valueType === 'datetime' ? 'day' : ''
    };

    setGroupBy('global', newGroupByState, index);

    const ddVis = [...isDDVisible];
    ddVis[index] = false;
    setDDVisible(ddVis);

    const valDD = [...isValueDDVisible];
    valDD[index] = true;
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
          {
            <Button
              className={`fa-button--truncate`}
              type='text'
              onClick={() => triggerDropDown(index)}
              icon={<SVG name='plus' />}
            >
              Add new
            </Button>
          }
          {isDDVisible[index] ? (
            <div className={`${styles.group_block__event_selector}`}>
              <GroupSelect
                options={filterOptions}
                searchPlaceHolder='Select Property'
                optionClickCallback={(option, group) =>
                  onChange(option, group, index)
                }
                onClickOutside={() => triggerDropDown(index, true)}
                allowSearch={true}
                extraClass={styles.group_block__event_selector__select}
                allowSearchTextSelection={false}
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

  const renderGroupDisplayName = (opt, index) => {
    let propertyName = '';
    if (opt.property && opt.prop_category === 'user') {
      propertyName = userPropNames[opt.property]
        ? userPropNames[opt.property]
        : PropTextFormat(opt.property);
    }
    if (opt.property && opt.prop_category === 'group') {
      propertyName = groupPropNames[opt.property]
        ? groupPropNames[opt.property]
        : PropTextFormat(opt.property);
    }
    if (!opt.property) {
      propertyName = 'Select user property';
    }
    return (
      <Tooltip title={propertyName} color={TOOLTIP_CONSTANTS.DARK}>
        <Button
          icon={<SVG name={opt.prop_category} size={16} color={'purple'} />}
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
    if (groupByState.global.length < 1) return;
    return groupByState.global.map((opt, index) => (
      <div key={index} className={`flex relative items-center mt-2`}>
        {
          <>
            <div className={`flex relative`}>
              {renderGroupDisplayName(opt, index)}
              {isDDVisible[index] ? (
                <div className={`${styles.group_block__event_selector}`}>
                  <GroupSelect
                    options={filterOptions}
                    searchPlaceHolder='Select Property'
                    optionClickCallback={(option, group) =>
                      onChange(option, group, index)
                    }
                    onClickOutside={() => triggerDropDown(index, true)}
                    allowSearch={true}
                    extraClass={styles.group_block__event_selector__select}
                    allowSearchTextSelection={false}
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
    <div className={'flex flex-col justify-start'}>
      {renderExistingBreakdowns()}
      {renderInitGroupSelect(groupByState.global.length)}
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
  groupOpts: state.groups.data
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(GroupBlock);
