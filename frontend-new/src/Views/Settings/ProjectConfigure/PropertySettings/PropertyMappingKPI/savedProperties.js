import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import {
  Row,
  Col,
  Form,
  Button,
  Input,
  Select,
  Table,
  notification,
  Dropdown,
  Menu,
  message,
  Modal
} from 'antd';
import { MoreOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import {
  fetchPropertyMappings,
  removePropertyMapping
} from 'Reducers/settings/middleware';
import { getPropertyDisplayName} from './utils';
import _ from 'lodash';

const SavedProperties = ({
  activeProject,
  propertyMapping,
  fetchPropertyMappings,
  removePropertyMapping,
  KPI_config
}) => {

  const [mappedProperties, setMappedProperties] = useState([]);

  useEffect(() => {
    const propertyMaps = [];
    propertyMapping.forEach((prop) => {
      propertyMaps.push({
        name: prop.display_name,
        properties: prop.properties,
        actions: prop,
      });
    });
    setMappedProperties(propertyMaps);
  }, [propertyMapping]);

  const removeProperty = (item) => {
    Modal.confirm({
      title: 'Do you want to remove this property map?',
      content: 'Please confirm to proceed',
      okText: 'Yes',
      cancelText: 'Cancel',
      onOk: () => {
        removePropertyMapping(activeProject?.id, item?.id).then(() => {
          fetchPropertyMappings(activeProject?.id);
          message.success('Property Map removed!')
        }).catch((err) => {
          message.error(err?.data?.error);
          console.log('Property Map remove failed-->', err);
        });
      }
    });
  }

  const menu = (item) => {
    return (
      <Menu>
        <Menu.Item key='0'
          onClick={() => removeProperty(item)}
        >
          <a>Remove Property</a>
        </Menu.Item>
      </Menu>
    );
  };

  const columns = [
    {
      title: 'Diplay name',
      dataIndex: 'name',
      key: 'name',
      width: '200px',
      render: (text) => <span className={'capitalize'}>{text}</span>,
    },
    {
      title: 'Mapped properties',
      dataIndex: 'properties',
      key: 'properties',
      render: (items) => (
        <div className='flex items-center flex-wrap'>
      {  items.map((item,index) => { 
          return (<div className='flex items-center'>
            { index!=0 && <SVG name={'DoubeEndedArrow'} color={'grey'} size={24} extraClass={'mr-1 ml-1'} /> }
            <div 
                className={`fa-div--truncate btn-total-round py-1 px-2 background-color--mono-color-1`} 
              >
                {`${getPropertyDisplayName(KPI_config,item?.dc,item?.name)} (${_.startCase(item?.dc)})`}
              </div>
          </div>
          )
        })}
        </div>
        // <span className={'capitalize'}>{text ? text.replace('_', ' ') : ""}</span>
      ),
    },
    {
      title: '',
      dataIndex: 'actions',
      key: 'actions',
      render: (obj) => (
        <div className={`flex justify-end`}>
          <Dropdown
            overlay={() => menu(obj)}
            trigger={['click']}>
            <Button size={'large'} type='text' icon={<MoreOutlined />} />
          </Dropdown>
        </div>
      ),
    },
  ];

  return (
    <>
      <Table
        className='fa-table--basic mt-4'
        columns={columns}
        dataSource={mappedProperties}
        pagination={false}
      // loading={tableLoading}
      />
    </>
  )
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  propertyMapping: state.settings.propertyMapping,
  KPI_config: state.kpi?.config,
});

export default connect(mapStateToProps, { fetchPropertyMappings, removePropertyMapping })(SavedProperties)