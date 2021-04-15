import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';

import { Text, SVG } from 'factorsComponents'; 
import {
    Row, Col, Button, Tabs, Table, Dropdown, Menu, notification
  } from 'antd';
import { MoreOutlined } from "@ant-design/icons";

import { fetchSmartProperties, deleteSmartProperty } from 'Reducers/settings/middleware';
import SmartProperties from './SmartProperties';

const { TabPane } = Tabs;

function Properties ({activeProject, smartProperties, fetchSmartProperties, deleteSmartProperty}) {

    const [selectedProperty, setSelectedProperty] = useState(null);
    const [showPropertyForm, setShowPropertyForm] = useState(false);
    const [smartPropData, setSmartPropData] = useState([]);

    useEffect(() => {
        if(activeProject?.id) {
            fetchSmartProperties(activeProject.id);
        }
        
    }, [activeProject]);

    useEffect(() => {
        const smrtProperties = [];
        smartProperties.forEach((prop) => {
            //harcoded type
            smrtProperties.push({name: prop.name, type: prop.type_alias, actions: prop})
        })
        setSmartPropData(smrtProperties);
    }, [smartProperties])

    const columns = [

        {
          title: 'Diplay name',
          dataIndex: 'name',
          key: 'name', 
          render: (text) => <span className={'capitalize'}>{text}</span>
        },
        {
          title: 'Type',
          dataIndex: 'type',
          key: 'type', 
          render: (text) => <span className={'capitalize'}>{text.replace('_', ' ')}</span>
        },
        {
            title: '',
            dataIndex: 'actions',
            key: 'actions', 
            render: (obj) => (
                <div className={`flex justify-end`}>
                    <Dropdown overlay={() => menu(obj)} trigger={['click']}>
                    <Button size={'large'} type="text" icon={<MoreOutlined />} />
                    </Dropdown>
                </div>
              )
          }
    ];

    const menu = (obj) => {
        return (
        <Menu>
          <Menu.Item key="0" onClick={() => confirmRemove(obj.id)}>
            <a>Remove Property</a>
          </Menu.Item>
          <Menu.Item key="0" onClick={() => editProp(obj)}>
            <a>Edit Property</a>
          </Menu.Item>
        </Menu>
        );
    };

    const editProp = (obj) => {
        setSelectedProperty(obj);
        setShowPropertyForm(true)
    }

    const confirmRemove = (id) => {
        deleteSmartProperty(activeProject.id, id).then(res => {
            fetchSmartProperties(activeProject.id);
            notification.success({
                message: "Success",
                description: "Deleted property successfully ",
                duration: 5,
            })
        }, err => {
            notification.error({
                message: "Error",
                description: err.data.error,
                duration: 5,
            })
        });
    }

    const renderSmartPropertyTable = () => {
        return (
            <>
            <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Properties</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} onClick={() => setShowPropertyForm(true)}><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-4'}>
            <Col span={24}>  
            <div className={'mt-6'}>
                <Tabs defaultActiveKey="1" >
                            <TabPane tab="Smart Properties" key="1">
                                    <Table className="fa-table--basic mt-4" 
                                    columns={columns} 
                                    dataSource={smartPropData} 
                                    pagination={false}
                                    />
                            </TabPane>
                </Tabs> 
            </div>  
        </Col> 
        </Row>
        </>
        )
    }

    const renderSmartPropertyDetail = () => {
        return (
            <SmartProperties smartProperty={selectedProperty}  setShowSmartProperty={(showVal) => {
                setShowPropertyForm(showVal);
                setSelectedProperty(null);
                fetchSmartProperties(activeProject.id);
            }}></SmartProperties>
        )
    }
    
    return (<div className={'mb-10 pl-4'}>
        {!showPropertyForm ? renderSmartPropertyTable() : renderSmartPropertyDetail()}
    </div>)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    smartProperties: state.settings.smartProperties
  });

export default connect(mapStateToProps, {fetchSmartProperties, deleteSmartProperty})(Properties);