import React, { useEffect, useCallback, useState } from 'react';
import {   Layout, Row, Col, Modal, Input, Select, Form, Button, message } from 'antd';
import { SVG, Text} from 'factorsComponents';
import { Link } from 'react-router-dom';
import {saveGoalInsights} from 'Reducers/factors';
import {connect} from 'react-redux';

function Header({saveGoalInsights, activeProject, factors_insight_rules}) {
  const { Header } = Layout;


  const [showSaveModal, setshowSaveModal] = useState(false);
  const [errorInfo, seterrorInfo] = useState(null);
  const [isLoading, setisLoading] = useState(false);
  const [form] = Form.useForm();

  const saveGoal = (payload) => {   
    setshowSaveModal(false);
    setisLoading(true);
    if(factors_insight_rules){ 
      let factorsData = {
        ...factors_insight_rules, 
        name: payload.name
      } 
      saveGoalInsights(activeProject.id, factorsData).then(() => {
        setshowSaveModal(false);
        setisLoading(false);
        message.success('Invitation sent successfully!');
      }).catch((err) => {
        console.log('invite error', err);
        form.resetFields();
        seterrorInfo(err);
        setisLoading(false); 
      });
    }

  }; 
  const onChange = () => {
    seterrorInfo(null);
  };
  const onReset = () => {
    setshowSaveModal(false);
    seterrorInfo(null);
    form.resetFields();
  };

  const addShadowToHeader = useCallback(() => {
    const scrollTop = (window.pageYOffset !== undefined) ? window.pageYOffset : (document.documentElement || document.body.parentNode || document.body).scrollTop;
    if (scrollTop > 0) {
      document.getElementById('app-header').style.filter = 'drop-shadow(0px 2px 0px rgba(200, 200, 200, 0.25))';
    } else {
      document.getElementById('app-header').style.filter = 'none';
    }
  }, []);

  useEffect(() => {
    document.addEventListener('scroll', addShadowToHeader);
    return () => {
      document.removeEventListener('scroll', addShadowToHeader);
    };
  }, [addShadowToHeader]);

  return (
    <Header id="app-header" className="ant-layout-header--custom bg-white w-full z-20 fixed p-0 top-0" >

         <div className="flex py-4 justify-between items-center"> 
                    <div className="flex items-center items-center">
                        <div>
                            <Link to="/"><SVG name={'brand'} color="#0B1E39" size={32} /></Link>
                        </div>
                        <div style={{ color: '#0E2647', opacity: 0.56, fontSize: '14px' }} className="font-bold leading-5 ml-2">  <Link to="/explain" style={{ color: '#0E2647', fontSize: '14px' }} >Factors</Link> / New Goal</div>
                    </div>
                    <div className="flex items-center items-center">
                    <Button   
                    onClick={()=>setshowSaveModal(true)}
                      className="items-center"
                      type="primary"
                      icon={<SVG extraClass="mr-1" name={"save"} size={24} color="#FFFFFF" />}
                    >
                      Save
                    </Button>
              </div>  
        </div>

        <Modal
        visible={showSaveModal}
        zIndex={1020}
        onCancel={()=>setshowSaveModal(false)}
        className={'fa-modal--regular fa-modal--slideInDown'}
        footer={false} 
        centered={true}
        maskClosable={false}
        afterClose={onReset}
        transitionName=""
        maskTransitionName=""
        
      >
        <div className={'p-4'}>
          <Row className={'mb-6'}>
            <Col span={24}>
              <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Save Goal</Text>
            </Col>
          </Row>
         <Form
            form={form}
            name="inviteUser"
            onFinish={saveGoal}
            onChange={onChange}
            className={'w-full'}
          >
            <Row gutter={[24, 24]}>

                <Col span={24}>
                <Text type={'title'} level={7} extraClass={'m-0'}>Title</Text>
                  <Form.Item name="name" rules={[ { required: true}]} className={'m-0'} >
                    <Input size="large" className={'fa-input w-full'} />
                  </Form.Item>
                </Col> 
                {errorInfo && <Col span={24}>
                    <div className={'flex flex-col justify-center items-center mt-1'} >
                        <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                    </div>
                </Col>
                }
                <Col span={24}>
                  <div className={'flex justify-end'}>
                    <Button size={'large'} onClick={onReset} className={'mr-2'}>Cancel</Button>
                    <Button size={'large'} loading={isLoading} type="primary" htmlType="submit">Save</Button>
                  </div>
                </Col>

            </Row>
                </Form>

        </div>

      </Modal>


    </Header>
  );
}

const mapStateToProps = (state) => {
  return {
    activeProject: state.global.active_project, 
    factors_insight_rules: state.factors.factors_insight_rules
  };
};
export default connect(mapStateToProps, {saveGoalInsights})(Header);
