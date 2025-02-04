import React, { useState, useEffect } from 'react';
import { connect, useDispatch } from 'react-redux';
import { SVG } from 'factorsComponents';
import { bindActionCreators } from 'redux';

import { Button, Tooltip } from 'antd';

import _ from 'lodash';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';
import {
  groupKPIPropertiesOnCategory,
  processProperties
} from 'Utils/dataFormatter';
import { ReactSortable } from 'react-sortablejs';
import { setGroupByActionList } from 'Reducers/coreQuery/actions';
import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';
import FaSelect from '../../FaSelect';
import { setGroupBy, delGroupBy } from '../../../reducers/coreQuery/middleware';
import styles from './index.module.scss';

function GroupBlock({
  groupByState,
  setGroupBy,
  delGroupBy,
  userPropNames,
  eventPropNames,
  KPIConfigProps,
  textStartCase,
  propertyMaps,
  isSameKPIGrp,
  selectedMainCategory
}) {
  const [isDDVisible, setDDVisible] = useState([false]);
  const [isValueDDVisible, setValueDDVisible] = useState([false]);
  const [propSelVis, setSelVis] = useState([false]);
  const [filterOptions, setFilterOptions] = useState([]);
  const dispatch = useDispatch();
  useEffect(() => {
    let commonProperties = [];
    if (propertyMaps) {
      commonProperties =
        propertyMaps?.map((item) => [
          item?.display_name,
          item?.name,
          item?.data_type,
          'propMap',
          item?.category
        ]) || [];
    }
    const kpiProperties = !isSameKPIGrp
      ? commonProperties
      : KPIConfigProps || [];
    const kpiItemsgroupedByCategoryProperty = groupKPIPropertiesOnCategory(
      kpiProperties,
      'user',
      selectedMainCategory?.group
    );
    const propertyArrays = Object.values(kpiItemsgroupedByCategoryProperty);

    const modifiedFilterOpts = propertyArrays?.map((opt) => ({
      iconName: opt?.icon,
      label: opt?.label,
      values: processProperties(opt?.values, opt?.propertyType)
    }));
    setFilterOptions(modifiedFilterOpts);
  }, [KPIConfigProps, propertyMaps]);

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
    const newGroupByState = { ...groupByState.global[index] };
    newGroupByState.prop_category = option?.extraProps?.queryType;
    newGroupByState.eventName = option?.extraProps?.valueType;
    newGroupByState.property = option?.value;
    newGroupByState.prop_type = option?.extraProps?.valueType;
    newGroupByState.display_name = option?.label;
    if (newGroupByState.prop_type === 'numerical') {
      newGroupByState.gbty = 'raw_values';
    }
    if (newGroupByState.prop_type === 'datetime') {
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

  const matchEventName = (item) => {
    const findItem = eventPropNames?.[item] || userPropNames?.[item];
    return findItem || item;
  };

  const renderGroupDisplayName = (opt, index) => {
    let propertyName = '';
    if (opt?.property) {
      propertyName = opt?.display_name
        ? opt.display_name
        : matchEventName(opt.property);
    }
    if (!opt.property) {
      propertyName = 'Select user property';
    }
    return (
      <Tooltip title={propertyName} color={TOOLTIP_CONSTANTS.DARK}>
        <Button
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
    if (groupByState.global.length < 1) return;
    return (
      <ReactSortable
        list={groupByState.global}
        setList={(listItems) => {
          dispatch(setGroupByActionList(listItems));
        }}
        style={{ marginLeft: '-20px' }}
      >
        {groupByState.global.map((opt, index) => (
          <div
            key={index}
            className={`flex relative items-center mt-2 ${styles.draghandleparent}`}
          >
            <div className='flex relative'>
              <div
                className={styles.draghandle}
                style={{ cursor: 'pointer', margin: 'auto 2px' }}
              >
                <SVG name='drag' />
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
      {/* <div className={`${styles.group_block__event} flex justify-start items-center`}>
        <div className={'fa--query_block--add-event inactive flex justify-center items-center mr-2'}><SVG name={'groupby'} size={24} color={'purple'}/></div>
        <Text type={'title'} level={6} weight={'thin'} extraClass={'m-0'}>Breakdown</Text>
      </div> */}

      {renderExistingBreakdowns()}
      {renderInitGroupSelect(groupByState.global.length)}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  userPropNames: state.coreQuery.userPropNames,
  eventPropNames: state.coreQuery.eventPropNames,
  groupByState: state.coreQuery.groupBy
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
