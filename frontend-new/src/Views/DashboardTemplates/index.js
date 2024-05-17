import React, { useCallback, useEffect, useState } from 'react';
import { Link, useHistory, useLocation } from 'react-router-dom';
import { SVG, Text } from '../../components/factorsComponents';
import SelectTemplates from './SelectTemplates';
import { useDispatch, useSelector } from 'react-redux';
import styles from './index.module.scss';
import AddDashboard from '../Dashboard/AddDashboard';
import { StartFreshImage } from 'Constants/templates.constants';
import TemplatesThumbnail from 'Constants/templates.constants';
import {
  ADD_DASHBOARD_MODAL_OPEN,
  NEW_DASHBOARD_TEMPLATES_MODAL_OPEN,
  UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE
} from 'Reducers/types';
import EmptyScreen from 'Components/EmptyScreen';
import { PlayCircleFilled } from '@ant-design/icons';
import { Button, Modal } from 'antd';
import HowToCreateDashboardWebM from './../../assets/images/illustrations/HowToCreateDashboard.webm';
function DashboardTemplates() {
  const dispatch = useDispatch();
  const history = useHistory();
  const { state } = useLocation();
  const [howToModal, setHowToModal] = useState(false);
  const [showTemplates, setShowTemplates] = useState(
    state?.fromSelectTemplateBtn ? true : false
  );
  const { templates } = useSelector((state) => state.dashboardTemplates);
  const [selectedTemplateFirst, setSelectedTemplateFirst] = useState('');
  const onClick = (templateID) => {
    dispatch({
      type: UPDATE_PICKED_FIRST_DASHBOARD_TEMPLATE,
      payload: templateID
    });

    dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
    // setSelectedTemplateFirst();
  };
  const templatesToShow = [
    {
      title: 'Blank',
      description: 'Start fresh and add your own widgets',
      image: StartFreshImage,
      onClick: () => {
        dispatch({ type: ADD_DASHBOARD_MODAL_OPEN });
      }
    },
    {
      title: 'Web Analytics',
      description: 'Track your main Web KPIs and more with one click',
      image: TemplatesThumbnail.get('webanalytics')?.image,

      id: BUILD_CONFIG.firstTimeDashboardTemplates?.webanalytics
    },
    {
      title: 'Website Visitor Identification',
      description: 'See which companies are on your page.',
      image: TemplatesThumbnail.get('websitevisitoridentification')?.image,

      id: BUILD_CONFIG.firstTimeDashboardTemplates?.websitevisitoridentification
    },
    {
      title: 'All Paid Marketing',
      description: 'Keep track of your marketing spends performance',
      image: TemplatesThumbnail.get('allpaidmarketing')?.image,

      id: BUILD_CONFIG.firstTimeDashboardTemplates?.allpaidmarketing
    }
  ];
  const handleOpenTemplateModal = useCallback(() => {
    dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN });
  }, []);

  return (
    <>
      {/* {showTemplates && (
        <div className='ant-modal-wrap bg-white'>
          <SelectTemplates
            setShowTemplates={setShowTemplates}
            templates={templates}
          />
        </div>
      )} */}
      <EmptyScreen
        image={
          <SVG
            name={'selectTemplatesBackgroundChart'}
            height='160'
            width='250'
          />
        }
        title={
          <Text
            type={'title'}
            level={4}
            weight={'bold'}
            color={'grey-2'}
            extraClass={'m-0'}
          >
            Organise your reports into dashboards
          </Text>
        }
        description={
          <>
            <span>Create a dashboard to add reports to it</span>
            <br />
            <div className='pt-4'>
              <span>
                {' '}
                <Button
                  type='link'
                  icon={<PlayCircleFilled style={{fontSize:'18px'}} />}
                  onClick={() => setHowToModal(true)}
                >
                  How to create a dashboard?
                </Button>{' '}
                <Modal
                  visible={howToModal}
                  onCancel={() => setHowToModal(false)}
                  footer={null}
                  width={'60vw'}
                  className='fa-modal--regular'
                >
                  <Text
                    type={'title'}
                    level={5}
                    weight={'bold'}
                    extraClass={'mb-4'}
                  >
                    Create a dashboard or scratch or start with a template
                  </Text>
                  <video
                    autoPlay
                    loop
                    style={{
                      height: '100%',
                      width: '100%',
                      objectFit: 'cover',
                      clipPath: 'inset(1px 1px)',
                      borderRadius: '16px'
                    }}
                  >
                    <source src={HowToCreateDashboardWebM} type='video/webm' />
                  </video>
                </Modal>
              </span>
            </div>
          </>
        }
        ActionButton={{
          text: 'New Dashboard',
          props: { type: 'dashed', style: { width: 232 } },
          onClick: handleOpenTemplateModal,
          tooltip: {
            title:
              'You can start creating dashboard from a blank slate or choose from our curated template list',
            placement: 'bottom'
          }
        }}
      />
      {/* <div
        className={`flex justify-center flex-col items-center m-auto ${styles.contentClass}`}
      >
        <div className='text-center'>
          <div style={{ width: 'min-content', margin: '0 auto' }}>
            <SVG
              name={'selectTemplatesBackgroundChart'}
              height='160'
              width='250'
            />
          </div>
          <Text
            type={'title'}
            level={4}
            weight={'bold'}
            color={'grey-2'}
            extraClass={'m-0'}
          >
            Hey there 👋
          </Text>
          <Text
            type={'title'}
            level={6}
            weight={'bold'}
            color={'grey-2'}
            extraClass={'m-0'}
          >
            Start fresh or choose from templates
          </Text>
        </div>
        <div className={`${styles.firstDashboardChoicesContainer}`}>
          <div className={`${styles.firstViewMoreTemplates}`}>
            <Link
              onClick={() =>
                dispatch({ type: NEW_DASHBOARD_TEMPLATES_MODAL_OPEN })
              }
            >
              View more templates →
            </Link>
          </div>
          <div className={`flex flex-row my-3 justify-center`}>
            {templatesToShow.map((eachTemplate, eachIndex) => {
              return (
                <div
                  className={styles.firstChoiceTemplatesItem}
                  key={eachIndex}
                  onClick={
                    eachIndex === 0
                      ? eachTemplate.onClick
                      : () => onClick(eachTemplate.id)
                  }
                >
                  <div>
                    {' '}
                    <img src={eachTemplate.image} alt={eachTemplate.title} />
                  </div>
                  <div className={styles.firstChoiceTemplatesItemContent}>
                    <Text
                      type={'title'}
                      level={5}
                      weight={'bold'}
                      color={'grey-2'}
                      extraClass={'m-0'}
                    >
                      {eachTemplate.title}
                    </Text>
                    <div>{eachTemplate.description}</div>
                  </div>
                </div>
              );
            })}
          </div>
        </div>
        <div className='text-center pt-2'>
          Learn{' '}
          <Link href='https://help.factors.ai/en/articles/6294988-dashboards'>
            Dashboard Basics
          </Link>
        </div>
      </div> */}
      <AddDashboard />
    </>
  );
}

export default DashboardTemplates;
