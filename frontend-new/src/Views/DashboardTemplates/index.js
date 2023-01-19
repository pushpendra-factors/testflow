import React, { useEffect, useState } from 'react';
import { Link, useHistory, useLocation } from 'react-router-dom';
import { SVG, Text } from '../../components/factorsComponents';
import SelectTemplates from './SelectTemplates';
import { useDispatch, useSelector } from 'react-redux';
import styles from './index.module.scss';
import AddDashboard from '../Dashboard/AddDashboard';
import { ADD_DASHBOARD_MODAL_OPEN } from 'Reducers/types';

function DashboardTemplates() {
  const history = useHistory();
  const dispatch = useDispatch();
  const { state } = useLocation();
  const [showTemplates, setShowTemplates] = useState(
    state?.fromSelectTemplateBtn ? true : false
  );
  const { templates } = useSelector((state) => state.dashboardTemplates);
  const handleCreateNewDashboard = () => {
    dispatch({ type: ADD_DASHBOARD_MODAL_OPEN });
  };
  return (
    <>
      {showTemplates && (
        <div className='ant-modal-wrap bg-white'>
          <SelectTemplates
            setShowTemplates={setShowTemplates}
            templates={templates}
          />
        </div>
      )}
      <div
        className={`flex justify-center flex-col items-center m-auto ${styles.contentClass}`}
      >
        <div className='mb-2'>
          <SVG
            name={'selectTemplatesBackgroundChart'}
            height='160'
            width='250'
          />
        </div>
        <Text
          type={'title'}
          level={6}
          weight={'bold'}
          color={'grey-2'}
          extraClass={'m-0'}
        >
          Create a dashboard to monitor your metrics in one place.
        </Text>
        <div className='flex flex-row my-3 justify-center'>
          <div
            onClick={handleCreateNewDashboard}
            className={`flex flex-row ${styles.cardnew} w-1/3 mr-6`}
          >
            <div className='px-6 py-4 flex flex-col items-center justify-center background-color--brand-color-1'>
              <SVG
                name={'addNew'}
                extraClass={'mr-2'}
                width='3rem'
                height='3rem'
                color={'grey'}
              />
            </div>
            <div className='flex flex-col py-4 px-2 justify-start items-start'>
              <Text type={'title'} level={7} weight={'bold'}>
                Create New
              </Text>
              <Text
                type={'title'}
                level={7}
                color={'grey'}
                extraClass={'m-0 mb-2'}
              >
                Build a new Dashborad that stores all your reports in one place.
              </Text>
            </div>
          </div>
          <div
            onClick={() => setShowTemplates(true)}
            className={`flex flex-row ${styles.cardnew} w-1/3 ml-6`}
          >
            <div className='px-6 py-4 flex flex-col items-center justify-center background-color--brand-color-1'>
              <SVG
                name={'selectFromTemplates'}
                extraClass={'mr-2'}
                width='3rem'
                height='3rem'
                color={'grey'}
              />
            </div>
            <div className='flex flex-col py-4 px-2 justify-start'>
              <Text type='title' level={7} weight={'bold'}>
                Select From Templates
              </Text>
              <Text
                type='title'
                level={7}
                color={'grey'}
                extraClass={'m-0 mb-2'}
              >
                Pick from pre-built dashboard templates to analyse overall
                marketing performance.
              </Text>
            </div>
          </div>
        </div>
        {/* <Text type={'title'} level={7} color={'grey'} weight={'bold'} extraClass={'m-0'}>
                        Learn <Link>Dashboard Basics</Link>
                    </Text> */}
      </div>
      <AddDashboard />
    </>
  );
}

export default DashboardTemplates;
