import React, { useEffect, useState } from 'react';
import { connect, useSelector } from 'react-redux';

import styles from './index.module.scss';
import { SVG } from 'factorsComponents';
import { Button } from 'antd';

import GlobalFilterBlock from './GlobalFilterBlock';
import ORButton from '../../ORButton';
import { compareFilters, groupFilters } from '../../../utils/global';

const GLobalFilter = ({
  filters = [],
  setGlobalFilters,
  onFiltersLoad = [],
  KPIConfigProps,
  selectedMainCategory,
  viewMode = false,
  propertyMaps,
  isSameKPIGrp
}) => {
  const userPropertiesV2 = useSelector(
    (state) => state.coreQuery.userPropertiesV2
  );
  const activeProject = useSelector((state) => state.global.active_project);

  const [filterProps, setFilterProperties] = useState({});
  const [filterDD, setFilterDD] = useState(false);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  useEffect(() => {
    let commonProperties = [];
    if (propertyMaps) {
      commonProperties = propertyMaps?.map((item) => {
        return [item?.display_name, item?.name, item?.data_type, 'propMap'];
      });
    }
    const props = Object.assign({}, filterProps);
    props['user'] = !isSameKPIGrp
      ? commonProperties
      : KPIConfigProps
      ? KPIConfigProps
      : [];
    setFilterProperties(props);
  }, [KPIConfigProps, propertyMaps, isSameKPIGrp]);

  // useEffect(() => {
  //   if (onFiltersLoad.length) {
  //     onFiltersLoad.forEach((fn) => fn());
  //   }
  // }, [filters]);

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
        if (filtersGr.length == 1) {
          const filt = filtersGr[0];
          filtrs.push(
            <div className={`fa--query_block--filters flex flex-wrap`}>
              <div key={index} className={`mt-2`}>
                <GlobalFilterBlock
                  isSameKPIGrp={isSameKPIGrp}
                  activeProject={activeProject}
                  index={index}
                  filterType={'analytics'}
                  filter={filt}
                  extraClass={styles.filterSelect}
                  delIcon={`remove`}
                  deleteFilter={delFilter}
                  insertFilter={(val, index) => editFilter(index, val)}
                  closeFilter={closeFilter}
                  selectedMainCategory={selectedMainCategory}
                  filterProps={filterProps}
                  propsConstants={['user']}
                  refValue={refValue}
                  viewMode={viewMode}
                ></GlobalFilterBlock>
              </div>
              {index !== orFilterIndex && (
                <div className={`mt-2`}>
                  {!viewMode && (
                    <ORButton
                      index={index}
                      setOrFilterIndex={setOrFilterIndex}
                    />
                  )}
                </div>
              )}
              {index === orFilterIndex && (
                <div key={'init'} className={`mt-2`}>
                  <GlobalFilterBlock
                    isSameKPIGrp={isSameKPIGrp}
                    activeProject={activeProject}
                    blockType={'global'}
                    filterType={'analytics'}
                    extraClass={styles.filterSelect}
                    delBtnClass={styles.filterDelBtn}
                    filterProps={filterProps}
                    propsConstants={Object.keys(filterProps)}
                    insertFilter={addFilter}
                    deleteFilter={() => closeFilter()}
                    closeFilter={closeFilter}
                    selectedMainCategory={selectedMainCategory}
                    refValue={refValue}
                    showOr={true}
                    viewMode={viewMode}
                  ></GlobalFilterBlock>
                </div>
              )}
            </div>
          );
          index += 1;
        } else {
          filtrs.push(
            <div className={'fa--query_block--filters flex flex-wrap'}>
              <div key={index} className={`mt-2`}>
                <GlobalFilterBlock
                  isSameKPIGrp={isSameKPIGrp}
                  activeProject={activeProject}
                  index={index}
                  filterType={'analytics'}
                  filter={filtersGr[0]}
                  extraClass={styles.filterSelect}
                  delIcon={`remove`}
                  deleteFilter={delFilter}
                  insertFilter={(val, index) => editFilter(index, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  selectedMainCategory={selectedMainCategory}
                  propsConstants={['user']}
                  refValue={refValue}
                  viewMode={viewMode}
                ></GlobalFilterBlock>
              </div>
              <div key={index + 1} className={`mt-2`}>
                <GlobalFilterBlock
                  isSameKPIGrp={isSameKPIGrp}
                  activeProject={activeProject}
                  index={index + 1}
                  filterType={'analytics'}
                  filter={filtersGr[1]}
                  extraClass={styles.filterSelect}
                  delIcon={`remove`}
                  deleteFilter={delFilter}
                  insertFilter={(val, index) => editFilter(index, val)}
                  closeFilter={closeFilter}
                  filterProps={filterProps}
                  propsConstants={['user']}
                  selectedMainCategory={selectedMainCategory}
                  refValue={refValue}
                  showOr={true}
                  viewMode={viewMode}
                ></GlobalFilterBlock>
              </div>
            </div>
          );
          index += 2;
        }
      });
    }
    if (filterDD) {
      filtrs.push(
        <div key={filtrs.length} className={`mt-2`}>
          <GlobalFilterBlock
            isSameKPIGrp={isSameKPIGrp}
            activeProject={activeProject}
            blockType={'global'}
            filterType={'analytics'}
            extraClass={styles.filterSelect}
            delBtnClass={styles.filterDelBtn}
            filterProps={filterProps}
            propsConstants={Object.keys(filterProps)}
            insertFilter={addFilter}
            deleteFilter={() => closeFilter()}
            closeFilter={closeFilter}
            selectedMainCategory={selectedMainCategory}
            refValue={lastRef + 1}
            viewMode={viewMode}
          ></GlobalFilterBlock>
        </div>
      );
    } else {
      !viewMode &&
        filtrs.push(
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
