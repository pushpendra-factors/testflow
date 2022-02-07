import React, { useState } from 'react';
import { Button } from 'antd';
import { Link, useHistory } from 'react-router-dom';
// import NoDataChart from '../../../components/NoDataChart';
import { SVG, Text } from '../../../components/factorsComponents';
import { PlusOutlined } from '@ant-design/icons'
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import InviteUsers from '../../Settings/ProjectSettings/UserSettings/InviteUsers';

function EmptyDashboard() {
    const [visible, setVisible] = useState(false);
    const history = useHistory();

    const handleClick = () => {
        setVisible(true);
    }

    const onCancel = () => {
        setVisible(false);
    }

    return (
        <>
            <ErrorBoundary
                fallback={
                <FaErrorComp
                    size={'medium'}
                    title={'Dashboard Overview Error'}
                    subtitle={
                    'We are facing trouble loading dashboards overview. Drop us a message on the in-app chat.'
                    }
                />
                }
                onError={FaErrorLog}
            >
                <div
                    style={{marginTop:'20em'}}
                    className={
                    'flex justify-center flex-col items-center fa-dashboard--no-data-container'
                    }
                >
                    <img alt='no-data' src='assets/images/no-data.png' className={'mb-6'} />
                    <Text type={'title'} level={5} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>
                        We don’t have enough data yet
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mt-2'}>
                        But we're almost there
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0'}>
                        Connect to at least one data source using
                    </Text>
                    <Button type={'link'} style={{backgroundColor:'white'}} onClick={()=> history.push('/project-setup')}>Setup Assist<SVG name={'Arrowright'} size={16} extraClass={'ml-1'} color={'blue'} /></Button>
                    <Button type={'text'} icon={<PlusOutlined style={{color:'gray', fontSize:'18px'}} />} onClick={handleClick}>Invite a teammate for help</Button>
                </div>

                {/* Trigger Invite Modal */}
                <InviteUsers visible = {visible} onCancel = {onCancel} />
                
            </ErrorBoundary>
        </>
    );

}

export default EmptyDashboard;