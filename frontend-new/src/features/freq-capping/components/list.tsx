import React, { useCallback, useEffect, useReducer, useState } from 'react';
import { Text } from 'Components/factorsComponents';
import { Badge, Button, Dropdown, Menu, Row, Table, notification } from 'antd';
import { MoreOutlined } from '@ant-design/icons';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { useSelector } from 'react-redux';
import { FrequencyCap } from '../types';
import styles from '../index.module.scss';
import { deleteLinkedinFreqCapRules } from '../state/service';

interface FreqCapListProps {
  freqCapRules: Array<FrequencyCap>;
  deleteCallBack: () => any;
}
const FrequencyCappingList = ({
  freqCapRules,
  deleteCallBack
}: FreqCapListProps) => {
  const history = useHistory();

  const { active_project } = useSelector((state: any) => state.global);

  const columns = [
    {
      title: 'Name',
      dataIndex: 'display_name',
      key: 'display_name',
      width: '700px',
      render: (item: FrequencyCap) => (
        <Text
          type='title'
          level={7}
          truncate
          charLimit={50}
          extraClass='cursor-pointer m-0'
        >
          {item}
        </Text>
      )
    },
    {
      title: 'Status',
      dataIndex: 'status',
      key: 'status',
      render: (item: FrequencyCap) => (
        <div className='flex items-center'>
          {item?.status === 'paused' ? (
            <Badge
              className='fa-custom-badge fa-custom-badge--orange'
              status='default'
              text='unpublished'
            />
          ) : (
            <Badge
              className='fa-custom-badge fa-custom-badge--green'
              status='processing'
              text='published'
            />
          )}
        </div>
      )
    },
    {
      title: '',
      dataIndex: 'id',
      key: 'id',
      align: 'right',
      width: 75,
      render: (obj) => (
        <Dropdown
          trigger={['click']}
          overlay={menu(obj)}
          placement='bottomRight'
        >
          <Button
            type='text'
            icon={
              <MoreOutlined
                rotate={90}
                style={{ color: 'gray', fontSize: '18px' }}
              />
            }
          />
        </Dropdown>
      )
    }
  ];

  const menu = (item: string) => (
    <Menu className={`${styles.antdActionMenu}`}>
      <Menu.Item
        key='0'
        onClick={() => {
          // Edit rule
          history.replace(`${PathUrls.FreqCap}/${item}`);
        }}
      >
        <a>Edit rule</a>
      </Menu.Item>
      <Menu.Item
        key='1'
        onClick={(e) => {
          // crearte copy
        }}
      >
        <a>Create copy</a>
      </Menu.Item>
      <Menu.Divider />
      <Menu.Item
        key='2'
        onClick={() => {
          // delete action call
          deleteRule(item);
        }}
      >
        <a>
          <span style={{ color: 'red' }}>Remove rule</span>
        </a>
      </Menu.Item>
    </Menu>
  );

  const deleteRule = async (id: string) => {
    const response = await deleteLinkedinFreqCapRules(active_project.id, id);
    if (response?.status === 200) {
      notification.success({
        message: 'Success',
        description: 'Rule Successfully Deleted!',
        duration: 3
      });

      deleteCallBack();
    }
  };

  const moveToRuleView = (obj: any) => {
    console.log(obj);
  };

  return (
    <div>
      <Row>
        <Text type='title' level={7} weight='bold' extraClass='m-0'>
          Frequency capping rules
        </Text>
      </Row>
      <Row>
        <Table
          className='fa-table--basic mt-6'
          onRow={(record, rowIndex) => ({
            onClick: (event) => {
              // SetViewMode(true);
              // setAlertDetails(record.actions);
            } // click row
          })}
          columns={columns}
          dataSource={freqCapRules}
          pagination={false}
          loading={false}
          tableLayout='fixed'
          rowClassName='cursor-pointer'
        />
      </Row>
    </div>
  );
};

export default FrequencyCappingList;
