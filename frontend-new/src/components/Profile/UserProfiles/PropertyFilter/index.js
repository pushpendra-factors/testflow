import React, { useEffect, useState } from 'react';
import { useSelector } from 'react-redux';
import { SVG } from 'factorsComponents';
import { Button } from 'antd';
import { compareFilters } from '../../../../utils/global';
import PropFilterBlock from './PropFilterBlock';

const PropertyFilter = ({
  profileType,
  source,
  filters = [],
  setFilters,
  onFiltersLoad = [],
}) => {
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const groupProperties = useSelector(
    (state) => state.coreQuery.groupProperties
  );
  const activeProject = useSelector((state) => state.global.active_project);

  const [filterProps, setFilterProperties] = useState({ user: [], group: [] });
  const [filterDD, setFilterDD] = useState(false);

  useEffect(() => {
    const props = Object.assign({}, filterProps);
    if (profileType === 'account') {
      if (source === 'All') {
        props.group = [
          ...(groupProperties['$hubspot_company']
            ? groupProperties['$hubspot_company']
            : []),
          ...(groupProperties['$salesforce_account']
            ? groupProperties['$salesforce_account']
            : []),
        ];
      } else props.group = groupProperties[source];
    } else if (profileType === 'user') props.user = userProperties;
    setFilterProperties(props);
  }, [userProperties, groupProperties]);

  useEffect(() => {
    if (onFiltersLoad.length) {
      onFiltersLoad.forEach((fn) => fn());
    }
  }, [filters]);

  const delFilter = (index) => {
    const filtersSorted = [...filters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.filter((f, i) => i !== index);
    setFilters(fltrs);
  };
  const editFilter = (id, filter) => {
    const filtersSorted = [...filters];
    filtersSorted.sort(compareFilters);
    const fltrs = filtersSorted.map((f, i) => (i === id ? filter : f));
    setFilters(fltrs);
  };
  const addFilter = (filter) => {
    const fltrs = [...filters];
    fltrs.push(filter);
    setFilters(fltrs);
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
            <PropFilterBlock
              activeProject={activeProject}
              index={id}
              filter={filter}
              deleteFilter={delFilter}
              insertFilter={(val) => editFilter(id, val)}
              closeFilter={closeFilter}
              filterProps={filterProps}
              propsConstants={['user']}
            ></PropFilterBlock>
          </div>
        );
      });
      if (filters.length < 3) {
        if (filterDD) {
          list.push(
            <div key={list.length} className='m-0 mr-2 mb-2'>
              <PropFilterBlock
                activeProject={activeProject}
                index={list.length}
                deleteFilter={() => closeFilter()}
                insertFilter={addFilter}
                closeFilter={closeFilter}
                filterProps={filterProps}
                propsConstants={['user']}
              ></PropFilterBlock>
            </div>
          );
        } else {
          list.push(
            <div key={list.length} className='flex m-0 mr-2 mb-2'>
              <Button
                className={`fa-button--truncate`}
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
      return <div className={'flex flex-wrap'}>{list}</div>;
    }
  };
  return <>{filterList()}</>;
};
export default PropertyFilter;
