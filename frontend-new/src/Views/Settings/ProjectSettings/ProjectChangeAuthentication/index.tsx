import {
  ArrowLeftOutlined,
  BackwardOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
  PlusOutlined
} from '@ant-design/icons';
import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import { SVG, Text } from 'Components/factorsComponents';
import { Avatar, Button, Col, Divider, Form, Input, Row, message } from 'antd';
import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { Link, useHistory, useLocation } from 'react-router-dom';
import { toggleFaHeader } from 'Reducers/global/actions';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
  fetchProjectSettings,
  fetchProjectsList,
  getActiveProjectDetails
} from 'Reducers/global';
import { SSO_LOGIN_URL } from 'Utils/sso';
import { SearchProjectsList } from 'Components/ProjectModal/ProjectsListsPopoverContent';
import { RESET_GROUPBY } from 'Reducers/coreQuery/actions';
import { redirectSAMLProject } from 'Utils/saml';
import { PathUrls } from 'Routes/pathUrls';
import {
  fetchAgentInfo,
  login,
  signout,
  updateAgentLoginMethod
} from 'Reducers/agentActions';
import logger from 'Utils/logger';
import styles from './index.module.scss';
import { getBackendHost } from '../IntegrationSettings/util';

function ProjectChangeAuthentication({
  fetchProjectSettings,
  getActiveProjectDetails,
  login,
  fetchProjects,
  updateAgentLoginMethod,
  fetchAgentInfo,
  signout
}) {
  const location = useLocation();
  const history = useHistory();
  const dispatch = useDispatch();
  const form = Form.useForm();
  const [type, setType] = useState(null);
  const { projects } = useSelector((state) => state.global);
  const { loginMethod, agent_details } = useSelector((state) => state.agent);

  const showProjectsList = useMemo(
    () => location.state?.showProjects,
    [location.state]
  );
  useEffect(() => {
    dispatch(toggleFaHeader(false));
    if (Object.keys(location.state || {}).length === 0) {
      history.replace('/');
    } else if (location?.state?.selectedProject) {
      // coming from the Change Project Screen
      const currMethod = loginMethod;
      const selectedMethod = location?.state?.selectedProject?.login_method;
      const selectedProjectID = location?.state?.selectedProject?.id;

      // DEFAULT REDIRECT RULES
      if (currMethod === 1) {
        // EMAIL
        // go tp email directly
        if (selectedMethod === 1) {
          getActiveProjectDetails(selectedProjectID);
          fetchProjectSettings(selectedProjectID);
          localStorage.setItem('activeProject', selectedProjectID);
          history.replace('/');
        } else {
          setType(selectedMethod);
        }
      } else if (currMethod === 2) {
        setType(selectedMethod);
      } else if (currMethod === 3) {
        // GOOGLE
        if (selectedMethod === 1 || selectedMethod === 3) {
          getActiveProjectDetails(selectedProjectID);
          fetchProjectSettings(selectedProjectID);
          localStorage.setItem('activeProject', selectedProjectID);
          history.replace('/');
          dispatch({ type: RESET_GROUPBY });
        } else {
          setType(selectedMethod);
        }
      } else {
        logger.error('Unknown Login Method Found');
        // setType(selectedMethod);
      }
    }
    return () => {
      dispatch(toggleFaHeader(true));
    };
  }, [loginMethod]);
  const handleSAMLLogin = () => {
    redirectSAMLProject(agent_details?.email);
  };
  const redirectGoogleLogin = () => {
    message.loading('Redirecting to Google Login');
    window.location.href = SSO_LOGIN_URL;
  };
  const handleProjectSelectHandle = (project: {
    login_method: any;
    id: string;
  }) => {
    switch (project.login_method) {
      case 1: // This case only happens for SAML cases
        break;
      case 2:
        redirectSAMLProject(agent_details?.email);
        break;
      case 3:
        redirectGoogleLogin();
        break;
      default:
        break;
    }
  };

  const renderProjectsList = () => (
    <div className={styles.projectsList}>
      <SearchProjectsList
        projects={projects}
        active_project={null}
        handleProjectListItemClick={handleProjectSelectHandle}
      />

      <Button
        type='default'
        className='w-3/4'
        onClick={() => {
          history.push(`${PathUrls.Onboarding}?setup=new`);
        }}
      >
        <PlusOutlined /> Create New Project
      </Button>
    </div>
  );

  const renderSSOForm = () => (
    <>
      <div className='flex justify-center items-center my-4 '>
        <Text
          type='title'
          level={5}
          extraClass='m-0'
          weight='bold'
          color='character-primary'
        >
          Your project admin has mandated login through SAML
        </Text>
      </div>
      <Button htmlType='submit' size='large' onClick={handleSAMLLogin}>
        Log in with Single sign-on
      </Button>
    </>
  );
  const handleEmailForm = async (values: any) => {
    const messageHandle = message.loading('Attempting Login', 0);
    try {
      await login(values.email, values.password);

      message.success('Login Success');
      updateAgentLoginMethod(1);
      window.location.reload();
    } catch (error) {
      form[0].resetFields();
      logger.error(error);
      message.error(error);
    } finally {
      messageHandle();
    }
  };
  const renderEmailForm = () => (
    <div className='w-full mt-2'>
      <Form
        validateTrigger='SDF'
        form={form[0]}
        className='flex flex-col gap-y-3'
        onFinish={handleEmailForm}
      >
        <Form.Item
          name='email'
          rules={[
            {
              required: true,
              type: 'email',
              message: 'Please enter a valid email'
            }
          ]}
        >
          <Input
            size='large'
            type='Email'
            className='fa-input'
            placeholder='Email'
          />
        </Form.Item>
        <Form.Item
          name='password'
          rules={[
            {
              required: true,
              message: 'Please enter password'
            }
          ]}
        >
          <Input.Password
            size='large'
            iconRender={(visible) =>
              visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
            }
            placeholder='Password'
            className='fa-input'
          />
        </Form.Item>
        <Form.Item>
          <Button
            htmlType='submit'
            className='w-full'
            type='primary'
            size='large'
          >
            Log In
          </Button>
        </Form.Item>
      </Form>
      <div className='font-medium flex items-center justify-center mt-4'>
        <Link
          to={{
            pathname: '/forgotpassword'
          }}
        >
          Forgot Password ?
        </Link>
      </div>
    </div>
  );

  const handleLogout = async () => {
    await signout();
  };
  return (
    <div>
      <div className='fa-container h-screen relative'>
        <Row justify='center'>
          <Col span={24}>
            <LoggedOutScreenHeader />
          </Col>
          {!showProjectsList && (
            <Col className='justify-center mt-10'>
              <Text
                color='character-primary'
                type='title'
                level={2}
                weight='bold'
                extraClass='m-0'
              >
                Now Requires{' '}
                {type === 3 ? 'Google' : type === 2 ? 'SAML' : 'Email'} Login !
              </Text>
            </Col>
          )}
          <Col span={24} className='mt-8 justify-center'>
            <div className='w-full flex items-center justify-center'>
              <div
                className='flex flex-col justify-center items-center'
                style={{
                  width: '512px',
                  padding: '40px 48px',
                  borderRadius: 8,
                  border: '1px solid  #D9D9D9',
                  minWidth: '512px'
                }}
              >
                <Text
                  color='character-primary'
                  type='title'
                  level={5}
                  weight='bold'
                  extraClass='m-0'
                >
                  {showProjectsList ? (
                    `Select the project to get started`
                  ) : (
                    <>
                      Your project admin has mandated login through{' '}
                      {type === 3 ? 'Google' : type === 2 ? 'SAML' : 'Email'}
                    </>
                  )}
                </Text>
                {showProjectsList ? (
                  renderProjectsList()
                ) : type === 3 ? (
                  <a href={SSO_LOGIN_URL} className='w-full'>
                    <Button
                      type='default'
                      size='large'
                      style={{
                        background: '#fff',
                        boxShadow: '0px 2px 0px 0px #0000000D'
                      }}
                      onClick={() => {
                        localStorage.setItem(
                          'selectedProjectID',
                          location.state?.selectedProject?.id
                        );
                      }}
                      className='w-full'
                    >
                      <SVG name='Google' size={24} />
                      Continue with Google
                    </Button>
                  </a>
                ) : type === 2 ? (
                  renderSSOForm()
                ) : (
                  renderEmailForm()
                )}
              </div>
            </div>
          </Col>
          <Col span={8}>
            <div className='flex flex-col justify-center items-center '>
              <Row>
                <Col span={24} />
              </Row>
              <Row>
                {showProjectsList ? (
                  <div>
                    <Button danger type='default' onClick={handleLogout}>
                      {' '}
                      Logout{' '}
                    </Button>
                  </div>
                ) : (
                  <Col span={24}>
                    <div className='flex flex-col justify-center items-center mt-5'>
                      <Text type='paragraph' mini color='grey'>
                        <Link
                          to={{
                            pathname: '/'
                          }}
                        >
                          <ArrowLeftOutlined /> Go back to the list of projects
                        </Link>
                      </Text>
                      {/* <Text type={'paragraph'} mini color={'grey'}>Want to try out Factors.AI? <a href={'https://www.factors.ai/schedule-a-demo'} target="_blank">Request A Demo</a></Text> */}
                    </div>
                  </Col>
                )}
              </Row>
            </div>
          </Col>
        </Row>
      </div>
    </div>
  );
}

export default connect(null, {
  getActiveProjectDetails,
  fetchProjectSettings,
  login,
  fetchProjects: fetchProjectsList,
  updateAgentLoginMethod,
  fetchAgentInfo,
  signout
})(ProjectChangeAuthentication);
