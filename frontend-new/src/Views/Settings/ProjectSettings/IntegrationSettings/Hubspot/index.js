import React, { useState } from 'react';
import { useEffect } from 'react';
import {connect} from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import {
    Row, Col, Modal, Input, Form, Button, notification, message
  } from 'antd';
  import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
  import {ErrorBoundary} from 'react-error-boundary'


const HubspotIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings, 
    setIsActive
}) =>{ 
    const [form] = Form.useForm();
    const [errorInfo, seterrorInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [showForm, setShowForm] = useState(false);

useEffect(()=>{
    fetchProjectSettings(activeProject.id).then(()=>{
      if(currentProjectSettings?.int_hubspot){
        setIsActive(true);
      }
    })
},[]);

const onFinish = values => { 
    setLoading(true);
        udpateProjectSettings(activeProject.id, 
        { 'int_hubspot_api_key': values.api_key, 
        'int_hubspot' : true 
    }).then(() => {
        setLoading(false);
        setShowForm(false); 
        setTimeout(() => {
            message.success('Hubspot integration successful'); 
        }, 500);
        setIsActive(true);
    }).catch((err) => {
        setShowForm(false);
        setLoading(false);
        console.log('change password failed-->', err);
        seterrorInfo(err.error);
        setIsActive(false);
    });
  };

  const onDisconnect = () =>{
    setLoading(true);
        udpateProjectSettings(activeProject.id, 
        { 'int_hubspot_api_key': '', 
        'int_hubspot' : false 
    }).then(() => {
        setLoading(false);
        setShowForm(false); 
        setTimeout(() => {
            message.success('Hubspot integration disconnected!'); 
        }, 500);
        setIsActive(false);
    }).catch((err) => {
        message.error(`${err?.data?.error}`);  
        setShowForm(false);
        setLoading(false);
        console.log('change password failed-->', err); 
    });
  }



const onReset = () => {
    seterrorInfo(null);
    setShowForm(false);
    form.resetFields();
  };
  const onChange = () => {
    seterrorInfo(null);
  };


  

return (
    <>
    <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with Hubspot integrations'} />} onError={FaErrorLog}>
  
    <Modal
        visible={showForm}
        zIndex={1020}
        onCancel={onReset}
        afterClose={()=>setShowForm(false)}
        className={'fa-modal--regular fa-modal--slideInDown'}  
        centered={true}
        footer={null}
        transitionName=""
        maskTransitionName=""
    >
         <div className={'p-4'}>
          <Form
          form={form}
          onFinish={onFinish}
          className={'w-full'}
          onChange={onChange}
          >
            <Row>
              <Col span={24}>
                <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Integrate with Hubspot</Text>
                <Text type={'title'} level={7} color={'grey'}  extraClass={'m-0 mt-2'}>Add your API key from Hubspot to connect with your hubspot account. You can find a detailed guide in our documentation section.</Text>
              </Col>
            </Row>
            <Row className={'mt-6'}>
              <Col span={24}> 
                <Form.Item
                    name="api_key"
                    rules={[
                      {
                        required: true,
                        message: 'Please input your Hubspot API Key'
                      }
                    ]}

                    >
                      <Input size="large" className={'fa-input w-full'} placeholder="Hubspot API Key" />
                      </Form.Item>
              </Col>
                {errorInfo && <Col span={24}>
                    <div className={'flex flex-col justify-center items-center mt-1'} >
                        <Text type={'title'} color={'red'} size={'7'} className={'m-0'}>{errorInfo}</Text>
                    </div>
                </Col>
                }
            </Row>
            <Row className={'mt-6'}>
              <Col span={24}>
                <div className={'flex justify-end'}>
                  <Button disabled={loading} size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button>
                  <Button loading={loading} type="primary" size={'large'} htmlType="submit"> Connect Now </Button>
                </div>
              </Col>
            </Row>
            </Form>
        </div>
    </Modal>
    {
        currentProjectSettings?.int_hubspot &&  <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Connected Account</Text>
        <Text type={'title'} level={7} color={'grey'}  extraClass={'m-0 mt-2'}>API Key</Text>
        <Input size="large" disabled={true} placeholder="API Key" value={currentProjectSettings.int_hubspot_api_key} style={{width:'400px'}}/>
    </div>
    }
    <div className={'mt-4 flex'}>
    {currentProjectSettings?.int_hubspot ? <Button loading={loading} onClick={()=>onDisconnect()}>Disconnect</Button> : <Button type={'primary'} loading={loading} onClick={()=>setShowForm(!showForm)}>Connect</Button>
    }
        <Button className={'ml-2 '}>View documentation</Button> 
    </div>
    </ErrorBoundary>
    </>
)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings
  });
  
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(HubspotIntegration)