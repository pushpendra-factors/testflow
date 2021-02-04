import React, { useState, useEffect } from 'react';
import {
  Row, Col, Form, Input, Button, Tabs, Select, message, Radio
} from 'antd';
import { Text } from 'factorsComponents'; 
import { connect } from 'react-redux'; 
import {saveSmartEvents, fetchSmartEvents, fetchObjectPropertiesbySource} from 'Reducers/events';
import { fetchEventNames, getUserProperties } from 'Reducers/coreQuery/middleware'; 
import _ from 'lodash';
const { TabPane } = Tabs; 
const { Option, OptGroup } = Select;

function SmartEventsForm({smart_events, objPropertiesSource, fetchSmartEvents, fetchObjectPropertiesbySource, setShowSmartEventForm, saveSmartEvents, activeProject, events}) { 
    
    const [loading, setLoading] = useState(false); 
    const [errorInfo, seterrorInfo] = useState(null);
    const [dataObjectSource, setDataObjectSource] = useState(null);
    const [dataObject, setDataObject] = useState(null);
    const [dataObjectProperty, setDataObjectProperty] = useState('');
    const [timestampReferenceOthers, setTimestampReferenceOthers] = useState(false);
    const [objPropertiesSourceArr, setobjPropertiesSourceArr] = useState(null);
    const [objPropertiesSourceArrDT, setobjPropertiesSourceArrDT] = useState(null);
    
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
                {
                  "gen": "curr",
                  "op": "EQUAL",
                  "value": {}
                }
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
                          "op": "EQUAL",
                          "value": {}
                        }
                      ]
                    }
                  ], 

            }

        } 
        saveSmartEvents(activeProject.id,finalData).then((data)=>{ 
            message.success('Smart Event Added!');
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
      const settimestampReference = (e) => {
          if(e.target.value == 'other'){
            setTimestampReferenceOthers(true)
        }
        else{
              setTimestampReferenceOthers(false) 
          }
      }

      const dataObjectConstants =  dataObjectSource === 'hubspot' ? ['contact', 'deal'] : ['account', 'contact', 'lead'] 


      const onSelectDataObjectChange = (value) =>{ 
        setDataObjectProperty('');
        setDataObject(value);
      }
      const onSelectObjectProperty =(value) =>{  
        setDataObjectProperty(value)
      }
 

      useEffect(()=>{ 
        setLoading(true); 
        fetchObjectPropertiesbySource(activeProject.id,dataObjectSource, dataObject).then((data)=>{ 
            setLoading(false); 
        }).catch((err)=>{    
        const ErrMsg = err?.data?.error ? err.data.error : `Oops! Something went wrong!`;
        message.error(ErrMsg); 
        setLoading(false);  
        });
      },[dataObjectSource,dataObject ])

      useEffect(()=>{
          let objPropArr = []; 
          let objPropArrDateTime = []; 
          objPropertiesSource && Object.keys(objPropertiesSource)?.map((key) =>   {
              if(!_.isEmpty(objPropertiesSource[key])){ 
                    objPropertiesSource[key]?.sort().map((item)=>{
                      objPropArr = [...objPropArr, item];
                    })
                    if(key=='datetime'){
                        objPropertiesSource[key]?.sort().map((item)=>{
                            objPropArrDateTime = [...objPropArrDateTime, item];
                        }) 
                    }
                } 
            });
            setobjPropertiesSourceArr(objPropArr);
            setobjPropertiesSourceArrDT(objPropArrDateTime);
                                            

      },[objPropertiesSource]) 

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
                          initialValues = {{ 
                            event_type: 'crm', 
                            property_evaluation_type: 'any',
                          }}
                        >
                            <Row>
                                <Col span={12}>
                                    <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>New Smart Event</Text>
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
                                <Text type={'title'} level={7} extraClass={'m-0'}>Display Name</Text>
                                <Form.Item
                                        name="name"
                                        rules={[{ required: true, message: 'Please input displayt name.' }]}
                                >
                                <Input disabled={loading} size="large" className={'fa-input w-full'} placeholder="Display Name" />
                                        </Form.Item>
                                </Col> 
                            </Row>

                            <Row className={'mt-8'}>
                                <Col span={18}>
                                <Text type={'title'} level={7} extraClass={'m-0'}>Description </Text>
                                <Form.Item
                                    name="description" 
                                >
                                <Input disabled={loading} size="large" className={'fa-input w-full'} placeholder="Description" />
                                </Form.Item>
                                </Col> 
                            </Row>

                            <Row className={'mt-8'}>
                                <Col span={18}>
                                    <Text type={'title'} level={7} extraClass={'m-0'}>Event type</Text>
                                    <Form.Item
                                    name="event_type"
                                    className={'m-0'}  
                                    >
                                    <Select className={'fa-select w-full'} disabled size={'large'}>
                                        <Option value="crm">CRM transition based event</Option> 
                                    </Select>
                                    </Form.Item>
                                </Col>
                            </Row>


                            <Row className={'mt-8'}>
                                <Col span={18}>
                                    <div className={'border-top--thin pt-5 mt-5'}>
                                        <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Event Rule</Text> 

                                    </div>
                                </Col>
                            </Row>
                            <Row className={'mt-8'}>
                                <Col span={18}>
                                    <Text type={'title'} level={7} extraClass={'m-0'}>Select data source</Text>
                                    <Form.Item
                                    name="source"
                                    className={'m-0'} 
                                    rules={[{ required: true, message: 'Please select a data source.' }]}
                                    >
                                    <Select className={'fa-select w-full'} onChange={(value)=>setDataObjectSource(value)} placeholder={'Select data source'} size={'large'}>
                                        <Option value="hubspot">Hubspot</Option> 
                                        <Option value="salesforce">Salesforce</Option> 
                                    </Select>
                                    </Form.Item>
                                </Col>
                            </Row>
                            {dataObjectSource &&
                            <Row className={'mt-8'}>
                                <Col span={18}>
                                    <Text type={'title'} level={7} extraClass={'m-0'}>Select data Object</Text>
                                    <Form.Item
                                    name="object_type"
                                    className={'m-0'} 
                                    rules={[{ required: true, message: 'Please select a data object.' }]}
                                    >
                                    <Select onChange={(value)=>{onSelectDataObjectChange(value)}} className={'fa-select w-full'} placeholder={'Select data Object'} size={'large'}>
                                        {
                                        dataObjectConstants?.map((item)=>{ 
                                            return <Option key={item} value={item}>{item}</Option>  
                                        })} 
                                    </Select>
                                    </Form.Item>
                                </Col>
                            </Row>
                            }
                            <Row className={'mt-8'}>
                                <Col span={18}> 
                                    <Text type={'title'} level={7} extraClass={'m-0'}>Rule for evaluation</Text>
                                    <Form.Item
                                    name="property_evaluation_type"
                                    className={'m-0'} 
                                    > 
                                                <Radio.Group>
                                                    <Radio value={'any'}>Any change in property</Radio>
                                                    <Radio value={'specific'} disabled>Specific change in property</Radio> 
                                                </Radio.Group> 
                                    </Form.Item>
                                </Col>
                            </Row>
                            {dataObject && <>
                            <Row className={'mt-8'}>
                                <Col span={18}>
                                    <Text type={'title'} level={7} extraClass={'m-0'}>Select object property</Text>
                                    <Form.Item
                                    name="property_name"
                                    className={'m-0'}
                                    rules={[{ required: true, message: 'Please select an object property.' }]}
                                    >
                                    <Select 
                                    value={dataObjectProperty} 
                                    onChange={(value)=>{onSelectObjectProperty(value)}}
                                     className={'fa-select w-full'} 
                                     placeholder={'Select object property'} 
                                     size={'large'}
                                     showSearch
                                     optionFilterProp="children"
                                     filterOption={(input, option) =>
                                        option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                                     }
                                     filterSort={(optionA, optionB) =>
                                        optionA.children.toLowerCase().localeCompare(optionB.children.toLowerCase())
                                     }
                                    >

                                        {objPropertiesSourceArr?.sort().map((item)=>{
                                            return <Option value={item}>{item}</Option> 
                                        })}

                                    </Select>
                                    </Form.Item>
                                </Col>
                            </Row>
                            </>
                            }

                            <Row className={'mt-8'}>
                                <Col span={18}> 
                                    <Text type={'title'} level={7} extraClass={'m-0'}>Select time of event</Text>
                                    <Form.Item 
                                    name="timestamp_reference_field"
                                    className={'m-0'} 
                                    rules={[{ required: true }]}
                                    > 
                                                <Radio.Group onChange={(value)=>settimestampReference(value)}>
                                                    <Radio value={'timestamp_in_track'}>Factors recieved time</Radio>
                                                    <Radio value={'other'}>Select an object property</Radio> 
                                                </Radio.Group> 
                                    </Form.Item>
                                </Col>
                            </Row>

                            {timestampReferenceOthers &&
                            <Row className={'mt-0'}>
                                <Col span={18}> 
                                    <Form.Item
                                    name="datetime_objProperty"
                                    className={'m-0'} 
                                    rules={[{ required: true, message: 'Please select a date type property.' }]}
                                    >
                                    <Select className={'fa-select w-full mt-2'} placeholder={'List all the date type proprties  '} size={'large'}
                                        showSearch
                                        optionFilterProp="children"
                                        filterOption={(input, option) =>
                                           option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
                                        }
                                        filterSort={(optionA, optionB) =>
                                           optionA.children.toLowerCase().localeCompare(optionB.children.toLowerCase())
                                        }
                                    >

                                        {objPropertiesSourceArrDT?.sort().map((item)=>{
                                            return <Option value={item}>{item}</Option> 
                                        })}

                                    </Select>
                                    </Form.Item>
                                </Col>
                            </Row> 
                            }

                
                        </Form>
            </div>  
        </Col> 
        </Row> 
        
    </> 
  );
}

const mapStateToProps = (state) => ({
    smart_events: state.events.smart_events,
    objPropertiesSource: state.events.objPropertiesSource,
    activeProject: state.global.active_project, 
    events: state.coreQuery.eventOptions
  });

  export default connect(mapStateToProps, {saveSmartEvents, fetchSmartEvents, fetchEventNames, fetchObjectPropertiesbySource})(SmartEventsForm); 