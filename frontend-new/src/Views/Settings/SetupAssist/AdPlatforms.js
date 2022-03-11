import React, { useState, useEffect } from "react";
import { Row, Col, Tag, Avatar, Skeleton, Button, Tooltip } from "antd";
import { Text, SVG, FaErrorComp, FaErrorLog } from "factorsComponents";
import { connect } from "react-redux";
import { fetchProjectSettings } from "Reducers/global";
import retryDynamicImport from 'Utils/dynamicImport';
import 'animate.css';


import GoogleAdWords from "../ProjectSettings/IntegrationSettings/GoogleAdWords";
import FacebookIntegration from "../ProjectSettings/IntegrationSettings/Facebook";
import LinkedInIntegration from "../ProjectSettings/IntegrationSettings/LinkedIn";
import BingIntegration from '../ProjectSettings/IntegrationSettings/Bing';

import { ErrorBoundary } from "react-error-boundary";

const IntegrationProviderData = [
  {
    name: 'Google Ads',
    desc: 'Integrate reporting from Google Search, Youtube and Display Network',
    icon: 'Google_ads',
    kbLink:'https://help.factors.ai/en/articles/5099504-google-ads',
  },
  {
    name: 'Facebook',
    desc: 'Pull in reports from Facebook, Instagram and Facebook Audience Network',
    icon: 'Facebook_ads',
    kbLink:'https://help.factors.ai/en/articles/5099507-facebook-ads'
  },
  {
    name: 'LinkedIn',
    desc: 'Sync LinkedIn ads reports with Factors for performance reporting',
    icon: 'Linkedin_ads',
    kbLink:'https://help.factors.ai/en/articles/5099508-linkedin',
  },
  {
    name: 'Bing Ads',
    desc:
      'Sync Bing ads reports with Factors for performance reporting',
    icon: 'Bing',
    kbLink: false,
  },
];



const IntegrationCard = ({ item, index }) => {
  const [showForm, setShowForm] = useState(false);
  const [isActive, setIsActive] = useState(false);
  const [toggle, setToggle] = useState(false);
  const [isStatus, setIsStatus] = useState('');

  const loadIntegrationForm = (item) => {
    switch (item?.name) {
      case 'Facebook':
        return (
          <FacebookIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Google Ads':
        return <GoogleAdWords kbLink={item.kbLink} setIsStatus={setIsStatus} />;
      case 'LinkedIn':
        return (
          <LinkedInIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Bing Ads':
        return (
          <BingIntegration kbLink={item.kbLink} setIsStatus={setIsStatus} />
        );
      default:
        return (
          <>
            <Tag color='orange' style={{ marginTop: '8px' }}>
              Enable from{' '}
              <a href='https://app-old.factors.ai/' target='_blank'>
                here
              </a>
            </Tag>{' '}
          </>
        );
    }
  };

  useEffect(() => {
    setToggle(!(isActive || isStatus === 'Active'));
  }, [isActive, isStatus]);

  return (
    <div key={index} className='fa-intergration-card'>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size={'medium'}
            title={'Bundle Error:02'}
            subtitle={
              'We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
            }
          />
        }
        onError={FaErrorLog}
      >
        <div className={'flex justify-between'}>
          <div className={'flex'}>
            <Avatar
              size={40}
              shape={'square'}
              icon={<SVG name={item.icon} size={40} color={'purple'} />}
              style={{ backgroundColor: '#F5F6F8' }}
            />
          </div>
          <div
            className={'flex flex-col justify-start items-start ml-4 w-full'}
          >
            <div className={'flex items-center'}>
              <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>
                {item.name}
              </Text>
              {(isActive || isStatus === 'Active') && (
                <Tag color='green' style={{ marginLeft: '8px' }}>
                  Active
                </Tag>
              )}
              {isStatus === 'Pending' && (
                <Tooltip
                  title={
                    item.name === 'Google Ads'
                      ? 'Account(s) Selection Pending.'
                      : 'URL(s) Selection Pending.'
                  }
                >
                  <Tag color='orange' style={{ marginLeft: '8px' }}>
                    Pending!
                  </Tag>
                </Tooltip>
              )}
            </div>
            <Text
              type={'paragraph'}
              mini
              extraClass={'m-0 w-9/12'}
              color={'grey'}
              lineHeight={'medium'}
            >
              {item.desc}
            </Text>
            {toggle && loadIntegrationForm(item)}
          </div>
          {(isActive || isStatus === 'Active') && (
            <div className={'flex flex-col items-start'}>
              <Button
                type={'text'}
                onClick={() => setToggle(!toggle)}
                icon={
                  toggle ? (
                    <SVG size={16} name={'ChevronDown'} />
                  ) : (
                    <SVG size={16} name={'ChevronRight'} />
                  )
                }
              />
            </div>
          )}
        </div>
      </ErrorBoundary>
    </div>
  );
};
function IntegrationSettings({ currentProjectSettings, activeProject, fetchProjectSettings, currentAgent }) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false); 
    }); 
  },[activeProject]);

  return (
    <>
    <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Integrations Error'} subtitle={'We are facing some issues with the integrations. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>
      <div className={'animate__animated animate__fadeInUpBig animate__fast mb-10 pl-4'}>
        <Row gutter={[24, 24]} justify={'space-between'} className={'pb-2 mt-0 '}>
          <Col span={17}>
            <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Connect with your Ad platforms</Text>
          </Col>
          <Col>
            {/* <Text type={'title'} size={8} color={'grey'} extraClass={'m-0'}>LEARN MORE</Text> */}
          </Col>
        </Row>
        <Row className={'mt-4'}>
          <Col span={24}>
            {dataLoading ? <Skeleton active paragraph={{ rows: 4 }} />
              : IntegrationProviderData.map((item, index) => {
                return (
                  <IntegrationCard item={item} index={index} key={index} currentProjectSettings={currentProjectSettings} />
                );
              })
            }
          </Col>
        </Row>

      </div>
      </ErrorBoundary>
    </> 
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details,
});


export default connect(mapStateToProps, {fetchProjectSettings})(IntegrationSettings);
