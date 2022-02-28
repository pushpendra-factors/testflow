import React, { useState } from 'react';
import { useEffect } from 'react';
import {connect} from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings } from 'Reducers/global';
import {
    Row, Col, Modal, Input, Form, Button, notification, message
  } from 'antd';
  import { Text, FaErrorComp, FaErrorLog } from 'factorsComponents';
  import {ErrorBoundary} from 'react-error-boundary'
  import factorsai from 'factorsai';


const HubspotIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings, 
    setIsActive,
    kbLink = false,
    currentAgent
}) =>{ 
    const [form] = Form.useForm();
    const [errorInfo, seterrorInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [showForm, setShowForm] = useState(false);

    useEffect(() => {
      if (currentProjectSettings?.int_hubspot) {
        setIsActive(true);
      }
    }, [currentProjectSettings]);

    const sendSlackNotification = () => {
      let webhookURL = 'https://hooks.slack.com/services/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';
      let data = {
          "text": `User ${currentAgent.email} from Project "${activeProject.name}" Activated Integration: Hubspot`,
          "username" : "Signup User Actions",
          "icon_emoji" : ":golf:"
      }
      let params = {
          method: 'POST',
          body: JSON.stringify(data)
      }
  
      fetch(webhookURL, params)
      .then((response) => response.json())
      .then((response) => {
          console.log(response);
      })
      .catch((err) => {
          console.log('err',err);
      });
    }

const onFinish = values => { 
    setLoading(true);

    //Factors INTEGRATION tracking
    factorsai.track('INTEGRATION',{'name': 'hubspot','activeProjectID': activeProject.id});

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
        sendSlackNotification();
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
    <div className={'mt-4 flex'} data-tour = 'step-11'>
    {currentProjectSettings?.int_hubspot ? <Button loading={loading} onClick={()=>onDisconnect()}>Disconnect</Button> : <Button type={'primary'} loading={loading} onClick={()=>setShowForm(!showForm)}>Connect</Button>
    }
        {kbLink && <a className={'ant-btn ml-2 '} target={"_blank"} href={kbLink}>View documentation</a>}
    </div>
    </ErrorBoundary>
    </>
)
}

const mapStateToProps = (state) => ({
    activeProject: state.global.active_project,
    currentProjectSettings: state.global.currentProjectSettings,
    currentAgent: state.agent.agent_details,
  });
  
export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings })(HubspotIntegration)