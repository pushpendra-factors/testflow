import React, { useState } from 'react';
import { Row, Col, Button, Input, Form, Divider, message } from 'antd';
import { Text } from 'factorsComponents';
import { Link } from 'react-router-dom';
import { connect } from 'react-redux';
import { forgotPassword } from 'Reducers/agentActions';
import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import ForgotPasswordIllustration from '../../assets/images/forgot_password_illustration.png';
import ForgotPasswordSuccessIllustration from '../../assets/images/forgot_password_success.png';
import styles from './index.module.scss';

function ForgotPassword({ forgotPassword }) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [userEmail, setUserEmail] = useState(null);

  const SendResetLink = () => {
    setDataLoading(true);
    form
      .validateFields()
      .then((value) => {
        setDataLoading(true);
        setTimeout(() => {
          forgotPassword(userEmail ? userEmail : value.email)
            .then(() => {
              if (!userEmail) setUserEmail(value.email);
              else message.success('Email resent successfully');
              setDataLoading(false);
              // history.push('/');
            })
            .catch((err) => {
              setDataLoading(false);
              form.resetFields();
              seterrorInfo(err);
            });
        }, 200);
      })
      .catch((info) => {
        setDataLoading(false);
        form.resetFields();
        seterrorInfo(info);
      });
  };

  const onChange = () => {
    seterrorInfo(null);
  };

  return (
    <>
      <div className={'fa-container'}>
        <Row justify={'center'}>
          <Col span={24}>
            <LoggedOutScreenHeader />
          </Col>
          <Col span={24}>
            <div className='w-full flex items-center justify-center mt-6'>
              <div
                className='flex flex-col justify-center items-center'
                style={{
                  width: 400,
                  padding: '40px 48px',
                  borderRadius: 8,
                  border: '1px solid  #D9D9D9'
                }}
              >
                <div className='py-4'>
                  <img
                    src={
                      userEmail
                        ? ForgotPasswordSuccessIllustration
                        : ForgotPasswordIllustration
                    }
                    alt='illustration'
                    className={styles.forgotPasswordIllustration}
                  />
                </div>
                <div className={'flex justify-center items-center mt-4'}>
                  <Text
                    type={'title'}
                    level={3}
                    extraClass={'m-0'}
                    weight={'bold'}
                    color='character-title'
                  >
                    {userEmail ? 'Please check your email' : 'Forgot password?'}
                  </Text>
                </div>
                <Form
                  form={form}
                  name='login'
                  validateTrigger
                  initialValues={{ remember: false }}
                  onFinish={SendResetLink}
                  onChange={onChange}
                >
                  {!userEmail && (
                    <>
                      <div
                        className={
                          'flex justify-center items-center text-center mt-4'
                        }
                      >
                        <Text
                          type={'title'}
                          level={7}
                          color={'character-primary'}
                          extraClass={'m-0 desc-text'}
                        >
                          Please enter the email address. We will send an email
                          with a reset link and instructions to reset your
                          password
                        </Text>
                      </div>
                      <div
                        className={
                          'flex flex-col justify-center items-center mt-6'
                        }
                      >
                        <Form.Item
                          label={null}
                          name='email'
                          rules={[
                            {
                              required: true,
                              type: 'email',
                              message: 'Please enter your email'
                            }
                          ]}
                          className='w-full'
                        >
                          <Input
                            className={'fa-input w-full'}
                            loading={dataLoading}
                            size={'large'}
                            placeholder='Enter your email'
                          />
                        </Form.Item>
                      </div>
                      <div
                        className={
                          'flex flex-col justify-center items-center mt-6'
                        }
                      >
                        <Form.Item
                          className={'m-0 w-full'}
                          loading={dataLoading}
                        >
                          <Button
                            htmlType='submit'
                            loading={dataLoading}
                            type={'primary'}
                            size={'large'}
                            className={'w-full'}
                          >
                            Send Reset Link
                          </Button>
                        </Form.Item>
                      </div>
                    </>
                  )}
                  {userEmail && (
                    <div
                      className={
                        'flex flex-col justify-center items-center my-4 text-center'
                      }
                    >
                      <Text
                        type={'title'}
                        size={'6'}
                        color={'character-secondary'}
                        align={'center'}
                        extraClass={'m-0 desc-text'}
                      >
                        An email has been sent from{' '}
                        <span style={{ fontWeight: 600 }}>
                          support@factors.ai
                        </span>
                        . Please follow the link in the email to reset your
                        password
                      </Text>
                      <Text
                        type={'title'}
                        size={'6'}
                        color={'character-primary'}
                        align={'center'}
                        weight={'bold'}
                        extraClass={'m-0 desc-text mt-4 mb-3'}
                      >
                        {userEmail}
                      </Text>
                      <Divider />
                      <div className='flex justify-center items-center gap-2'>
                        <Text
                          type={'title'}
                          level={7}
                          color='character-primary'
                          extraClass='m-0'
                        >
                          Didn’t get it?
                        </Text>
                        <Button
                          htmlType='submit'
                          loading={dataLoading}
                          className={styles.resendButton}
                        >
                          Resend Email
                        </Button>
                      </div>
                    </div>
                  )}
                  {errorInfo && (
                    <Col span={24}>
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
                    </Col>
                  )}
                </Form>
              </div>
            </div>
          </Col>
          <Col span={24}>
            <div className={'w-full flex  justify-center items-center mt-8'}>
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
      </div>
    </>
  );
}

export default connect(null, { forgotPassword })(ForgotPassword);
