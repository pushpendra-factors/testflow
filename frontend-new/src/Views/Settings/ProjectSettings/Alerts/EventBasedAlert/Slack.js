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
import {PreviewCardSlack} from './PreviewCard';
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import { ConsoleSqlOutlined } from '@ant-design/icons';
import { getErrorMsg, SLACK } from './../utils';

const Slack = ({
    slackEnabled,
    setSlackEnabled,
    projectSettings,
    onConnectSlack,
    saveSelectedChannel,
    setSaveSelectedChannel,
    setShowSelectChannelsModal,
    viewAlertDetails,
    selectedMentions, 
    setSelectedMentions,
    slack_users,
    sendTestSlackMessage,
    alertMessage,
    alertName,
    groupBy,
    fetchSlackDetails
}) => {

    const [form] = Form.useForm();
    const [slackUsers, setSlackUsers] = useState([]);

    useEffect(() => {
        if (slack_users) {
            let slackUserList = slack_users?.filter((item) => !item?.deleted).map((item) => {
                return {
                    "value": item?.name,
                    "title": item?.real_name ? item?.real_name : item?.name,
                    "label": item?.real_name ? item?.real_name : item?.name,
                }
            })
            setSlackUsers(slackUserList);
        }
    }, [slack_users])

    const onPreventMouseDown = (event) => {
        event.preventDefault();
        event.stopPropagation();
    };

    const tagRender = (props) => {
        const { label, value, closable, onClose } = props;
        const onPreventMouseDown = (event) => {
            event.preventDefault();
            event.stopPropagation();
        };
        return (
            <Tag
                onMouseDown={onPreventMouseDown}
                closable={closable}
                onClose={onClose}
                style={{
                    marginRight: 3,
                }}
            >
                {label}
            </Tag>
        );
    };

    const onMentionChange = (value) => {
        setSelectedMentions(value);
    }; 

    const ErrorMsg = getErrorMsg(viewAlertDetails?.last_fail_details, SLACK);
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
                                    icon={<SVG name={'slack'} size={40} color='purple' />}
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
                                        Slack
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
                                    Get your alerts inside your Slack channel. You can also choose to send the alert to multiple channels.
                                </Text>
                            </div>
                        </div>
                    </Col>
                    <Col span={4} className={'m-0 mt-4 flex justify-end'}>
                        <Form.Item name='slack_enabled' className={'m-0'}>
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
                                        onChange={(checked) => setSlackEnabled(checked)}
                                        checked={slackEnabled}
                                    />
                                </span>{' '}
                            </div>
                        </Form.Item>
                    </Col>
                </Row>
            </div>
            {slackEnabled && !projectSettings?.int_slack && (
                <div className='p-4'>
                    <Row className={'mt-2 ml-2'}>
                        <Col span={10} className={'m-0'}>
                            <Text
                                type={'title'}
                                level={6}
                                color={'grey'}
                                extraClass={'m-0'}
                            >
                                Slack is not integrated, Do you want to integrate with
                                your slack account now?
                            </Text> 

                            <Button onClick={onConnectSlack} className='mt-2' icon={<SVG name={'slack'} size={16} />} >
                                Connect to slack
                            </Button>

                            <div className='flex items-center mt-4'>
                                    <Text
                                        type={'title'}
                                        level={7}
                                        weight={'regular'}
                                        extraClass={'m-0'}
                                    >
                                        Have you conneted with slack
                                    </Text>
                                    <Button ghost onClick={()=>fetchSlackDetails()} icon={<SVG name={'ArrowRotateRight'} size={16} />} className='ml-2'>
                                        Refresh to check
                                    </Button>
                            </div>
                        </Col>
                    </Row>
                </div>
            )}
            {slackEnabled && projectSettings?.int_slack && (
                <div>
                        <Row className='p-6'>
                            <Col span={12} className='pr-4'>

                                <Text
                                    type={'title'}
                                    level={7}
                                    weight={'regular'}
                                    extraClass={'m-0 mt-2 ml-2'}
                                >
                                    {saveSelectedChannel.length > 1
                                        ? 'Selected Channels'
                                        : 'Select Channel'}
                                </Text>
                                {saveSelectedChannel.length > 0 && (
                                    <div
                                        className={'rounded border border-gray-200 ml-2'}
                                        style={{'width':'375px'}}
                                    >
                                        <div className={'m-0'}>
                                            {saveSelectedChannel.map((channel, index) => (
                                                <div key={index}>
                                                    <Text
                                                        type={'title'}
                                                        level={7}
                                                        color={'grey'}
                                                        extraClass={'m-0 ml-4 my-2'}
                                                    >
                                                        {'#' + channel.name}
                                                    </Text>
                                                </div>
                                            ))}
                                        </div>
                                    </div>
                                )}

                                {!saveSelectedChannel.length > 0 ? (
                                    <div className={'mt-2 ml-2'}>
                                        <Button
                                            type={'link'}
                                            onClick={() => setShowSelectChannelsModal(true)}
                                        >
                                            Select Channel
                                        </Button>
                                    </div>
                                ) : (
                                    <div className={'mt-2 ml-2'}>
                                        <Button
                                            type={'link'}
                                            onClick={() => setShowSelectChannelsModal(true)}
                                        >
                                            {saveSelectedChannel.length > 1
                                                ? 'Manage Channels'
                                                : 'Manage Channel'}
                                        </Button>
                                    </div>
                                )}

{slackUsers?.length>0 ? <>
                                <div className={'mt-6 ml-2'}>
                                    <Text
                                        type={'title'}
                                        level={7}
                                        weight={'regular'}
                                        extraClass={'m-0 mt-2 ml-2'}
                                    >
                                        Mentions
                                    </Text>
                                    <div className='mr-4'>
                                        <Select
                                            mode="multiple"
                                            showArrow
                                            tagRender={tagRender}
                                            onChange={onMentionChange}
                                            size='large'
                                            options={slackUsers}
                                            className={'fa-select'}
                                            value={selectedMentions}
                                            style={{'width':'375px'}}
                                        />

                                    </div>
                                    <div className='mt-10 flex' style={{'width':'375px'}}>
                                        <SVG name={'InfoCircle'} size={24} color='#FAAD14' /> 
                                        <Text
                                            type={'title'}
                                            level={7}
                                            weight={'regular'}
                                            extraClass={'m-0 ml-1'}
                                        >
                                            Adding mentions for users will send them a notification for every alert. Tag responsibly :)
                                        </Text>
                                    </div>


                                </div>
                                </> : <>
                                
                                <div className={'mt-6 ml-2'}>
                                    <Text
                                        type={'title'}
                                        level={7}
                                        weight={'regular'}
                                        extraClass={'m-0 mt-2 ml-2'}
                                    >
                                        Your current Slack integration doesn’t have mentions. Simply reintegrate Slack with Factors to mention users in alerts.
                                    </Text>
                                    <Button className='mt-2' onClick={onConnectSlack} icon={<SVG name={'slack'} size={16} />}>Reintegrate now</Button>

                                    <div className='flex items-center mt-4'>
                                    <Text
                                        type={'title'}
                                        level={7}
                                        weight={'regular'}
                                        extraClass={'m-0'}
                                    >
                                        Have you reintegrated?
                                    </Text>
                                    <Button ghost onClick={()=>fetchSlackDetails()} icon={<SVG name={'ArrowRotateRight'} size={16} />} className='ml-2'>
                                        Refresh to check
                                    </Button>
                            </div>
                                </div>
                                </>
                                
                                }
                                

                            </Col>

                            <Col span={12} className={'m-0 pl-4'}>
                                <div className='flex w-full justify-end'>

                                <PreviewCardSlack 
                                    alertMessage={alertMessage}
                                    alertName={alertName}
                                    groupBy={groupBy}
                                    selectedMentions={selectedMentions}
                                    />
                                </div>
                            </Col>

                            


                        </Row>

                    <div className='border-top--thin-2 mt-4 p-4'>
                            <Button disabled={!saveSelectedChannel.length > 0} icon={<SVG name={'PaperPlane'} size={16} color='grey' />} ghost onClick={()=>sendTestSlackMessage()}>Send test message</Button>  
                        </div> 

                </div>
            )}
        </div>
    )
}


export default Slack