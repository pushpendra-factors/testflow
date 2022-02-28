import React, { useState, useEffect } from 'react';
import { SVG, Text } from '../../factorsComponents';
import styles from './index.module.scss';
import { Button, Tooltip } from 'antd';
import { connect } from 'react-redux';
import ProfileFilterWrapper from '../ProfileFilterWrapper';
import FaSelect from 'Components/FaSelect';
import { ProfileMapper, ReverseProfileMapper } from '../../../utils/constants';
import AliasModal from '../../QueryComposer/AliasModal';

function ProfileBlock({
  index,
  event,
  eventChange,
  queries,
  activeProject,
  userProperties,
  groupAnalysis,
}) {
  const [isDDVisible, setDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [moreOptions, setMoreOptions] = useState(false);
  const [filterProps, setFilterProperties] = useState({
    user: [],
  });

  const alphabetIndex = 'ABCDEF';

  const profileOptions = {
    users: [['Website Visitors'], ['Hubspot Contacts'], ['Salesforce Users']],
    $salesforce_opportunity: [['All Opportunities']],
    $hubspot_deal: [['All Deals']],
    $salesforce_account: [['All Accounts']],
    $hubspot_company: [['All Companies']],
  };

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

  const onChange = (value) => {
    const newEvent = { alias: '', label: '', filters: [] };
    newEvent.label = ProfileMapper[value] ? ProfileMapper[value] : value;
    setDDVisible(false);
    eventChange(newEvent, index - 1);
  };

  useEffect(() => {
    if (!event || event === undefined) {
      return undefined;
    }
    const assignFilterProps = Object.assign({}, filterProps);
    assignFilterProps.user = userProperties;
    setFilterProperties(assignFilterProps);
  }, [userProperties]);

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
      <ProfileFilterWrapper
        filterProps={filterProps}
        activeProject={activeProject}
        event={event}
        deleteFilter={() => setFilterDDVisible(false)}
        insertFilter={insertFilters}
        closeFilter={() => setFilterDDVisible(false)}
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

  const additionalActions = () => {
    return (
      <div className={`fa--query_block--actions-cols flex`}>
        <div className={`relative`}>
          <Button
            type='text'
            onClick={() => setMoreOptions(true)}
            className={`fa-btn--custom ml-1 mr-1`}
          >
            <SVG name='more'></SVG>
          </Button>

          {moreOptions ? (
            <FaSelect
              options={[
                ['Filter By', 'filter'],
                [!event?.alias?.length ? 'Create Alias' : 'Edit Alias', 'edit'],
              ]}
              optionClick={(val) => setAdditionalactions(val)}
              onClickOutside={() => setMoreOptions(false)}
            ></FaSelect>
          ) : (
            false
          )}
          <AliasModal
            visible={isModalVisible}
            event={ReverseProfileMapper[event.label][groupAnalysis]}
            onOk={handleOk}
            onCancel={handleCancel}
            alias={event.alias}
          ></AliasModal>
        </div>
        <Button type='text' onClick={deleteItem} className={`fa-btn--custom`}>
          <SVG name='trash'></SVG>
        </Button>
      </div>
    );
  };

  const eventFilters = () => {
    const filters = [];
    if (event && event?.filters?.length) {
      event.filters.forEach((filter, index) => {
        filters.push(
          <div key={index} className={'fa--query_block--filters'}>
            <ProfileFilterWrapper
              index={index}
              filter={filter}
              event={event}
              filterProps={filterProps}
              activeProject={activeProject}
              deleteFilter={removeFilters}
              insertFilter={insertFilters}
              closeFilter={() => setFilterDDVisible(false)}
            ></ProfileFilterWrapper>
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
            </Text>{' '}
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
          <div className='max-w-7xl'>
            <Tooltip title={ReverseProfileMapper[event.label][groupAnalysis]}>
              <Button
                icon={<SVG name='mouseevent' size={16} color={'purple'} />}
                className={``}
                type={'link'}
                onClick={triggerDropDown}
              >
                {ReverseProfileMapper[event.label][groupAnalysis]}
              </Button>
              {selectProfile()}
            </Tooltip>
          </div>
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
  eventNames: state.coreQuery.eventNames,
});

export default connect(mapStateToProps)(ProfileBlock);
