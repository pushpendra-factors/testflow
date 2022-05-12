import React, { useCallback, useState } from 'react';
import { Text, SVG } from 'factorsComponents';
import { Row, Col, Checkbox, Input } from 'antd';
import styles from './index.module.scss';
import { SearchOutlined } from '@ant-design/icons';

function SelectChannels({ channelOpts, selectedChannel, setSelectedChannel }) {
  const [searchVal, setSearchVal] = useState('');

  const handleSearchChange = useCallback((e) => {
    setSearchVal(e.target.value);
  }, []);

  const handleCheckBoxClick = useCallback(
    (q) => {
      const isSelected = selectedChannel.findIndex((sq) => sq.id === q.id) > -1;
      if (isSelected) {
        setSelectedChannel((currData) => {
          return currData.filter((c) => c.id !== q.id);
        });
      } else {
        setSelectedChannel((currData) => {
          return [...currData, { ...q, id: q.id }];
        });
      }
      console.log(selectedChannel)
    },
    [selectedChannel, setSelectedChannel]
  );

  const filteredQueries = channelOpts.filter(
    (q) => q.name.toLowerCase().indexOf(searchVal.toLowerCase()) > -1
  );

  return (
    <div className={`widget-selection ${styles.tabContent}`}>
      <div className={`${styles.searchBar} query-search`}>
        <Input
          onChange={handleSearchChange}
          value={searchVal}
          className={styles.searchInput}
          placeholder='Select channels'
          prefix={<SearchOutlined style={{ width: '1rem' }} color='#0E2647' />}
        />
      </div>

      <div className='queries-list'>
        {filteredQueries.map((q) => {

          const isSelected =
          selectedChannel.findIndex((sq) => sq.id === q.id) > -1;

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
                <Text mini extraClass={styles.queryTitle} type='paragraph'>
                  {'#'+ q.name}
                </Text>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default SelectChannels;
