import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, Input, Form, Progress, message, Select, Popconfirm, Upload, Checkbox
} from 'antd';
import { ExclamationCircleFilled } from '@ant-design/icons';
import { Text, SVG } from 'factorsComponents';
import { createProjectWithTimeZone, udpateProjectDetails } from 'Reducers/global';
import { projectAgentInvite, fetchProjectAgents } from 'Reducers/agentActions';
import { TimeZoneOffsetValueArr, getTimeZoneNameFromCity } from 'Utils/constants';
import Congrates from './Congrates';
import 'animate.css';
import factorsai from 'factorsai';
const { Option } = Select;
import styles from './index.module.scss';
import sanitizeInputString from 'Utils/sanitizeInputString';

function BasicDetails({ createProjectWithTimeZone, activeProject, handleCancel, udpateProjectDetails, projectAgentInvite, fetchProjectAgents }) {
  const [form] = Form.useForm();
  const [formData, setFormData] = useState(null);
  const [showProfile, setShowProfile] = useState(false);
  const [imageUrl, setImageUrl] = useState('');
  const [checkbox, setcheckbox] = useState(true);
  const [loading, setloading] = useState(false);

  const onFinish = values => {
    setloading(true);
    let projectName = sanitizeInputString(values?.projectName);
    let projectData = {
      name: projectName,
      time_zone: values?.time_zone
    };

    //Factors CREATE_PROJECT_TIMEZONE tracking
    factorsai.track('CREATE_PROJECT_TIMEZONE', { 'ProjectName': projectData?.name, 'time_zone': projectData?.time_zone });

    createProjectWithTimeZone(projectData).then((res) => {
      const projectId = res.data.id;
      if (checkbox) {
        projectAgentInvite(projectId, { 'email': 'solutions@factors.ai', 'role': 2 }).then(() => {
          message.success('Invitation sent successfully!');
        }).catch((err) => {
          message.error(err);
        });
      }
      if (imageUrl) {
        udpateProjectDetails(projectId, { 'profile_picture': imageUrl }).then(() => {
          message.success('Profile Image Uploaded')
        }).catch((err) => {
          message.error('error:', err)
        })
      }
      localStorage.setItem('activeProject', projectId);
      setloading(false);
      message.success('New Project Created!');
      setFormData(projectData);
    }).catch((err) => {
      setloading(false);
      message.error(err?.data?.error);
      console.log('createProject Failed:', err);
    });
  };

  const onSkip = () => {
    handleCancel();
    handleReset();
  };

  const handleReset = () => {
    setImageUrl('');
    setcheckbox(true);
    setShowProfile(false);
    setFormData(null);
    form.resetFields();
  }

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
    // if (info.file.status === 'uploading') {
    //   return;
    // }
    // if (info.file.status === 'done') {
    // Get this url from response in real world.
    getBase64(info.file.originFileObj, imageUrl => {
      setImageUrl(imageUrl);
    });
    // }
  };

  return (
    <>
      {!formData &&
        <div className={'fa-container'}>
          <Row justify={'center'} className={`${styles.start}`}>
            <Col span={7} >
              <div className={'flex flex-col justify-center mt-14'}>
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
                      initialValues={{
                        time_zone: `${TimeZoneOffsetValueArr[0]?.name} (UTC ${TimeZoneOffsetValueArr[0]?.offset})`
                      }}
                    >
                      <Row>
                        <Col span={24}>
                          <Text type={'title'} size={10} color={'grey'} extraClass={'m-0 ml-1 mb-1'}>Project Name</Text>
                          <Form.Item
                            label={null}
                            name="projectName"
                            rules={[{ required: true, message: 'Please input your Project Name!' }]}
                          >
                            <Input className={'fa-input'} size={'large'} placeholder={'eg. My Company Name'} />
                          </Form.Item>
                        </Col>
                        <Col span={24} className={'mt-4'}>
                          <Text type={'title'} size={10} color={'grey'} extraClass={'m-0 ml-1 mb-1'}>Select timezone</Text>
                          <Form.Item
                            name="time_zone"
                            className={'m-0'}
                            rules={[{ required: true, message: 'Please choose an option' }]}
                          >
                            <Select
                              className={'fa-select'} placeholder={'Time Zone'} size={'large'}>
                              {TimeZoneOffsetValueArr?.map((item) => {
                                return <Option value={item?.city}>{`${item?.name} (UTC ${item?.offset})`}</Option>
                              })}

                            </Select>
                          </Form.Item>
                          <Text type={'title'} size={8} color={'grey'} extraClass={'inline m-0 mt-4 ml-1 mb-2'}>This must reflect the same timezone as in your CRM</Text>
                          <Popconfirm placement="rightTop" title={<Text type={'title'} size={10} extraClass={'max-w-xs'}>This must reflect the same timezone as used in your CRM. Once selected, this action cannot be edited.</Text>} icon={<ExclamationCircleFilled style={{ color: '#1E89FF' }} />} okText="Got it" cancelText="Learn More" cancelButtonProps={{ type: 'text', style: { color: '#1E89FF', display: 'none' } }}>
                            <Button type={'text'} className={'m-0'} style={{ backgroundColor: 'white' }}><SVG name={'infoCircle'} size={18} color="gray" /></Button>
                          </Popconfirm>
                        </Col>
                        <Col span={24} className={'mt-4'}>
                          <Form.Item
                            label={null}
                            name="invite_support"
                          >
                            <div className='flex items-start'>
                              <Checkbox defaultChecked={checkbox} onChange={(e) => setcheckbox(e.target.checked)} className={'mt-1'}></Checkbox>
                              <Text type={'title'} size={10} color={'grey'} extraClass={'-mt-2 ml-3 mb-2'} >Invite <span className={'font-bold'}>solutions@factors.ai</span> into this project for ongoing support</Text>
                            </div>
                          </Form.Item>
                        </Col>
                        {showProfile ?
                          <Col span={24}>
                            <Row className={'animate__animated animate__fadeIn mt-4 border-t'}>
                              <Col span={6} className={'mt-6'}>
                                <Upload
                                  name="avatar"
                                  accept={''}
                                  showUploadList={false}
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
                          <Col span={24} className={'mt-6'}>
                            <div className={'m-0'}>
                              <Text type={'title'} size={8} color={'grey'} extraClass={'max-w-md m-0 ml-1'}>A logo helps personalize your Project. <a onClick={() => setShowProfile(true)}>Upload project thumbnail</a></Text>
                            </div>
                          </Col>
                        }
                        <Col span={24}>
                          <div className={'mt-8 flex justify-center'}>
                            <Form.Item className={'m-0'}>
                              <Button size={'large'} type="primary" loading={loading} style={{ width: '27vw', height: '36px' }} className={'m-0'} htmlType="submit">
                                Create
                              </Button>
                            </Form.Item>
                          </div>
                        </Col>
                        <Col span={24}>
                          <div className={'mt-4 flex justify-center'}>
                            <Form.Item className={'m-0'}>
                              <Button size={'large'} type={'text'} style={{ width: '27vw', height: '36px', color:'#40A9FF' }} htmlType="text" onClick={onSkip}>
                                Skip
                              </Button>
                            </Form.Item>
                          </div>
                        </Col>
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
          <div className={`${styles.hideSVG}`}>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
          </div>
        </div>
      }
      {formData && <Congrates handleCancel={handleCancel} handleReset={handleReset} />}
    </>

  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});

export default connect(mapStateToProps, { createProjectWithTimeZone, udpateProjectDetails, projectAgentInvite, fetchProjectAgents })(BasicDetails);
