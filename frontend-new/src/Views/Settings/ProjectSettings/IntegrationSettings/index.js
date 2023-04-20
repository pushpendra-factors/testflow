import React, { useState, useEffect } from 'react';
import {
  Row,
  Col,
  Tag,
  Avatar,
  Skeleton,
  Button,
  Tooltip,
  message
} from 'antd';
import { Text, SVG, FaErrorComp, FaErrorLog } from 'factorsComponents';
import { connect } from 'react-redux';
import { fetchProjectSettings, fetchProjectSettingsV1 } from 'Reducers/global';

// const HubspotIntegration = lazy(()=>retryDynamicImport(() => import("./Hubspot")));
// const SegmentIntegration = lazy(()=>retryDynamicImport(() => import("./Segment")));
// const DriftIntegration = lazy(()=>retryDynamicImport(() => import("./Drift")));
// const GoogleAdWords = lazy(()=>retryDynamicImport(() => import("./GoogleAdWords")));
// const FacebookIntegration = lazy(()=>retryDynamicImport(() => import("./Facebook")));
// const SalesForceIntegration = lazy(()=>retryDynamicImport(() => import("./Salesforce")));
// const LinkedInIntegration = lazy(()=>retryDynamicImport(() => import("./LinkedIn")));

import { ErrorBoundary } from 'react-error-boundary';
import HubspotIntegration from './Hubspot';
import SegmentIntegration from './Segment';
import DriftIntegration from './Drift';
import GoogleAdWords from './GoogleAdWords';
import FacebookIntegration from './Facebook';
import SalesForceIntegration from './Salesforce';
import LinkedInIntegration from './LinkedIn';
import GoogleSearchConsole from './GoogleSearchConsole';

import RevealIntegration from './Reveal';
import BingIntegration from './Bing';
import MarketoIntegration from './Marketo';
import SlackIntegration from './Slack';
import LeadSquaredIntegration from './LeadSquared';
import SixSignalIntegration from './SixSignal';
import SixSignalFactorsIntegration from './SixSignalFactors';
import RudderstackIntegration from './Rudderstack';
import MSTeamIntegration from './MSTeam';

import { ADWORDS_INTERNAL_REDIRECT_URI } from './util';
import { featureLock } from '../../../../routes/feature';

const IntegrationProviderData = [
  {
    name: 'Segment',
    desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Segment_ads',
    kbLink: 'https://help.factors.ai/en/articles/5835006-segment'
  },
  {
    name: 'Rudderstack',
    desc: 'Rudderstack is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Rudderstack_ads',
    kbLink: false
  },
  {
    name: 'Marketo',
    desc: 'Marketo is a leader in marketing automation. Using our Marketo source, we will ingest your Program, Campaign, Person and List records into Factors',
    icon: 'Marketo',
    kbLink: false
  },
  {
    name: 'Slack',
    desc: 'Does your team live on Slack? Set up alerts that track KPIs and marketing data. Nudge your team to take the right actions.',
    icon: 'Slack',
    kbLink: false
  },
  {
    name: 'Microsoft Teams',
    desc: 'Does your team live on Teams? Set up alerts that track KPIs and marketing data. Nudge your team to take the right actions.',
    icon: 'MSTeam',
    kbLink: false
  },
  {
    name: 'Hubspot',
    desc: 'Sync your Contact, Company and Deal objects with Factors on a daily basis',
    icon: 'Hubspot_ads',
    kbLink: 'https://help.factors.ai/en/articles/5099532-hubspot'
  },
  {
    name: 'Salesforce',
    desc: 'Sync your Leads, Contact, Account, Opportunity and Campaign objects with Factors on a daily basis',
    icon: 'Salesforce_ads',
    kbLink: 'https://help.factors.ai/en/articles/5099533-salesforce'
  },
  {
    name: 'Google Ads',
    desc: 'Integrate reporting from Google Search, Youtube and Display Network',
    icon: 'Google_ads',
    kbLink: 'https://help.factors.ai/en/articles/5099504-google-ads'
  },
  {
    name: 'Facebook',
    desc: 'Pull in reports from Facebook, Instagram and Facebook Audience Network',
    icon: 'Facebook_ads',
    kbLink: 'https://help.factors.ai/en/articles/5099507-facebook-ads'
  },
  {
    name: 'LinkedIn',
    desc: 'Sync LinkedIn ads reports with Factors for performance reporting',
    icon: 'Linkedin_ads',
    kbLink: 'https://help.factors.ai/en/articles/5099508-linkedin'
  },
  {
    name: 'Drift',
    desc: 'Track events and conversions from Drift’s chat solution on the website',
    icon: 'DriftLogo',
    kbLink: false
  },
  {
    name: 'Google Search Console',
    desc: 'Track organic search impressions, clicks and position from Google Search',
    icon: 'Google',
    kbLink: 'https://help.factors.ai/en/articles/5576963-google-search-console'
  },
  {
    name: 'Bing Ads',
    desc: 'Sync Bing ads reports with Factors for performance reporting',
    icon: 'Bing',
    kbLink: false
  },
  {
    name: 'Clearbit Reveal',
    desc: 'Take action as soon as a target account hits your site',
    icon: 'ClearbitLogo',
    kbLink: false
  },
  {
    name: 'LeadSquared',
    desc: 'Leadsquared is a leader in marketing automation. Using our Leadsquared source, we will ingest your Program, Campaign, Person and List records into Factors.',
    icon: 'LeadSquared',
    kbLink: false
  },
  {
    name: '6Signal by 6Sense',
    desc: 'Gain insight into who is visiting your website and where they are in the buying journey',
    icon: 'SixSignalLogo',
    kbLink: false
  },
  {
    name: 'Factors Website De-anonymization',
    desc: 'Gain insight into who is visiting your website and where they are in the buying journey',
    icon: 'Brand',
    kbLink: false
  }
];

