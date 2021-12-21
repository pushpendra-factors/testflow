import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { SVG, Text } from '../../../components/factorsComponents';
import { FaErrorComp, FaErrorLog } from '../../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { Button } from 'antd';
import Header from '../../AppLayout/Header';
import AddDashboard from '../AddDashboard';

function DashboardAfterIntegration({setaddDashboardModal}) {
    // const [addDashboardModal, setaddDashboardModal] = useState(false);
    const [dataLoading, setdataLoading] = useState(true);

    useEffect(() => {
        setTimeout(() => {
            setdataLoading(false)
        },3000);
    }, []);

    // const addDashboard = () => {
    //     setaddDashboardModal(true);
    // }

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
                <Header>
                    <div className={'rounded-lg border-2 border-gray-200 w-full h-24 mt-8'}>
                            <div className='w-20 float-left mt-2 ml-4 mr-4 mb-1'>
                                <img src='assets/images/NoData.png'/>
                            </div>
                            <div className={'mt-4 mb-4'}>
                                <Text type={'title'} level={4} color={'grey-2'} weight={'bold'} extraClass={'m-0 mt-2 mb-1'}>
                                    Complete project setup
                                </Text>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mb-1'}>
                                    A few things pending under your project setup
                                </Text>
                            </div>
                            <div className={'float-right -mt-20 pt-2 mr-8'}>
                                <Link to='/project-setup' className={'text-base font-semibold'}>Setup Assist<span style={{fontSize:'30px'}}>→</span> </Link>
                            </div>
                    </div>
                </Header>

                <div
                    style={{marginTop:'24em'}}
                    className={
                    'flex justify-center flex-col items-center fa-dashboard--no-data-container'
                    }
                >
                    <img alt='no-data' src='assets/images/Group 880.png' className={'mb-2'} />
                    <Text type={'title'} level={6} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>
                        Create a dashboard to moniter your metrics in one place.
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>
                        Learn  <Link to='#' className={'text-sm font-semibold'}>Dashboard Basics {dataLoading? <span style={{fontSize:'20px'}}>→</span>:null} </Link>
                    </Text>
                    { dataLoading ? 
                    <div className={'rounded-lg border-2 border-gray-400 w-11/12 mt-6'}>
                        <Text type={'title'} level={6} color={'grey'} extraClass={'m-0 mt-2 -mb-4'}>
                           We don’t have any data yet. While we fetching your metrics,
                        </Text>
                        <Text type={'title'} level={6} color={'grey-2'} weight={'bold'} extraClass={'m-0 mb-2'}>
                           Explore our Demo Project <span style={{fontSize:'30px'}}>→</span>
                        </Text>
                    </div>
                    :
                    <div className={'mt-6'}>
                        <Button type={'primary'} size={'large'} className={'w-full'} onClick={() => setaddDashboardModal(true)}>Create your first dashboard</Button>
                        <Text type={'title'} level={7} weight={'bold'} color={'grey'} extraClass={'m-0 mt-2 mb-2'}>
                            or
                        </Text>
                        <Button type={'default'} size={'large'} className={'w-full'}>Choose from templates</Button>
                    </div>
                    }
                </div>

                {/* Add dashboard modal */}
                {/* <AddDashboard
                    addDashboardModal={addDashboardModal}
                    setaddDashboardModal={setaddDashboardModal}
                /> */}
                
            </ErrorBoundary>
        </>
    );

}

export default DashboardAfterIntegration;