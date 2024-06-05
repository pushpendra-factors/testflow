import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Input, Select, Form, message, Upload } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { LoadingOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';
import { udpateProjectDetails, udpateProjectSettings } from 'Reducers/global';
import { getTimeZoneNameFromCity } from 'Utils/constants';
import sanitizeInputString from 'Utils/sanitizeInputString';
import { Currency } from 'Utils/currency';
import _ from 'lodash';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';
import style from './index.module.scss';

const { Option } = Select;

function EditBasicSettings({
  activeProject,
  setEditMode,
  udpateProjectDetails,
  agent,
  udpateProjectSettings,
  currentProjectSettings
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [imageUrl, setImageUrl] = useState('');
  const [form] = Form.useForm();
  const [isLoaded, setIsLoaded] = useState(false);
  const [currencyVal, setCurrencyVal] = useState();

  useEffect(() => {
    if (currentProjectSettings && currentProjectSettings?.currency) {
      setCurrencyVal(currentProjectSettings?.currency);
    }
    setTimeout(() => {
      setDataLoading(false);
      setIsLoaded(true);
    }, 200);
  }, [currentProjectSettings]);

  const onFinish = (values) => {
    setDataLoading(true);
    const projectName = sanitizeInputString(values?.name);
    const projectData = {
      ...values,
      name: projectName,
      profile_picture: imageUrl,
      time_zone: '' // disabling sending timezone in project settings update
      // time_zone: values?.time_zone
    };

    udpateProjectDetails(activeProject.id, projectData)
      .then(() => {
        setDataLoading(false);
        message.success('Project details updated!');
        setEditMode(false);
      })
      .catch((err) => {
        setDataLoading(false);
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

  const handleChange = (info) => {
    // if (info.file.status === 'uploading') {
    //   // setLoading(true);
    //   return;
    // }
    // if (info.file.status === 'done') {
    // Get this url from response in real world.
    getBase64(info.file.originFileObj, (imageUrl) => {
      setImageUrl(imageUrl);
      // setLoading(false);
    });
    // }
  };

  const setCurrencyFn = (value) => {
    console.log('setCurrency value', value);
    setCurrencyVal(value);
    const data = {
      currency: value
    };
    udpateProjectSettings(activeProject.id, data)
      .then(() => {
        message.success('Currency details updated!');
      })
      .catch((err) => {
        console.log('err->', err);
        message.error(err.data.error);
      });
  };

  return (
    <div className='mb-10 pl-4'>
      <Form
        form={form}
        onFinish={onFinish}
        className='w-full'
        initialValues={{
          name: activeProject?.name,
          project_uri: activeProject?.project_uri,
          date_format: activeProject?.date_format,
          time_format: activeProject?.time_format,
          time_zone: !_.isEmpty(activeProject?.time_zone)
            ? activeProject?.time_zone
            : ''
        }}
      >
        <CommonSettingsHeader
          title='Basic Details'
          actionsNode={
            <div className='flex justify-end'>
              <Button
                size='large'
                disabled={dataLoading}
                onClick={() => setEditMode(false)}
              >
                Cancel
              </Button>
              <Button
                size='large'
                type='primary'
                disabled={dataLoading}
                className='ml-2'
                htmlType='submit'
              >
                {dataLoading && isLoaded ? <LoadingOutlined /> : ''}
                Save
              </Button>
            </div>
          }
        />

        <div className='animate__animated animate__fadeIn'>
          <Row className='mt-2'>
            <Col xs={12} sm={12} md={24}>
              <Upload
                name='avatar'
                accept=''
                showUploadList={false}
                listType='picture'
                beforeUpload={beforeUpload}
                onChange={handleChange}
              >
                {imageUrl || activeProject?.profile_picture ? (
                  <div
                    className={`flex justify-center items-center ${style.projectImageContainer}`}
                    style={{
                      width: 145,
                      height: 145,
                      borderRadius: 11,
                      border: '0.978px dashed #D9D9D9',
                      background: '#FAFAFA'
                    }}
                  >
                    <img
                      src={imageUrl || activeProject?.profile_picture}
                      alt='avatar'
                      style={{ width: 145, height: 145, borderRadius: 11 }}
                    />
                    <div className={style.editImageIcon}>
                      <SVG name='ImageEdit' size='22' color='#40A9FF' />
                    </div>
                  </div>
                ) : (
                  <div
                    className={`flex justify-center items-center ${style.projectImageContainer}`}
                    style={{
                      width: 145,
                      height: 145,
                      borderRadius: 11,
                      border: '0.978px dashed #D9D9D9',
                      background: '#FAFAFA'
                    }}
                  >
                    <SVG
                      name='ImageBackground'
                      extraClass={style.projectImage}
                    />
                    <div className={style.editImageIcon}>
                      <SVG name='ImageEdit' size='22' color='#40A9FF' />
                    </div>
                  </div>
                )}
              </Upload>
              <Text type='paragraph' mini extraClass='m-0 mt-2' color='grey'>
                A logo helps personalise your Project
              </Text>
            </Col>
          </Row>
        </div>
        <Row className='mt-6'>
          <Col span={24}>
            <Text type='title' level={7} extraClass='m-0'>
              Project Name
            </Text>
            <Form.Item
              name='name'
              rules={[{ required: true, message: 'Please enter project name' }]}
              className='m-0'
            >
              <Input
                className='fa-input'
                size='large'
                placeholder='Project Name'
              />
            </Form.Item>
          </Col>
        </Row>
        <Row className='mt-6'>
          <Col span={24}>
            <Text type='title' level={7} extraClass='m-0'>
              Project URL
            </Text>
            <Form.Item name='project_uri' className='m-0'>
              <Input
                className='fa-input'
                size='large'
                placeholder='Project URL'
              />
            </Form.Item>
          </Col>
        </Row>

        <Row className='mt-6'>
          <Col span={24}>
            <Text type='title' level={7} extraClass='m-0'>
              Time Zone
            </Text>
            <Text type='title' level={6} extraClass='m-0' weight='bold'>
              {!_.isEmpty(activeProject?.time_zone)
                ? `${getTimeZoneNameFromCity(activeProject?.time_zone)?.text}`
                : '---'}
            </Text>
          </Col>
        </Row>
      </Form>

      <Row className='mt-6 mb-10'>
        <Col span={24}>
          <Text type='title' level={7} extraClass='m-0'>
            Currency
          </Text>
          <Select
            className='fa-select w-full'
            value={currencyVal}
            placeholder='Currency'
            size='large'
            onChange={setCurrencyFn}
            showSearch
            optionFilterProp='children'
            filterOption={(input, option) =>
              option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
            }
            filterSort={(optionA, optionB) =>
              optionA.children
                .toLowerCase()
                .localeCompare(optionB.children.toLowerCase())
            }
          >
            {Currency &&
              Object.keys(Currency).map((key, index) => (
                <Option value={`${key}`}>{`${Currency[key]} (${key})`}</Option>
              ))}
          </Select>
        </Col>
      </Row>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  agent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  udpateProjectDetails,
  udpateProjectSettings
})(EditBasicSettings);
