import React, { useState, useEffect, useCallback } from 'react';
import { connect } from 'react-redux';

import { Text, SVG } from 'factorsComponents';
import {
    Row, Col, Button, Tabs, Table, Dropdown, Menu, notification, Tooltip
} from 'antd';
import { MoreOutlined } from "@ant-design/icons";

import { fetchSmartProperties, deleteSmartProperty } from 'Reducers/settings/middleware';
import SmartProperties from './SmartProperties';
import DCG from './DCG';
import DCGTable from './DCG/DCGTable';
import PropetyValueModalDCG from './DCG/PropetyValueModalDCG';

import ConfirmationModal from "../../../../components/ConfirmationModal";

const { TabPane } = Tabs;

function Properties({ 
    activeProject,
    smartProperties,
    fetchSmartProperties,
    deleteSmartProperty,
    agents,
    currentAgent,
 }) {

    const [selectedProperty, setSelectedProperty] = useState(null);
    const [showPropertyForm, setShowPropertyForm] = useState(false);
    const [showDCGForm, setShowDCGForm] = useState(false);
    const [smartPropData, setSmartPropData] = useState([]);
    const [deleteWidgetModal, showDeleteWidgetModal] = useState(false);
    const [deleteApiCalled, setDeleteApiCalled] = useState(false);

    const [editProperty, setEditProperty] = useState(null);
    const [isModalVisible, setShowModalVisible] = useState(false);

    const [tableLoading, setTableLoading] = useState(false);
    const [enableEdit, setEnableEdit] = useState(false);

    useEffect(() => {
        setEnableEdit(false);
        agents && currentAgent && agents.map((agent) => {
          if (agent.uuid === currentAgent.uuid) {
            if (agent.role === 1) {
              setEnableEdit(true);
            }
          }
        }); 
      }, [activeProject, agents, currentAgent]);


    const [tabNo, setTabNo] = useState(1);
    useEffect(() => {
        if (activeProject?.id) {
            setTableLoading(true);
            fetchSmartProperties(activeProject.id).then(() => {
                setTableLoading(false);
            })
        }

    }, [activeProject]);

    useEffect(() => {
        const smrtProperties = [];
        smartProperties.forEach((prop) => {
            //harcoded type
            smrtProperties.push({ name: prop.name, type: prop.type_alias, actions: prop })
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
                <Menu.Item key="0" onClick={() => showDeleteWidgetModal(obj.id)}>
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
        return deleteSmartProperty(activeProject.id, id).then(res => {
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
    function callback(key) {
        setTabNo(key);
    }



    const confirmDelete = useCallback(async () => {
        try {
            setDeleteApiCalled(true);
            await confirmRemove(deleteWidgetModal);
            setDeleteApiCalled(false);
            showDeleteWidgetModal(false);
        } catch (err) {
            console.log(err);
            console.log(err.response);
        }
    }, [
        deleteWidgetModal
    ]);
    const renderSmartPropertyTable = () => {
        return (
            <>
                <Row>
                    <Col span={12}>
                        <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Properties</Text>
                    </Col>
                    <Col span={12}>
                        <div className={'flex justify-end'}>
                            {tabNo == 1 &&
                                <Button size={'large'} onClick={() => {
                                    //   setTabNo(1);
                                    setShowPropertyForm(true)
                                }}><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>

                            }
                            {tabNo == 2 &&
                            <> 
                <Tooltip placement="top" trigger={'hover'} title={enableEdit ? 'Only Admin can edit' : null}>
                  <Button size={'large'} disabled={enableEdit} onClick={() => { 
                                    setShowDCGForm(true)
                                    setShowModalVisible(true)
                                }}
                   ><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>
                </Tooltip>  
              </>
                                // <Button size={'large'} className={'ml-2'} onClick={() => {
                                //     //   setTabNo(2);
                                //     setShowDCGForm(true)
                                //     setShowModalVisible(true)
                                // }
                                // }><SVG name={'plus'} extraClass={'mr-2'} size={16} />Add New</Button>
                            }
                        </div>
                    </Col>
                </Row>
                <Row className={'mt-4'}>
                    <Col span={24}>
                        <div className={'mt-6'}>
                            <Tabs activeKey={`${tabNo}`} onChange={callback} >
                                <TabPane tab="Custom Dimensions" key="1">
                                    <Table className="fa-table--basic mt-4"
                                        columns={columns}
                                        dataSource={smartPropData}
                                        pagination={false}
                                        loading={tableLoading}
                                    />
                                </TabPane>
                                <TabPane tab="Default Channel Group" key="2">
                                    <DCGTable
                                        setEditProperty={setEditProperty}
                                        setShowModalVisible={setShowModalVisible}
                                        enableEdit={enableEdit}
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
            <SmartProperties smartProperty={selectedProperty} setShowSmartProperty={(showVal) => {
                setShowPropertyForm(showVal);
                setSelectedProperty(null);
                fetchSmartProperties(activeProject.id);
            }}></SmartProperties>
        )
    }

    return (<div className={'mb-10 pl-4'}>
        {tabNo == 1 && <>
            {!showPropertyForm ? renderSmartPropertyTable() : renderSmartPropertyDetail()}
        </>}
        {tabNo == 2 && <>
            {!showDCGForm && renderSmartPropertyTable()}

            <PropetyValueModalDCG
                isModalVisible={isModalVisible}
                setShowModalVisible={setShowModalVisible}
                setShowDCGForm={setShowDCGForm}
                setTabNo={setTabNo}
                editProperty={editProperty}
                setEditProperty={setEditProperty}
            />

            {/* {!showPropertyForm ? renderSmartPropertyTable() : renderSmartPropertyDetail()} */}
        </>} 
        <ConfirmationModal
            visible={deleteWidgetModal ? true : false}
            confirmationText="Do you really want to remove this property?"
            onOk={confirmDelete}
            onCancel={showDeleteWidgetModal.bind(this, false)}
            title="Remove Property"
            okText="Confirm"
            cancelText="Cancel"
            confirmLoading={deleteApiCalled}
        />
    </div>)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    smartProperties: state.settings.smartProperties,
    agents: state.agent.agents, 
    currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, { fetchSmartProperties, deleteSmartProperty })(Properties);