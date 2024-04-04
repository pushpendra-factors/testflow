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
import { TOOLTIP_CONSTANTS } from 'Constants/tooltips.constans';
import { getErrorMsg, MS_TEAMS } from './../utils';
import {PreviewCardTeams} from './PreviewCard';

const Teams = ({
    viewAlertDetails,
    setTeamsEnabled,
    teamsEnabled,
    projectSettings,
    onConnectMSTeams,
    teamsSaveSelectedChannel,
    selectedWorkspace,
    setTeamsShowSelectChannelsModal,
    alertMessage,
    alertName,
    groupBy,
    sendTestTeamsMessage,
    matchEventName,
    teamsTestMsgLoading,
    teamsTestMsgTxt,
    fetchTeamsDetails
}) => {


    const [loading, setLoading] = useState(false);

    const refreshTeamsDetails = () => {
        setLoading(true);
        fetchTeamsDetails();
        setTimeout(() => {
            setLoading(false); 
        }, 5000);
    }


    const ErrorMsg = getErrorMsg(viewAlertDetails?.last_fail_details, MS_TEAMS);
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
                                    icon={<SVG name={'MSTeam'} size={40} color='purple' />}
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
                                        Teams
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
                                    Get your alerts inside Microsoft Teams. You can also choose to send the alert to multiple channels.
                                </Text>
                            </div>
                        </div>
                    </Col>
                    <Col span={4} className={'m-0 mt-4 flex justify-end'}>
                        <Form.Item name='teams_enabled' className={'m-0'}>
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
                                        onChange={(checked) => setTeamsEnabled(checked)}
                                        checked={teamsEnabled}
                                    />
                                </span>{' '}
                            </div>
                        </Form.Item>
                    </Col>
                </Row>
            </div>
            
            <div>
            <Row>

               <Col span={12} className={'m-0'}>

               {teamsEnabled && !projectSettings?.int_teams && (
                <div className='pr-4 p-6'>
                    <Row className={'m-0'}>
                        <Col  className={'m-0'}>
                            <Text
                                type={'title'}
                                level={6}
                                color={'grey'}
                                extraClass={'m-0'}
                            >
                                Teams is not integrated, Do you want to integrate with
                                your Microsoft Teams account now?
                            </Text>
                        </Col>
                    </Row>
                    <Row className={'mt-2'}>
                        <Col className={'m-0'}>
                            <Button onClick={onConnectMSTeams}>
                                <SVG name={'MSTeam'} size={20} />
                                Connect to Teams
                            </Button>

                            <div className='flex items-center mt-4'>
                                    <Text
                                        type={'title'}
                                        level={7}
                                        weight={'regular'}
                                        extraClass={'m-0'}
                                    >
                                        Have you conneted with Teams
                                    </Text>
                                    <Button type={'link'} loading={loading} onClick={()=>refreshTeamsDetails()} icon={<SVG name={'ArrowRotateRight'} size={16} />} className='fa-button-ghost ml-2'>
                                        Refresh to check
                                    </Button>
                            </div>

                        </Col>
                    </Row>
                </div>
            )}

            {teamsEnabled && projectSettings?.int_teams && (<>

                <div className='pr-4 p-6'>
                    
                            <Row>
                                <Col>
                                    <Text
                                        type={'title'}
                                        level={7}
                                        weight={'regular'}
                                        extraClass={'m-0 mt-2'}
                                    >
                                        {teamsSaveSelectedChannel?.length > 1
                                            ? `Selected channels from "${selectedWorkspace?.name}"`
                                            : teamsSaveSelectedChannel?.length == 0 ? `Select Channel` : `Selected channel from "${selectedWorkspace?.name}"`}
                                    
                                    </Text>
                                </Col>
                            </Row>

                        {teamsSaveSelectedChannel.length > 0 && (
                            <Row
                                className={'rounded border border-gray-200'}
                                style={{'width':'375px'}}
                            >
                                <Col className={'m-0'}>
                                    {teamsSaveSelectedChannel.map((channel, index) => (
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
                                </Col>
                            </Row>
                        
                        )}
                    {!teamsSaveSelectedChannel.length > 0 ? (
                        <Row className={'mt-2'}>
                            <Col span={10} className={'m-0'}>
                                <Button
                                    type={'link'}
                                    className='fa-button-ghost'
                                    onClick={() => setTeamsShowSelectChannelsModal(true)}
                                >
                                    Select Channel
                                </Button>
                            </Col>
                        </Row>
                    ) : (
                        <Row className={'mt-2'}>
                            <Col span={10} className={'m-0'}>
                                <Button
                                    type={'link'}
                                    className='fa-button-ghost'
                                    onClick={() => setTeamsShowSelectChannelsModal(true)}
                                >
                                    {teamsSaveSelectedChannel.length > 1
                                        ? 'Manage Channels'
                                        : 'Manage Channel'}
                                </Button>
                            </Col>
                        </Row>
                    )}
                    </div>
</>
)}
                </Col>
                {teamsEnabled && 
                <Col span={12} className={'m-0'}>
                                <div className='pl-4 flex w-full justify-end p-6'>
                                <PreviewCardTeams
                                    alertMessage={alertMessage}
                                    alertName={alertName}
                                    groupBy={groupBy}
                                    matchEventName={matchEventName}
                                    />
                                </div>
                            </Col>
                }


            </Row>

            {(teamsEnabled && projectSettings?.int_teams) && <>
                <div className='border-top--thin-2 mt-4 p-4'>
                            <Button disabled={!teamsSaveSelectedChannel.length > 0} loading={teamsTestMsgLoading} icon={teamsTestMsgTxt ? <SVG name='Checkmark' size={16}  color='grey' /> : <SVG name={'PaperPlane'} size={16} color='grey' />} onClick={()=>sendTestTeamsMessage()}>{ teamsTestMsgLoading ? 'Sending...' : teamsTestMsgTxt ? 'Message sent!' : 'Send test message'}</Button>
                        </div> 
            </>}

            </div>
            
        </div>

    )
}


export default Teams


