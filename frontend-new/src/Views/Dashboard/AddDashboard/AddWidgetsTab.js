import React, { useCallback, useState } from 'react';
import { Text, SVG } from '../../../components/factorsComponents';
import {
  Row, Col, Checkbox, Input
} from 'antd';
import styles from './index.module.scss';
import { SearchOutlined } from '@ant-design/icons';

function AddWidgetsTab({ queries, selectedQueries, setSelectedQueries }) {
  const [searchVal, setSearchVal] = useState('');

  const handleSearchChange = useCallback((e) => {
    setSearchVal(e.target.value);
  }, []);

  const handleCheckBoxClick = useCallback((q) => {
    const isSelected = selectedQueries.findIndex(sq => sq.id === q.id) > -1;
    if (isSelected) {
      setSelectedQueries(currData => {
        return currData.filter(c => c.query_id !== q.id);
      });
    } else {
      setSelectedQueries(currData => {
        return [...currData, { ...q, query_id: q.id }];
      });
    }
  }, [selectedQueries, setSelectedQueries]);

  if (!queries.length) {
    return (
      <Row className={'py-20'} justify={'center'} gutter={[24, 24]}>
        <Col span={18}>
          <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Widgets make up your dashboard and are created from saved queries You must create a query first and save it before you can add it here. <a href="#!">Learn more.</a></Text>
        </Col>
      </Row>
    );
  }

  const filteredQueries = queries.filter(q => q.title.toLowerCase().indexOf(searchVal.toLowerCase()) > -1);

  return (
    <div className="widget-selection">

      <div className="query-search">
        <Input
          onChange={handleSearchChange}
          value={searchVal}
          className={styles.searchInput}
          placeholder="Make widgets from saved queries"
          prefix={<SearchOutlined style={{ width: '1rem' }}
            color="#0E2647" />}
        />
      </div>

      <div className="queries-list">
        {filteredQueries.map(q => {
          let svgName = 'funnels_cq';
          const requestQuery = q.query;
          if (requestQuery.query_group) {
            svgName = 'events_dashboard_cq';
          }

          const isSelected = selectedQueries.findIndex(sq => sq.query_id === q.id) > -1;

          return (
            <div key={q.id} className={`flex items-center justify-between px-1 py-3 cursor-pointer ${styles.queryRow} ${isSelected ? styles.selected : ''}`}>
              <div className="flex justify-start items-center">
                <div className="mr-2">
                  <Checkbox checked={isSelected} onChange={handleCheckBoxClick.bind(this, q)} />
                </div>
                <Text mini extraClass={styles.queryTitle} type="paragraph">{q.title}</Text>
              </div>
              <div className="flex">
                <SVG name={svgName} size={24} />
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

export default AddWidgetsTab;
