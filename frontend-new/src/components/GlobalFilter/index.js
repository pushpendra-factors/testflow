import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { SVG } from 'Components/factorsComponents';
import { Button } from 'antd';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { IsDomainGroup } from 'Components/Profile/utils';
import ORButton from '../ORButton';
import { compareFilters, groupFilters } from '../../utils/global';
import styles from './index.module.scss';
import { GROUP_NAME_DOMAINS } from './FilterWrapper/utils';

function GlobalFilter({
  filters = [],
  setGlobalFilters,
  groupName = 'users',
  event
}) {
  const { groups, groupProperties, userPropertiesV2, eventPropertiesV2 } =
    useSelector((state) => state.coreQuery);
  const activeProject = useSelector((state) => state.global.active_project);
  const [filterProps, setFilterProperties] = useState({});
  const [filterDD, setFilterDD] = useState(false);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);
  const predefinedProperty = useSelector(
    (state) => state.preBuildDashboardConfig.config.data.result
  );

  useEffect(() => {
    const props = {};
    if (event?.label) {
      props.event = eventPropertiesV2[event.label];
    }
    if (groupName === 'predefined') {
      props.user = predefinedProperty?.pr;
    } else if (groupName === 'users' || groupName === 'events') {
      props.user = userPropertiesV2;
    } else if (IsDomainGroup(groupName)) {
      props[GROUP_NAME_DOMAINS] = groupProperties[GROUP_NAME_DOMAINS];
      Object.entries(groupProperties || {}).forEach(([group, properties]) => {
        if (Object.keys(groups?.all_groups || {}).includes(group)) {
          props[group] = properties;
        }
      });
    } else {
      props[groupName] = groupProperties[groupName];
    }
    setFilterProperties(props);
  }, [groupProperties, event, groupName, eventPropertiesV2, userPropertiesV2]);

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
            <div className='fa--query_block--filters flex flex-row flex-wrap'>
              <div key={index} className='mt-2'>
                <FilterWrapper
                  event={event}
                  projectID={activeProject?.id}
                  index={index}
                  filter={filt}
                  deleteFilter={delFilter}
                  insertFilter={(val, ind) => editFilter(ind, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  refValue={refValue}
                />
              </div>
              {index !== orFilterIndex && (
                <div className='mt-2'>
                  <ORButton index={index} setOrFilterIndex={setOrFilterIndex} />
                </div>
              )}
              {index === orFilterIndex && (
                <div key='init' className='mt-2'>
                  <FilterWrapper
                    event={event}
                    projectID={activeProject?.id}
                    filterProps={filterProps}
                    insertFilter={addFilter}
                    deleteFilter={() => closeFilter()}
                    closeFilter={closeFilter}
                    refValue={refValue}
                    showOr
                  />
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filtrs.push(
            <div className='fa--query_block--filters flex flex-row flex-wrap'>
              <div key={index} className='mt-2'>
                <FilterWrapper
                  event={event}
                  projectID={activeProject?.id}
                  index={index}
                  filter={filtersGr[0]}
                  deleteFilter={delFilter}
                  insertFilter={(val, ind) => editFilter(ind, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  refValue={refValue}
                />
              </div>
              <div key={index + 1} className='mt-2'>
                <FilterWrapper
                  event={event}
                  projectID={activeProject?.id}
                  index={index + 1}
                  filter={filtersGr[1]}
                  deleteFilter={delFilter}
                  insertFilter={(val, ind) => editFilter(ind, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  refValue={refValue}
                  showOr
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
        <div key={filtrs.length} className='mt-2'>
          <FilterWrapper
            event={event}
            projectID={activeProject?.id}
            filterProps={filterProps}
            insertFilter={addFilter}
            deleteFilter={() => closeFilter()}
            closeFilter={closeFilter}
            refValue={lastRef + 1}
          />
        </div>
      );
    } else {
      filtrs.push(
        <div key={filtrs.length} className='flex mt-2'>
          <Button
            className='fa-button--truncate'
            type='text'
            onClick={() => setFilterDD(true)}
            icon={<SVG name='plus' />}
          >
            Add new
          </Button>
        </div>
      );
    }
    return <div className={styles.block}>{filtrs}</div>;
  }
}

export default GlobalFilter;
