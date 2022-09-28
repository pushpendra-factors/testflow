import React, { useEffect, useState } from 'react';
import { connect, useSelector } from 'react-redux';

import styles from './index.module.scss';
import { SVG } from 'factorsComponents';
import { Button } from 'antd';
import GlobalFilterBlock from './GlobalFilterBlock';

const GLobalFilter = ({
  filters = [],
  setGlobalFilters,
  onFiltersLoad = [],
  selKPICategory,
  selectedMainCategory,
  DDKPIValues,
  viewMode = false
}) => {
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const activeProject = useSelector((state) => state.global.active_project);

  const [filterProps, setFilterProperties] = useState({});
  const [filterDD, setFilterDD] = useState(false);

  useEffect(() => {
    const props = Object.assign({}, filterProps);
    props['user'] = DDKPIValues ? DDKPIValues : [];
    setFilterProperties(props);
  }, [DDKPIValues, selKPICategory]);

  useEffect(() => {
    if (onFiltersLoad.length) {
      onFiltersLoad.forEach((fn) => fn());
    }
  }, [filters]);

  const delFilter = (index) => {
    const fltrs = [...filters].filter((f, i) => i !== index);
    setGlobalFilters(fltrs);
  };
  const editFilter = (id, filter) => {
    const fltrs = [...filters].map((f, i) => (i === id ? filter : f));
    setGlobalFilters(fltrs);
  };
  const addFilter = (filter) => {
    const fltrs = [...filters];
    fltrs.push(filter);
    setGlobalFilters(fltrs);
  };
  const closeFilter = () => {
    setFilterDD(false);
  };

  if (filterProps) {
    const filtrs = [];

    filters.forEach((filt, id) => {
      filtrs.push(
        <div key={id} className={'mt-2'}>
          <GlobalFilterBlock
            activeProject={activeProject}
            index={id}
            filterType={'analytics'}
            filter={filt}
            extraClass={styles.filterSelect}
            delIcon={`remove`}
            deleteFilter={delFilter}
            insertFilter={(val) => editFilter(id, val)}
            closeFilter={closeFilter}
            filterProps={filterProps}
            propsConstants={['user']}
            selectedMainCategory={selectedMainCategory}
            viewMode={viewMode}
          ></GlobalFilterBlock>
        </div>
      );
    });

    if (filterDD) {
      filtrs.push(
        <div key={filtrs.length} className={`mt-2`}>
          <GlobalFilterBlock
            activeProject={activeProject}
            blockType={'global'}
            filterType={'analytics'}
            extraClass={styles.filterSelect}
            delBtnClass={styles.filterDelBtn}
            propsConstants={['user']}
            filterProps={filterProps}
            // propsConstants={Object.keys(filterProps)}
            insertFilter={addFilter}
            deleteFilter={() => closeFilter()}
            selectedMainCategory={selectedMainCategory}
            closeFilter={closeFilter}
            viewMode={viewMode}
          ></GlobalFilterBlock>
        </div>
      );
    } else {
      !viewMode && filtrs.push(
        <div key={filtrs.length} className={`flex mt-2`}>
          <Button
            className={`fa-button--truncate`}
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
};

export default GLobalFilter;