function IntegrationCard({ item, index, defaultOpen }) {
  const [isActive, setIsActive] = useState(false);
  const [toggle, setToggle] = useState(false);
  const [isStatus, setIsStatus] = useState('');

  const loadIntegrationForm = (item) => {
    switch (item?.name) {
      case 'Hubspot':
        return (
          <HubspotIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Segment':
        return (
          <SegmentIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Rudderstack' :
        return (
          <RudderstackIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Drift':
        return (
          <DriftIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Facebook':
        return (
          <FacebookIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Salesforce':
        return (
          <SalesForceIntegration
            kbLink={item.kbLink}
            setIsActive={setIsActive}
          />
        );
      case 'Google Ads':
        return <GoogleAdWords kbLink={item.kbLink} setIsStatus={setIsStatus} />;
      case 'LinkedIn':
        return (
          <LinkedInIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'Google Search Console':
        return (
          <GoogleSearchConsole kbLink={item.kbLink} setIsStatus={setIsStatus} />
        );
      case 'Bing Ads':
        return (
          <BingIntegration kbLink={item.kbLink} setIsStatus={setIsStatus} />
        );
      case 'Marketo':
        return (
          <MarketoIntegration kbLink={item.kbLink} setIsStatus={setIsStatus} />
        );
      case 'Slack':
        return (
          <SlackIntegration kbLink={item.kbLink} setIsStatus={setIsStatus} />
        );
      case 'Microsoft Teams':
        return (
          <MSTeamIntegration kbLink={item.kbLink} setIsStatus={setIsStatus} />
        );
      case 'Clearbit Reveal':
        return (
          <RevealIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      case 'LeadSquared':
        return (
          <LeadSquaredIntegration
            kbLink={item.kbLink}
            setIsActive={setIsActive}
          />
        );
      case '6Signal by 6Sense':
        return (
          <SixSignalIntegration
            kbLink={item.kbLink}
            setIsActive={setIsActive}
          />
        );
      case 'Factors Website De-anonymization':
        return (
          <SixSignalFactorsIntegration
            kbLink={item.kbLink}
            setIsActive={setIsActive}
          />
        );
      default:
        return (
          <>
            <Tag color='orange' style={{ marginTop: '8px' }}>
              Enable from{' '}
              <a
                href='https://app-old.factors.ai/'
                target='_blank'
                rel='noreferrer'
              >
                here
              </a>
            </Tag>{' '}
          </>
        );
    }
  };

  useEffect(() => {
    setToggle(!(isActive || isStatus === 'Active'));

    if (defaultOpen) {
      setToggle(true);
    }
  }, [isActive, isStatus]);

  return (
    <div key={index} className='fa-intergration-card'>
      <ErrorBoundary
        fallback={
          <FaErrorComp
            size='medium'
            title='Bundle Error:02'
            subtitle='We are facing trouble loading App Bundles. Drop us a message on the in-app chat.'
          />
        }
        onError={FaErrorLog}
      >
        <div>
          <div
            className='flex justify-between cursor-pointer'
            onClick={() =>
              isActive || isStatus === 'Active' ? setToggle(!toggle) : null
            }
          >
            <div className='flex'>
              <Avatar
                size={40}
                shape='square'
                icon={<SVG name={item.icon} size={40} color='purple' />}
                style={{ backgroundColor: '#F5F6F8' }}
              />
            </div>
            <div className='flex flex-col justify-start items-start ml-4 w-full'>
              <div className='flex flex-row items-center justify-start'>
                <Text type='title' level={5} weight='bold' extraClass='m-0'>
                  {item.name}
                </Text>
                {(isActive || isStatus === 'Active') && (
                  <Tag color='green' style={{ marginLeft: '8px' }}>
                    Active
                  </Tag>
                )}
              </div>

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
              <Text
                type='paragraph'
                mini
                extraClass='m-0 w-9/12'
                color='grey'
                lineHeight='medium'
              >
                {item.desc}
              </Text>
            </div>
            {(isActive || isStatus === 'Active') && (
              <Button
                type='text'
                onClick={() => setToggle(!toggle)}
                icon={
                  toggle ? (
                    <SVG size={16} name='ChevronDown' />
                  ) : (
                    <SVG size={16} name='ChevronRight' />
                  )
                }
              />
            )}
          </div>
          <div className='ml-16 flex flex-col items-start'>
            {toggle && loadIntegrationForm(item)}
          </div>
        </div>
      </ErrorBoundary>
    </div>
  );
}
function IntegrationSettings({
  currentProjectSettings,
  activeProject,
  fetchProjectSettings,
  currentAgent,
  fetchProjectSettingsV1
}) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false);
    });
    fetchProjectSettingsV1(activeProject.id);
  }, [activeProject]);

  useEffect(() => {
    if (window.location.href.indexOf('?error=') > -1) {
      const searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        const error = searchParams.get('error');
        const str = error.replace('_', ' ');
        const finalmsg = str.toLocaleLowerCase();
        if (finalmsg) {
          message.error(finalmsg);
        }
      }
    }

    if (window.location.href.indexOf('status=') > -1) {
      const searchParams = new URLSearchParams(window.location.search);
      if (searchParams) {
        const error = searchParams.get('status');
        const str = error.replace('_', ' ');
        const finalmsg = str.toLocaleLowerCase();
        if (finalmsg) {
          message.error(
            `Error: ${finalmsg}. Sorry! That doesn’t seem right. Please try again`
          );
        }
      }
    }
  }, []);

  return (
    <ErrorBoundary
      fallback={
        <FaErrorComp
          size='medium'
          title='Integrations Error'
          subtitle='We are facing some issues with the integrations. Drop us a message on the in-app chat.'
        />
      }
      onError={FaErrorLog}
    >
      <div className='fa-container mt-32 mb-12 min-h-screen'>
        <Row gutter={[24, 24]} justify='center'>
          <Col span={18}>
            <div className='mb-10 pl-4'>
              <Row>
                <Col span={12}>
                  <Text type='title' level={3} weight='bold' extraClass='m-0'>
                    Integrations
                  </Text>
                </Col>
              </Row>
              <Row className='mt-4'>
                <Col span={24}>
                  {dataLoading ? (
                    <Skeleton active paragraph={{ rows: 4 }} />
                  ) : (
                    IntegrationProviderData.map((item, index) => {
                      let defaultOpen = false;
                      if (
                        window.location.href.indexOf(
                          ADWORDS_INTERNAL_REDIRECT_URI
                        ) > -1
                      ) {
                        defaultOpen = true;
                      }
                      // Flag for 6Signal Factors key
                      if (
                        (item.name === 'Factors Website De-anonymization' &&
                        !featureLock(currentAgent.email)) || (item.name === 'Microsoft Teams' && !featureLock(currentAgent.email))
                      ) {
                        return null;
                      }
                      return (
                        <IntegrationCard
                          item={item}
                          index={index}
                          key={index}
                          defaultOpen={defaultOpen}
                          currentProjectSettings={currentProjectSettings}
                        />
                      );
                    })
                  )}
                </Col>
              </Row>
            </div>
          </Col>
        </Row>
      </div>
    </ErrorBoundary>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  currentProjectSettings: state.global.currentProjectSettings,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps, {
  fetchProjectSettings,
  fetchProjectSettingsV1
})(IntegrationSettings);
