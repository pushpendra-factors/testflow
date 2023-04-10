/* eslint-disable */
import React, { useState, useEffect } from 'react';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { setGroupBy, delGroupBy } from 'Reducers/coreQuery/middleware';

import FilterBlock from 'Components/KPIComposer/FilterBlock';
import EventFilterWrapper from 'Components/KPIComposer/EventFilterWrapper';

import GroupSelect2 from 'Components/KPIComposer/GroupSelect2';
import EventGroupBlock from 'Components/KPIComposer/EventGroupBlock';
import { QUERY_TYPE_FUNNEL } from 'Utils/constants';

import FaSelect from 'Components/FaSelect';
import AliasModal from 'Components/KPIComposer/AliasModal';
import ORButton from 'Components/ORButton';
import { getNormalizedKpi } from '../utils';
import { get } from 'lodash';
import { compareFilters, groupFilters } from 'Utils/global';

import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';

function QueryBlock({
  index,
  event,
  eventChange,
  queries,
  queryType,
  // eventOptions,
  eventNames,
  activeProject,
  groupBy,
  setGroupBy,
  delGroupBy,
  // userProperties,
  // eventProperties,
  // setSelectedMainCategory,
  kpi,
  KPIConfigProps,
  selectedMainCategory
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState(false);
  const [urlList, setUrlList] = useState(false);
  const [pageUrlDD, setPageUrlDD] = useState(false);

  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  const [isModalVisible, setIsModalVisible] = useState(false);

  const showModal = () => {
    setIsModalVisible(true);
  };

  const handleCancel = () => {
    setIsModalVisible(false);
  };
  const handleOk = (alias) => {
    const newEvent = Object.assign({}, event);
    newEvent.alias = alias;
    setIsModalVisible(false);
    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const alphabetIndex = 'ABCDEFGHIJK';

  const setPageURL = (value) => {
    const newEvent = Object.assign({}, event);
    newEvent.pageViewVal = value[1] ? value[1] : value[0];
    eventChange(newEvent, index - 1, 'filters_updated');
    setPageUrlDD(false);
  };

  const onChange = (value, group, category, type) => {
    let qt;
    for (let item of kpi?.config) {
      for (let it of item.metrics) {
        if (it?.name === value[1]) qt = it?.kpi_query_type;
      }
    }
    const newEvent = { alias: '', label: '', filters: [], group: '' };
    newEvent.label = value[0];
    newEvent.name = value[1];
    newEvent.objType = value[2];
    newEvent.entity = value[4];
    newEvent.group = group;
    newEvent.qt = value[3];
    if (category) {
      newEvent.category = category;
    }
    setDDVisible(false);
    if (group === 'page_views') {
      eventChange(newEvent, index - 1, 'add', 'select_url');
    } else {
      eventChange(newEvent, index - 1, 'add');
    }
  };

  // useEffect(() => {
  //   if (!event || event === undefined) {
  //     return undefined;
  //   } // Akhil please check this line
  //   const assignFilterProps = Object.assign({}, filterProps);

  //   if (eventProperties[event.label]) {
  //     assignFilterProps.event = eventProperties[event.label];
  //   }
  //   assignFilterProps.user = userProperties;
  //   setFilterProperties(assignFilterProps);
  // }, [userProperties, eventProperties]);

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const deleteItem = () => {
    eventChange(event, index - 1, 'delete');
  };

  const kpiEvents = kpi?.config?.map((item) => {
    return getNormalizedKpi({ kpi: item });
  }); 

  const selectEvents = () => {
    return (
      <>
        {isDDVisible ? (
          <div className={styles.query_block__event_selector}>
            <GroupSelect2
              groupedProperties={kpiEvents ? kpiEvents : []}
              placeholder='Select Event'
              optionClick={(group, val, category) =>
                onChange(val, group, category)
              }
              onClickOutside={() => setDDVisible(false)}
              allowEmpty={true}
            />
          </div>
        ) : null}
      </>
    );
  };

  const selectPageUrls = () => {
    let KPI_PageUrls = kpi?.page_urls?.event_names;
    let pageURLs = KPI_PageUrls ? KPI_PageUrls.map((item) => [item, item]) : [];
    return (
      <>
        {
          <div className={'flex items-center'}>
            <Text type={'title'} level={8} extraClass={'m-0 mx-2'}>
              {'from'}
            </Text>
            <div className={'relative'}>
              <Button onClick={() => setPageUrlDD(true)}>
                {event?.pageViewVal
                  ? event?.pageViewVal == 'select_url'
                    ? 'Select URL'
                    : event?.pageViewVal
                  : 'Select URL'}
              </Button>
              {pageUrlDD ? (
                <FaSelect
                  options={pageURLs || []}
                  placeholder='Select Event'
                  optionClick={(val) => setPageURL(val)}
                  onClickOutside={() => setPageUrlDD(false)}
                  allowSearch={true}
                />
              ) : null}
            </div>
          </div>
        }
      </>
    );
  };

  const addGroupBy = () => {
    setGroupByDDVisible(true);
  };

  const addFilter = () => {
    setFilterDDVisible(true);
  };

  const insertFilters = (filter, filterIndex) => {
    const newEvent = Object.assign({}, event);
    const filtersSorted = newEvent.filters;
    filtersSorted.sort(compareFilters);
    if (filterIndex >= 0) {
      newEvent.filters = filtersSorted.map((filt, i) => {
        if (i === filterIndex) {
          return filter;
        }
        return filt;
      });
    } else {
      newEvent.filters.push(filter);
    }

    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const removeFilters = (i) => {
    const newEvent = Object.assign({}, event);
    const filtersSorted = newEvent.filters;
    filtersSorted.sort(compareFilters);
    if (filtersSorted[i]) {
      filtersSorted.splice(i, 1);
      newEvent.filters = filtersSorted;
    }
    eventChange(newEvent, index - 1, 'filters_updated');
  };
  const closeFilter = () => {
    setFilterDDVisible(false);
    setOrFilterIndex(-1);
  };

  const selectEventFilter = (index) => {
    return (
      <EventFilterWrapper
        filterProps={filterProps}
        activeProject={activeProject}
        event={event}
        deleteFilter={closeFilter}
        insertFilter={insertFilters}
        closeFilter={closeFilter}
        selectedMainCategory={selectedMainCategory}
        refValue={index}
      />
    );
  };

  const deleteGroupBy = (groupState, id, type = 'event') => {
    delGroupBy(type, groupState, id);
  };

  const pushGroupBy = (groupState, index) => {
    let ind;
    index >= 0 ? (ind = index) : (ind = groupBy.length);
    setGroupBy('event', groupState, ind);
  };

  const selectGroupByEvent = () => {
    if (isGroupByDDVisible) {
      return (
        <EventGroupBlock
          eventIndex={index}
          event={event}
          setGroupState={pushGroupBy}
          closeDropDown={() => setGroupByDDVisible(false)}
          KPIConfigProps={KPIConfigProps}
        ></EventGroupBlock>
      );
    }
  };
 

  const additionalActions = () => {
    // Kept Filter by only, as it was previously, just changed Filter Menu

    return (
      <div className={'flex'} style={{'margin-left':'1px'}}>
        <Tooltip title='Delete this KPI' color={TOOLTIP_CONSTANTS.DARK}>
          <Button  
            onClick={deleteItem}
            className={`btn-right-round`}
          >
            <SVG name='close' size={18} color={'grey'} />
          </Button>
        </Tooltip>
      </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    let index = 0;
    let lastRef = 0;
    if (event && event?.filters?.length) {
      const group = groupFilters(event.filters, 'ref');
      const filtersGroupedByRef = Object.values(group);
      const refValues = Object.keys(group);
      lastRef = parseInt(refValues[refValues.length - 1]);

      filtersGroupedByRef.forEach((filtersGr) => {
        const refValue = filtersGr[0].ref;
        if (filtersGr.length == 1) {
          const filter = filtersGr[0];
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                <EventFilterWrapper
                  index={index}
                  filter={filter}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  selectedMainCategory={selectedMainCategory}
                  refValue={refValue}
                />
              </div>
              {index !== orFilterIndex && (
                <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
              )}
              {index === orFilterIndex && (
                <div key={'init'}>
                  <EventFilterWrapper
                    filterProps={filterProps}
                    activeProject={activeProject}
                    event={event}
                    deleteFilter={closeFilter}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    selectedMainCategory={selectedMainCategory}
                    refValue={refValue}
                    showOr={true}
                  />
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                <EventFilterWrapper
                  index={index}
                  filter={filtersGr[0]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  selectedMainCategory={selectedMainCategory}
                  refValue={refValue}
                />
              </div>
              <div key={index + 1}>
                <EventFilterWrapper
                  index={index + 1}
                  filter={filtersGr[1]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  selectedMainCategory={selectedMainCategory}
                  refValue={refValue}
                  showOr={true}
                />
              </div>
            </div>
          );
          index += 2;
        }
      });
    }

    if (isFilterDDVisible) {
      filters.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectEventFilter(lastRef + 1)}
        </div>
      );
    }

    return filters;
  };

  const groupByItems = () => {
    const groupByEvents = [];
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
      groupBy
        .map((gbp, index) => {
          return { ...gbp, groupByIndex: index };
        })
        .filter((gbp) => {
          return gbp.eventName === event.label && gbp.eventIndex === index;
        })
        .forEach((gbp, gbpIndex) => {
          const { groupByIndex, ...orgGbp } = gbp;
          groupByEvents.push(
            <div key={gbpIndex} className={'fa--query_block--filters'}>
              <EventGroupBlock
                index={gbp.groupByIndex}
                grpIndex={gbpIndex}
                eventIndex={index}
                groupByEvent={orgGbp}
                event={event}
                delGroupState={(ev) => deleteGroupBy(ev, gbpIndex)}
                setGroupState={pushGroupBy}
                closeDropDown={() => setGroupByDDVisible(false)}
                selectedMainCategory={selectedMainCategory}
              />
            </div>
          );
        });
    }

    if (isGroupByDDVisible) {
      groupByEvents.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectGroupByEvent()}
        </div>
      );
    }

    return groupByEvents;
  };

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
          {
            <Button
              type='text'
              onClick={triggerDropDown}
              icon={<SVG name={'plus'} color={'grey'} />}
            >
              {ifQueries ? 'Add New' : 'Add new'}
            </Button>
          }
          {selectEvents()}
        </div>
      </div>
    );
  }

  let KPIFilterOptions = [];
  return (
    <div
      className={`${styles.query_block} borderless no-padding`}
    > 
      <div
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${
          styles.query_block__event
        }  items-center`}
      > 
      { index!=1 && <SVG name={'DoubeEndedArrow'} color={'grey'} size={24} extraClass={'mr-1 ml-1'} /> }
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-2'}`}>
          <div className='max-w-7xl'>
            <Tooltip
              title={
                eventNames[event.label] ? eventNames[event.label] : event.label
              }
              color={TOOLTIP_CONSTANTS.DARK}
            >
              <Button
                // icon={<SVG name='mouseevent' size={16} color={'purple'} />}
                className={`fa-button--truncate fa-button--truncate-lg btn-left-round`}
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

          <div>
            {additionalActions()}
          </div>
        </div>
      </div>
      {eventFilters()} 
    </div>
  );
}

const mapStateToProps = (state) => ({
  eventOptions: state.coreQuery.eventOptions,
  activeProject: state.global.active_project,
  userProperties: state.coreQuery.userProperties,
  eventProperties: state.coreQuery.eventProperties,
  groupBy: state.coreQuery.groupBy.event,
  eventNames: state.coreQuery.eventNames,
  kpi: state.kpi
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
