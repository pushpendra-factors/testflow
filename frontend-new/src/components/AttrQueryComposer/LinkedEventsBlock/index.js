import React, { useState, useEffect } from 'react';
import styles from './index.module.scss';
import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';
import FilterWrapper from '../../GlobalFilter/FilterWrapper';
import { Button, Dropdown, Menu, Tooltip } from 'antd';
import { SVG } from 'factorsComponents';
import { isArray } from 'lodash';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import GroupSelect from 'Components/GenericComponents/GroupSelect';
import getGroupIcon from 'Utils/getGroupIcon';

const LinkedEventsBlock = ({
  linkEvent,
  linkEventChange,
  delLinkEvent,
  eventNameOptions,
  activeProject,
  eventPropertiesV2,
  eventNames,
  eventUserPropertiesV2
}) => {
  const [selectVisible, setSelectVisible] = useState(false);
  const [filterBlockVisible, setFilterBlockVisible] = useState(false);

  const [filterProps, setFilterProperties] = useState({
    event: [],
    user: []
  });

  useEffect(() => {
    if (!linkEvent || !linkEvent?.label?.length) {
      return;
    }
    const assignFilterProps = Object.assign({}, filterProps);

    if (eventPropertiesV2[linkEvent.label]) {
      assignFilterProps.event = eventPropertiesV2[linkEvent.label];
    }
    assignFilterProps.user = eventUserPropertiesV2;
    setFilterProperties(assignFilterProps);
  }, [eventUserPropertiesV2, eventPropertiesV2, linkEvent]);

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
        hasPrefix
        filterProps={filterProps}
        projectID={activeProject?.id}
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
              hasPrefix
              index={index}
              filter={filterContent}
              filterProps={filterProps}
              projectID={activeProject?.id}
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

  const onEventSelect = (option, group) => {
    const currentLinkEvent = Object.assign({}, linkEvent);
    currentLinkEvent.label = option?.value;
    currentLinkEvent.filters = [];
    linkEventChange(currentLinkEvent);
    setSelectVisible(false);
  };
  const setAdditionalactions = (opt) => {
    if (opt[1] === 'filter') {
      addFilterBlock();
    }
  };

  const getMenu = (filterOptions) => (
    <Menu style={{ minWidth: '200px', padding: '10px' }}>
      {filterOptions.map((eachFilter, eachIndex) => {
        return (
          <Menu.Item
            icon={
              <SVG
                name={eachFilter[1]}
                extraClass={'self-center'}
                style={{ marginRight: '10px' }}
              ></SVG>
            }
            style={{ display: 'flex', padding: '10px', margin: '5px' }}
            key={eachIndex}
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
      <div className={'fa--query_block--actions--cols flex relative ml-2'}>
        <div className={`relative flex`}>
          <Tooltip
            title='Filter this Linked Event'
            color={TOOLTIP_CONSTANTS.DARK}
          >
            <Button
              className={'fa-btn--custom btn-total-round'}
              type='text'
              onClick={addFilterBlock}
            >
              <SVG name='filter'></SVG>
            </Button>
          </Tooltip>
        </div>
        <Tooltip
          title='Delete this Linked Event'
          color={TOOLTIP_CONSTANTS.DARK}
        >
          <Button
            className={'fa-btn--custom btn-total-round'}
            type='text'
            onClick={deleteItem}
          >
            <SVG name='trash'></SVG>
          </Button>
        </Tooltip>
      </div>
    );
  };

  const selectEvents = () => {
    let orderedEventOptions = eventNameOptions.map((groupOpt) => {
      return {
        iconName: getGroupIcon(groupOpt?.icon),
        label: groupOpt?.label,
        value: groupOpt?.label,
        values: groupOpt?.values?.map((op) => {
          return {
            value: op[1],
            label: op[0]
          };
        })
      };
    });
    const mostRecentGroupindex = orderedEventOptions
      ?.map((opt) => opt.label)
      ?.indexOf('Most Recent');
    if (mostRecentGroupindex > 0) {
      orderedEventOptions = [
        orderedEventOptions[mostRecentGroupindex],
        ...orderedEventOptions.slice(0, mostRecentGroupindex),
        ...orderedEventOptions.slice(mostRecentGroupindex + 1)
      ];
    }
    return (
      <div className={styles.block__event_selector}>
        {selectVisible ? (
          <div className={styles.block__event_selector__btn}>
            <GroupSelect
              options={orderedEventOptions}
              onClickOutside={() => setSelectVisible(false)}
              optionClickCallback={onEventSelect}
              placeholder='Select Event'
              allowSearch={true}
              extraClass={styles.block__event_selector__select}
              allowSearchTextSelection={false}
            />
          </div>
        ) : null}
      </div>
    );
  };

  const renderLinkEventBlockContent = () => {
    let filterOptions = [];
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
        {filterOptions.length != 0 ? (
          <Dropdown
            placement='bottomLeft'
            overlay={getMenu(filterOptions)}
            trigger={['hover']}
          >
            <Button
              type='text'
              size={'large'}
              className={`fa-btn--custom mr-1 btn-total-round`}
            >
              <SVG name='more' />
            </Button>
          </Dropdown>
        ) : (
          ''
        )}
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
};

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  eventPropertiesV2: state.coreQuery.eventPropertiesV2,
  eventUserPropertiesV2: state.coreQuery.eventUserPropertiesV2,
  eventNameOptions: state.coreQuery.eventOptions,
  eventNames: state.coreQuery.eventNames
});

const mapDispatchToProps = (dispatch) => bindActionCreators({}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(LinkedEventsBlock);
