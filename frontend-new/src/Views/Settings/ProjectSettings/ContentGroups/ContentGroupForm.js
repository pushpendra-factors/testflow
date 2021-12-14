import React, { useState, useEffect } from 'react';
import {
  Row, Col, Form, Input, Button, Tabs, Select, message, Table, Menu, Dropdown, notification
} from 'antd'; 
import { MoreOutlined, PlusOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents'; 
import { connect } from 'react-redux'; 
import {addContentGroup, updateContentGroup} from 'Reducers/global';
import _ from 'lodash';
import AddEditValue from './AddEditValue';

function ContentGroupsForm({activeProject, setShowSmartForm, addContentGroup, updateContentGroup}) { 
  const [form] = Form.useForm();
  const [formState, setFormState] = useState('add');
  const [showAddValueModal, setshowAddValueModal] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);
  const [smartPropState, setSmartPropState] = useState({});
  const [rulesState, setRulesState] = useState([]);
  const [rulesData, setRulesData] = useState([]);
     

      const renderRuleViewButtons = (rules) => {
        return rules.map((obj, i) => {
            return (<div className={`flex justify-center ${i > 0 && 'mt-4'}`}>
            <Button 
                type={'text'}
                size={'large'}
                style={{color:'gray'}}
                className={`fa-button--truncate pointer-events-none`} 
                > {obj?.lop} 
            </Button>
    
            <Button 
                size={'large'}
                className={`fa-button--truncate ml-2 pointer-events-none`} 
                > {obj?.op} 
            </Button>
    
            <Button 
                size={'large'}
                className={`fa-button--truncate ml-1 pointer-events-none`} 
                > {obj?.va} 
            </Button>
        </div>)
        })
    }

      const columns = [

        {
          title: 'Order',
          dataIndex: 'content_group_value',
          key: 'content_group_value', 
          render: (text) => <span className={'capitalize'}>{text}</span>
        },
        {
          title: 'Rule',
          dataIndex: 'rule',
          key: 'rule',
          align: 'center', 
          render: (rules) => renderRuleViewButtons(rules)
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
          {/* <Menu.Item key="0" onClick={() => confirmRemove(obj)}>
            <a>Remove</a>
          </Menu.Item>
          <Menu.Item key="0" onClick={() => editProp(obj)}>
            <a>Edit</a>
          </Menu.Item> */}
        </Menu>
        );
    };


    
    const createForm = (smrtProp) => {
      addContentGroup(activeProject.id, smrtProp).then(res => {
            smrtProp.id = res.data.id;
            setSmartPropState({...smrtProp});
            setFormState('view');
            setShowModalVisible(false);
            notification.success({
                message: "Success",
                description: "Custom Dimension rules created successfully ",
                duration: 5,
              });
        }, err => {
            notification.error({
                message: "Error",
                description: err.data.error,
                duration: 5,
              });
        });
    }

    const updateForm = (smrtProp) => {
      updateContentGroup(activeProject.id, smrtProp).then(res => {
            smrtProp.id = res.data.id;
            setSmartPropState({...smrtProp});
            setRulesState(smrtProp.rule);
            setFormState('view');
            setShowModalVisible(false);
            notification.success({
                message: "Success",
                description: "Custom Dimension rules updated successfully ",
                duration: 5,
              });
        }, err => {
            notification.error({
                message: "Error",
                description: err.data.error,
                duration: 5,
              });
        });
    }

    const onFinish = (data) => {
        if(data) {
            // Save with data
            // Close modal
            const smrtProp = {project_id: activeProject.id, content_group_name: data.content_group_name, content_group_description: data.content_group_description, rule:rulesState};
            console.log(smrtProp);
            if(formState !== 'add') {
                updateForm(smrtProp);
            } else {
                delete smrtProp.id;
                createForm(smrtProp)
            }
              
        }
    }

    const onChange = () => {
    };

    const handleValuesSubmit = (data) => {
      const rule = {...data};
      const rulesToUpdate = [...rulesState.filter((rl) => JSON.stringify(rl) !== JSON.stringify(data))];
      rulesToUpdate.push(rule);
      setRulesState(rulesToUpdate);
      setshowAddValueModal(false);
      console.log(data);
    }

    const handleCancel = () => {
      setshowAddValueModal(false)
    }
    

    useEffect(() => {
      const columData = [];
      console.log(rulesData)
      rulesState.forEach((rl) => {
          columData.push({value: rl.content_group_value, rule: rl.rule, actions: rl});
      })
      setRulesData(rulesState);
  }, [rulesState])

  return (
    <>
     
        <Row>
            <Col span={24}>  
            <div> 
                    <Form
                        form={form}
                        onFinish={onFinish}
                        className={'w-full'}
                        onChange={onChange}
                        >
                            <Row>
                                <Col span={12}>
                                    <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}><PlusOutlined style={{color:'grey', fontSize:'16px',marginRight:'8px'}}/> New Content Group</Text>
                                </Col>
                                <Col span={12}>
                                    <div className={'flex justify-end'}>
                                    <Button size={'large'} onClick={() => setShowSmartForm(false)}>Cancel</Button>
                                    <Button size={'large'} className={'ml-2'} type={'primary'}  htmlType="submit">Save</Button>
                                    </div>
                                </Col>
                            </Row> 
                                            
                            <Row className={'mt-8'}>
                                <Col span={18}>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Name</Text>
                                <Form.Item
                                        name="content_group_name"
                                        rules={[{ required: true, message: 'Please input display name.' }]}
                                >
                                <Input size="large" className={'fa-input w-full'} />
                                        </Form.Item>
                                </Col> 
                            </Row>

                            <Row className={'mt-6'}>
                                <Col span={18}>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Description </Text>
                                <Form.Item
                                    name="content_group_description" 
                                    rules={[{ required: true, message: 'Please input description.' }]}
                                >
                                <Input size="large" className={'fa-input w-full'} />
                                </Form.Item>
                                </Col> 
                            </Row>

                            <Row className={'mt-6'}>
                                <Col span={12}>
                                    <div className={'border-top--thin pt-5 mt-5'}>
                                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Values</Text> 

                                    </div>
                                </Col>
                                <Col span={12}>
                                    <div className={'flex justify-end border-top--thin pt-5 mt-5'}>
                                        <Button type={'text'} size={'middle'} onClick={()=> setshowAddValueModal(true)}><SVG name={'plus'} extraClass={'m-0'} size={18} />New value</Button>
                                    </div>
                                </Col>
                            </Row>
                            {!rulesData[0] &&
                            <Row className={'mt-8'}>
                                <Col span={24}>
                                    <div className={'flex justify-center'}>
                                        <img src='assets/images/NoData.png' className={'w-20'}/>
                                    </div>
                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 flex justify-center mt-4'}>Create Values that defines this content group</Text>
                                    <div className={'flex justify-center mt-4'}>
                                        <Button size={'middle'} onClick={()=> setshowAddValueModal(true)}><SVG name={'plus'} extraClass={'m-0'} size={18} />Add new value</Button>
                                    </div>
                                </Col>
                            </Row>
                            }
                            {rulesData[0] &&
                            <Row>
                                <Col span={24}>
                                <Table className="fa-table--basic mt-2" 
                                columns={columns} 
                                dataSource={rulesData} 
                                pagination={false}
                                />
                                </Col>
                            </Row>
                            }           
                        </Form>
            </div>  
        </Col> 
        </Row> 
        
        {/* Add/Edit value modal */}
        <AddEditValue visible={showAddValueModal} handleCancel={handleCancel} submitValues={handleValuesSubmit}/>
    </> 
  );
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project, 
  });

  export default connect(mapStateToProps, {addContentGroup, updateContentGroup})(ContentGroupsForm); 