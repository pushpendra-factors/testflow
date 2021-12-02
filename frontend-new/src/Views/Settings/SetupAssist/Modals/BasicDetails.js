import React, { useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Progress, message, Select
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { createProject, udpateProjectDetails } from '../../../../reducers/global';
import { TimeZoneOffsetValues } from 'Utils/constants'; 
import InviteMembers from './InviteMembers';
const { Option } = Select;

const getKeyByValue = (obj, value) =>  Object.keys(obj).find(key => obj[key]?.city === value);

const TimeZoneName = 
{
  "IST":'IST',
  "PT" :'PT (Pacific Time)',
  "CT" :'CT (Central Time)',
  "ET" :'ET (Eastern Time)',
  "GMT" :'GMT',
  "AEST" :'AEST (Australia Eastern Standard Time)', 
}

function BasicDetails({ createProject, activeProject, setEditMode, udpateProjectDetails, handleCancel }) {
  const [form] = Form.useForm();
  const [formData, setFormData] = useState(null);

  const onFinish = values => {
      createProject(values.projectName).then(() => {
        setFormData(values);
        message.success('New Project Created!');
      }).catch((err) => {
        message.error('Oops! Something went wrong.');
        console.log('createProject Failed:', err);
      });

    // let projectData = {
    //     ...values,
    //     time_zone: TimeZoneOffsetValues[values.time_zone]?.city
    //   }; 
  
    //   udpateProjectDetails(activeProject.id, projectData).then(() => {
    //     message.success('Project details updated!'); 
    //     setEditMode(false);
    //   }).catch((err) => {
    //     console.log('err->', err);
    //     message.error(err.data.error); 
    //   });
  };

//   const onSkip = () => {
//     form.resetFields();
//     setFormData(true)
//   };

  return (
    <>
    {!formData &&
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={7} >
                    <div className={'flex flex-col justify-center mt-20'}>
                        <Row className={'mb-20'}>
                            <Col span={24} >
                                <Text type={'title'} level={3} color={'grey-2'} weight={'bold'}>Basic Details</Text>
                                <Progress percent={33.33} strokeWidth={3} showInfo={false} />
                            </Col>
                        </Row>
                        <Row>
                    <Col span={24}>
                    <Form
                        name="createNewProject"
                        onFinish={onFinish}
                        form={form}
                        initialValues={{ time_zone: TimeZoneName[getKeyByValue(TimeZoneOffsetValues,activeProject?.time_zone)] }}
                    >
                    <Row>
                        <Col span={24}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mb-2'}>Project name</Text>
                            <Form.Item
                                label={null}
                                name="projectName"
                                rules={[{ required: true, message: 'Please input your Project Name!' }]}
                            >
                            <Input className={'fa-input'} size={'large'} placeholder={'eg. Marketing Analytics'} />
                            </Form.Item>
                        </Col>
                        <Col span={24}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mt-6 mb-2'}>Select Your timezone</Text>
                            <Form.Item
                                name="time_zone"
                                className={'m-0'}
                                rules={[{ required: true, message: 'Please choose an option' }]}
                                // disabled={!activeProject?.is_multiple_project_timezone_enabled}
                            >
                                <Select 
                                // disabled={!activeProject?.is_multiple_project_timezone_enabled}
                                className={'fa-select'} placeholder={'Time Zone'} size={'large'}>
                                { Object.keys(TimeZoneName).map((item)=>{
                                    return  <Option value={item}>{TimeZoneName[item]}</Option> 
                                })} 
                                </Select>
                            </Form.Item>
                        </Col>
                        <Col span={24}>
                            <div className={'mt-20 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type="primary" style={{width:'280px', height:'36px'}} className={'ml-2'} htmlType="submit">
                                    Next
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col>
                        {/* <Col span={24}>
                            <div className={'mt-4 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type={'text'} style={{width:'280px', height:'36px'}} htmlType="text" onClick={onSkip}>
                                    Skip
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col> */}
                    </Row>
                    </Form>
                        
                        </Col>
                        <Col span={24} className={'mt-20'}>
                            <Text type={'title'} level={6} align={'center'} color={'grey-2'}>or Explore our demo project for now</Text>
                        </Col>
                        </Row>
                    </div>
                </Col>
            </Row>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
      </div>
    }
    {formData && <InviteMembers handleCancel = {handleCancel} />}
    </>

  );
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project
  });

export default connect(mapStateToProps, { createProject, udpateProjectDetails })(BasicDetails);
