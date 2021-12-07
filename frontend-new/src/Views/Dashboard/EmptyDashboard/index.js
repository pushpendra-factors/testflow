import { Anchor, Button } from 'antd';
import React from 'react';
import { Link } from 'react-router-dom';
// import NoDataChart from '../../../components/NoDataChart';
import { SVG, Text } from '../../../components/factorsComponents';
import { ArrowRightOutlined } from '@ant-design/icons'

function EmptyDashboard() {

    return (
        <>
            {/* <ErrorBoundary
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
            > */}

                {/* <div className='flex justify-center items-center w-full h-full'>
                    <NoDataChart />
                </div> */}
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
                        Few more steps to go
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mb-4'}>
                        To Start moniter your metrics, complete setting using
                    </Text>
                    <Link to='/project-setup' className={'text-base font-semibold'}>Setup Assist  <ArrowRightOutlined style={{fontSize:'20px'}}/></Link>
                </div>
                
            {/* </ErrorBoundary> */}
        </>
    );

}

export default EmptyDashboard;