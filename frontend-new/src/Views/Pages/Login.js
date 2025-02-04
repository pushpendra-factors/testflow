import React, { useEffect, useState } from 'react';
import { connect, useSelector } from 'react-redux';
import { Row, Col, Button, Input, Form, message, Divider } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { Link, useHistory } from 'react-router-dom';
import { EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import factorsai from 'factorsai';
import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import useScript from 'hooks/useScript';
import { PathUrls } from 'Routes/pathUrls';
import { fetchProjectsList } from 'Reducers/global';
import { login, updateAgentLoginMethod } from '../../reducers/agentActions';
import styles from './index.module.scss';
import { SSO_LOGIN_URL } from '../../utils/sso';
import LoginIllustration from '../../assets/images/login_Illustration.png';
import logger from 'Utils/logger';

export function LoggedOutFooter() {
  return (
    <>
      <div className={`${styles.hide} select-none`}>
        <img
          src={LoginIllustration}
          className={styles.loginIllustration}
          alt='illustration'
        />
      </div>
      <div className='text-center mt-10'>
        <Text type='title' level={8} color='grey' extraClass='text-center'>
          By logging in, I accept the Factors.ai{' '}
          <a
            href='https://www.factors.ai/terms-of-use'
            target='_blank'
            rel='noreferrer'
          >
            Terms of Use
          </a>{' '}
          and acknowledge having read through the{' '}
          <a
            href='https://www.factors.ai/privacy-policy'
            target='_blank'
            rel='noreferrer'
          >
            Privacy policy
          </a>
        </Text>
      </div>
    </>
  );
}

function Login(props) {
  const { login, fetchProjects, updateAgentLoginMethod } = props;
  const { projects } = useSelector((state) => state.global);
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);

  const history = useHistory();

  useScript({
    url: 'https://tribl.io/footer.js?orgId=kBXL81Dq42snObK5D623',
    async: false
  });

  const checkError = () => {
    const url = new URL(window.location.href);
    const error = url.searchParams.get('error');
    if (error) {
      if (error === 'INVALID_AGENT') {
        message.error('Account doesn’t exist, please sign up');
        return;
      }
      if (error === 'AGENT_NOT_ACTIVE') {
        message.error('Already a user. Verify your email or Signup again.');
        return;
      }
      const str = error.replace('_', ' ');
      const finalmsg = str.toLocaleLowerCase();
      message.error(finalmsg);
    }
  };

  useEffect(() => {
    if (props.isAgentLoggedIn) {
      history.push({
        pathname: '/',
        state: { navigatedFromLoginPage: true }
      });
    }
    checkError();
  }, []);
  const checkSSOProjectsFlow = () => {
    if (projects.length === 0) return;
    if (!props.isAgentLoggedIn) return;

    const isAnyEmail =
      projects &&
      projects.filter((e) => e.login_method === 1 || e.login_method === 0)
        .length > 0;

    if (isAnyEmail) {
      // Should Redirect to any of the project
      const firstEmailProject = projects.find((e) => e.login_method === 1);
      if (firstEmailProject) {
        localStorage.setItem('activeProject', firstEmailProject?.id);
      }
      history.push({
        pathname: '/',
        state: { navigatedFromLoginPage: true }
      });
    } else {
      // show project_change
      history.push({
        pathname: PathUrls.ProjectChangeAuthentication,
        state: {
          showProjects: true
        }
      });
    }
  };
  useEffect(() => {
    checkSSOProjectsFlow();
  }, [projects]);

  const CheckLogin = () => {
    setDataLoading(true);
    form
      .validateFields()
      .then((value) => {
        setDataLoading(true);

        // Factors LOGIN tracking
        factorsai.track('LOGIN', { username: value?.form_username });

        setTimeout(() => {
          login(value.form_username, value.form_password)
            .then(async () => {
              await fetchProjects();
              // checkSSOProjectsFlow();
              updateAgentLoginMethod(1);
              setDataLoading(false);
            })
            .catch((err) => {
              setDataLoading(false);
              form.resetFields();
              seterrorInfo(err);
              logger.error(err);
            });
        }, 200);
      })
      .catch((info) => {
        setDataLoading(false);
        form.resetFields();
        logger.error(info);
      });
  };

  const onChange = () => {
    seterrorInfo(null);
  };

  return (
    <>
      <div className='fa-container h-screen relative'>
        <Row justify='center' className={`${styles.start}`}>
          <Col span={24}>
            <LoggedOutScreenHeader />
          </Col>
          <Col justify='center' className='mt-10'>
            <Text
              color='character-primary'
              type='title'
              level={2}
              weight='bold'
              extraClass='m-0'
            >
              Welcome back !
            </Text>
          </Col>
          <Col span={24} justify='center' className='mt-8'>
            <div className='w-full flex items-center justify-center'>
              <div
                className='flex flex-col justify-center items-center'
                style={{
                  width: 400,
                  padding: '40px 48px',
                  borderRadius: 8,
                  border: '1px solid  #D9D9D9'
                }}
              >
                <Form
                  form={form}
                  name='login'
                  validateTrigger
                  initialValues={{ remember: false }}
                  onFinish={CheckLogin}
                  onChange={onChange}
                  className='w-full'
                >
                  <div className='flex justify-center items-center  '>
                    <Text
                      type='title'
                      level={5}
                      extraClass='m-0'
                      weight='bold'
                      color='character-primary'
                    >
                      Login to Continue
                    </Text>
                  </div>
                  <div className='mt-8'>
                    <Form.Item
                      label={null}
                      name='form_username'
                      rules={[
                        {
                          required: true,
                          type: 'email',
                          message: 'Please enter a valid email'
                        }
                      ]}
                      className='w-full'
                    >
                      <Input
                        className='fa-input w-full'
                        disabled={dataLoading}
                        size='large'
                        placeholder='Email'
                      />
                    </Form.Item>
                  </div>
                  <div className='mt-4'>
                    <Form.Item
                      label={null}
                      name='form_password'
                      rules={[
                        {
                          required: true,
                          message: 'Please enter password'
                        }
                      ]}
                      className='w-full'
                    >
                      <Input.Password
                        className='fa-input w-full'
                        size='large'
                        placeholder='Password'
                        disabled={dataLoading}
                        iconRender={(visible) =>
                          visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />
                        }
                      />
                    </Form.Item>
                  </div>
                  <div className=' mt-6'>
                    <Form.Item className='m-0' loading={dataLoading}>
                      <Button
                        htmlType='submit'
                        loading={dataLoading}
                        type='primary'
                        size='large'
                        className='w-full'
                      >
                        LOG IN
                      </Button>
                    </Form.Item>
                  </div>
                  {errorInfo && (
                    <div className='flex flex-col justify-center items-center mt-1'>
                      <Text type='title' color='red' size='7' className='m-0'>
                        {errorInfo}
                      </Text>
                    </div>
                  )}
                  <div className='font-medium flex items-center justify-center mt-4'>
                    <Link
                      disabled={dataLoading}
                      to={{
                        pathname: '/forgotpassword'
                      }}
                    >
                      Forgot Password ?
                    </Link>
                  </div>
                  <Divider className='my-6'>
                    <Text type='title' level={7} extraClass='m-0' color='grey'>
                      OR
                    </Text>
                  </Divider>

                  <Form.Item className='m-0'>
                    <a href={SSO_LOGIN_URL}>
                      <Button
                        type='default'
                        size='large'
                        style={{
                          background: '#fff',
                          boxShadow: '0px 2px 0px 0px #0000000D'
                        }}
                        className='w-full'
                      >
                        <SVG name='Google' size={24} />
                        Continue with Google
                      </Button>
                    </a>
                  </Form.Item>
                  <div className='font-medium flex items-center justify-center mt-4'>
                    <Link
                      disabled={dataLoading}
                      to={{
                        pathname: '/sso'
                      }}
                    >
                      Login with single sign-on
                    </Link>
                  </div>
                </Form>
              </div>
            </div>
          </Col>

          <Col span={8}>
            <div className='flex flex-col justify-center items-center '>
              <Row>
                <Col span={24} />
              </Row>
              <Row>
                <Col span={24}>
                  <div className='flex flex-col justify-center items-center mt-5'>
                    <Text type='paragraph' mini color='grey'>
                      Don’t have an account?{' '}
                      <Link
                        disabled={dataLoading}
                        to={{
                          pathname: '/signup'
                        }}
                      >
                        Sign Up
                      </Link>
                    </Text>
                    {/* <Text type={'paragraph'} mini color={'grey'}>Want to try out Factors.AI? <a href={'https://www.factors.ai/schedule-a-demo'} target="_blank">Request A Demo</a></Text> */}
                  </div>
                </Col>
              </Row>
            </div>
          </Col>
        </Row>

        <LoggedOutFooter />
      </div>

      {/* <MoreAuthOptions showModal={showModal} setShowModal={setShowModal}/> */}
    </>
  );
}

const mapStateToProps = (state) => ({
  isAgentLoggedIn: state.agent.isLoggedIn
});

export default connect(mapStateToProps, {
  login,
  fetchProjects: fetchProjectsList,
  updateAgentLoginMethod
})(Login);
