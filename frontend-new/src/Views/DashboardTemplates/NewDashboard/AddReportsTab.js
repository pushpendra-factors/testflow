import React, { useCallback, useState } from 'react';
import { Row, Col, Checkbox, Input, Modal} from 'antd';
import { Text, SVG } from '../../../components/factorsComponents';
import styles from './index.module.scss';
import { SearchOutlined } from '@ant-design/icons';
import { getQueryType } from '../../../utils/dataFormatter';
import { QUERY_TYPE_PROFILE } from '../../../utils/constants';

function AddReportsTab({queries,AddReportsVisible,setAddReportsVisible,selectedQueries, setSelectedQueries}){
  const [searchVal, setSearchVal] = useState('');
  const handleSearchChange = useCallback((e) => {
    setSearchVal(e.target.value);
  }, []);

  const handleCheckBoxClick = useCallback(
    (q) => {
      const isSelected = selectedQueries.findIndex((sq) => sq.id === q.id) > -1;
      if (isSelected) {
        setSelectedQueries((currData) => {
          return currData.filter((c) => c.id !== q.id);
        });
      } else {
        setSelectedQueries((currData) => {
          return [...currData, { ...q, id: q.id }];
        });
      }
    },
    [selectedQueries, setSelectedQueries]
  );

  const handleOk = ()=>{
    setAddReportsVisible(false);
    alert('New Dashboard Created!');
  }
  const handleCancel=()=>{
      setAddReportsVisible(false);
  }
  const filteredQueries = queries.filter(
    (q) => q.title.toLowerCase().indexOf(searchVal.toLowerCase()) > -1
  );

  return(
    <Modal        
      title={"Add Reports"}
      centered={true}
      zIndex={1005}
      width={700}
      onCancel={handleCancel}
      onOk={handleOk}
      className={"fa-modal--regular p-4 fa-modal--slideInDown"}
      // confirmLoading={apisCalled}
      closable={false}
      okText={"Add Reports"}
      cancelText={"Close"}
      transitionName=""
      maskTransitionName=""
      okButtonProps={{ size: "large" }}
      cancelButtonProps={{ size: "large" }}
      visible={AddReportsVisible}
    >
      <div className={`${styles.searchBar} query-search`}>
        <Input
          onChange={handleSearchChange}
          value={searchVal}
          className={styles.searchInput}
          placeholder='Search Reports'
          prefix={<SearchOutlined style={{ width: '1rem' }} color='#0E2647' />}
        />
      </div>
      
      <div className={`queries-list overflow-auto ${styles.lists}`}>
        {filteredQueries.map((q) => {

          const isSelected =
            selectedQueries.findIndex((sq) => sq.id === q.id) > -1;

          return (
            <div
              key={q.id}
              className={`flex items-center justify-between px-1 py-3 cursor-pointer ${
                styles.queryRow
              } ${isSelected ? styles.selected : ''}`}
            >
              <div className='flex justify-start items-center'>
                <div className='mr-2'>
                  <Checkbox
                    checked={isSelected}
                    onChange={handleCheckBoxClick.bind(this, q)}
                  />
                </div>
                <Text   type='paragraph'>
                  {q.title}
                </Text>
              </div>
              {/* <div className='flex'>
                <SVG name={svgName} size={24} />
              </div> */}
            </div>
          );
        })}
      </div>                
    </Modal>
  );
}

export default AddReportsTab;