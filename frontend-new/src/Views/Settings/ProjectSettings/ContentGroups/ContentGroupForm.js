import React, { useState, useEffect } from 'react';
import {
  Row, Col, Form, Input, Button, Tabs, Select, message, Radio, Menu, Dropdown
} from 'antd'; 
import { DownOutlined, PlusOutlined } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents'; 
import { connect } from 'react-redux'; 
import {saveSmartEvents, fetchSmartEvents, fetchObjectPropertiesbySource, fetchSpecificPropertiesValue} from 'Reducers/events';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware'; 
import _ from 'lodash';
import AddEditValue from './AddEditValue';
const { TabPane } = Tabs; 
const { Option, OptGroup } = Select;

function ContentGroupsForm({smart_events, objPropertiesSource, specificPropertiesData, fetchSmartEvents, fetchSpecificPropertiesValue, fetchObjectPropertiesbySource, setShowSmartEventForm, saveSmartEvents, activeProject, events}) { 
    
    const [loading, setLoading] = useState(false); 
    const [showAddValueModal, setshowAddValueModal] = useState(false);
    
    // Specific Rules
    const [currOperator, setCurrOperator] = useState('EQUAL'); 
    const [lastOperator, setLastOperator] = useState('EQUAL');
    const [currVal, setCurrVal] = useState('');
    const [lastVal, setLastVal] = useState(''); 
    
    const [form] = Form.useForm();
     
    const onChange = () => {
        seterrorInfo(null);
      };
    
    const postDataFormat = {
        "expr": {
          "description": "string",
          "filters": [
            {
              "logical_op": "AND",
              "property_name": "string",
              "rules": [
                 
              ]
            }
          ],
          "logical_op": "AND",
          "object_type": "salesforce",
          "property_evaluation_type": "specific",
          "source": "salesforce",
          "timestamp_reference_field": "string"
        },
        "name": "string"
      };

    
    const onFinish = data => {
        setLoading(true); 
        const finalData = {
            ...postDataFormat,
            "name": data.name,
            expr: {
                ...postDataFormat.expr,
                description: data.description,
                property_evaluation_type: data.property_evaluation_type,
                source: data.source,
                object_type: data.object_type,
                timestamp_reference_field: data.timestamp_reference_field === 'other' ?  data.datetime_objProperty : data.timestamp_reference_field,
                filters: [
                    {
                      "logical_op": "AND",
                      "property_name": data.property_name,
                      "rules": data.property_evaluation_type == 'any' ? [] : [
                        {
                          "gen": "curr",
                          "op": currOperator,
                          "value": currVal
                        },
                        {
                          "gen": "last",
                          "op": lastOperator,
                          "value": lastVal
                        }
                      ]
                    }
                  ], 

            }

        }  

        saveSmartEvents(activeProject.id,finalData).then((data)=>{ 
            message.success('Custom Event Added!');
            fetchSmartEvents(activeProject.id);
            setShowSmartEventForm(false);
            setLoading(false); 
            }).catch((err)=>{
              console.log("SmartEventsSave catch",err);
              const ErrMsg = err?.data?.error ? err.data.error : `Oops! Something went wrong!`;
              message.error(ErrMsg); 
              setLoading(false); 
          }); 

      };
    

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
                        loading={true}
                        //   initialValues = {{ 
                        //     event_type: 'crm', 
                        //     property_evaluation_type: 'any',
                        //   }}
                        >
                            <Row>
                                <Col span={12}>
                                    <Text type={'title'} level={4} weight={'bold'} extraClass={'m-0'}><PlusOutlined style={{color:'grey', fontSize:'16px',marginRight:'8px'}}/> New Content Group</Text>
                                </Col>
                                <Col span={12}>
                                    <div className={'flex justify-end'}>
                                    <Button size={'large'} disabled={loading} onClick={() => setShowSmartEventForm(false)}>Cancel</Button>
                                    <Button size={'large'} disabled={loading}  className={'ml-2'} type={'primary'}  htmlType="submit">Save</Button>
                                    </div>
                                </Col>
                            </Row> 
                                            
                            <Row className={'mt-8'}>
                                <Col span={18}>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Name</Text>
                                <Form.Item
                                        name="name"
                                        rules={[{ required: true, message: 'Please input display name.' }]}
                                >
                                <Input disabled={loading} size="large" className={'fa-input w-full'} />
                                        </Form.Item>
                                </Col> 
                            </Row>

                            <Row className={'mt-6'}>
                                <Col span={18}>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>Description </Text>
                                <Form.Item
                                    name="description" 
                                >
                                <Input disabled={loading} size="large" className={'fa-input w-full'} />
                                </Form.Item>
                                </Col> 
                            </Row>

                            <Row className={'mt-6'}>
                                <Col span={18}>
                                    <div className={'border-top--thin pt-5 mt-5'}>
                                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Values</Text> 

                                    </div>
                                </Col>
                                <Col span={4}>
                                    <div className={'flex justify-end border-top--thin pt-5 mt-5'}>
                                        <Button type={'text'} size={'middle'} onClick={()=> setshowAddValueModal(true)}><SVG name={'plus'} extraClass={'m-0'} size={18} />New value</Button>
                                    </div>
                                </Col>
                            </Row>
                            <Row className={'mt-8'}>
                                <Col span={22}>
                                    <div className={'flex justify-center'}>
                                        <img src='assets/images/NoData.png' className={'w-20'}/>
                                    </div>
                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 flex justify-center mt-4'}>Create Values that defines this content group</Text>
                                    <div className={'flex justify-center mt-4'}>
                                        <Button size={'middle'} ><SVG name={'plus'} extraClass={'m-0'} size={18} />Add new value</Button>
                                    </div>
                                </Col>
                            </Row>
                
                        </Form>
            </div>  
        </Col> 
        </Row> 
        
        {/* Add/Edit value modal */}
        <AddEditValue visible={showAddValueModal} setshowAddValueModal={setshowAddValueModal}/>
    </> 
  );
}

const mapStateToProps = (state) => ({
    smart_events: state.events.smart_events,
    objPropertiesSource: state.events.objPropertiesSource,
    specificPropertiesData: state.events.specificPropertiesData,
    activeProject: state.global.active_project, 
    events: state.coreQuery.eventOptions
  });

  export default connect(mapStateToProps, {saveSmartEvents, fetchSmartEvents, fetchEventNames, fetchSpecificPropertiesValue, fetchObjectPropertiesbySource})(ContentGroupsForm); 