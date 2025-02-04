import { SVG, Text } from 'Components/factorsComponents';
import {
  Button,
  Checkbox,
  Col,
  Divider,
  Form,
  Input,
  Popconfirm,
  Row,
  Select,
  Upload,
  message
} from 'antd';
import React, { useCallback, useEffect, useState } from 'react';
import { ExclamationCircleFilled } from '@ant-design/icons';
import sanitizeInputString from 'Utils/sanitizeInputString';
import factorsai from 'factorsai';
import {
  createProjectWithTimeZone,
  udpateProjectDetails,
  udpateProjectSettings
} from 'Reducers/global';
import { ConnectedProps, connect, useSelector } from 'react-redux';
import { bindActionCreators } from 'redux';
import { TimeZoneOffsetValueArr } from 'Utils/constants';
import { projectAgentInvite, fetchProjectAgents } from 'Reducers/agentActions';
import useMobileView from 'hooks/useMobileView';
import logger from 'Utils/logger';
import useQuery from 'hooks/useQuery';
import { useHistory } from 'react-router-dom';
import { PathUrls } from 'Routes/pathUrls';
import { AdminLock } from 'Routes/feature';
import style from './index.module.scss';
import {
  CommonStepsProps,
  OnboardingStepsConfig,
  PROJECT_CREATED
} from '../../types';
import {
  extractDomainFromUrl,
  getCompanyDomainFromEmail
} from '../../../utils';

const { Option } = Select;

