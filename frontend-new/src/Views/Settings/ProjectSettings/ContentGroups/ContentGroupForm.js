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

function ContentGroupsForm({activeProject, selectedGroup, setShowSmartProperty, addContentGroup, updateContentGroup}) { 
  const [form] = Form.useForm();
  const [formState, setFormState] = useState('add');
  const [showAddValueModal, setshowAddValueModal] = useState(false);
  const [selectedRule, setSelectedRule] = useState(null);
  const [smartPropState, setSmartPropState] = useState({});
  const [rulesState, setRulesState] = useState([]);
  const [rulesData, setRulesData] = useState([]);
     

      const renderRuleViewButtons = (rules) => {
        return rules.map((obj, i) => {
            return (<div className={`flex justify-start -mr-48 ${i > 0 && 'mt-4'}`}>
            <Button 
                type={'text'}
                size={'large'}
                style={{color:'gray'}}
                className={`fa-button--truncate pointer-events-none w-16`} 
                > {i==0?'URL':obj?.lop} 
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
          render: (text) => <span className={'text-gray-600 text-sm font-bold'}>{text}</span>
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
                  <Button size={'large'} type="text" icon={<MoreOutlined rotate={90}/>} />
                  </Dropdown>
              </div>
            )
        }
      ];


      const editProp = (obj) => {
        setSelectedRule(obj);
        setshowAddValueModal(true);
    }

    const confirmRemove = (obj) => {
        if (formState == 'add') {
            const rulesToUpdate = [...rulesState.filter((rl) => JSON.stringify(rl) !== JSON.stringify(obj))];
            setRulesState(rulesToUpdate);
        }

        const rulesToUpdate = [...smartPropState.rule.filter((rule) => JSON.stringify(rule) !== JSON.stringify(obj))];
        
        if(formState!=='add') {
            const smrtProp = Object.assign({}, smartPropState);
            smrtProp.rule = rulesToUpdate;
            updateForm(smrtProp);
        }
    }

    useEffect(() => {
        if(selectedGroup) {
            setSmartPropState(selectedGroup);
            setFormState('view');
            setRulesState(selectedGroup.rule);
        }
    }, [selectedGroup])

      const menu = (obj) => {
        return (
        <Menu>
          <Menu.Item key="0" onClick={() => confirmRemove(obj)}>
            <a>Remove</a>
          </Menu.Item>
          {/* <Menu.Item key="1" onClick={() => editProp(obj)}>
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
            setShowSmartProperty(false);
            setshowAddValueModal(false);
            notification.success({
                message: "Success",
                description: "Content Group rules created successfully ",
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
            // setShowSmartProperty(false);
            setshowAddValueModal(false);
            notification.success({
                message: "Success",
                description: "Content Group rules updated successfully ",
                duration: 5,
              });
        }, err => {
            setRulesState(rulesState);
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
            setShowSmartProperty(false);
            const smrtProp = {id: smartPropState.id?smartPropState.id: '', project_id: activeProject.id, content_group_name: data.content_group_name, content_group_description: data.content_group_description, rule:rulesState};
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

    const handleValuesSubmit = (data, oldRule) => {
        if(data) {
            const rule = {...data};
            const rulesToUpdate = [...rulesState.filter((rl) => JSON.stringify(rl) !== JSON.stringify(oldRule))];
            rulesToUpdate.push(rule);
            setRulesState(rulesToUpdate);
            setshowAddValueModal(false);
            setSelectedRule(null);
            if(formState === 'view') {
                const smrtProp = {id: smartPropState.id ,project_id: smartPropState.project_id, content_group_name: smartPropState.content_group_name, content_group_description: smartPropState.content_group_description, rule:rulesToUpdate};
                updateForm(smrtProp);
            }

        }
    }

    const handleCancel = () => {
      setshowAddValueModal(false)
      setSelectedRule(null);
    }
    

    useEffect(() => {
      const columData = [];
      rulesState.forEach((rl) => {
          columData.push({content_group_value: rl.content_group_value, rule: rl.rule, actions: rl});
      })
      setRulesData(columData);
  }, [rulesState])

  const renderContentGroupDeatails = () => {
      return (
          <>
            <Row className={'mt-8'}>
                <Col span={18}>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Name</Text>
                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{smartPropState.content_group_name}</Text>
                </Col> 
            </Row>

            <Row className={'mt-6'}>
                <Col span={18}>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Description </Text>
                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{smartPropState.content_group_description}</Text>
                    {formState === 'view' ? 
                    <Button size={'large'} className={'m-0 mt-2'} type={'primary'}  onClick={() => setFormState('edit')}>Edit</Button>
                    : null}
                </Col> 
            </Row>
          </>
      );
  }

  const renderContentGroupForm = () => {
      return (
          <>
          {smartPropState.content_group_name?
            <Row className={'mt-8'}>
                <Col span={18}>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Name</Text>
                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{smartPropState.content_group_name}</Text>
                </Col> 
            </Row>
            :
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
            }

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
          </>
      );
  }

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
                        initialValues={
                            {
                            content_group_name: selectedGroup?.content_group_name? selectedGroup.content_group_name : '',
                            content_group_description: selectedGroup?.content_group_description? selectedGroup.content_group_description : ''
                        }
                    }
                        >
                            <Row>
                                <Col span={12}>
                                    <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}>{formState === 'add'?<PlusOutlined style={{color:'grey', fontSize:'16px',marginRight:'8px'}}/>:null}{formState === 'add'? 'New Content Group' : 'Content Group Details' }</Text>
                                </Col>
                                <Col span={12}>
                                    <div className={'flex justify-end'}>
                                    <Button size={'large'} onClick={() => setShowSmartProperty(false)}>Cancel</Button>
                                    {/* {formState === 'view' ? 
                                    <Button size={'large'} className={'ml-2'} type={'primary'}  onClick={() => setFormState('edit')}>Edit</Button>
                                    : null} */}
                                    {formState !== 'view' ?  
                                    <Button size={'large'} className={'ml-2'} type={'primary'}  htmlType="submit">Save</Button>
                                    : <Button size={'large'} className={'ml-2'} type={'primary'}  onClick={() => setShowSmartProperty(false)}>Close</Button>}
                                    </div>
                                </Col>
                            </Row> 
                        {formState !== 'view'? renderContentGroupForm(): renderContentGroupDeatails()}
                            
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
                                        <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/NoData.png' className={'w-20'}/>
                                    </div>
                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 flex justify-center mt-4'}>Create Values that defines this content group</Text>
                                    <div className={'flex justify-center mt-8'}>
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
        {showAddValueModal && <AddEditValue selectedRule={selectedRule} handleCancel={handleCancel} submitValues={handleValuesSubmit}/>}
    </> 
  );
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project, 
  });

  export default connect(mapStateToProps, {addContentGroup, updateContentGroup})(ContentGroupsForm); 