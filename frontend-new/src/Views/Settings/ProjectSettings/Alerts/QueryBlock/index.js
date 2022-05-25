/* eslint-disable */
import React, { useState, useEffect } from 'react';
import { SVG, Text } from 'factorsComponents';
import styles from './index.module.scss';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { setGroupBy, delGroupBy } from '../../../../../reducers/coreQuery/middleware';

import FaSelect from 'Components/FaSelect';
import GroupSelect2 from '../../../../../components/KPIComposer/GroupSelect2';
import EventFilterWrapper from '../../../../../components/KPIComposer/EventFilterWrapper';
import EventGroupBlock from '../../../../../components/KPIComposer/EventGroupBlock';
import { QUERY_TYPE_FUNNEL } from '../../../../../utils/constants';
import AliasModal from '../../../../../components/KPIComposer/AliasModal';
import { getNormalizedKpi } from '../../../../../utils/kpiQueryComposer.helpers';

function QueryBlock({
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
  userProperties,
  eventProperties,
  setSelectedMainCategory,
  kpi,
  KPIConfigProps,
  selectedMainCategory,
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [isGroupByDDVisible, setGroupByDDVisible] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState(false);
  const [urlList, setUrlList] = useState(false);
  const [pageUrlDD, setPageUrlDD] = useState(false);

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

  const setPageURL = (value) => {
    const newEvent = Object.assign({}, event);
    newEvent.pageViewVal = value[1] ? value[1] : value[0];
    eventChange(newEvent, index - 1, 'filters_updated');
    setPageUrlDD(false);
  };

  const onChange = (value, group, category) => {
    const newEvent = { alias: '', label: '', filters: [], group: '' };
    newEvent.label = value[0];
    newEvent.metric = value[1];
    newEvent.group = group;
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
    return getNormalizedKpi({kpi:item})
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
    if (filterIndex >= 0) {
      newEvent.filters = newEvent.filters.map((filt, i) => {
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
    if (newEvent.filters[i]) {
      newEvent.filters.splice(i, 1);
    }
    eventChange(newEvent, index - 1, 'filters_updated');
  };

  const selectEventFilter = () => {
    return (
      <EventFilterWrapper
        filterProps={filterProps}
        activeProject={activeProject}
        event={event}
        deleteFilter={() => setFilterDDVisible(false)}
        insertFilter={insertFilters}
        closeFilter={() => setFilterDDVisible(false)}
        selectedMainCategory={selectedMainCategory}
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

  const setAdditionalactions = (opt) => {
    if (opt[1] === 'filter') {
      addFilter();
    } else if (opt[1] === 'groupby') {
      addGroupBy();
    } else {
      showModal();
    }
    setMoreOptions(false);
  };

  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions-cols flex'}>
        <div className={`relative`}>
          <Button
            type='text'
            onClick={() => setAdditionalactions(['Filter By', 'filter'])}
            className={`-ml-2`}
          >
            Add filter
          </Button>

          <AliasModal
            visible={isModalVisible}
            event={
              eventNames[event.label] ? eventNames[event.label] : event.label
            }
            onOk={handleOk}
            onCancel={handleCancel}
            alias={event.alias}
          ></AliasModal>
        </div>
      </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    if (event && event?.filters?.length) {
      event.filters.forEach((filter, index) => {
        filters.push(
          <div key={index} className={'fa--query_block--filters'}>
            <EventFilterWrapper
              index={index}
              filter={filter}
              event={event}
              filterProps={filterProps}
              activeProject={activeProject}
              deleteFilter={removeFilters}
              insertFilter={insertFilters}
              closeFilter={() => setFilterDDVisible(false)}
              selectedMainCategory={selectedMainCategory}
            />
          </div>
        );
      });
    }

    if (isFilterDDVisible) {
      filters.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectEventFilter()}
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
        className={`${styles.query_block} fa--query_block my-1 ${
          ifQueries ? 'borderless no-padding' : 'borderless no-padding'
        }`}
      >
        <div
          className={`${styles.query_block__event} flex justify-start items-center`}
        >
          {
            <Button
              type='link'
              onClick={triggerDropDown}
            >
              {ifQueries ? 'Select another KPI' : 'Select KPI'}
            </Button>
          }
          {selectEvents()}
        </div>
      </div>
    );
  }

  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding mt-0`}
    >
      <div
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${
          styles.query_block__event
        } block_section items-center`}
      >
        <div className={'flex items-center'}>
          {event?.alias?.length ? (
            <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>
              {event?.alias}
              <Tooltip title={'Edit Alias'}>
                <Button
                  className={`${styles.custombtn} mx-1`}
                  type='text'
                  onClick={showModal}
                >
                  <SVG size={20} name='edit' color={'grey'} />
                </Button>
              </Tooltip>
            </Text>
          ) : null}
        </div>
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-0'}`}>
          <div className='max-w-7xl'>
            <Tooltip
              title={
                eventNames[event.label] ? eventNames[event.label] : event.label
              }
            >
              <Button
                // icon={<SVG name='mouseevent' size={16} color={'purple'} />}
                className={`fa-button--truncate fa-button--truncate-lg`}
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
          {(event?.pageViewVal || event?.group == 'page_views') &&
            selectPageUrls()}
        </div>
      </div>
      <div className={'mt-4 mb-1'}>
        <Text type={'title'} level={7} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>FILTER BY</Text>
      </div>
      {eventFilters()}
      <div className={'mt-2'}>
        {additionalActions()}
      </div>
      {/* {groupByItems()}  */}
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
  kpi: state.kpi,
});

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      setGroupBy,
      delGroupBy,
    },
    dispatch
  );

export default connect(mapStateToProps, mapDispatchToProps)(QueryBlock);
