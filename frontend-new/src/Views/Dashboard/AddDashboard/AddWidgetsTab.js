import React, { useCallback, useState } from 'react';
import { Text, SVG } from '../../../components/factorsComponents';
import { Row, Col, Checkbox, Input } from 'antd';
import styles from './index.module.scss';
import { SearchOutlined } from '@ant-design/icons';
import { getQueryType } from '../../../utils/dataFormatter';
import { QUERY_TYPE_PROFILE } from '../../../utils/constants';
import VirtualList from 'rc-virtual-list';
import useAutoFocus from 'hooks/useAutoFocus';

const itemHeight = 48;
const ContainerHeight = 382;

function AddWidgetsTab({ queries, selectedQueries, setSelectedQueries }) {
  const inputReference = useAutoFocus();
  const [searchVal, setSearchVal] = useState('');

  const handleSearchChange = useCallback((e) => {
    setSearchVal(e.target.value);
  }, []);

  const handleCheckBoxClick = useCallback(
    (q) => {
      const isSelected = selectedQueries.findIndex((sq) => sq.id === q.id) > -1;
      if (isSelected) {
        setSelectedQueries((currData) => {
          return currData.filter((c) => c.query_id !== q.id);
        });
      } else {
        setSelectedQueries((currData) => {
          return [...currData, { ...q, query_id: q.id }];
        });
      }
    },
    [selectedQueries, setSelectedQueries]
  );

  if (!queries.length) {
    return (
      <Row className={'py-20'} justify={'center'} gutter={[24, 24]}>
        <Col span={18}>
          <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>
            Widgets make up your dashboard and are created from saved queries
            You must create a query first and save it before you can add it
            here. <a href='#!'>Learn more.</a>
          </Text>
        </Col>
      </Row>
    );
  }

  const filteredQueries = queries.filter(
    (q) => q.title.toLowerCase().indexOf(searchVal.toLowerCase()) > -1
  );

  return (
    <div className='widget-selection'>
      <div className={`${styles.searchBar} query-search`}>
        <Input
          onChange={handleSearchChange}
          value={searchVal}
          className={styles.searchInput}
          placeholder='Make widgets from saved queries'
          ref={inputReference}
          prefix={<SearchOutlined style={{ width: '1rem' }} color='#0E2647' />}
        />
      </div>

      <div
        className='queries-list'
        style={{
          maxHeight: '500px',
          overflow: 'auto'
        }}
      >
        <VirtualList
          data={filteredQueries}
          height={ContainerHeight}
          itemHeight={itemHeight}
          itemKey='id'
        >
          {(q) => {
            const queryType = getQueryType(q.query);
            const queryTypeName = {
              events: 'events_cq',
              funnel: 'funnels_cq',
              channel_v1: 'campaigns_cq',
              attribution: 'attributions_cq',
              profiles: 'profiles_cq',
              kpi: 'KPI_cq'
            };
            let svgName = '';
            Object.entries(queryTypeName).forEach(([k, v]) => {
              if (queryType === k) {
                svgName = v;
              }
            });

            const isSelected =
              selectedQueries.findIndex((sq) => sq.query_id === q.id) > -1;
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
                    {q.title}
                  </Text>
                </div>
                <div className='flex'>
                  <SVG name={svgName} size={24} />
                </div>
              </div>
            );
          }}
        </VirtualList>
      </div>
    </div>
  );
}

export default AddWidgetsTab;
