import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton, Input, Select, Form, message, Upload
} from 'antd';
import { Text } from 'factorsComponents';
import { UserOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';
import { udpateProjectDetails } from 'Reducers/global';
import { TimeZoneOffsetValues } from 'Utils/constants';  

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
const { Option } = Select;

function EditBasicSettings({ activeProject, setEditMode, udpateProjectDetails }) {
  const [dataLoading, setDataLoading] = useState(true);
  const [imageUrl, setImageUrl] = useState('');
  const [form] = Form.useForm();

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 200);
  }, []);

  const onFinish = values => { 
    let projectData = {
      ...values,
      profile_picture: imageUrl,
      time_zone: TimeZoneOffsetValues[values.time_zone]?.city
    }; 

    udpateProjectDetails(activeProject.id, projectData).then(() => {
      message.success('Project details updated!'); 
      setEditMode(false);
    }).catch((err) => {
      console.log('err->', err);
      message.error(err.data.error); 
    });
  };


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
      // setLoading(true);
      return;
    }
    if (info.file.status === 'done') {
      // Get this url from response in real world.
      getBase64(info.file.originFileObj, imageUrl => {
        setImageUrl(imageUrl);
        // setLoading(false);
      });
    }
  };


  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          initialValues={{
            name: activeProject?.name,
            project_uri: activeProject?.project_uri,
            date_format: activeProject?.date_format,
            time_format: activeProject?.time_format,
            time_zone: TimeZoneName[getKeyByValue(TimeZoneOffsetValues,activeProject?.time_zone)]
          }}
        >

          <Row>
            <Col span={12}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Basic Details</Text>
            </Col>
            <Col span={12}>
              <div className={'flex justify-end'}>
                <Button size={'large'} disabled={dataLoading} onClick={() => setEditMode(false)}>Cancel</Button>
                <Button size={'large'} type="primary" disabled={dataLoading} className={'ml-2'}
                  htmlType="submit"
                >Save</Button>
              </div>
            </Col>
          </Row>
          <Row className={'mt-2'}>
            <Col>
              <Upload
                  name="avatar"
                  accept={''}
                  showUploadList={false}
                  action="https://www.mocky.io/v2/5cc8019d300000980a055e76"
                  beforeUpload={beforeUpload}
                  onChange={handleChange}
              >
              {dataLoading ? <Skeleton.Avatar active={true} size={104} shape={'square'} />
                : imageUrl ? <img src={imageUrl} alt="avatar" style={{width:'105px'}} /> : <Avatar size={104} shape={'square'} icon={<UserOutlined />} />
              }
              </Upload>
              <Text type={'paragraph'} mini extraClass={'m-0 mt-1'} color={'grey'} >A logo helps personalise your Project</Text>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Project Name</Text>
              <Form.Item
                name="name"
                rules={[{ required: true, message: 'Please enter project name' }]}
                className={'m-0'}
              >
                <Input className={'fa-input'} size={'large'} placeholder="Project Name" />
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Project URL</Text>
              <Form.Item
                name="project_uri"
                className={'m-0'}
              >
                <Input className={'fa-input'} size={'large'} placeholder="Project URL" />
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Date Format</Text>
              <Form.Item
                name="date_format"
                className={'m-0'}
                rules={[{ required: true, message: 'Please choose an option' }]}
              >
                <Select className={'fa-select w-full'} placeholder={'Date Format'} size={'large'}>
                  <Option value="DD-MM-YYYY">DD-MM-YYYY</Option>
                  <Option value="YYYY-MM-DD">YYYY-MM-DD</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>
          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Time Format</Text>
              <Form.Item
                name="time_format"
                className={'m-0'}
                rules={[{ required: true, message: 'Please choose an option' }]}
              >
                <Select className={'fa-select w-full'} placeholder={'Time Format'} size={'large'}>
                  <Option value="12 Hours">12 Hours</Option>
                  <Option value="24 Hours">24 Hours</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Row className={'mt-6'}>
            <Col span={24}>
              <Text type={'title'} level={7} extraClass={'m-0'}>Time Zone</Text>
              <Form.Item
                name="time_zone"
                className={'m-0'}
                rules={[{ required: true, message: 'Please choose an option' }]}
                disabled={!activeProject?.is_multiple_project_timezone_enabled}
              >
                <Select 
                disabled={!activeProject?.is_multiple_project_timezone_enabled}
                className={'fa-select w-full'} placeholder={'Time Zone'} size={'large'}>
                   { Object.keys(TimeZoneName).map((item)=>{
                    return  <Option value={item}>{TimeZoneName[item]}</Option> 
                  })} 
                </Select>
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </div>

    </>

  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

export default connect(mapStateToProps, { udpateProjectDetails })(EditBasicSettings);
