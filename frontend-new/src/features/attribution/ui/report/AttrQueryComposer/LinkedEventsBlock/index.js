import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import GroupSelect2 from 'Components/QueryComposer/GroupSelect2';
import { Button, Tooltip } from 'antd';
import { SVG } from 'factorsComponents';
import { isArray } from 'lodash';
import FaSelect from 'Components/FaSelect';
import styles from './index.module.scss';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';

function LinkedEventsBlock({
  linkEvent,
  linkEventChange,
  delLinkEvent,
  eventNameOptions,
  activeProject,
  eventProperties,
  eventNames,
  userProperties
}) {
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterBlockVisible, setFilterBlockVisible] = useState(false);

  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
  });

  const [moreOptions, setMoreOptions] = useState(false);

  useEffect(() => {
    if (!linkEvent || !linkEvent?.label?.length) {
      return;
    }
    const assignFilterProps = Object.assign({}, filterProps);

    if (eventProperties[linkEvent.label]) {
      assignFilterProps.event = eventProperties[linkEvent.label];
    }
    assignFilterProps.user = userProperties;
    setFilterProperties(assignFilterProps);
  }, [userProperties, eventProperties, linkEvent]);

  const toggleEventSelect = () => {
    setSelectVisible(!selectVisible);
  };

  const addFilter = (val) => {
    const updatedEvent = Object.assign({}, linkEvent);
    const filt = updatedEvent.filters.filter(
      (fil) => JSON.stringify(fil) === JSON.stringify(val)
    );
    if (filt && filt.length) return;
    updatedEvent.filters.push(val);
    linkEventChange(updatedEvent);
  };

  const editFilter = (index, val) => {
    const updatedEvent = Object.assign({}, linkEvent);
    // const filt = updatedEvent.filters.filter(fil => JSON.stringify(fil) === JSON.stringify(val));
    updatedEvent.filters[index] = val;
    linkEventChange(updatedEvent);
  };

  const delFilter = (val) => {
    const updatedEvent = Object.assign({}, linkEvent);
    const filt = updatedEvent.filters.filter((v, i) => i !== val);
    updatedEvent.filters = filt;
    linkEventChange(updatedEvent);
  };

  const closeFilter = () => {
    setFilterBlockVisible(false);
  };

  const deleteItem = () => {
    delLinkEvent();
    closeFilter();
  };

  const addFilterBlock = () => {
    setFilterBlockVisible(true);
  };

  const selectEventFilter = () => {
    return (
      <FilterWrapper
        filterProps={filterProps}
        activeProject={activeProject}
        event={linkEvent}
        deleteFilter={() => closeFilter()}
        insertFilter={addFilter}
        closeFilter={closeFilter}
      />
    );
  };

  const eventFilters = () => {
    const filters = [];
    if (linkEvent && linkEvent?.filters?.length) {
      linkEvent.filters.forEach((filter, index) => {
        let filterContent = filter;
        filterContent.values =
          filter.props[1] === 'datetime' && isArray(filter.values)
            ? filter.values[0]
            : filter.values;
        filters.push(
          <div key={index} className={'fa--query_block--filters'}>
            <FilterWrapper
              index={index}
              filter={filterContent}
              filterProps={filterProps}
              activeProject={activeProject}
              event={linkEvent}
              deleteFilter={delFilter}
              insertFilter={(val) => editFilter(index, val)}
              closeFilter={closeFilter}
            />
          </div>
        );
      });
    }

    if (filterBlockVisible) {
      filters.push(
        <div key={'init'} className={'fa--query_block--filters'}>
          {selectEventFilter()}
        </div>
      );
    }

    return filters;
  };

  const onEventSelect = (val) => {
    const currentLinkEvent = Object.assign({}, linkEvent);
    currentLinkEvent.label = val;
    currentLinkEvent.filters = [];
    linkEventChange(currentLinkEvent);
    setSelectVisible(false);
  };

  const additionalActions = () => {
    return (
      <div className={'fa--query_block--actions--cols flex relative ml-2'}>
        <div className={`relative flex`}>
          <Button
            type='text'
            onClick={() => setMoreOptions(true)}
            className={'fa-btn--custom mr-1'}
          >
            <SVG name='more'></SVG>
          </Button>

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
        <Button className={'fa-btn--custom'} type='text' onClick={deleteItem}>
          <SVG name='trash'></SVG>
        </Button>
      </div>
    );
  };

  const selectEvents = () => {
    return (
      <div className={styles.block__event_selector}>
        {selectVisible ? (
          <div className={styles.block__event_selector__btn}>
            <GroupSelect2
              groupedProperties={eventNameOptions}
              placeholder='Select Event'
              optionClick={(group, val) =>
                onEventSelect(val[1] ? val[1] : val[0])
              }
              onClickOutside={() => setSelectVisible(false)}
              useCollapseView
            ></GroupSelect2>
          </div>
        ) : null}
      </div>
    );
  };

  const renderLinkEventBlockContent = () => {
    return (
      <div
        className={`${styles.block__content} fa--query_block_section--basic mt-4 relative`}
      >
        {
          <Tooltip
            title={
              eventNames[linkEvent?.label]
                ? eventNames[linkEvent?.label]
                : linkEvent?.label
            }
          >
            <Button
              type='link'
              className={`fa-button--truncate fa-button--truncate-lg`}
              onClick={toggleEventSelect}
            >
              <SVG name='mouseevent' extraClass={'mr-1'}></SVG>
              {eventNames[linkEvent?.label]
                ? eventNames[linkEvent?.label]
                : linkEvent?.label}
            </Button>
          </Tooltip>
        }

        {selectEvents()}

        <div className={styles.block__additional_actions}>
          {additionalActions()}
        </div>
      </div>
    );
  };

  const renderLinkEventSelect = () => {
    return (
      <div className={`${styles.block__content} mt-4`}>
        {
          <Button
            type='text'
            onClick={toggleEventSelect}
            icon={<SVG name={'plus'} color={'grey'} />}
          >
            Add new
          </Button>
        }
        {selectEvents()}
      </div>
    );
  };

  return (
    <div className={styles.block}>
      {linkEvent?.label?.length
        ? renderLinkEventBlockContent()
        : renderLinkEventSelect()}
      {eventFilters()}
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventProperties: state.coreQuery.eventProperties,
  userProperties: state.coreQuery.userProperties,
  eventNameOptions: state.coreQuery.eventOptions,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) => bindActionCreators({}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(LinkedEventsBlock);
