import React, { useState } from 'react';
import { connect } from 'react-redux';
import { Row, Col, Button, message, Divider } from 'antd';
import { Text } from 'factorsComponents';
import { signup } from 'Reducers/agentActions';
import styles from './index.module.scss';
import LoggedOutScreenHeader from 'Components/GenericComponents/LoggedOutScreenHeader';
import ForgotPasswordSuccessIllustration from '../../assets/images/forgot_password_success.png';

function Congrats({ signup, data }) {
  const [dataLoading, setDataLoading] = useState(false);

  const resendEmail = () => {
    setDataLoading(true);
    signup(data)
      .then(() => {
        message.success('Email resent!');
        setDataLoading(false);
      })
      .catch((err) => {
        setDataLoading(false);
        message.success('Email resent!');
      });
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
                  width: 450,
                  padding: '40px 48px 16px 48px',
                  borderRadius: 8,
                  border: '1px solid  #D9D9D9'
                }}
              >
                <div className='py-4'>
                  <img
                    src={ForgotPasswordSuccessIllustration}
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
                    Verify your email
                  </Text>
                </div>

                <div
                  className={
                    'flex flex-col justify-center items-center mt-4 text-center'
                  }
                >
                  <Text
                    type={'title'}
                    size={'6'}
                    color={'character-secondary'}
                    align={'center'}
                    extraClass={'m-0 desc-text'}
                  >
                    We’ve sent a confirmation link to your email. check for a
                    link from{' '}
                    <span style={{ fontWeight: 600 }}>support@factors.ai</span>{' '}
                    to activate your account and get started
                  </Text>
                  <Text
                    type={'title'}
                    size={'6'}
                    color={'character-primary'}
                    align={'center'}
                    weight={'bold'}
                    extraClass={'m-0 desc-text mt-4 mb-3'}
                  >
                    {data?.email}
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
                      onClick={resendEmail}
                      loading={dataLoading}
                      className={styles.resendButton}
                    >
                      Resend Email
                    </Button>
                  </div>
                </div>
              </div>
            </div>
          </Col>
        </Row>
      </div>
    </>
  );
}

export default connect(null, { signup })(Congrats);
