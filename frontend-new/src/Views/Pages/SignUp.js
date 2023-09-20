import React, { useEffect, useState } from 'react';
import { connect } from 'react-redux';
import {
  Row,
  Col,
  Button,
  Input,
  Form,
  message,
  Carousel,
  Divider
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { signup } from 'Reducers/agentActions';
import Congrats from './Congrats';
import { SSO_SIGNUP_URL } from '../../utils/sso';
import styles from './index.module.scss';
import HelpButton from 'Components/GenericComponents/HelpButton';
import { Link } from 'react-router-dom';
import Testimonial1 from '../../assets/images/testimonials/testimonial-1.png';
import Testimonial2 from '../../assets/images/testimonials/testimonial-2.png';
import Testimonial3 from '../../assets/images/testimonials/testimonial-3.png';
import Testimonial4 from '../../assets/images/testimonials/testimonial-4.png';
import Testimonial5 from '../../assets/images/testimonials/testimonial-5.png';
import HappyCustomers from '../../assets/images/happy_curtomers.png';

function SignUp({ signup }) {
  const [form] = Form.useForm();
  const [dataLoading, setDataLoading] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [formData, setformData] = useState(null);
  const [emailValidateType, setEmailValidateType] = useState('onBlur');

  const checkError = () => {
    const url = new URL(window.location.href);
    const error = url.searchParams.get('error');
    if (error) {
      let str = error.replace('_', ' ');
      let finalmsg = str.toLocaleLowerCase();
      message.error(finalmsg);
    }
  };

  useEffect(() => {
    checkError();
  }, []);

  const SignUpFn = () => {
    setDataLoading(true);
    form
      .validateFields()
      .then((values) => {
        setDataLoading(true);

        // //Factors SIGNUP tracking
        // factorsai.track('SIGNUP', {
        //   first_name: sanitizedValues?.first_name,
        //   email: sanitizedValues?.email
        // });

        signup(values)
          .then(() => {
            setDataLoading(false);
            setformData(values);
          })
          .catch((err) => {
            setDataLoading(false);
            form.resetFields();
            seterrorInfo(err);
          });
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
      {!formData && (
        <div
          className={
            'fa-content-container.no-sidebar fa-content-container--full-height w-full'
          }
        >
          {/* //parent container starts here */}

          <div className={'flex w-full h-full '}>
            {/* //left side content starts here */}
            <Col
              xs={{ span: 0 }}
              sm={{ span: 10 }}
              style={{ background: '#E6F7FF' }}
            >
              <div className='w-full h-screen flex items-center justify-center'>
                <div style={{ width: 448 }}>
                  <div className='flex justify-center items-center w-full'>
                    <SVG name={'BrandFull'} width={238} color='white' />
                  </div>
                  <div className='mt-10 flex justify-center'>
                    <Carousel
                      autoplay
                      autoplaySpeed={5000}
                      dots
                      style={{ width: 448 }}
                      className={styles.signupCarousel}
                    >
                      {[
                        Testimonial1,
                        Testimonial2,
                        Testimonial3,
                        Testimonial4,
                        Testimonial5
                      ].map((image, i) => (
                        <div key={i}>
                          <img
                            src={image}
                            className={'m-0'}
                            style={{ width: '100%' }}
                            alt='reviews'
                          />
                        </div>
                      ))}
                    </Carousel>
                  </div>
                  <div className='mt-10 flex justify-center'>
                    <img src={HappyCustomers} className={'m-0 '} alt='brands' />
                  </div>
                </div>
              </div>
            </Col>

            {/* //left side content ends here */}

            {/* //right side content starts here */}

            <Col xs={{ span: 24 }} sm={{ span: 14 }}>
              <div className='w-full h-full'>
                <div className='flex justify-end w-full items-center px-10 h-16'>
                  <HelpButton />
                </div>
                <div className='mt-20'>
                  <Text
                    type={'title'}
                    level={2}
                    weight={'bold'}
                    align={'center'}
                    color={'character-primary'}
                    extraClass={'m-0'}
                  >
                    Let’s create your account !
                  </Text>
                </div>
                <div className='w-full  flex  justify-center mt-8 '>
                  <div style={{ width: 428 }} className='px-12'>
                    <Form
                      form={form}
                      name='login'
                      // validateTrigger
                      initialValues={{ remember: false }}
                      onFinish={SignUpFn}
                      onChange={onChange}
                    >
                      <div>
                        <Text
                          type={'title'}
                          level={6}
                          align={'center'}
                          color={'character-primary'}
                          extraClass={'m-0'}
                          weight={'bold'}
                        >
                          Signup to continue
                        </Text>
                      </div>

                      <div className={' mt-8 w-full'}>
                        {/* <Text type={'title'} level={7} extraClass={'m-0'}>Work Email</Text> */}
                        <Form.Item
                          label={null}
                          name='email'
                          rules={[
                            {
                              required: true,
                              message:
                                'Please enter your business email address.'
                            },
                            ({ getFieldValue }) => ({
                              validator(rule, value) {
                                if (
                                  !value ||
                                  value.match(
                                    /^([A-Za-z0-9!'#$%&*+\/=?^_`{|}~-]+(?:\.[A-Za-z0-9!'#$%&*+\/=?^_`{|}~-]+)*@(?!gmail.com)(?!yahoo.com)(?!hotmail.com)(?!yahoo.co.in)(?!hey.com)(?!icloud.com)(?!me.com)(?!mac.com)(?!aol.com)(?!abc.com)(?!xyz.com)(?!pqr.com)(?!rediffmail.com)(?!live.com)(?!outlook.com)(?!msn.com)(?!ymail.com)([\w-]+\.)+[\w-]{2,})?$/
                                  )
                                ) {
                                  return Promise.resolve();
                                }
                                return Promise.reject(
                                  new Error(
                                    'Please enter your business email address.'
                                  )
                                );
                              }
                            })
                          ]}
                          validateTrigger={emailValidateType}
                        >
                          <Input
                            className={'fa-input w-full'}
                            disabled={dataLoading}
                            size={'large'}
                            placeholder='Work Email'
                            onBlur={() => setEmailValidateType('onChange')}
                          />
                        </Form.Item>
                      </div>
                      <div className={' mt-6'}>
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
                            Signup with Email
                          </Button>
                        </Form.Item>
                      </div>

                      <Row>
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
                      </Row>

                      <Divider className='my-6'>
                        <Text
                          type={'title'}
                          level={7}
                          extraClass={'m-0'}
                          color={'grey'}
                        >
                          OR
                        </Text>
                      </Divider>

                      <div className={' mt-6'}>
                        <Form.Item className={'m-0 w-full'}>
                          <a href={SSO_SIGNUP_URL}>
                            <Button
                              type={'default'}
                              size={'large'}
                              className='btn-custom--bordered w-full'
                            >
                              <SVG name={'Google'} size={24} />
                              Continue with Google
                            </Button>
                          </a>
                        </Form.Item>
                      </div>
                      <div className={'flex justify-center  mt-8 '}>
                        <Text
                          type={'title'}
                          level={7}
                          color={'character-primary'}
                        >
                          Already have an account?{' '}
                          <Link
                            disabled={dataLoading}
                            to={{
                              pathname: '/login'
                            }}
                          >
                            Log In
                          </Link>
                        </Text>
                      </div>
                    </Form>
                  </div>
                </div>
                <div style={{ marginTop: 110 }} className='flex justify-center'>
                  <div style={{ width: 428 }}>
                    <Text
                      type={'title'}
                      level={8}
                      color={'grey'}
                      extraClass={'text-center'}
                    >
                      By signing up, I accept the Factors.ai{' '}
                      <a
                        href={'https://www.factors.ai/terms-of-use'}
                        target='_blank'
                        rel='noreferrer'
                      >
                        Terms of Use
                      </a>{' '}
                      and acknowledge having read through the{' '}
                      <a
                        href={'https://www.factors.ai/privacy-policy'}
                        target='_blank'
                        rel='noreferrer'
                      >
                        Privacy policy
                      </a>
                    </Text>
                  </div>
                </div>
              </div>
            </Col>
            {/* //right side content ends here */}
          </div>
          {/* //parent container ends here */}
        </div>
      )}
      {formData && <Congrats data={formData} />}

      {/* <MoreAuthOptions showModal={showModal} setShowModal={setShowModal}/>   */}
    </>
  );
}

export default connect(null, { signup })(SignUp);
