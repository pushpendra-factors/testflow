import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import { Text } from 'Components/factorsComponents';
import { Button, Col, Form, Input, Row, message } from 'antd';
import React from 'react';
import { Link } from 'react-router-dom';
import { getBackendHost } from 'Views/Settings/ProjectSettings/IntegrationSettings/util';
import { getSAMLValidateURL, redirectSAMLProject } from 'Utils/saml';
import { LoggedOutFooter } from './Login';

function SingleSingOn() {
  const [form] = Form.useForm();
  const onFinish = (values: any) => {
    redirectSAMLProject(values?.email);
  };
  return (
    <div className='fa-container h-screen relative'>
      <Row justify='center'>
        <Col span={24}>
          <LoggedOutScreenHeader />
        </Col>
        <Col className='justify-center mt-10'>
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
        <Col span={24} className='mt-8 justify-center'>
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
                // validateTrigger
                initialValues={{ remember: false }}
                onFinish={onFinish}
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
                    Enter email to continue
                  </Text>
                </div>
                <div className='mt-8'>
                  <Form.Item
                    label={null}
                    name='email'
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
                      //   disabled={dataLoading}
                      size='large'
                      placeholder='Email'
                    />
                  </Form.Item>
                </div>
                <div className=' mt-6'>
                  <Form.Item className='m-0'>
                    <Button
                      htmlType='submit'
                      type='primary'
                      size='large'
                      className='w-full'
                    >
                      LOG IN
                    </Button>
                  </Form.Item>
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
  );
}

export default SingleSingOn;
