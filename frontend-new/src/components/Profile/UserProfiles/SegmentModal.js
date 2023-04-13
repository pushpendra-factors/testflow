import React, { useState, useEffect, useCallback } from 'react';
import { Modal, Button } from 'antd';
import {
  DefaultDateRangeForSegments,
  displayFilterOpts,
  getSegmentQuery
} from '../utils';
import { SVG, Text } from 'Components/factorsComponents';
import InputFieldWithLabel from '../MyComponents/InputFieldWithLabel/index';
import {
  QUERY_OPTIONS_DEFAULT_VALUE,
  QUERY_TYPE_SEGMENT,
  ReverseProfileMapper
} from 'Utils/constants';
import FaSelect from 'Components/FaSelect';
import { compareFilters, generateRandomKey } from 'Utils/global';
import { useSelector } from 'react-redux';
import EventsBlock from '../MyComponents/EventsBlock';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';

function SegmentModal({
  profileType,
  activeProject,
  type,
  typeOptions,
  editMode = false,
  visible,
  segment = {},
  onSave,
  onCancel,
  tableProps,
  caller
}) {
  const DEFAULT_SEGMENT_PAYLOAD = {
    name: '',
    description: '',
    query: {},
    type:
      type === 'All'
        ? profileType === 'user'
          ? 'web'
          : '$hubspot_company'
        : type
  };

  const DEFAULT_SEGMENT_QUERY_OPTIONS = {
    ...QUERY_OPTIONS_DEFAULT_VALUE,
    caller: caller,
    group_analysis: profileType === 'user' ? 'users' : '$hubspot_company',
    source: type,
    date_range: { ...DefaultDateRangeForSegments },
    table_props: tableProps
  };
  const [isEventDDVisible, setEventDDVisible] = useState(false);
  const [isUserDDVisible, setUserDDVisible] = useState(false);
  const [isConditionDDVisible, setConditionDDVisible] = useState(false);
  const [isFilterDDVisible, setFilterDDVisible] = useState(false);
  const [isCritDDVisible, setCritDDVisible] = useState(false);
  const [segmentPayload, setSegmentPayload] = useState(DEFAULT_SEGMENT_PAYLOAD);
  const [listEvents, setListEvents] = useState([]);
  const [queryOptions, setQueryOptions] = useState(
    DEFAULT_SEGMENT_QUERY_OPTIONS
  );
  const [criteria, setCriteria] = useState('any');
  const [filterProperties, setFilterProperties] = useState({
    user: [],
    group: []
  });
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );

  const CRITERIA_PERF_OPTIONS = [
    ['Any Event', 'any'],
    ['Each Event', 'each'],
    ['All Events', 'all']
  ];

  useEffect(() => {
    const props = { ...filterProperties };
    if (profileType === 'account') {
      props.group = groupProperties[segmentPayload.type];
    } else if (profileType === 'user') props.user = userProperties;
    setFilterProperties(props);
  }, [userProperties, groupProperties, segmentPayload.type, profileType]);

  useEffect(() => {
    const payload = { ...segmentPayload };
    payload.query = getSegmentQuery(listEvents, queryOptions, criteria);
    setSegmentPayload(payload);
  }, [listEvents, queryOptions, criteria]);

  const handleNameInput = (e) => {
    const payload = { ...segmentPayload };
    payload.name = e.target.value;
    setSegmentPayload(payload);
  };

  const handleDescInput = (e) => {
    const payload = { ...segmentPayload };
    payload.description = e.target.value;
    setSegmentPayload(payload);
  };

  const handleClickCancel = () => {
    onCancel();
    setListEvents([]);
    setSegmentPayload(DEFAULT_SEGMENT_PAYLOAD);
    setQueryOptions(DEFAULT_SEGMENT_QUERY_OPTIONS);
  };

  const setSegmentType = (val) => {
    if (val[1] !== segmentPayload.type) {
      const opts = { ...segmentPayload };
      opts.type = val[1];
      setSegmentPayload(opts);
      if (profileType === 'account') {
        const queryOpts = { ...queryOptions };
        queryOpts.group_analysis = val[1];
        setQueryOptions(queryOpts);
      }
    }
    setUserDDVisible(false);
  };

  const selectUsers = () => (
    <div className='absolute top-0'>
      {isUserDDVisible ? (
        <FaSelect
          options={typeOptions}
          onClickOutside={() => setUserDDVisible(false)}
          optionClick={(val) => setSegmentType(val)}
        />
      ) : null}
    </div>
  );

  const renderModalHeader = () => (
    <Text extraClass='m-0 p-4' type={'title'} level={5} weight={'bold'}>
      {editMode ? 'Edit Segment' : 'New Segment'}
    </Text>
  );

  const renderNameSection = () => (
    <InputFieldWithLabel
      extraClass='px-4 pb-4'
      inputClass='fa-input'
      title='Name'
      placeholder='Segment Name'
      value={segmentPayload.name}
      onChange={handleNameInput}
    />
  );

  const renderDescSection = () => (
    <InputFieldWithLabel
      isTextArea
      extraClass='px-4 pb-4'
      inputClass='fa-input'
      title='Description'
      placeholder='Description'
      value={segmentPayload.description}
      onChange={handleDescInput}
    />
  );

  const queryChange = useCallback(
    (newEvent, index, changeType = 'add', flag = null) => {
      const queryupdated = [...listEvents];
      if (queryupdated[index]) {
        if (changeType === 'add') {
          queryupdated[index] = newEvent;
        } else if (changeType === 'filters_updated') {
          // dont remove group by if filter is changed
          queryupdated[index] = newEvent;
        } else {
          queryupdated.splice(index, 1);
        }
      } else {
        if (flag) {
          Object.assign(newEvent, { pageViewVal: flag });
        }
        queryupdated.push(newEvent);
      }
      setListEvents(
        queryupdated.map((q) => {
          return {
            ...q,
            key: q.key || generateRandomKey()
          };
        })
      );
    },
    [listEvents]
  );

  const eventsList = () => {
    const blockList = [];
    listEvents.forEach((event, index) => {
      blockList.push(
        <div key={index}>
          <EventsBlock
            availableGroups={typeOptions}
            index={index + 1}
            queryType={QUERY_TYPE_SEGMENT}
            event={event}
            closeEvent={closeEvent}
            queries={listEvents}
            eventChange={queryChange}
            groupAnalysis={queryOptions.group_analysis}
          />
        </div>
      );
    });

    if (listEvents.length < 3) {
      if (isEventDDVisible) {
        blockList.push(
          <div key={blockList.length}>
            <EventsBlock
              availableGroups={typeOptions}
              queryType={QUERY_TYPE_SEGMENT}
              index={listEvents.length + 1}
              queries={listEvents}
              eventChange={queryChange}
              closeEvent={closeEvent}
              groupBy={queryOptions.groupBy}
              groupAnalysis={queryOptions.group_analysis}
            />
          </div>
        );
      }
    }

    return blockList.length ? (
      <div className='segment-query_block'>
        <h2 className='title'>Performed Events</h2>
        <div className='content'>{blockList}</div>
      </div>
    ) : null;
  };

  const setFilters = (filters) => {
    const opts = { ...queryOptions };
    opts.globalFilters = filters;
    setQueryOptions(opts);
  };

  const editFilter = (id, filter) => {
    const opts = { ...queryOptions };
    const filtersSorted = [...opts.globalFilters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.map((f, i) => (i === id ? filter : f));
    setFilters(fltrs);
  };

  const addFilter = (filter) => {
    const opts = { ...queryOptions };
    const fltrs = [...opts.globalFilters];
    fltrs.push(filter);
    setFilters(fltrs);
  };

  const removeFilters = (index) => {
    const opts = { ...queryOptions };
    const filtersSorted = [...opts.globalFilters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.filter((f, i) => i !== index);
    setFilters(fltrs);
  };

  const closeFilter = () => setFilterDDVisible(false);
  const closeEvent = () => setEventDDVisible(false);

  const filterList = () => {
    if (filterProperties) {
      const list = [];
      queryOptions.globalFilters.forEach((filter, id) => {
        list.push(
          <div key={id}>
            <FilterWrapper
              projectID={activeProject?.id}
              index={id}
              filter={filter}
              deleteFilter={removeFilters}
              insertFilter={(val) => editFilter(id, val)}
              closeFilter={closeFilter}
              filterProps={filterProperties}
              dropdownPlacement='top'
              dropdownMaxHeight={344}
            />
          </div>
        );
      });
      if (queryOptions.globalFilters.length < 3) {
        if (isFilterDDVisible) {
          list.push(
            <div key={list.length}>
              <FilterWrapper
                projectID={activeProject?.id}
                index={list.length}
                deleteFilter={() => closeFilter()}
                insertFilter={addFilter}
                closeFilter={closeFilter}
                filterProps={filterProperties}
                dropdownPlacement='top'
                dropdownMaxHeight={344}
              />
            </div>
          );
        }
      }

      return list.length ? (
        <div className='segment-query_block'>
          <h2 className='title'>With Properties</h2>
          <div className='content'>{list}</div>
        </div>
      ) : null;
    }
    return null;
  };

  const setActions = (opt) => {
    if (opt[1] === 'event') {
      setEventDDVisible(true);
    } else if (opt[1] === 'filter') {
      setFilterDDVisible(true);
    }
    setConditionDDVisible(false);
  };

  const generateConditionOpts = () => {
    const options = [];
    if (listEvents.length < 3) {
      options.push(['Performed Events', 'event']);
    }
    if (queryOptions.globalFilters.length < 3) {
      options.push(['With Properties', 'filter']);
    }
    return options;
  };

  const selectCondition = () => (
    <div className='absolute bottom-0'>
      {isConditionDDVisible ? (
        <FaSelect
          options={generateConditionOpts()}
          onClickOutside={() => setConditionDDVisible(false)}
          optionClick={(val) => setActions(val)}
          placement='top'
        />
      ) : null}
    </div>
  );

  const renderQuerySection = () => (
    <div className='p-4'>
      <div className='flex items-center mb-2'>
        <Text
          type={'title'}
          level={6}
          weight={'medium'}
          extraClass={`m-0 mr-3`}
        >
          Analyse
        </Text>
        <div className='relative mr-2'>
          <Button
            type='text'
            className='dropdown-btn'
            icon={<SVG name='user_friends' size={16} />}
            onClick={() => setUserDDVisible(!isUserDDVisible)}
          >
            {typeOptions?.find((elem) => elem[1] === segmentPayload?.type)?.[0]}
            <SVG name='caretDown' size={16} />
          </Button>
          {selectUsers()}
        </div>
      </div>
      <div className='segment-query_container'>
        <div className='segment-query_section'>
          {eventsList()}
          {filterList()}
          {(listEvents.length > 2 && queryOptions.globalFilters.length > 2) ||
          isEventDDVisible ||
          isFilterDDVisible ? null : (
            <div
              className={`relative ${
                listEvents.length || queryOptions.globalFilters.length
                  ? 'mt-2 ml-4'
                  : ''
              }`}
            >
              <Button
                type='text'
                icon={<SVG name='plus' size={16} />}
                onClick={() => setConditionDDVisible(!isConditionDDVisible)}
              >
                Add Condition
              </Button>
              {selectCondition()}
            </div>
          )}
        </div>

        {listEvents.length > 1 ? (
          <div style={{ borderTop: '2px solid #DBDBDB' }}>
            {selectCriteria()}
          </div>
        ) : null}
      </div>
    </div>
  );

  const selectCriteria = () => (
    <div className='flex items-center m-3'>
      <h2 className='whitespace-no-wrap line-height-8 m-0 mr-2'>
        {`See
      ${
        profileType === 'user'
          ? ReverseProfileMapper[segmentPayload.type]?.users
          : displayFilterOpts[segmentPayload.type]
      } in the last 28 days who performed `}
      </h2>
      <div className={`relative fa-button--truncate`}>
        <Button type='link' onClick={() => setCritDDVisible(!isCritDDVisible)}>
          {CRITERIA_PERF_OPTIONS.filter((op) => op[1] === criteria)[0][0]}
        </Button>

        {isCritDDVisible && (
          <FaSelect
            options={CRITERIA_PERF_OPTIONS}
            optionClick={(op) => {
              setCriteria(op[1]);
              setCritDDVisible(false);
            }}
            onClickOutside={() => setCritDDVisible(false)}
            placement='top'
          />
        )}
      </div>
    </div>
  );

  const renderModalFooter = () => (
    <div className={`segment-modal_footer`}>
      <div>
        <Button className='mr-1' type='default' onClick={handleClickCancel}>
          Cancel
        </Button>
        <Button
          className='ml-1'
          type='primary'
          onClick={() => onSave(segmentPayload)}
        >
          {editMode ? 'Save Changes' : 'Save Segments'}
        </Button>
      </div>
      {/* {editMode ? (
    <Button
      type='text'
      onClick={resetInputField}
      icon={<SVG size={16} name='trash' color={'grey'} />}
    >
      Delete Segment
    </Button>
  ) : null} */}
    </div>
  );

  return (
    <Modal
      title={null}
      width={1020}
      visible={visible}
      footer={null}
      className={'fa-modal--regular p-6'}
      closable={false}
    >
      <div className='segment-modal'>
        {renderModalHeader()}
        {renderNameSection()}
        {renderDescSection()}
        {renderQuerySection()}
      </div>
      {renderModalFooter()}
    </Modal>
  );
}

export default SegmentModal;
