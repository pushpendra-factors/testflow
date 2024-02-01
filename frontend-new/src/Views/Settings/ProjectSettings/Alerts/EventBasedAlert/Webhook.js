import React, {
    useState,
    useEffect,
    useCallback,
    useRef,
    useMemo
} from 'react';
import { connect, useDispatch, useSelector } from 'react-redux';
import {
    Row,
    Col,
    Select,
    Button,
    Form,
    Input,
    message,
    notification,
    Checkbox,
    Modal,
    Popover,
    Tooltip,
    Avatar,
    Switch,
    Menu,
    Dropdown,
    Tag,
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import PreviewCard from './PreviewCard';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import { ConsoleSqlOutlined } from '@ant-design/icons';
import { getErrorMsg, WEBHOOK } from './../utils';
import useFeatureLock from 'hooks/useFeatureLock';
import { FEATURES } from 'Constants/plans.constants';
import UpgradeButton from 'Components/GenericComponents/UpgradeButton';
import {PreviewCardWebhook} from './PreviewCard';

const Webhook = ({
    viewAlertDetails,
    groupBy,
    webhookEnabled,
    setWebhookEnabled,
    disbleWebhookInput,
    webhookRef,
    webhookUrl,
    setWebhookUrl,
    setConfirmBtn,
    setTestMessageBtn,
    showEditBtn,
    finalWebhookUrl,
    setHideTestMessageBtn,
    setDisbleWebhookInput,
    confirmedMessageBtn,
    handleClickConfirmBtn,
    testMessageResponse,
    testMessageBtn,
    handleTestWebhook,
    confirmBtn,
    hideTestMessageBtn,
    alertMessage,
    alertName,
    WHTestMsgTxt,
    WHTestMsgLoading,
    selectedEvent
}) => {

// Webhook support
const { isFeatureLocked: isWebHookFeatureLocked } = useFeatureLock(
    FEATURES.FEATURE_WEBHOOK
);

    const ErrorMsg = getErrorMsg(viewAlertDetails?.last_fail_details, WEBHOOK);
    return (
        <div className='border rounded mt-4' style={{ borderColor: ErrorMsg ? "red" : "auto" }}>
            <div style={{ borderRadius: '0.25rem 0.25rem 0 0', backgroundColor: '#fafafa' }}>
                <Row className={'ml-2 mr-2'}>
                    <Col span={20}>
                        <div className='flex justify-between p-3'>
                            <div className='flex'>
                                <Avatar
                                    size={40}
                                    shape='square'
                                    icon={<SVG name={'Webhook'} size={40} color='purple' />}
                                    style={{ backgroundColor: '#F5F6F8' }}
                                />
                            </div>
                            <div className='flex flex-col justify-start items-start ml-2 w-full'>
                                <div className='flex flex-row items-center justify-start'>
                                    <Text
                                        type='title'
                                        level={7}
                                        weight='medium'
                                        extraClass='m-0'
                                    >
                                        Webhook
                                    </Text>
                                    {
                                        ErrorMsg && <>
                                            <Tooltip title={ErrorMsg} color={TOOLTIP_CONSTANTS.DARK}>
                                                <div>
                                                    <SVG name={'InfoCircle'} extraClass={'ml-1'} size={16} color='red' />
                                                </div>
                                            </Tooltip>
                                        </>
                                    }
                                </div>
                                <Text
                                    type='paragraph'
                                    mini
                                    extraClass='m-0'
                                    color='grey'
                                    lineHeight='medium'
                                >
                                    Create a webhook with this event trigger and send the selected properties to other tools for automation.
                                </Text>
                                {/* <Text
                        type='paragraph'
                        mini
                        extraClass='m-0'
                        color='grey'
                        lineHeight='medium'
                      >
                        <span className='font-bold'>Note:</span> Please add
                        payload to enable this option.
                      </Text> */}
                            </div>

                        </div>
                    </Col>
                    <Col span={4} className={'m-0 mt-4 flex justify-end'}>

                        {isWebHookFeatureLocked ? (
                            <div className='p-2'>
                                <UpgradeButton featureName={FEATURES.FEATURE_WEBHOOK} />
                            </div>
                        ) : <Form.Item name='webhook_enabled' className={'m-0'}>
                            <div className={'flex flex-end items-center'}>
                                <Text
                                    type={'title'}
                                    level={7}
                                    weight='medium'
                                    extraClass={'m-0 mr-2'}
                                >
                                    Enable
                                </Text>
                                <span style={{ width: '50px' }}>
                                    <Switch
                                        checkedChildren='On'
                                        unCheckedChildren='OFF'
                                        disabled={
                                            !(
                                                groupBy &&
                                                groupBy.length &&
                                                groupBy[0] &&
                                                groupBy[0].property
                                            ) || isWebHookFeatureLocked
                                        }
                                        onChange={(checked) => setWebhookEnabled(checked)}
                                        checked={webhookEnabled}
                                    />
                                </span>{' '}
                            </div>
                        </Form.Item>
                        }
                    </Col>
                </Row>
            </div>
            {webhookEnabled && (
                <div>

                    <Row className='p-6'>
                        <Col span={12} className={'m-0'}>


                    <Row>
                        <Col span={12} className={'m-0'}>
                            <Text
                                type={'title'}
                                level={7}
                                weight='medium'
                                extraClass={'m-0'}
                            >
                                Paste your webhook URL here
                            </Text>
                        </Col>
                    </Row>
                    <Row className={'mt-1'}>
                        <Col span={16}>
                            <Input
                                className='fa-input'
                                size='large'
                                placeholder='Webhook URL'
                                disabled={disbleWebhookInput}
                                ref={webhookRef}
                                value={webhookUrl}
                                onChange={(e) => {
                                    setWebhookUrl(e.target.value);
                                    setConfirmBtn(false);
                                    setTestMessageBtn(false);
                                }}
                                onBlur={() => {
                                    if (webhookUrl === '') {
                                        setTestMessageBtn(true);
                                        setConfirmBtn(true);
                                    }
                                    if (showEditBtn && webhookUrl === finalWebhookUrl) {
                                        setHideTestMessageBtn(true);
                                        setConfirmBtn(false);
                                        setDisbleWebhookInput(true);
                                    }
                                }}
                            ></Input>
                        </Col>
                        <Col span={6} className={'m-0'}>
                            {!confirmedMessageBtn && !showEditBtn ? (
                                <Button
                                    type='link'
                                    disabled={confirmBtn}
                                    onClick={() => handleClickConfirmBtn()}
                                    size='large'
                                >
                                    Confirm
                                </Button>
                            ) : confirmedMessageBtn && !showEditBtn ? (
                                <Button
                                    type='link'
                                    disabled
                                    onClick={() => handleClickConfirmBtn()}
                                    size='large'
                                    icon={
                                        <SVG
                                            name={'Checkmark'}
                                            size={16}
                                            color={'#52C41A'}
                                            extraClass={'m-0'}
                                        />
                                    }
                                >
                                    Confirmed
                                </Button>
                            ) : (
                                <Button
                                    type='link'
                                    disabled={confirmBtn}
                                    onClick={() => {
                                        setDisbleWebhookInput(false);
                                        setConfirmBtn(true);
                                        setHideTestMessageBtn(true);
                                        setTimeout(() => {
                                            webhookRef.current.focus();
                                        }, 200);
                                    }}
                                    size='large'
                                >
                                    Edit
                                </Button>
                            )}
                        </Col>
                    </Row>
                    {hideTestMessageBtn && (
                        <Row className={'mt-2'}>
                            <Col span={24} className={'m-0'}>
                                {testMessageResponse ? (
                                    <div>
                                        <div className='inline'>
                                            <SVG
                                                name={'CheckCircle'}
                                                size={16}
                                                extraClass={'m-0 inline'}
                                            />
                                            <Text
                                                type={'title'}
                                                level={7}
                                                extraClass={'m-0 ml-1 inline'}
                                            >
                                                We've sent a sample message to this endpoint.
                                                Check and hit 'Confirm' if everything is alright!
                                            </Text>
                                        </div>
                                        <div className='inline'>
                                            <Button
                                                type='link'
                                                style={{
                                                    backgroundColor: 'white',
                                                    borderStyle: 'none'
                                                }}
                                                size='small'
                                                disabled={testMessageBtn}
                                                onClick={() => handleTestWebhook()}
                                                icon={
                                                    <SVG
                                                        name={'PaperPlane'}
                                                        size={18}
                                                        color={
                                                            testMessageBtn ? '#00000040' : '#1e89ff'
                                                        }
                                                        extraClass={'-mt-1'}
                                                    />
                                                }
                                            >
                                                Try Again
                                            </Button>
                                        </div>
                                    </div>
                                ) : (
                                    <Button
                                        type='link'
                                        disabled={testMessageBtn}
                                        style={{
                                            backgroundColor: 'white',
                                            borderStyle: 'none'
                                        }}
                                        size='small'
                                        onClick={() => handleTestWebhook()}
                                        icon={
                                            <SVG
                                                name={'PaperPlane'}
                                                size={18}
                                                color={testMessageBtn ? '#00000040' : '#1e89ff'}
                                                extraClass={'-mt-1'}
                                            />
                                        }
                                    >
                                        Test this with a sample message
                                    </Button>
                                )}
                            </Col>
                        </Row>
                    )}
                    <Row className='mt-3'>
                        <Col>
                            <Text
                                type='paragraph'
                                mini
                                extraClass='m-0'
                                color='grey'
                                lineHeight='medium'
                            >
                                Note that if you edit this alert or its payload in the
                                future, you must reconfigure the flows to support these
                                changes
                            </Text>
                        </Col>
                    </Row>


                    </Col>

                    <Col span={12} className={'m-0'}>
                        <div className='flex w-full justify-end'>
                            <PreviewCardWebhook 
                                alertMessage={alertMessage}
                                alertName={alertName}
                                groupBy={groupBy}
                                selectedEvent={selectedEvent}
                                />
                        </div>
                    </Col>

                </Row>

                <div className='border-top--thin-2 mt-4 p-4'> 
                            <Button disabled={!webhookUrl.length > 0} loading={WHTestMsgLoading} icon={WHTestMsgTxt ?  <SVG name='Checkmark' size={16}  color='grey' /> : <SVG name={'PaperPlane'} size={16} color='grey' />} ghost onClick={()=>handleTestWebhook()}>{ WHTestMsgLoading ? 'Sending...' : WHTestMsgTxt ? 'Message sent!' : 'Send test message'}</Button>  
                        </div> 
                </div>
            )}
        </div>
    )
}


export default Webhook