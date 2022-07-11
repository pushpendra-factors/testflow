import React, { useEffect, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import { SVG, Text } from '../../components/factorsComponents';
import { FaErrorComp, FaErrorLog } from '../../components/factorsComponents';
import { ErrorBoundary } from 'react-error-boundary';
import { Button} from 'antd';
import NewDashboard from './NewDashboard';
import SelectTemplates from './SelectTemplates';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';


function DashboardTemplates() {
    const history = useHistory();
    const [AddDashboardDetailsVisible, setAddDashboardDetailsVisible] = useState(false);
    const [showTemplates,setShowTemplates]=useState(false);
    const {templates}  = useSelector((state) => state.dashboardTemplates);
    return (
        <>
            {showTemplates&&
            <div className="ant-modal-wrap bg-white">
                <SelectTemplates setShowTemplates={setShowTemplates} templates={templates}/>
            </div>
            }
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
                {/* <Header> */}
                    <div className={'rounded-lg border-2 border-gray-200 w-4/5 mx-auto my-16'}>
                            <div className='w-20 float-left mt-2 ml-4 mr-4 mb-1'>
                                <img src='assets/images/NoData.png'/>
                            </div>
                            <div className={'mt-4 mb-4'}>
                                <Text type={'title'} level={4} color={'grey-2'} weight={'bold'} extraClass={'m-0 mt-2 mb-1'}>
                                    Complete Project Setup
                                </Text>
                                <Text type={'title'} level={7} color={'grey'} extraClass={'m-0 mb-1'}>
                                    A few things pending under your project setup
                                </Text>
                            </div>
                            <div className={'float-right -mt-20 pt-2 mr-8'}>
                                <Button type={'link'} style={{backgroundColor:'white'}} className={'mt-2'} onClick={()=> history.push('/project-setup')}>Setup Assist<SVG name={'Arrowright'} size={16} extraClass={'ml-1'} color={'blue'} /></Button>
                            </div>
                    </div>
                {/* </Header> */}
                <div className={`flex justify-center flex-col items-center m-auto ${styles.contentClass}`} >
                    <img alt='no-data' src='assets/images/Group 880.png' className={'mb-2 opacity-0.8'} />
                    <Text type={'title'} level={6} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>
                        Create a dashboard to moniter your metrics in one place.
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>
                        Learn <Link>Dashboard Basics</Link>
                    </Text>
                    <div className='flex flex-row mt-6 justify-center'>
                        <div onClick={() => setAddDashboardDetailsVisible(true)} className={`flex flex-row ${styles.cardnew} w-1/3 mr-6`}>
                            <div className=''>
                                <SVG name={'addNew'} extraClass={'mx-4 my-4'} width="3rem" height="3rem"/>
                            </div>
                            <div className='flex flex-col mt-4 ml-2 justify-start'>
                                <Text type='title'>
                                    Create New
                                </Text>
                                <Text type='paragraph'>
                                    Build a new Dashborad that stores all your reports in one place.
                                </Text>
                            </div>
                        </div>
                        <div onClick={() => setShowTemplates(true)} className={`flex flex-row ${styles.cardnew} w-1/3 ml-6`}>
                            <div>
                                <SVG name={'selectFromTemplates'}  extraClass={'mx-4 my-4'} width="3rem" height="3rem"/>
                            </div>
                            <div className='flex flex-col mt-4 ml-2 justify-start'>
                                <Text type='title'>
                                    Select From Templates
                                </Text>
                                <Text type='paragraph' extraClass={'mb-2'}>
                                    Pick from pre-built dashboard templates to analyse overall marketing performance.
                                </Text> 
                            </div>
                        </div>
                    </div>
                </div>
                  <NewDashboard 
                    AddDashboardDetailsVisible={AddDashboardDetailsVisible}
                    setAddDashboardDetailsVisible={setAddDashboardDetailsVisible}
                  />
            </ErrorBoundary>
        </>
    );

}

export default DashboardTemplates;

