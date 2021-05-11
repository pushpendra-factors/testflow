import React, { useState, useEffect, lazy } from "react";
import { Row, Col, Tag, Avatar, Skeleton, Button } from "antd";
import { Text, SVG, FaErrorComp, FaErrorLog } from "factorsComponents";
import { connect } from "react-redux";
import { fetchProjectSettings } from "Reducers/global";
import retryDynamicImport from 'Utils/dynamicImport';

// const HubspotIntegration = lazy(()=>retryDynamicImport(() => import("./Hubspot")));
// const SegmentIntegration = lazy(()=>retryDynamicImport(() => import("./Segment")));
// const DriftIntegration = lazy(()=>retryDynamicImport(() => import("./Drift")));
// const GoogleAdWords = lazy(()=>retryDynamicImport(() => import("./GoogleAdWords")));
// const FacebookIntegration = lazy(()=>retryDynamicImport(() => import("./Facebook")));
// const SalesForceIntegration = lazy(()=>retryDynamicImport(() => import("./Salesforce")));
// const LinkedInIntegration = lazy(()=>retryDynamicImport(() => import("./LinkedIn")));

import HubspotIntegration from "./Hubspot";
import SegmentIntegration from "./Segment";
import DriftIntegration from "./Drift";
import GoogleAdWords from "./GoogleAdWords";
import FacebookIntegration from "./Facebook";
import SalesForceIntegration from "./Salesforce";
import LinkedInIntegration from "./LinkedIn";

import { ErrorBoundary } from "react-error-boundary";

const IntegrationProviderData = [
  {
    name: 'Segment',
    desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Segment_ads'
  },
  {
    name: 'Hubspot',
    desc: 'Sync your Contact, Company and Deal objects with Factors on a daily basis',
    icon: 'Hubspot_ads'
  },
  {
    name: 'Salesforce',
    desc: 'Sync your Leads, Contact, Account, Opportunity and Campaign objects with Factors on a daily basis',
    icon: 'Salesforce_ads'
  },
  {
    name: 'Google',
    desc: 'Integrate reporting from Google Search, Youtube and Display Network',
    icon: 'Google_ads'
  },
  {
    name: 'Facebook',
    desc: 'Pull in reports from Facebook, Instagram and Facebook Audience Network',
    icon: 'Facebook_ads'
  },
  {
    name: 'LinkedIn',
    desc: 'Sync LinkedIn ads reports with Factors for performance reporting',
    icon: 'Linkedin_ads'
  },
  {
    name: 'Drift',
    desc: 'Track events and conversions from Drift’s chat solution on the website',
    icon: 'DriftLogo'
  },
];



const IntegrationCard = ({ item, index }) => {
  const [showForm, setShowForm] = useState(false);
  const [isActive, setIsActive] = useState(false);
  const [toggle, setToggle] = useState(false);

  const loadIntegrationForm = (name) => {
    switch (name) {
      case 'Hubspot': return <HubspotIntegration setIsActive={setIsActive} />;
      case 'Segment': return <SegmentIntegration setIsActive={setIsActive} />;
      case 'Drift': return <DriftIntegration setIsActive={setIsActive} />;
      case 'Facebook': return <FacebookIntegration setIsActive={setIsActive} />;
      case 'Salesforce': return <SalesForceIntegration setIsActive={setIsActive} />;
      case 'Google': return <GoogleAdWords setIsActive={setIsActive} />;
      case 'LinkedIn': return <LinkedInIntegration setIsActive={setIsActive} />;
      default: return <><Tag color="orange" style={{ marginTop: '8px' }}>Enable from <a href="https://app-old.factors.ai/" target="_blank">here</a></Tag> </>
    }
  }

  useEffect(() => {
    setToggle(!isActive);
  }, [isActive]);

  return (
    <div key={index} className="fa-intergration-card">
      <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Bundle Error:02'} subtitle={ "We are facing trouble loading App Bundles. Drop us a message on the in-app chat."} />} onError={FaErrorLog}> 
      <div className={"flex justify-between"}>
        <div className={"flex"}>
          <Avatar
            size={40}
            shape={"square"}
            icon={<SVG name={item.icon} size={40} color={"purple"} />}
            style={{ backgroundColor: "#F5F6F8" }}
          />
        </div>
        <div className={'flex flex-col justify-start items-start ml-4 w-full'}>
          <div className={'flex items-center'}>
            <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>{item.name}</Text>
            {isActive && <Tag color="green" style={{ marginLeft: '8px' }}>Active</Tag>}
          </div>
          <Text type={'paragraph'} mini extraClass={'m-0 w-9/12'} color={'grey'} lineHeight={'medium'}>{item.desc}</Text>
          {toggle && loadIntegrationForm(item.name)}
        </div>
        {isActive &&
          <div className={'flex flex-col items-start'}>
            <Button type={'text'} onClick={() => setToggle(!toggle)} icon={toggle ? <SVG size={16} name={'ChevronDown'} /> : <SVG size={16} name={'ChevronRight'} />} />
          </div>
        }
      </div>
      </ErrorBoundary>
    </div>
  );
};
function IntegrationSettings({ currentProjectSettings, activeProject, fetchProjectSettings }) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchProjectSettings(activeProject.id).then(() => {
      setDataLoading(false); 
    }); 
  },[activeProject]);

  return (
    <>
    <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Integrations Error'} subtitle={'We are facing some issues with the integrations. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Integrations</Text>
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
  currentProjectSettings: state.global.currentProjectSettings
});


export default connect(mapStateToProps, {fetchProjectSettings})(IntegrationSettings);