function Step1({
  createProjectWithTimeZone,
  udpateProjectDetails,
  projectAgentInvite,
  udpateProjectSettings,
  incrementStepCount,
  fetchProjectAgents
}: OnboardingStep1Props) {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [imageUrl, setImageUrl] = useState('');
  const [isFormSubmitted, setIsFormSubmitted] = useState(false);
  const [projectName, setProjectName] = useState('');
  const [checkbox, setcheckbox] = useState(true);
  const history = useHistory();

  const { active_project, currentProjectSettings } = useSelector(
    (state: any) => state.global
  );
  const { agents, agent_details } = useSelector((state) => state.agent);
  const isSolutionsEmailInvited = agents?.find((agent) =>
    AdminLock(agent?.email)
  );

  const onboarding_steps: OnboardingStepsConfig =
    currentProjectSettings?.onboarding_steps;
  const isMobileView = useMobileView();
  const routerQuery = useQuery();
  const paramSetup = routerQuery.get('setup');
  const isNewSetup = paramSetup === 'new';

  const onFinish = async (values) => {
    try {
      setLoading(true);
      if (isFormSubmitted) {
        if (
          !AdminLock(agent_details?.email) &&
          checkbox &&
          !isSolutionsEmailInvited
        ) {
          await inviteUser(active_project?.id, 'solutions@factors.ai');
          await fetchProjectAgents(active_project?.id);
        }
        if (imageUrl) {
          await uploadImage(active_project?.id);
        }
        incrementStepCount();
        return;
      }
      const projectName = sanitizeInputString(values?.projectName);
      // extracting domain
      const ClearbitDomain = extractDomainFromUrl(values?.domainName);
      if (!ClearbitDomain) {
        message.error('Please Enter a valid domain!');
        setLoading(false);
        return false;
      }

      const projectData = {
        name: projectName,
        time_zone: values?.time_zone,
        clearbit_domain: ClearbitDomain
      };

      // Factors CREATE_PROJECT_TIMEZONE tracking
      factorsai.track('CREATE_PROJECT_TIMEZONE', {
        ProjectName: projectData?.name,
        time_zone: projectData?.time_zone,
        clearbit_domain: projectData?.clearbit_domain
      });
      let prevProjectId = '';
      if (isNewSetup) {
        prevProjectId = active_project?.id;
      }
      const createProjectRes = await createProjectWithTimeZone(projectData);
      const projectId = createProjectRes?.data?.id;
      if (checkbox && !AdminLock(agent_details?.email)) {
        await inviteUser(projectId, 'solutions@factors.ai');
        await fetchProjectAgents(projectId);
      }
      if (imageUrl) {
        await uploadImage(projectId);
      }
      const updatedOnboardingConfig = { [PROJECT_CREATED]: true };

      await udpateProjectSettings(projectId, {
        onboarding_steps: updatedOnboardingConfig
      });
      localStorage.setItem('activeProject', projectId);

      if (prevProjectId)
        localStorage.setItem('prevActiveProject', prevProjectId);

      setIsFormSubmitted(true);
      setLoading(false);
      incrementStepCount();
      history.replace(PathUrls.Onboarding);
    } catch (error) {
      setLoading(false);
      message.error(error?.data?.error || 'Something went wrong');
      logger.log('createProject Failed:', error);
    }
  };

  const uploadImage = async (projectId: string) => {
    try {
      await udpateProjectDetails(projectId, {
        profile_picture: imageUrl
      });
      setImageUrl('');
    } catch (error) {
      logger.error('Error in uploading Image', error);
      message.error('Error in uploading image');
    }
  };

  const inviteUser = async (projectId: string, user: string) => {
    try {
      await projectAgentInvite(projectId, {
        email: user,
        role: 2
      });
    } catch (error) {
      logger.error('Error in Inviting user', error);
      message.error('Error in Inviting User!');
    }
  };

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

  function getBase64(img, callback: (url: string) => void) {
    const reader = new FileReader();
    reader.addEventListener('load', () => callback(reader.result));
    reader.readAsDataURL(img);
  }

  const handleChange = (info) => {
    getBase64(info.file.originFileObj, (imageUrl) => {
      setImageUrl(imageUrl);
    });
  };

  const handleFormValuesChange = (_changedValues, allValues) => {
    setProjectName(allValues?.projectName || '');
  };

  const getInitialFormValues = useCallback(() => {
    if (isFormSubmitted) {
      return {
        time_zone: active_project?.time_zone || '',
        projectName: active_project?.name || '',
        domainName: active_project?.clearbit_domain || ''
      };
    }
    return {
      time_zone: `${TimeZoneOffsetValueArr[0]?.text}`,
      projectName: '',
      domainName: getCompanyDomainFromEmail(agent_details?.email)
    };
  }, [isFormSubmitted, active_project, agent_details?.email]);

  useEffect(() => {
    if (isNewSetup) {
      if (isFormSubmitted === true) setIsFormSubmitted(false);
    } else if (onboarding_steps?.[PROJECT_CREATED]) {
      if (isFormSubmitted === false) setIsFormSubmitted(true);
    }
  }, [onboarding_steps, isNewSetup, isFormSubmitted]);

  useEffect(() => {
    form.setFieldsValue(getInitialFormValues());
    setProjectName(getInitialFormValues().projectName);
  }, [isFormSubmitted, form, getInitialFormValues]);

  useEffect(() => {
    if (isFormSubmitted) {
      // check if solutions was invited before
      if (isSolutionsEmailInvited) {
        setcheckbox(true);
      } else {
        setcheckbox(false);
      }
    }
  }, [isFormSubmitted, isSolutionsEmailInvited]);

  return (
    <div>
      <Row>
        <Col
          xs={24}
          sm={24}
          md={24}
          className={`${isMobileView ? 'text-center' : ''}`}
        >
          <Text
            type='title'
            level={3}
            color='character-primary'
            extraClass='m-0'
            weight='bold'
          >
            Create a New Project
          </Text>
          <Text
            type='title'
            level={6}
            extraClass='m-0 mt-1'
            color='character-secondary'
          >
            Let's get started by creating a project for your organisation
          </Text>
        </Col>
        <Form
          name='createNewProject'
          onFinish={onFinish}
          form={form}
          initialValues={getInitialFormValues()}
          onValuesChange={handleFormValuesChange}
        >
          <Row className='mt-8'>
            <Col xs={24} sm={24} md={12}>
              <div>
                <Text
                  type='title'
                  size={10}
                  color='character-primary'
                  extraClass='m-0 ml-1 mb-1'
                >
                  Project Name
                </Text>
                <Form.Item
                  label={null}
                  name='projectName'
                  rules={[
                    {
                      required: true,
                      message: 'Please input your Project Name!'
                    }
                  ]}
                >
                  <Input
                    className='fa-input'
                    size='large'
                    placeholder='eg. My Company Name'
                    disabled={isFormSubmitted}
                    autoFocus
                  />
                </Form.Item>
              </div>
              <div className='mt-6'>
                <Text
                  type='title'
                  size={10}
                  color='character-primary'
                  extraClass='m-0 ml-1 mb-1'
                >
                  Company domain
                </Text>
                <Form.Item
                  label={null}
                  name='domainName'
                  rules={[
                    {
                      required: true,
                      message: 'Please input your Domain Name!'
                    }
                  ]}
                >
                  <Input
                    className='fa-input'
                    size='large'
                    placeholder={"Company's domain name"}
                    disabled={isFormSubmitted}
                  />
                </Form.Item>
              </div>
              <div className='mt-6'>
                <Text
                  type='title'
                  size={10}
                  color='character-primary'
                  extraClass='m-0 ml-1 mb-1'
                >
                  Select timezone
                </Text>
                <Form.Item
                  name='time_zone'
                  className='m-0'
                  rules={[
                    {
                      required: true,
                      message: 'Please choose an option'
                    }
                  ]}
                >
                  <Select
                    className='fa-select'
                    placeholder='Time Zone'
                    size='large'
                    disabled={isFormSubmitted}
                  >
                    {TimeZoneOffsetValueArr?.map((item) => {
                      if (item?.city != 'Europe/Berlin') {
                        return (
                          <Option value={item?.city}>{`${item?.text}`}</Option>
                        );
                      }
                    })}
                  </Select>
                </Form.Item>
                <Text
                  type='title'
                  size={8}
                  color='disabled-color'
                  extraClass='inline m-0 mt-4 ml-1 mb-2'
                >
                  Set it to what is there in your CRM
                </Text>
                <Popconfirm
                  placement='rightTop'
                  title={
                    <Text type='title' size={10} extraClass='max-w-sm'>
                      This must reflect the same timezone as used in your CRM.
                      Once selected, this action cannot be edited.
                    </Text>
                  }
                  icon={
                    <ExclamationCircleFilled style={{ color: '#1E89FF' }} />
                  }
                  okText='Got it'
                  cancelText='Learn More'
                  cancelButtonProps={{
                    type: 'text',
                    style: { color: '#1E89FF', display: 'none' }
                  }}
                >
                  <Button
                    type='text'
                    className='m-0'
                    style={{ backgroundColor: 'white' }}
                  >
                    <SVG name='infoCircle' size={18} color='gray' />
                  </Button>
                </Popconfirm>
              </div>
            </Col>
            <Col
              xs={24}
              sm={24}
              md={12}
              className={`${isMobileView ? 'mt-6' : ''}`}
            >
              <div className='animate__animated animate__fadeIn'>
                <div className='flex justify-center items-center h-full w-full'>
                  <div style={{ width: !isMobileView ? 145 : '100%' }}>
                    <Row>
                      <Col xs={12} sm={12} md={24}>
                        <Upload
                          name='avatar'
                          accept=''
                          showUploadList={false}
                          listType='picture'
                          beforeUpload={beforeUpload}
                          onChange={handleChange}
                        >
                          {imageUrl ||
                          (isFormSubmitted &&
                            active_project?.profile_picture) ? (
                            <img
                              src={imageUrl || active_project?.profile_picture}
                              alt='avatar'
                            />
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
                                <SVG
                                  name='ImageEdit'
                                  size='22'
                                  color='#40A9FF'
                                />
                              </div>
                            </div>
                          )}
                        </Upload>
                      </Col>
                      <Col xs={12} sm={12} md={24}>
                        <Text
                          type='title'
                          level={8}
                          extraClass={`m-0 ${isMobileView ? 'mt-10' : 'mt-4'}`}
                          color='character-secondary'
                        >
                          A company logo helps personalise your Project
                        </Text>
                      </Col>
                    </Row>
                  </div>
                </div>
              </div>
            </Col>
            <Col xs={24} sm={24} className='mt-6'>
              <Form.Item label={null} name='invite_support'>
                <div className='flex items-center'>
                  <Checkbox
                    checked={checkbox}
                    disabled={isFormSubmitted && isSolutionsEmailInvited}
                    onChange={(e) => setcheckbox(e.target.checked)}
                    className='mt-1'
                  />
                  <Text
                    type='title'
                    size={10}
                    color='grey'
                    extraClass=' m-0 ml-3'
                  >
                    Invite{' '}
                    <span className='font-bold'>solutions@factors.ai</span> into
                    this project for ongoing support
                  </Text>
                </div>
              </Form.Item>
            </Col>

            {!isMobileView && <Divider className='mt-10 mb-6' />}

            <Col span={24}>
              <div
                className={`flex ${
                  isMobileView ? 'justify-center mt-6' : 'justify-end'
                }`}
              >
                <Form.Item className='m-0'>
                  <Button
                    type='primary'
                    loading={loading}
                    style={{ height: '36px' }}
                    className='m-0'
                    htmlType='submit'
                    disabled={projectName === ''}
                  >
                    Create and continue
                  </Button>
                </Form.Item>
              </div>
            </Col>
          </Row>
        </Form>
      </Row>
    </div>
  );
}

const mapDispatchToProps = (dispatch) =>
  bindActionCreators(
    {
      createProjectWithTimeZone,
      udpateProjectDetails,
      projectAgentInvite,
      udpateProjectSettings,
      fetchProjectAgents
    },
    dispatch
  );

const connector = connect(null, mapDispatchToProps);
type ReduxProps = ConnectedProps<typeof connector>;
type OnboardingStep1Props = ReduxProps & CommonStepsProps;

export default connector(Step1);
