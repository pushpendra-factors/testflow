import React, { useState, useEffect } from 'react';
import { SVG, Text } from '../../factorsComponents';
import styles from './index.module.scss';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { connect } from 'react-redux';
import ProfileFilterWrapper from '../ProfileFilterWrapper';
import FaSelect from 'Components/FaSelect';
import {
  ProfileMapper,
  ReverseProfileMapper,
  profileOptions
} from '../../../utils/constants';
import AliasModal from '../../QueryComposer/AliasModal';
import { INITIALIZE_GROUPBY } from '../../../reducers/coreQuery/actions';
import { useDispatch } from 'react-redux';
import ORButton from '../../ORButton';
import { compareFilters, groupFilters } from '../../../utils/global';
import { TOOLTIP_CONSTANTS } from '../../../constants/tooltips.constans';

function ProfileBlock({
  index,
  event,
  eventChange,
  queries,
  activeProject,
  userProperties,
  groupProperties,
  groupAnalysis,
  queryOptions,
  setQueryOptions
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    user: []
  });
  const dispatch = useDispatch();

  const alphabetIndex = 'ABCDEF';

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

  /* need confirmation */
  const resetQueryOptions = () => {
    const opts = Object.assign({}, queryOptions);
    opts.globalFilters = [];
    dispatch({
      type: INITIALIZE_GROUPBY,
      payload: {
        global: [],
        event: []
      }
    });
    setQueryOptions(opts);
  };

  const onChange = (value) => {
    const newEvent = { alias: '', label: '', filters: [] };
    newEvent.label = ProfileMapper[value] || value;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
    // resetQueryOptions();
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return undefined;
    }
    const assignFilterProps = Object.assign({}, filterProps);
    if (groupAnalysis === 'users') {
      assignFilterProps['user'] = userProperties;
      assignFilterProps['group'] = [];
    } else {
      assignFilterProps['user'] = [];
      assignFilterProps['group'] = groupProperties[groupAnalysis];
    }
    setFilterProperties(assignFilterProps);
  }, [userProperties, groupProperties, groupAnalysis]);

  const deleteItem = () => {
    eventChange(event, index - 1, 'delete');
  };

  const triggerDropDown = () => {
    setDDVisible(true);
  };

  const selectProfile = () => {
    return (
      <div className={`${styles.query_block__event_selector}`}>
        {isDDVisible ? (
          <FaSelect
            options={profileOptions[groupAnalysis]}
            onClickOutside={() => setDDVisible(false)}
            optionClick={(val) => onChange(val)}
            extraClass={styles.faselect}
          ></FaSelect>
        ) : null}
      </div>
    );
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

  const selectEventFilter = (refValue) => {
    return (
      <ProfileFilterWrapper
        filterProps={filterProps}
        activeProject={activeProject}
        event={event}
        deleteFilter={closeFilter}
        insertFilter={insertFilters}
        closeFilter={closeFilter}
        refValue={refValue}
        groupName={groupAnalysis}
      ></ProfileFilterWrapper>
    );
  };

  const setAdditionalactions = (opt) => {
    if (opt[1] === 'filter') {
      addFilter();
    } else {
      showModal();
    }
    setMoreOptions(false);
  };
  const getMenu = (filterOptions) => (
    <Menu style={{ minWidth: '200px', padding: '10px' }}>
      {filterOptions.map((eachFilter, eachIndex) => {
        return (
          <Menu.Item
            key={eachIndex}
            icon={
              <SVG
                name={eachFilter[1]}
                extraClass={'self-center'}
                style={{ marginRight: '10px' }}
              ></SVG>
            }
            style={{ display: 'flex', padding: '10px', margin: '5px' }}
            onClick={() => setAdditionalactions(eachFilter)}
          >
            <span style={{ paddingLeft: '5px' }}>{eachFilter[0]}</span>
          </Menu.Item>
        );
      })}
    </Menu>
  );
  const additionalActions = () => {
    return (
      <div className={`fa--query_block--actions-cols flex`}>
        <div className={`relative`}>
          <Tooltip title='Filter this Profile' color={TOOLTIP_CONSTANTS.DARK}>
            <Button
              type='text'
              onClick={addFilter}
              className={`fa-btn--custom mr-1 btn-total-round`}
            >
              <SVG name='filter'></SVG>
            </Button>
          </Tooltip>

          {/* {moreOptions ? (
            <FaSelect
              options={[
                ['Filter By', 'filter'],
                [!event?.alias?.length ? 'Create Alias' : 'Edit Alias', 'edit']
              ]}
              optionClick={(val) => setAdditionalactions(val)}
              onClickOutside={() => setMoreOptions(false)}
            ></FaSelect>
          ) : (
            false
          )} */}
          <AliasModal
            visible={isModalVisible}
            event={ReverseProfileMapper[event.label][groupAnalysis]}
            onOk={handleOk}
            onCancel={handleCancel}
            alias={event.alias}
          ></AliasModal>
        </div>
        <Tooltip title='Delete this Profile' color={TOOLTIP_CONSTANTS.DARK}>
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
                <ProfileFilterWrapper
                  index={index}
                  filter={filter}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                ></ProfileFilterWrapper>
              </div>
              {index !== orFilterIndex && (
                <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
              )}
              {index === orFilterIndex && (
                <div key={'init'}>
                  <ProfileFilterWrapper
                    filterProps={filterProps}
                    activeProject={activeProject}
                    event={event}
                    deleteFilter={closeFilter}
                    insertFilter={insertFilters}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr={true}
                  ></ProfileFilterWrapper>
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filters.push(
            <div className={'fa--query_block--filters flex flex-row'}>
              <div key={index}>
                <ProfileFilterWrapper
                  index={index}
                  filter={filtersGr[0]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                ></ProfileFilterWrapper>
              </div>
              <div key={index + 1}>
                <ProfileFilterWrapper
                  index={index + 1}
                  filter={filtersGr[1]}
                  event={event}
                  filterProps={filterProps}
                  activeProject={activeProject}
                  deleteFilter={removeFilters}
                  insertFilter={insertFilters}
                  closeFilter={closeFilter}
                  refValue={refValue}
                  showOr={true}
                ></ProfileFilterWrapper>
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
              {'Add New'}
            </Button>
          }
          {selectProfile()}
        </div>
      </div>
    );
  }
  let filterOptions = [
    ['Filter By', 'filter'],
    [!event?.alias?.length ? 'Create Alias' : 'Edit Alias', 'edit']
  ];
  return (
    <div
      className={`${styles.query_block} fa--query_block_section borderless no-padding mt-2`}
    >
      <div
        className={`${!event?.alias?.length ? 'flex justify-start' : ''} ${
          styles.query_block__event
        } block_section items-center`}
      >
        <div className={'flex items-center'}>
          <div
            className={
              'fa--query_block--add-event active flex justify-center items-center mr-2'
            }
          >
            <Text
              type={'title'}
              level={7}
              weight={'bold'}
              color={'white'}
              extraClass={'m-0'}
            >
              {alphabetIndex[index - 1]}
            </Text>
          </div>
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
        <div className={`flex ${!event?.alias?.length ? '' : 'ml-8 mt-1'}`}>
          <div className='relative ml-2'>
            <Tooltip title={ReverseProfileMapper[event.label][groupAnalysis]}>
              <Button
                icon={<SVG name='mouseevent' size={16} color={'purple'} />}
                className={`btn-total-round`}
                type={'link'}
                onClick={triggerDropDown}
              >
                {ReverseProfileMapper[event.label][groupAnalysis]}
              </Button>
              {selectProfile()}
            </Tooltip>
          </div>
          {filterOptions.length != 0 ? (
            <Dropdown
              placement='bottomLeft'
              overlay={getMenu(filterOptions)}
              trigger={['hover']}
            >
              <Button
                type='text'
                size={'large'}
                className={`fa-btn--custom mr-1 btn-total-round ml-2`}
              >
                <SVG name='more' />
              </Button>
            </Dropdown>
          ) : (
            ''
          )}
          <div className={styles.query_block__additional_actions}>
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
  groupProperties: state.coreQuery.groupProperties,
  eventNames: state.coreQuery.eventNames
});

export default connect(mapStateToProps)(ProfileBlock);
