import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { Button } from 'antd';
import { SVG } from '../../../factorsComponents';
import { compareFilters } from '../../../../utils/global';
import FilterWrapper from 'Components/GlobalFilter/FilterWrapper';

function PropertyFilter({
  viewMode,
  filtersLimit = 3,
  profileType,
  source,
  filters = [],
  setFilters,
  onFiltersLoad = []
}) {
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const activeProject = useSelector((state) => state.global.active_project);

  const [filterProps, setFilterProperties] = useState({ user: [], group: [] });
  const [filterDD, setFilterDD] = useState(false);

  useEffect(() => {
    const props = { ...filterProps };
    if (profileType === 'account') {
      if (source === 'All') {
        props.group = [
          ...(groupProperties.$hubspot_company
            ? groupProperties.$hubspot_company
            : []),
          ...(groupProperties.$salesforce_account
            ? groupProperties.$salesforce_account
            : []),
          ...(groupProperties.$6signal ? groupProperties.$6signal : [])
        ];
      } else props.group = groupProperties[source];
    } else if (profileType === 'user') props.user = userProperties;
    setFilterProperties(props);
  }, [userProperties, groupProperties, source]);

  useEffect(() => {
    if (onFiltersLoad.length) {
      onFiltersLoad.forEach((fn) => fn());
    }
  }, [filters]);

  const delFilter = (index) => {
    if (!viewMode) {
      const filtersSorted = [...filters];
      filtersSorted.sort(compareFilters);
      const fltrs = filtersSorted.filter((f, i) => i !== index);
      setFilters(fltrs);
    }
  };
  const editFilter = (id, filter) => {
    if (!viewMode) {
      const filtersSorted = [...filters];
      filtersSorted.sort(compareFilters);
      const fltrs = filtersSorted.map((f, i) => (i === id ? filter : f));
      setFilters(fltrs);
    }
  };
  const addFilter = (filter) => {
    if (!viewMode) {
      const fltrs = [...filters];
      fltrs.push(filter);
      setFilters(fltrs);
    }
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
              groupName={source}
              viewMode={viewMode}
              projectID={activeProject?.id}
              index={id}
              filter={filter}
              deleteFilter={delFilter}
              insertFilter={(val) => editFilter(id, val)}
              closeFilter={closeFilter}
              filterProps={filterProps}
            />
          </div>
        );
      });
      if (filters.length < filtersLimit) {
        if (filterDD) {
          list.push(
            <div key={list.length} className='m-0 mr-2 mb-2'>
              <FilterWrapper
                groupName={source}
                viewMode={viewMode}
                projectID={activeProject?.id}
                index={list.length}
                deleteFilter={() => closeFilter()}
                insertFilter={addFilter}
                closeFilter={closeFilter}
                filterProps={filterProps}
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
