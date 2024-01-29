import React, { useEffect, useState } from 'react';
import { Row, Col, Button, Input, Form, message, Divider } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { useHistory } from 'react-router-dom';
import queryString from 'query-string';
import { activate } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import PasswordChecks from 'Components/GenericComponents/PasswordChecks';
import sanitizeInputString from 'Utils/sanitizeInputString';
import useScript from 'hooks/useScript';
import CountryPhoneInput from 'Components/GenericComponents/CountryPhoneInput';
import LoginIllustration from '../../assets/images/login_Illustration.png';
import { SSO_ACTIVATE_URL } from '../../utils/sso';
import styles from './index.module.scss';

function Activate(props) {
  const [form] = Form.useForm();
  const [firstPassword, setFirstPassword] = useState('');
  const [errorInfo, seterrorInfo] = useState(null);
  const [dataLoading, setDataLoading] = useState(false);
  const [validationMode, setValidationMode] = useState({
    password: 'onBlur',
    confirmPassword: 'onBlur'
  });

  useScript({
    url: 'https://js.hs-scripts.com/6188127.js',
    async: true,
    defer: true,
    id: 'hs-script-loader'
  });

  const history = useHistory();

  const checkError = () => {
    const url = new URL(window.location.href);
    const error = url.searchParams.get('error');
    if (error) {
      const str = error.replace('_', ' ');
      const finalmsg = str.toLocaleLowerCase();
      message.error(finalmsg);
    }
  };

  useEffect(() => {
    checkError();
  }, []);

  const ResetPassword = () => {
    setDataLoading(true);
    const tokenFromUrl = queryString.parse(props.location.search)?.token;

    form
      .validateFields()
      .then((values) => {
        setDataLoading(true);
        setTimeout(() => {
          const sanitizedValues = {
            first_name: sanitizeInputString(values?.first_name),
            last_name: values?.last_name
              ? sanitizeInputString(values.last_name)
              : '',
            password: values?.password
          };

          if (values?.phone?.phone && values?.phone?.code) {
            sanitizedValues.phone = `${values.phone.code}-${values.phone.phone}`;
          }

          props
            .activate(sanitizedValues, tokenFromUrl)
            .then(() => {
              setDataLoading(false);
              history.push('/');
              message.success('Account activated!');
            })
            .catch((err) => {
              setDataLoading(false);
              setFirstPassword('');
              form.resetFields();
              seterrorInfo(err);
            });
        }, 200);
      })
      .catch((info) => {
        setDataLoading(false);
        setFirstPassword('');
        form.resetFields();
        seterrorInfo(info);
      });
  };

  const onChange = (_changedValues, allValues) => {
    seterrorInfo(null);
    setFirstPassword(allValues.password);
  };

  return (
    <>
      <div className='fa-container h-screen relative'>
        <LoggedOutScreenHeader />
        <div className='text-center'>
          <Text
            color='character-primary'
            type='title'
            level={2}
            weight='bold'
            extraClass='m-0 text-center'
          >
            You’re almost there !
          </Text>
        </div>
        <Row justify='center' className={`${styles.start} pb-4`}>
          <Col span={24}>
            <div className='w-full flex items-center justify-center mt-4'>
              <div
                className='flex flex-col justify-center items-center '
                style={{
                  width: 400,
                  padding: '40px 48px',
                  borderRadius: 8,
                  border: '1px solid  #D9D9D9'
                }}
              >
                <Form
                  form={form}
                  onFinish={ResetPassword}
                  className='w-full'
                  onValuesChange={onChange}
                  initialValues={{ first_name: '' }}
                >
                  <div className='flex justify-center items-center'>
                    <Text
                      type='title'
                      level={5}
                      extraClass='m-0'
                      weight='bold'
                      color='character-primary'
                    >
                      Activate your account{' '}
                    </Text>
                  </div>
                  <div className='flex flex-col mt-8'>
                    <Form.Item
                      label={null}
                      name='first_name'
                      rules={[
                        { required: true, message: 'Please enter first name' }
                      ]}
                      className='w-full'
                    >
                      <Input
                        className='fa-input w-full'
                        disabled={dataLoading}
                        size='large'
                        placeholder='First Name'
                      />
                    </Form.Item>
                  </div>
                  <div className='flex flex-col mt-4'>
                    <Form.Item label={null} name='last_name' className='w-full'>
                      <Input
                        className='fa-input w-full'
                        disabled={dataLoading}
                        size='large'
                        placeholder='Last Name'
                      />
                    </Form.Item>
                  </div>
                  <div className='mt-4'>
                    <Form.Item
                      label={null}
                      name='phone'
                      rules={[
                        ({ getFieldValue }) => ({
                          validator(rule, value) {
                            if (
                              !value ||
                              !value?.phone ||
                              value?.phone?.match(/^[0-9\b]+$/)
                            ) {
                              return Promise.resolve();
                            }
                            return Promise.reject(
                              new Error('Please enter valid phone number.')
                            );
                          }
                        })
                      ]}
                    >
                      <CountryPhoneInput
                        className='fa-input w-full'
                        size='large'
                        placeholder='Phone Number (optional)'
                        type='tel'
                        allowClear
                      />
                    </Form.Item>
                  </div>
                  <div className='flex flex-col justify-center items-center mt-4 w-full'>
                    <Form.Item
                      name='password'
                      rules={[
                        {
                          required: true,
                          message: 'Please enter your password.'
                        },
                        ({ getFieldValue }) => ({
                          validator(rule, value) {
                            if (
                              !value ||
                              value.match(
                                /^(?=.*?[A-Z])(?=.*?[a-z])(?=.*?[0-9])(?=.*?[#?!@$%^&*-]).{8,}$/
                              )
                            ) {
                              return Promise.resolve();
                            }
                            return Promise.reject(
                              new Error('Enter a valid password!')
                            );
                          }
                        })
                      ]}
                      className='w-full'
                      validateTrigger={validationMode.password}
                    >
                      <Input.Password
                        disabled={dataLoading}
                        size='large'
                        className='fa-input w-full'
                        placeholder='Enter New Password'
                        onBlur={() =>
                          setValidationMode({
                            ...validationMode,
                            password: 'onChange'
                          })
                        }
                      />
                    </Form.Item>
                  </div>
                  <div className='flex flex-col justify-center items-center mt-4 w-full'>
                    <Form.Item
                      name='confirm_password'
                      dependencies={['password']}
                      rules={[
                        {
                          required: true,
                          message: 'Please confirm your password.'
                        },
                        ({ getFieldValue }) => ({
                          validator(rule, value) {
                            if (!value || getFieldValue('password') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(
                              new Error(
                                'The password that you entered do not match!'
                              )
                            );
                          }
                        })
                      ]}
                      className='w-full'
                      validateTrigger={validationMode.confirmPassword}
                    >
                      <Input.Password
                        disabled={dataLoading}
                        size='large'
                        className='fa-input w-full'
                        placeholder='Confirm New Password'
                        onBlur={() =>
                          setValidationMode({
                            ...validationMode,
                            confirmPassword: 'onChange'
                          })
                        }
                      />
                    </Form.Item>
                  </div>

                  <PasswordChecks password={firstPassword} />
                  <div className='flex flex-col justify-center items-center mt-5 w-full'>
                    <Form.Item className='m-0 w-full' loading={dataLoading}>
                      <Button
                        htmlType='submit'
                        loading={dataLoading}
                        type='primary'
                        size='large'
                        className='w-full'
                      >
                        Let's get started
                      </Button>
                    </Form.Item>
                  </div>
                  <Row>
                    {errorInfo && (
                      <Col span={24}>
                        <div className='flex flex-col justify-center items-center mt-1'>
                          <Text
                            type='title'
                            color='red'
                            size='7'
                            className='m-0'
                          >
                            {errorInfo}
                          </Text>
                        </div>
                      </Col>
                    )}
                  </Row>
                  <Divider className='my-6'>
                    <Text type='title' level={7} extraClass='m-0' color='grey'>
                      OR
                    </Text>
                  </Divider>
                  <div className='flex flex-col justify-center items-center mt-5 w-full'>
                    <Form.Item className='m-0 w-full' loading={dataLoading}>
                      <a href={SSO_ACTIVATE_URL}>
                        <Button
                          loading={dataLoading}
                          type='default'
                          size='large'
                          style={{
                            background: '#fff',
                            boxShadow: '0px 0px 2px rgba(0, 0, 0, 0.3)'
                          }}
                          className='w-full'
                        >
                          <SVG name='Google' size={24} />
                          Continue with Google
                        </Button>
                      </a>
                    </Form.Item>
                  </div>
                </Form>
              </div>
            </div>
          </Col>
        </Row>
        <div className={`${styles.hide}`}>
          <img
            src={LoginIllustration}
            className={styles.loginIllustration}
            alt='illustration'
          />
        </div>
      </div>

      {/* <MoreAuthOptions showModal={showModal} setShowModal={setShowModal}/> */}
    </>
  );
}

export default connect(null, { activate })(Activate);
