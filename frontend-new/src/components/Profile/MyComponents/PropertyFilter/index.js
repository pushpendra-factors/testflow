import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Button } from 'antd';
import { SVG } from '../../../factorsComponents';
import { compareFilters } from '../../../../utils/global';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';
import { GROUP_NAME_DOMAINS } from 'Components/GlobalFilter/FilterWrapper/utils';
import { IsDomainGroup } from 'Components/Profile/utils';

function PropertyFilter({
  viewMode,
  filtersLimit = 3,
  profileType,
  source,
  filters = [],
  setFilters
}) {
  const userPropertiesV2 = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const availableGroups = useSelector((state) => state.coreQuery.groups);
  const activeProject = useSelector((state) => state.global.active_project);
  const predefinedProperty = useSelector((state) => state.preBuildDashboardConfig.config.data.result);

  const [filterProps, setFilterProperties] = useState({});
  const [filterDD, setFilterDD] = useState(false);

  useEffect(() => {
    const props = {};
    if (profileType === 'account') {
      if (IsDomainGroup(source)) {
        props[GROUP_NAME_DOMAINS] = groupProperties[GROUP_NAME_DOMAINS];
        Object.keys(availableGroups?.account_groups).forEach((group) => {
          props[group] = groupProperties[group];
        });
      } else props[source] = groupProperties[source];
      props.user = userPropertiesV2;
    } else if (profileType === 'user') {
      props.user = userPropertiesV2;
    } else if (profileType === 'predefined') {
      props.user = predefinedProperty?.pr;
    }
    setFilterProperties(props);
  }, [userPropertiesV2, groupProperties, availableGroups, profileType, source, predefinedProperty?.pr]);

  const updateFilters = (newFilters) => {
    if (viewMode) return;
    const sortedFilters = [...newFilters].sort(compareFilters);
    setFilters(sortedFilters);
  };

  const delFilter = (index) => {
    updateFilters(filters.filter((f, i) => i !== index));
  };

  const editFilter = (id, filter) => {
    updateFilters(filters.map((f, i) => (i === id ? filter : f)));
  };

  const addFilter = (filter) => {
    updateFilters([...filters, filter]);
  };

  const closeFilter = () => {
    setFilterDD(false);
  };

  const filterList = () => {
    if (filterProps) {
      const list = [];
      filters.forEach((filter, id) => {
        list.push(
          <div key={id} className='m-0 mr-2 mb-2'>
            <FilterWrapper
              viewMode={viewMode}
              projectID={activeProject?.id}
              index={id}
              filter={filter}
              deleteFilter={delFilter}
              insertFilter={(val) => editFilter(id, val)}
              closeFilter={closeFilter}
              filterProps={filterProps}
              minEntriesPerGroup={3}
              profileType={profileType}
            />
          </div>
        );
      });
      if (filters.length < filtersLimit) {
        if (filterDD) {
          list.push(
            <div key={list.length} className='m-0 mr-2 mb-2'>
              <FilterWrapper
                viewMode={viewMode}
                projectID={activeProject?.id}
                index={list.length}
                deleteFilter={() => closeFilter()}
                insertFilter={addFilter}
                closeFilter={closeFilter}
                filterProps={filterProps}
                minEntriesPerGroup={3}
                profileType={profileType}
              />
            </div>
          );
        } else if (!viewMode) {
          list.push(
            <div key={list.length} className='flex m-0 mr-2 mb-2'>
              <Button
                className='fa-button--truncate'
                type='link'
                onClick={() => setFilterDD(true)}
                icon={<SVG name='plus' color='purple' />}
              >
                {filters.length ? null : 'Add Filter'}
              </Button>
            </div>
          );
        }
      }
      return (
        <div className={`flex ${viewMode ? 'flex-col' : 'flex-wrap'}`}>
          {list}
        </div>
      );
    }
    return null;
  };
  return <>{filterList()}</>;
}
export default PropertyFilter;