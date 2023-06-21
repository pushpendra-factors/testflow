import React, { useEffect, useState } from 'react';
import { connect, useSelector } from 'react-redux';

import styles from 'Components/GlobalFilter/index.module.scss';
import { Text, SVG } from 'factorsComponents';
import { Button } from 'antd';

import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';

const GlobalFilter = ({
  filters = [],
  setGlobalFilters,
  groupName = 'users',
  event =
  {
    "alias": "",
    "label": "$session",
    "filters": [],
    "group": "Most Recent",
    "key": "9EpYUyLk"
  },
  filterDD,
  setFilterDD,
  eventTypeName
}) => {
  const { userProperties, groupProperties, eventProperties, buttonClickPropNames,pageViewPropNames } = useSelector(
    (state) => state.coreQuery
  );
  const activeProject = useSelector((state) => state.global.active_project);
  const [filterProps, setFilterProperties] = useState({});

  const [orFilterIndex, setOrFilterIndex] = useState(-1); 

  useEffect(() => {
    const props = {}; 

    if (eventTypeName == 'Sessions') {
      props.event = eventProperties[event?.label]; 
    }
    if (eventTypeName == 'CRM Events') {
      props.user = userProperties; 
    }
    if(eventTypeName == 'Page Views'){
      props.page_view = pageViewPropNames; 
    }
    if(eventTypeName == 'Button Clicks'){
      props.button_click = buttonClickPropNames; 
    }

    // if (groupName === 'users') {
    //   props.user = userProperties;
    //   props.group = [];
    // } else {
    //   props.user = [];
    //   props.group = groupProperties[groupName];
    // }

    setFilterProperties(props); 
  }, [userProperties, groupProperties, eventProperties, groupName, eventTypeName, buttonClickPropNames]);

  const delFilter = (index) => {
    const filtersSorted = [...filters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.filter((f, i) => i !== index);
    setGlobalFilters(fltrs);
  };
  const editFilter = (id, filter) => {
    const filtersSorted = [...filters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.map((f, i) => (i === id ? filter : f));
    setGlobalFilters(fltrs);
  };
  const addFilter = (filter) => {
    const fltrs = [...filters];
    fltrs.push(filter);
    setGlobalFilters(fltrs);
  };
  const closeFilter = () => {
    setFilterDD(false);
    setOrFilterIndex(-1);
  };

  if (filterProps) {
    const filtrs = [];
    let index = 0;
    let lastRef = 0;
    if (filters?.length) {
      const group = groupFilters(filters, 'ref');
      const filtersGroupedByRef = Object.values(group);
      const refValues = Object.keys(group);
      lastRef = parseInt(refValues[refValues.length - 1]);

      filtersGroupedByRef.forEach((filtersGr) => {
        const refValue = filtersGr[0].ref;
        if (filtersGr.length === 1) {
          const filt = filtersGr[0];
          filtrs.push(
            <div className={'fa--query_block--filters flex flex-row items-center'}>
              <Text type={'title'} level={8} extraClass={`m-0 mt-2 mr-4`}>Filter by</Text>
              <div key={index} className={`mt-2`}>
                <FilterWrapper
                  event={event}
                  projectID={activeProject?.id}
                  index={index}
                  filter={filt}
                  deleteFilter={delFilter}
                  insertFilter={(val, index) => editFilter(index, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  refValue={refValue}
                  groupName={groupName}
                />
              </div>
              {index !== orFilterIndex && (
                <div className={`mt-2`}>
                  <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
                </div>
              )}
              {index === orFilterIndex && (
                <div key={'init'} className={`mt-2`}>
                  <FilterWrapper
                    event={event}
                    projectID={activeProject?.id}
                    filterProps={filterProps}
                    insertFilter={addFilter}
                    deleteFilter={() => closeFilter()}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr={true}
                    groupName={groupName}
                  />
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filtrs.push(
            <div className={'fa--query_block--filters flex flex-row items-center'}>
              <Text type={'title'} level={8} extraClass={`m-0 mt-2 mr-4`}>Filter by</Text>
              <div key={index} className={`mt-2`}>
                <FilterWrapper
                  event={event}
                  projectID={activeProject?.id}
                  index={index}
                  filter={filtersGr[0]}
                  deleteFilter={delFilter}
                  insertFilter={(val, index) => editFilter(index, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  refValue={refValue}
                  groupName={groupName}
                />
              </div>
              <div key={index + 1} className={`mt-2`}>
                <FilterWrapper
                  event={event}
                  projectID={activeProject?.id}
                  index={index + 1}
                  filter={filtersGr[1]}
                  deleteFilter={delFilter}
                  insertFilter={(val, index) => editFilter(index, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  refValue={refValue}
                  showOr={true}
                  groupName={groupName}
                />
              </div>
            </div>
          );
          index += 2;
        }
      });
    }
    if (filterDD) {
      filtrs.push(
        <div key={filtrs.length} className={`mt-2 flex items-center`}>
          <Text type={'title'} level={8} extraClass={`m-0 mr-4`}>Filter by</Text>
          <FilterWrapper
            event={event}
            projectID={activeProject?.id}
            filterProps={filterProps}
            insertFilter={addFilter}
            deleteFilter={() => closeFilter()}
            closeFilter={closeFilter}
            refValue={lastRef + 1}
            groupName={groupName}
          />
        </div>
      );
    }
    // else {
    //   filtrs.push(
    //     <div key={filtrs.length} className={`flex mt-2`}>
    //       <Button
    //         className={`fa-button--truncate`}
    //         type='text'
    //         onClick={() => setFilterDD(true)}
    //         icon={<SVG name='plus' />}
    //       >
    //         Add new
    //       </Button>
    //     </div>
    //   );
    // }
    return (<div className={`flex flex-col items-start ml-20`}>{filtrs}</div>
    );
  }
};

export default GlobalFilter;
