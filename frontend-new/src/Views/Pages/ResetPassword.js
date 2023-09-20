import React, { useState } from 'react';
import { Row, Col, Button, Input, Form, message } from 'antd';
import { Text, SVG } from 'factorsComponents';
import { Link, useHistory } from 'react-router-dom';
import queryString from 'query-string';
import { setPassword } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import LoginIllustration from '../../assets/images/login_Illustration.png';
import styles from './index.module.scss';
import PasswordChecks from 'Components/GenericComponents/PasswordChecks';

function ResetPassword(props) {
  const [form] = Form.useForm();
  const [errorInfo, seterrorInfo] = useState(null);
  const [firstPassword, setFirstPassword] = useState('');
  const [dataLoading, setDataLoading] = useState(false);
  const [validationMode, setValidationMode] = useState({
    password: 'onBlur',
    confirmPassword: 'onBlur'
  });

  const history = useHistory();

  const ResetPassword = () => {
    setDataLoading(true);
    const tokenFromUrl = queryString.parse(props.location.search)?.token;

    form
      .validateFields()
      .then((value) => {
        setDataLoading(true);
        setTimeout(() => {
          props
            .setPassword(value.password, tokenFromUrl)
            .then(() => {
              setDataLoading(false);
              history.push('/');
              message.success('Password Changed!');
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
      <div className={'fa-container'}>
        <Row justify={'center'}>
          <Col span={24}>
            <LoggedOutScreenHeader />
          </Col>
          <Col justify='center' className='mt-10'>
            <Text
              color='character-primary'
              type={'title'}
              level={2}
              weight={'bold'}
              extraClass='m-0'
            >
              You’re almost there !
            </Text>
          </Col>
          <Col span={24} justify='center' className='mt-8'>
            <div className='w-full flex items-center justify-center '>
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
                  onFinish={ResetPassword}
                  className={'w-full'}
                  onValuesChange={onChange}
                >
                  <div className={'flex justify-center items-center'}>
                    <Text
                      type={'title'}
                      level={6}
                      extraClass={'m-0'}
                      weight={'bold'}
                      color='character-primary'
                    >
                      Reset your Password
                    </Text>
                  </div>
                  <div className={'w-full mt-8'}>
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
                        className={'fa-input w-full'}
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
                  <div className='mt-4 w-full'>
                    <Form.Item
                      name='confirm_password'
                      dependencies={['password']}
                      rules={[
                        {
                          required: true,
                          message: 'Please confirm your new password.'
                        },
                        ({ getFieldValue }) => ({
                          validator(rule, value) {
                            if (!value || getFieldValue('password') === value) {
                              return Promise.resolve();
                            }
                            return Promise.reject(
                              new Error(
                                'The new password that you entered do not match!'
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
                        className={'fa-input w-full'}
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

                  <div className='w-full mt-6'>
                    <Form.Item className={'m-0'} loading={dataLoading}>
                      <Button
                        htmlType='submit'
                        loading={dataLoading}
                        type={'primary'}
                        size={'large'}
                        className={'w-full'}
                      >
                        Set new password
                      </Button>
                    </Form.Item>
                  </div>
                  {errorInfo && (
                    <div
                      className={
                        'flex flex-col justify-center items-center mt-1'
                      }
                    >
                      <Text
                        type={'title'}
                        color={'red'}
                        size={'7'}
                        className={'m-0'}
                      >
                        {errorInfo}
                      </Text>
                    </div>
                  )}
                </Form>
              </div>
            </div>
          </Col>
          <Col span={24}>
            <div className={'flex flex-col justify-center items-center mt-8'}>
              <Link
                to={{
                  pathname: '/login'
                }}
              >
                Go back to login
              </Link>
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
    </>
  );
}

export default connect(null, { setPassword })(ResetPassword);
