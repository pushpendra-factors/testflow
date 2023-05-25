import React, { useState, useEffect } from "react";
import { Row, Col, Tag, Avatar, Skeleton, Button, Tooltip } from "antd";
import { Text, SVG, FaErrorComp, FaErrorLog } from "factorsComponents";
import { connect } from "react-redux";
import { fetchProjectSettings } from "Reducers/global";
import retryDynamicImport from 'Utils/dynamicImport';
import 'animate.css';

import HubspotIntegration from "../ProjectSettings/IntegrationSettings/Hubspot";
import SalesForceIntegration from "../ProjectSettings/IntegrationSettings/Salesforce";
import LeadSquaredIntegration from "../ProjectSettings/IntegrationSettings/LeadSquared";

import { ErrorBoundary } from "react-error-boundary";

const IntegrationProviderData = [
  {
    name: 'Hubspot',
    desc: 'Sync your Contact, Company and Deal objects with Factors on a daily basis',
    icon: 'Hubspot_ads',
    kbLink: 'https://help.factors.ai/en/articles/7261985-hubspot-integration',
  },
  {
    name: 'Salesforce',
    desc: 'Sync your Leads, Contact, Account, Opportunity and Campaign objects with Factors on a daily basis',
    icon: 'Salesforce_ads',
    kbLink: 'https://help.factors.ai/en/articles/7261989-salesforce-integration',
  },
  {
    name: 'LeadSquared',
    desc: 'Leadsquared is a leader in marketing automation. Using our Leadsquared source, we will ingest your Program, Campaign, Person and List records into Factors.',
    icon: 'LeadSquared',
    kbLink: 'https://help.factors.ai/en/articles/7283684-leadsquared-integration',
  },
];



const IntegrationCard = ({ item, index }) => {
  const [showForm, setShowForm] = useState(false);
  const [isActive, setIsActive] = useState(false);
  const [toggle, setToggle] = useState(false);

  const loadIntegrationForm = (item) => {
    switch (item?.name) {
      case 'Hubspot': return <HubspotIntegration kbLink={item.kbLink} setIsActive={setIsActive} />;
      case 'Salesforce': return <SalesForceIntegration kbLink={item.kbLink} setIsActive={setIsActive} />;
      case 'LeadSquared':
        return (
          <LeadSquaredIntegration kbLink={item.kbLink} setIsActive={setIsActive} />
        );
      default: return <><Tag color="orange" style={{ marginTop: '8px' }}>Enable from <a href="https://app-old.factors.ai/" target="_blank">here</a></Tag> </>
    }
  }

  useEffect(() => {
    setToggle(!isActive);
  }, [isActive]);

  return (
    <div key={index} className="fa-intergration-card">
      <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Bundle Error:02'} subtitle={"We are facing trouble loading App Bundles. Drop us a message on the in-app chat."} />} onError={FaErrorLog}>
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
            {toggle && loadIntegrationForm(item)}
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
  }, [activeProject]);

  return (
    <>
      <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Integrations Error'} subtitle={'We are facing some issues with the integrations. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>
        <div className={'animate__animated animate__fadeInUpBig animate__fast mb-10 pl-4'}>
          <Row gutter={[24, 24]} justify={'space-between'} className={'pb-2 mt-0 '}>
            <Col span={17}>
              <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Connect with your CRMs</Text>
            </Col>
            <Col>
              {/* <Text type={'title'} size={8} color={'grey'} extraClass={'m-0'}>LEARN MORE</Text> */}
            </Col>
          </Row>
          <Row>
            <Col span={24}>
              <Text type={'title'} level={7} color={'grey-2'} extraClass={'m-0 mb-1'}>Your CRM data will be available on the platform between 24-48 hours from the point of integration. Once it’s available, you’ll have the last 30 days of data to start working with.</Text>
              <Text type={'title'} level={7} color={'grey-2'} extraClass={'m-0'}>You can choose up to 1 CRM account to integrate with. </Text>
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


export default connect(mapStateToProps, { fetchProjectSettings })(IntegrationSettings);
