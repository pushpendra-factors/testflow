import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Progress, message, Select, Popconfirm, Upload
} from 'antd';
import { ExclamationCircleFilled } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import { createProjectWithTimeZone, udpateProjectDetails } from 'Reducers/global';
import { TimeZoneOffsetValues } from 'Utils/constants'; 
import Congrates from './Congrates';
import 'animate.css';
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

function BasicDetails({ createProjectWithTimeZone, activeProject, handleCancel, udpateProjectDetails }) {
  const [form] = Form.useForm();
  const [formData, setFormData] = useState(null);
  const [showProfile, setShowProfile] = useState(false);
  const [imageUrl, setImageUrl] = useState('');

  const onFinish = values => {
       let projectData = {
        name: values.projectName,
        time_zone: TimeZoneOffsetValues[values.time_zone]?.city
      }; 
      createProjectWithTimeZone(projectData).then((res) => {
        const projectId = res.data.id;
        if(imageUrl) {
          udpateProjectDetails(projectId, {'profile_picture':imageUrl}).then(() => {
            message.success('Profile Image Uploaded')
            setFormData(projectData);
          }).catch((err) => {
            message.error('error:',err)
          })
        } else {
          setFormData(projectData);
        }
        message.success('New Project Created!');
      }).catch((err) => {
        message.error('Oops! Something went wrong.');
        console.log('createProject Failed:', err);
      });
  };

//   const onSkip = () => {
//     form.resetFields();
//     setFormData(true)
//   };

  function getBase64(img, callback) {
    const reader = new FileReader();
    reader.addEventListener('load', () => callback(reader.result));
    reader.readAsDataURL(img);
  }
  
  function beforeUpload(file) {
    const isJpgOrPng = file.type === 'image/jpeg' || file.type === 'image/png';
    if (!isJpgOrPng) {
      message.error('You can only upload JPG/PNG file!');
    }
    const isLt2M = file.size / 1024 / 1024 < 2;
    if (!isLt2M) {
      message.error('Image must smaller than 2MB!');
    }
    return isJpgOrPng && isLt2M;
  }

  const handleChange = info => {
    if (info.file.status === 'uploading') {
      return;
    }
    if (info.file.status === 'done') {
      // Get this url from response in real world.
      getBase64(info.file.originFileObj, imageUrl => {
        setImageUrl(imageUrl);
      });
    }
  };

  return (
    <>
    {!formData &&
      <div className={'fa-container'}>
            <Row justify={'center'}>
                <Col span={7} >
                    <div className={'flex flex-col justify-center mt-16'}>
                        <Row className={'mb-4'}>
                            <Col span={24} >
                                <Text type={'title'} level={3} color={'grey-2'} align={'center'} weight={'bold'}>Basic Details</Text>
                                {/* <Progress percent={33.33} strokeWidth={3} showInfo={false} /> */}
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
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 ml-1 mb-1'}>Project Name</Text>
                            <Form.Item
                                label={null}
                                name="projectName"
                                rules={[{ required: true, message: 'Please input your Project Name!' }]}
                            >
                            <Input className={'fa-input'} size={'large'} placeholder={'eg. My Company Name'} />
                            </Form.Item>
                        </Col>
                        <Col span={24}>
                            <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 ml-1 mt-6 mb-1'}>Select timezone</Text>
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
                            <Text type={'title'} size={10} color={'grey'} extraClass={'inline m-0 mt-4 ml-1 mb-2'}>This must reflect the same timezone as in your CRM</Text>
                            <Popconfirm placement="rightTop" title={<Text type={'title'} size={10} extraClass={'max-w-xs'}>This must reflect the same timezone as used in your CRM. Once selected, this action cannot be edited.</Text>} icon={<ExclamationCircleFilled style={{color:'#1E89FF'}}/>} okText="Got it" cancelText="Learn More" cancelButtonProps={{ type: 'text', style:{color:'#1E89FF', display:'none'}}}>
                                <Button type={'text'} className={'m-0'} style={{backgroundColor:'white'}}><SVG name={'infoCircle'} size={18} color="gray"/></Button>
                            </Popconfirm>
                        </Col>
                        {showProfile?
                        <Col span={24}>
                            <Row className={'animate__animated animate__fadeIn mt-4 border-t'}>
                                <Col span={6} className={'mt-6'}>
                                    <Upload
                                        name="avatar"
                                        accept={''}
                                        showUploadList={false}
                                        action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
                                        beforeUpload={beforeUpload}
                                        onChange={handleChange}
                                    >
                                        {imageUrl ? <img src={imageUrl} alt="avatar" /> : <img src='../../../../assets/avatar/upload-picture.png'></img>}
                                    </Upload>
                                </Col>
                                <Col span={17} className={'mt-6'}>
                                    <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>Upload project thumbnail</Text>
                                    <Text type={'title'} size={10} extraClass={'inline m-0'} color={'grey'}>You can skip this now and do it later by going to the project settings.</Text>
                                </Col>
                                <Col span={1} className={'mt-6'}>
                                    <Button type='text' className={'m-0'} onClick={() => setShowProfile(false)}><SVG name={'times'} size={12} /></Button>
                                </Col>
                            </Row>
                        </Col>
                        : 
                        <Col span={24}>
                            <div className={'mt-2'}>
                                <Text type={'title'} size={8} color={'grey'} extraClass={'max-w-md m-0 ml-1'}>A logo helps personalize your Project. <a onClick={() => setShowProfile(true)}>Upload project thumbnail</a></Text>
                            </div>
                        </Col>
                        }
                        <Col span={24}>
                            <div className={'mt-8 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type="primary" style={{width:'27vw', height:'36px'}} className={'m-0'} htmlType="submit">
                                    Create
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col>
                        {/* <Col span={24}>
                            <div className={'mt-4 flex justify-center'}>
                                <Form.Item className={'m-0'}>
                                    <Button size={'large'} type={'text'} style={{width:'28vw', height:'36px'}} htmlType="text" onClick={onSkip}>
                                    Skip
                                    </Button>
                                </Form.Item>
                            </div>
                        </Col> */}
                    </Row>
                    </Form>
                        
                        </Col>
                        {/* <Col span={24} className={'mt-8'}>
                            <a href='#!'><Text type={'title'} level={6} align={'center'} weight={'bold'} color={'brand-color'}>or Explore our demo project for now</Text></a>
                        </Col> */}
                        </Row>
                    </div>
                </Col>
            </Row>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
      </div>
    }
    {formData && <Congrates handleCancel = {handleCancel} />}
    </>

  );
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
  });

export default connect(mapStateToProps, { createProjectWithTimeZone, udpateProjectDetails })(BasicDetails);
