import React, { useState } from 'react';
import { useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectSettings, udpateProjectSettings, enableLeadSquaredIntegration, disableLeadSquaredIntegration } from 'Reducers/global';
import {
    Row, Col, Modal, Input, Form, Button, notification, message, Avatar
} from 'antd';
import { Text, FaErrorComp, FaErrorLog, SVG } from 'factorsComponents';
import { ErrorBoundary } from 'react-error-boundary'
import factorsai from 'factorsai';
import { sendSlackNotification } from '../../../../../utils/slack';


const LeadSquaredIntegration = ({
    fetchProjectSettings,
    udpateProjectSettings,
    activeProject,
    currentProjectSettings,
    setIsActive,
    kbLink = false,
    currentAgent,
    enableLeadSquaredIntegration,
    disableLeadSquaredIntegration
}) => {
    const [form] = Form.useForm();
    const [errorInfo, seterrorInfo] = useState(null);
    const [loading, setLoading] = useState(false);
    const [showForm, setShowForm] = useState(false);

    useEffect(() => {
        if (currentProjectSettings?.lead_squared_config !== null) {
            setIsActive(true);
        }
    }, [currentProjectSettings]);

    const onFinish = values => {
        setLoading(true);

        //Factors INTEGRATION tracking
        factorsai.track('INTEGRATION', { 'name': 'leadSquared', 'activeProjectID': activeProject.id });

        enableLeadSquaredIntegration(activeProject.id,
            {
                "access_key": values.access_key,
                "secret_key": values.secret_key,
                "host": values.host,
            }).then(() => {
                setLoading(false);
                fetchProjectSettings(activeProject.id);
                setShowForm(false);
                setTimeout(() => {
                    message.success('LeadSquared integration successful');
                }, 500);
                setIsActive(true);
                sendSlackNotification(currentAgent.email, activeProject.name, 'Leadsquared');
            }).catch((err) => {
                setShowForm(false);
                setLoading(false);
                seterrorInfo(err.error);
                setIsActive(false);
            });
    };

    const onDisconnect = () => {
        setLoading(true);
        disableLeadSquaredIntegration(activeProject.id).then(() => {
            setLoading(false);
            fetchProjectSettings(activeProject.id);
            setShowForm(false);
            setTimeout(() => {
                message.success('LeadSquared integration disconnected!');
            }, 500);
            setIsActive(false);
        }).catch((err) => {
            message.error(`${err?.data?.error}`);
            setShowForm(false);
            setLoading(false);
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
            <ErrorBoundary fallback={<FaErrorComp subtitle={'Facing issues with LeadSquared integrations'} />} onError={FaErrorLog}>

                <Modal
                    visible={showForm}
                    zIndex={1020}
                    onCancel={onReset}
                    afterClose={() => setShowForm(false)}
                    className={'fa-modal--regular fa-modal--slideInDown'}
                    centered={true}
                    closable={false}
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
                                    <Avatar
                                        size={40}
                                        shape={'square'}
                                        icon={<SVG name={'LeadSquared'} size={40} color={'purple'} />}
                                        style={{ backgroundColor: '#F5F6F8' }}
                                    />
                                </Col>
                            </Row>
                            <Row>
                                <Col span={24}>
                                    <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Integrate with LeadSquared</Text>
                                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Add your Access key, Secret Key and Host from Leadsquared to connect with your Leadsquared account.</Text>
                                </Col>
                            </Row>
                            <Row className={'mt-6'}>
                                <Col span={24} className={'mb-2'}>
                                    <Form.Item
                                        name="access_key"
                                        rules={[
                                            {
                                                required: true,
                                                message: 'Please input your Leadsquared Access Key'
                                            }
                                        ]}

                                    >
                                        <Input size="large" className={'fa-input w-full'} placeholder="Access Key" />
                                    </Form.Item>
                                </Col>
                                <Col span={24} className={'mb-2'}>
                                    <Form.Item
                                        name="secret_key"
                                        rules={[
                                            {
                                                required: true,
                                                message: 'Please input your Leadsquared Secret Key'
                                            }
                                        ]}

                                    >
                                        <Input size="large" className={'fa-input w-full'} placeholder="Secret Key" />
                                    </Form.Item>
                                </Col>
                                <Col span={24} className={'mb-2'}>
                                    <Form.Item
                                        name="host"
                                        rules={[
                                            {
                                                required: true,
                                                message: 'Please input your Leadsquared Host'
                                            }
                                        ]}

                                    >
                                        <Input size="large" className={'fa-input w-full'} placeholder="Host" />
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
                                        {/* <Button disabled={loading} size={'large'} onClick={onReset} className={'mr-2'}> Cancel </Button> */}
                                        <Button loading={loading} type="primary" size={'large'} htmlType="submit"> Connect Now </Button>
                                    </div>
                                </Col>
                            </Row>
                        </Form>
                    </div>
                </Modal>
                {
                    currentProjectSettings?.lead_squared_config !== null && <div className={'mt-4 flex flex-col border-top--thin py-4 mt-2 w-full'}>
                        <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>Connected Account</Text>
                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Access Key</Text>
                        <Input size="middle" disabled={true} placeholder="Access Key" value={currentProjectSettings?.lead_squared_config?.access_key} style={{ width: '400px' }} />
                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Secret Key</Text>
                        <Input size="middle" disabled={true} placeholder="Secret Key" value={currentProjectSettings?.lead_squared_config?.secret_key} style={{ width: '400px' }} />
                        <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>Host</Text>
                        <Input size="middle" disabled={true} placeholder="Host" value={currentProjectSettings?.lead_squared_config?.host} style={{ width: '400px' }} />
                    </div>
                }
                <div className={'mt-4 flex'} data-tour='step-11'>
                    {currentProjectSettings?.lead_squared_config !== null ? <Button loading={loading} onClick={() => onDisconnect()}>Disconnect</Button> : <Button type={'primary'} loading={loading} onClick={() => setShowForm(!showForm)}>Connect</Button>
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

export default connect(mapStateToProps, { fetchProjectSettings, udpateProjectSettings, enableLeadSquaredIntegration, disableLeadSquaredIntegration })(LeadSquaredIntegration)