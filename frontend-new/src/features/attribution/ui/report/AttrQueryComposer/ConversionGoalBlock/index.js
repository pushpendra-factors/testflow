import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { Button, Tooltip } from 'antd';
import { SVG, Text } from 'factorsComponents';
import { isArray } from 'lodash';
import FaSelect from 'Components/FaSelect';
import ORButton from 'Components/ORButton';
import { getNormalizedKpi } from 'Utils/kpiQueryComposer.helpers';
import { compareFilters, groupFilters } from 'Utils/global';
import { fetchKPIConfigWithoutDerivedKPI } from 'Reducers/kpi';

import { TOOLTIP_CONSTANTS } from '../../../../../../constants/tooltips.constans';
import EventFilterWrapper from 'Components/KPIComposer/EventFilterWrapper';

function ConversionGoalBlock({
  eventGoal,
  eventGoalChange,
  delEvent,
  eventNameOptions,
  eventNames,
  activeProject,
  eventProperties,
  userProperties,
  group_analysis = 'users',
  KPI_config,
  KPI_config_without_derived_kpi,
  showDerivedKPI = false,
  fetchKPIConfigWithoutDerivedKPI,
  currentProjectSettings
}) {
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterBlockVisible, setFilterBlockVisible] = useState(false);

  const [moreOptions, setMoreOptions] = useState(false);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const [groupProps, setGroupProps] = useState();

  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
  });

  const attrGroupNameMap = {
    hs_kpi: { label: 'Hubspot Deals', value: 'hubspot_deals' },
    sf_kpi: {
      label: 'Salesforce Opportunities',
      value: 'salesforce_opportunities'
    },
    user_kpi: { label: 'Users', value: 'user_kpi' }
  };

  useEffect(() => {
    if (eventGoal) {
      setFilterPropsforKpiGroups();
    }
  }, [userProperties, eventProperties, group_analysis, eventGoal]);

  useEffect(() => {
    if (
      currentProjectSettings.attribution_config &&
      currentProjectSettings.attribution_config.kpis_to_attribute
    ) {
      const kpiList =
        currentProjectSettings.attribution_config.kpis_to_attribute;
      const groupedList = [];
      Object.keys(kpiList).forEach((grpName) => {
        const propertyObj = { label: '', icon: '', values: [] };
        propertyObj.label = attrGroupNameMap[grpName]?.label;
        propertyObj.icon = grpName;
        kpiList[grpName].forEach((item) => {
          propertyObj.values.push([
            item.label,
            item.value,
            item.category,
            item.group
          ]);
        });
        groupedList.push(propertyObj);
      });
      setGroupProps(groupedList);
    }
  }, [activeProject, showDerivedKPI, currentProjectSettings]);

  const setEventPropsForUserGroup = () => {
    if (!eventGoal || !eventGoal?.label?.length) {
      return;
    }
    const assignFilterProps = Object.assign({}, filterProps);

    if (eventProperties[eventGoal.label]) {
      assignFilterProps.event = eventProperties[eventGoal.label];
    }
    assignFilterProps.user = userProperties;
    setFilterProperties(assignFilterProps);
  };

  const setFilterPropsforKpiGroups = () => {
    const assignFilterProps = Object.assign({}, filterProps);
    assignFilterProps.event = getKPIProps(group_analysis);
    setFilterProperties(assignFilterProps);
  };

  const getKPIProps = () => {
    let KPIlist = KPI_config || [];
    let selGroup = KPIlist.find((item) => {
      return item?.display_category == eventGoal.group;
    });

    let DDvalues = selGroup?.properties?.map((item) => {
      if (item == null) return;
      let ddName = item.display_name ? item.display_name : item.name;
      let ddtype =
        selGroup?.category == 'channels'
          ? item.object_type
          : item.entity
          ? item.entity
          : item.object_type;
      return [ddName, item.name, item.data_type, ddtype];
    });
    return DDvalues;
  };

  const getKpiGroupList = (groupName) => {
    let KPIlist = showDerivedKPI
      ? KPI_config
      : KPI_config_without_derived_kpi || [];
    let selGroup = KPIlist.find((item) => {
      return item?.display_category == groupName;
    });

    const group = ((selGroup) => {
      return getNormalizedKpi({ kpi: selGroup });
    })(selGroup);
    return [group];
  };

  const toggleEventSelect = () => {
    setSelectVisible(!selectVisible);
  };

  const addFilter = (val) => {
    const updatedEvent = Object.assign({}, eventGoal);

    const filtersSorted = updatedEvent.filters;
    filtersSorted.sort(compareFilters);
    const filt = filtersSorted.filter(
      (fil) => JSON.stringify(fil) === JSON.stringify(val)
    );
    if (filt && filt.length) return;

    updatedEvent.filters.push(val);
    eventGoalChange(updatedEvent);
  };

  const editFiler = (index, val) => {
    let updatedEvent = Object.assign({}, eventGoal);
    const filt = Object.assign({}, val);
    filt.operator = isArray(val.operator) ? val.operator[0] : val.operator;
    const filtersSorted = updatedEvent.filters;
    filtersSorted.sort(compareFilters);
    filtersSorted[index] = filt;
    updatedEvent.filters = filtersSorted;
    eventGoalChange(updatedEvent);
  };

  const delFilter = (val) => {
    const updatedEvent = Object.assign({}, eventGoal);
    const filtersSorted = updatedEvent.filters;
    filtersSorted.sort(compareFilters);
    const filt = filtersSorted.filter((v, i) => i !== val);
    updatedEvent.filters = filt;
    eventGoalChange(updatedEvent);
  };

  const closeFilter = () => {
    setFilterBlockVisible(false);
    setOrFilterIndex(-1);
  };

  const deleteItem = () => {
    delEvent();
    closeFilter();
  };

  const addFilterBlock = () => {
    setFilterBlockVisible(true);
  };

  const selectEventFilter = (index) => {
    return (
      <FilterWrapper
        filterProps={filterProps}
        projectID={activeProject?.id}
        event={eventGoal}
        deleteFilter={() => closeFilter()}
        insertFilter={addFilter}
        closeFilter={closeFilter}
        refValue={index}
      />
    );
  };

  const renderFilterWrapper = (
    index,
    refValue,
    filter,
    showOr,
    inFilter,
    deleteFilter
  ) =>
    (
      <EventFilterWrapper
        index={index}
        filter={filter}
        event={eventGoal}
        filterProps={filterProps}
        activeProject={activeProject}
        deleteFilter={deleteFilter}
        insertFilter={inFilter}
        closeFilter={closeFilter}
        selectedMainCategory={eventGoal}
        showOr={showOr}
        refValue={refValue}
      />
    );

  const eventFilters = () => {
    const filters = [];
    let index = 0;
    let lastRef = 0;
    if (eventGoal && eventGoal?.filters?.length) {
      const group = groupFilters(eventGoal.filters, 'ref');
      const filtersGroupedByRef = Object.values(group);
      const refValues = Object.keys(group);
      lastRef = parseInt(refValues[refValues.length - 1]);

      filtersGroupedByRef.forEach((filtersGr) => {
        const refValue = filtersGr[0].ref;
        if (filtersGr.length == 1) {
          const filter = filtersGr[0];
          let filterContent = filter;
          filterContent.values =
            filter.props[1] === 'datetime' && isArray(filter.values)
              ? filter.values[0]
              : filter.values;
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
              {renderFilterWrapper(
                  index,
                  refValue,
                  filter,
                  false,
                  (val, index) => editFiler(index, val),
                  delFilter
                )}
              </div>
              {index !== orFilterIndex && (
                <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
              )}
              {index === orFilterIndex && (
                <div key={'init'}>
                  {renderFilterWrapper(
                    undefined,
                    refValue,
                    undefined,
                    true,
                    addFilter,
                    closeFilter
                  )}
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
              {renderFilterWrapper(
                  index,
                  refValue,
                  filtersGr[0],
                  false,
                  (val, index) => editFiler(index, val),
                  delFilter
                )}
              </div>
              <div key={index + 1}>
              {renderFilterWrapper(
                  index + 1,
                  refValue,
                  filtersGr[1],
                  true,
                  (val, index) => editFiler(index, val),
                  delFilter
                )}
              </div>
            </div>
          );
          index += 2;
        }
      });
    }

    if (filterBlockVisible) {
      filters.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectEventFilter(lastRef + 1)}
        </div>
      );
    }

    return filters;
  };

  const onEventSelect = (val, group, category) => {
    const currentEventGoal = Object.assign({}, eventGoal);
    currentEventGoal.label = val[1] ? val[1] : val[0];
    currentEventGoal.filters = [];
    if (group_analysis !== 'users') {
      currentEventGoal.label = val[0];
      currentEventGoal.metric = val[1] ? val[1] : val[0];
      const grp = Object.keys(attrGroupNameMap).find(
        (val) => attrGroupNameMap[val].label === group
      );
      currentEventGoal.group = attrGroupNameMap[grp].value;
      if (val[2]) {
        currentEventGoal.category = val[2];
      }
      if (val[3]) {
        currentEventGoal.group = val[3];
      }
    }
    eventGoalChange(currentEventGoal);
    setSelectVisible(false);
    closeFilter();
  };

  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions-cols flex relative ml-2'}>
        <div className={`relative flex`}>
          <Tooltip title='Filter this Attribute' color={TOOLTIP_CONSTANTS.DARK}>
            <Button
              type='text'
              onClick={() => setMoreOptions(true)}
              className={`fa-btn--custom mr-1 btn-total-round`}
            >
              <SVG name='more'></SVG>
            </Button>
          </Tooltip>

          {moreOptions ? (
            <FaSelect
              options={[[`Filter By`, 'filter']]}
              optionClick={(val) => {
                addFilterBlock();
                setMoreOptions(false);
              }}
              onClickOutside={() => setMoreOptions(false)}
              showIcon
            ></FaSelect>
          ) : (
            false
          )}
        </div>
        <Tooltip title='Delete this Attribute'>
          <Button
            type='text'
            onClick={deleteItem}
            className={`fa-btn--custom btn-total-round`}
          >
            <SVG name='trash'></SVG>
          </Button>
        </Tooltip>
      </div>
    );
  };

  const renderCountLabel = () => {
    return (
      <Text
        type={'title'}
        level={7}
        weight={'regular'}
        color={'grey'}
        extraClass={'m-0 ml-2'}
      >
        as count of unique users
      </Text>
    );
  };

  const getKpiGroupListAll = () => {
    const filterList = [
      'hubspot_deals',
      'salesforce_opportunities',
      'form_submission'
    ];
    let KPIlist = showDerivedKPI
      ? KPI_config
      : KPI_config_without_derived_kpi || [];
    let selGroup = KPIlist.find((item) => {
      return filterList.includes(item?.display_category);
    });

    const group = ((selGroup) => {
      return getNormalizedKpi({ kpi: selGroup });
    })(selGroup);
    return [group];
  };

  const getGroupedProps = () => {
    if (!group_analysis || group_analysis === 'users') return eventNameOptions;
    if (group_analysis === 'all') {
      return getKpiGroupListAll();
    } else {
      return getKpiGroupList(group_analysis);
    }
  };

  const selectEvents = () => {
    const groupedProps = getGroupedProps();
    return (
      <div className={styles.block__event_selector}>
        {selectVisible ? (
          <GroupSelect2
            groupedProperties={groupProps}
            placeholder='Select Event'
            optionClick={(group, val, category) =>
              onEventSelect(val, group, category)
            }
            onClickOutside={() => setSelectVisible(false)}
            useCollapseView
          ></GroupSelect2>
        ) : null}
      </div>
    );
  };

  const renderGoalBlockContent = () => {
    return (
      <div
        className={`${styles.block__content} flex items-center relative mt-4`}
      >
        {
          <Tooltip
            title={
              eventNames[eventGoal?.label]
                ? eventNames[eventGoal?.label]
                : eventGoal?.label
            }
          >
            <Button
              type='link'
              onClick={toggleEventSelect}
              icon={<SVG name='mouseevent' />}
              className={`fa-button--truncate fa-button--truncate-lg btn-total-round`}
            >
              {eventNames[eventGoal?.label]
                ? eventNames[eventGoal?.label]
                : eventGoal?.label}
            </Button>
          </Tooltip>
        }

        {selectEvents()}

        {(!group_analysis || group_analysis === 'users') && renderCountLabel()}

        <div className={styles.block__additional_actions}>
          {additionalActions()}
        </div>
      </div>
    );
  };

  const renderGoalSelect = () => {
    return (
      <div className={'flex justify-start items-center mt-4'}>
        {
          <Button
            type='text'
            onClick={toggleEventSelect}
            icon={<SVG name={'plus'} color={'grey'} />}
          >
            Add KPI
          </Button>
        }
        {selectEvents()}
      </div>
    );
  };

  return (
    <div className={`${styles.block} fa--query_block_section--basic relative`}>
      {eventGoal?.label?.length ? renderGoalBlockContent() : renderGoalSelect()}
      {eventFilters()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  eventProperties: state.coreQuery.eventProperties,
  userProperties: state.coreQuery.userProperties,
  eventNameOptions: state.coreQuery.eventOptions,
  eventNames: state.coreQuery.eventNames,
  KPI_config: state.kpi?.config,
  KPI_config_without_derived_kpi: state.kpi?.config_without_derived_kpi
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators({ fetchKPIConfigWithoutDerivedKPI }, dispatch);

export default connect(
  mapStateToProps,
  mapDispatchToProps
)(ConversionGoalBlock);
