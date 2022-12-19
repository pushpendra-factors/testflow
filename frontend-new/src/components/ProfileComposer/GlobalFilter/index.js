import React, { useEffect, useState } from "react";
import { useSelector } from "react-redux";
import styles from "./index.module.scss";
import { SVG } from "Components/factorsComponents";
import { Button } from "antd";
import GlobalFilterBlock from "./GlobalFilterBlock";
import ORButton from 'Components/ORButton';
import { compareFilters, groupFilters } from 'Utils/global';

const GLobalFilter = ({
  filters = [],
  setGlobalFilters,
  groupName
}) => {
  const userProperties = useSelector((state) => state.coreQuery.userProperties);
  const groupProperties = useSelector((state) => state.coreQuery.groupProperties);
  const activeProject = useSelector((state) => state.global.active_project);
  const [filterProps, setFilterProperties] = useState({});
  const [filterDD, setFilterDD] = useState(false);
  const [orFilterIndex, setOrFilterIndex] = useState(-1);

  useEffect(() => {
    const props = Object.assign({}, filterProps);
    if (groupName === 'users') {
      props['user'] = userProperties;
      props['group'] = [];
    } else {
      props['user'] = [];
      props['group'] = groupProperties[groupName];
    }

    setFilterProperties(props);
  }, [userProperties, groupProperties, groupName]);

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
    console.log(filter);
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
    if(filters?.length){
    const group = groupFilters(filters, 'ref');
    const filtersGroupedByRef = Object.values(group);
    const refValues = Object.keys(group);
    lastRef = parseInt(refValues[refValues.length-1]);

    filtersGroupedByRef.forEach((filtersGr)=>{
      const refValue = filtersGr[0].ref;
      if(filtersGr.length === 1){
        const filt = filtersGr[0];
        filtrs.push(
          <div className={'fa--query_block--filters flex flex-row'}>
            <div key={index} className={`mt-2`}>
              <GlobalFilterBlock
                activeProject={activeProject}
                index={index}
                filterType={'analytics'}
                filter={filt}
                extraClass={styles.filterSelect}
                delIcon={`remove`}
                deleteFilter={delFilter}
                insertFilter={(val,index) => editFilter(index, val)}
                closeFilter={closeFilter}
                filterProps={filterProps}
                propsConstants={['user']}
                refValue={refValue}
                groupName={groupName}
              ></GlobalFilterBlock>
            </div>
           {index !== orFilterIndex && (
             <div className={`mt-2`}>
             <ORButton index={index} setOrFilterIndex={setOrFilterIndex}/>
             </div>
            )}       
           {index === orFilterIndex && (
              <div key={'init'} className={`mt-2`}>
              <GlobalFilterBlock
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
                refValue={refValue}
                showOr = {true}
                groupName={groupName}
              ></GlobalFilterBlock>
              </div>              
            )}  
            </div>     
        );
        index+=1;
    }else{
      filtrs.push(
        <div className={'fa--query_block--filters flex flex-row'}>
          <div key={index} className={`mt-2`}>
              <GlobalFilterBlock
                activeProject={activeProject}
                index={index}
                filterType={'analytics'}
                filter={filtersGr[0]}
                extraClass={styles.filterSelect}
                delIcon={`remove`}
                deleteFilter={delFilter}
                insertFilter={(val,index) => editFilter(index, val)}
                closeFilter={closeFilter}
                filterProps={filterProps}
                propsConstants={['user', 'group']}
                refValue={refValue}
                groupName={groupName}
              ></GlobalFilterBlock>
            </div>
          <div key={index+1} className={`mt-2`}>
            <GlobalFilterBlock
                activeProject={activeProject}
                index={index+1}
                filterType={'analytics'}
                filter={filtersGr[1]}
                extraClass={styles.filterSelect}
                delIcon={`remove`}
                deleteFilter={delFilter}
                insertFilter={(val,index) => editFilter(index, val)}
                closeFilter={closeFilter}
                filterProps={filterProps}
                propsConstants={['user', 'group']}
                refValue={refValue}
                showOr = {true}
                groupName={groupName}
              ></GlobalFilterBlock>
          </div>
        </div>
      );
      index+=2;
    }
  })
}
    if (filterDD) {
      filtrs.push(
        <div key={filtrs.length} className={`mt-2`}>
          <GlobalFilterBlock
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
            refValue={lastRef + 1}
            groupName={groupName}
          ></GlobalFilterBlock>
        </div>
      );
    } else {
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
