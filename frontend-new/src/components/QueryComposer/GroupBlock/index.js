import React, { useState, useEffect } from 'react';
import { connect, useDispatch } from 'react-redux';
import { SVG } from 'Components/factorsComponents';
import { bindActionCreators } from 'redux';
import { Button, Tooltip } from 'antd';
import {
  PropTextFormat,
  convertAndAddPropertiesToGroupSelectOptions,
  processProperties
} from 'Utils/dataFormatter';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import { IsDomainGroup } from 'Components/Profile/utils';
import {
  CustomGroupDisplayNames,
  GROUP_NAME_DOMAINS
} from 'Components/GlobalFilter/FilterWrapper/utils';
import { invalidBreakdownPropertiesList } from 'Constants/general.constants';
import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';
import FaSelect from '../../FaSelect';
import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';
import styles from './index.module.scss';
import { ReactSortable } from 'react-sortablejs';
import { setGroupByActionList } from 'Reducers/coreQuery/actions';

function GroupBlock({
  groupByState,
  setGroupBy,
  delGroupBy,
  userPropertiesV2,
  groupProperties,
  userPropNames,
  groupPropNames,
  groupName,
  groups
}) {
  const dispatch = useDispatch();
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  const [propSelVis, setSelVis] = useState([false]);
  const [filterOptions, setFilterOptions] = useState([]);

  useEffect(() => {
    const filterOptsObj = {};

    const populateFilterOpts = (group, properties) => {
      const groupLabel =
        CustomGroupDisplayNames[group] ||
        groups?.all_groups?.[group] ||
        PropTextFormat(group);

      const groupValues = processProperties(properties, 'user', group);
      const groupPropIconName = getGroupIcon(groupLabel);

      filterOptsObj[groupLabel] = {
        iconName: groupPropIconName === 'NoImage' ? 'group' : groupPropIconName,
        label: groupLabel,
        values: groupValues
      };
    };

    if ((groupName === 'users' || groupName === 'events') && userPropertiesV2) {
      convertAndAddPropertiesToGroupSelectOptions(
        userPropertiesV2,
        filterOptsObj,
        'user'
      );
    } else if (!IsDomainGroup(groupName)) {
      const [group, properties] = [groupName, groupProperties[groupName]];
      populateFilterOpts(group, properties);
    } else {
      Object.entries(groupProperties || {}).forEach(([group, properties]) => {
        if (
          Object.keys(groups?.all_groups || {})
            .concat([GROUP_NAME_DOMAINS])
            .includes(group)
        ) {
          let filteredProperties = properties;
          if (group === GROUP_NAME_DOMAINS) {
            filteredProperties = properties.filter(
              (item) => !invalidBreakdownPropertiesList.includes(item[1])
            );
          }
          populateFilterOpts(group, filteredProperties);
        }
      });
    }

    setFilterOptions(Object.values(filterOptsObj));
  }, [userPropertiesV2, groupProperties, groupName]);

  const delOption = (index) => {
    delGroupBy('global', groupByState.global[index], index);
  };

  const onGrpPropChange = (val, index) => {
    const newGroupByState = { ...groupByState.global[index] };
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

  const renderInitGroupSelect = (index) => (
    <div key={0} className='m-0 mt-2'>
      <div className='flex relative'>
        <Button
          className='fa-button--truncate'
          type='text'
          onClick={() => triggerDropDown(index)}
          icon={<SVG name='plus' />}
        >
          Add new
        </Button>
        {isDDVisible[index] ? (
          <div className={`${styles.group_block__event_selector}`}>
            <GroupSelect
              options={filterOptions}
              searchPlaceHolder='Select Property'
              optionClickCallback={(option, group) =>
                onChange(option, group, index)
              }
              onClickOutside={() => triggerDropDown(index, true)}
              allowSearch
              extraClass={styles.group_block__event_selector__select}
              allowSearchTextSelection={false}
            />
          </div>
        ) : null}
      </div>
    </div>
  );

  const renderGroupPropertyOptions = (opt, index) => {
    if (!opt || opt.prop_type === 'categorical') return null;

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
        const propSel = propOpts.numerical.filter((v) => v[2] === opt.gbty);
        return propSel[0] ? propSel[0][0] : 'Select options';
      }
      if (opt.prop_type === 'datetime') {
        const propSel = propOpts.datetime.filter((v) => v[2] === opt.grn);
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
      <div className='flex items-center m-0 mx-2'>
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
            />
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
          icon={<SVG name={opt.prop_category} size={16} color='purple' />}
          className='fa-button--truncate fa-button--truncate-xs btn-left-round filter-buttons-margin'
          type='link'
          onClick={() => triggerDropDown(index)}
        >
          {!opt.property && <SVG name='plus' extraClass='mr-2' />}
          {propertyName}
        </Button>
      </Tooltip>
    );
  };

  const renderExistingBreakdowns = () => {
    if (groupByState.global.length < 1) return null;
    return (
      <ReactSortable
        list={groupByState.global}
        setList={(listItems) => {
          dispatch(setGroupByActionList(listItems));
        }}
      >
        {groupByState.global.map((opt, index) => (
          <div
            key={index}
            className={`flex relative items-center mt-2 ${styles['draghandleparent']}`}
          >
            <div className='flex relative'>
              <div
                style={{ cursor: 'pointer', margin: 'auto 2px' }}
                className={styles['draghandle']}
              >
                <SVG name='drag'></SVG>
              </div>
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
                    allowSearch
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
              size='small'
              className='fa-btn--custom filter-buttons-margin btn-right-round filter-remove-button'
            >
              <SVG name='remove' />
            </Button>
          </div>
        ))}
      </ReactSortable>
    );
  };

  return (
    <div className='flex flex-col justify-start'>
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
  groups: state.coreQuery.groups
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
