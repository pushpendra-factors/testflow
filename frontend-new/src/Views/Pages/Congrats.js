import React from 'react';
import { connect } from 'react-redux';
import {
  Row, Col, Button, message, Modal
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { signup } from 'Reducers/agentActions';
import styles from './index.module.scss';
import { meetLink } from '../../utils/hubspot';

function Congrats({ signup, data, setShowModal, showModal }) {

//   const popBookDemo = () => {
//     if(Calendly){ Calendly.initPopupWidget({ url: 'https://calendly.com/factorsai/demo' }); }
//   };

  const resendEmail = () => {
    console.log('resendEmail');
    signup(data).then(() => {
      message.success('Email resent!');
    }).catch((err) => {
      console.log('Signup-resent email err-->', err);
      message.success('Email resent!');
    });
  };


  return (
    <>
      <div className={'fa-container'}>
            <Row justify={'center'} className={`${styles.start}`}>
                <Col span={12}>
                    <div className={'flex flex-col justify-center items-center login-container'}>
                        <Row>
                            <Col span={24} >
                                <div className={'flex justify-center items-center'} >
                                    <SVG name={'BrandFull'} width={250} height={90} color="white"/>
                                </div>
                            </Col>
                        </Row>
                        <Row>
                            <Col span={24}>
                        
                        <Row>
                            <Col span={24}>
                                    <div className={'flex flex-col justify-center items-center mt-10 w-full'} >
                                        <img src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/pop-gift.png' style={{width:'100%', maxWidth:'160px', marginLeft:'20px'}} />
                                    </div>
                            </Col>

                            <Col span={24}>
                                <div className={'flex justify-center items-center mb-5'} >
                                    <Text type={'title'} level={3} extraClass={'m-0'} weight={'bold'}>Confirm your email to get started!</Text>
                                </div>
                            </Col>
                            
                            <Col span={24}>
                                <div className={'flex justify-center items-center mb-5'} >
                                    <Text type={'title'} level={6} extraClass={'m-0'} align={'center'} color={'grey'} weight={'bold'}>We’ve sent a confirmation link to your email. Check for a link from <span className={'text-black'}>support@factors.ai</span> to activate your account and get started</Text>
                                </div>
                            </Col>
                            <Col span={24}>
                                <div className={'flex justify-center items-center mb-5'} >
                                    <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{data.email}</Text>
                                </div>
                            </Col>
                           
                            <Col span={24}>
                                <div className={'flex flex-col justify-center items-center mt-10'}>
                                    {/* <Text type={'title'} level={6} color={'grey'} align="center" lineHeight={'large'} extraClass={'m-0 mb-4 w-3/5'}>Our team would be happy to walk you through the product and answer any questions </Text>
                                    <Button size={'large'} className={'w-full mt-4'} style={{ maxWidth: '280px' }} onClick={() => popBookDemo()}>Schedule a demo</Button> */}
                                    <Text type={'title'} level={7} align="center" extraClass={'m-0 mt-6'}>Didn’t get an email? <a onClick={() => resendEmail()} >Click to resend</a></Text>
                                </div>
                            </Col>
                        </Row>
                        </Col>
                        </Row>
                    </div>
                </Col>
            </Row>
            <div className={`${styles.hide}`}>
            <SVG name={'singlePages'} extraClass={'fa-single-screen--illustration'} />
            </div>
      </div>


        <Modal
        title={null}
        visible={showModal}
        footer={
            <Row className={'mb-2 mt-2'}>
            <Col span={24}>
                <div className={'flex flex-col justify-end items-end'} >
                    <Button type={'primary'} size={'large'} onClick={() => setShowModal(false)}>Done! Start setting up my Factors project</Button>
                </div>
            </Col>
            </Row>
        }
        centered={true}
        mask={true}
        maskClosable={false}
        maskStyle={{backgroundColor: 'rgb(0 0 0 / 70%)'}}
        closable={false}
        onCancel={()=> setShowModal(false)}
        className={`fa-modal--regular`}
        width={'70%'}
        >
            <div className={'m-0 mb-2'}>
                <Row justify={'center'} className={'mt-2'}>
                    <Col>
                        <Text type={'title'} level={5} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>Get a free product walkthrough</Text>
                    </Col>
                </Row>
                <Row className={'mt-4'}>
                    <iframe className={'w-full h-full'} style={{height: '450px'}} src={meetLink()}></iframe>
                </Row>
            </div>
        </Modal> 

    </>

  );
}

export default connect(null, { signup })(Congrats);
