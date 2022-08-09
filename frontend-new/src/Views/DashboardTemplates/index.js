import React, { useEffect, useState } from 'react';
import { Link, useHistory } from 'react-router-dom';
import { SVG, Text } from '../../components/factorsComponents';
import SelectTemplates from './SelectTemplates';
import { useSelector } from 'react-redux';
import styles from './index.module.scss';


function DashboardTemplates({setaddDashboardModal}) {
    const history = useHistory();
    const [showTemplates,setShowTemplates]=useState(false);
    const {templates}  = useSelector((state) => state.dashboardTemplates);
    return (
        <>
            {showTemplates&&
            <div className="ant-modal-wrap bg-white">
                <SelectTemplates setShowTemplates={setShowTemplates} templates={templates}/>
            </div>
            }
            <div className={`flex justify-center flex-col items-center m-auto ${styles.contentClass}`} >
                    <img alt='no-data' src='https://s3.amazonaws.com/www.factors.ai/assets/img/product/no-data-charts.png' className={'mb-2 opacity-0.8'} />
                    
                    <Text type={'title'} level={6} weight={'bold'} color={'grey-2'} extraClass={'m-0'}>
                        Create a dashboard to moniter your metrics in one place.
                    </Text>
                    <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>
                        Learn <Link>Dashboard Basics</Link>
                    </Text>
                    <div className='flex flex-row mt-6 justify-center'>
                        <div onClick={() => setaddDashboardModal(true)} className={`flex flex-row ${styles.cardnew} w-1/3 mr-6`}>
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
    </>
    );

}

export default DashboardTemplates;

