import React, { useState, useEffect, lazy } from "react";
import { Row, Col, Tag, Avatar, Skeleton, Button } from "antd";
import { Text, SVG, FaErrorComp, FaErrorLog } from "factorsComponents";
import { connect } from "react-redux";
import { fetchProjectSettings } from "Reducers/global";
import retryDynamicImport from 'Utils/dynamicImport';
import 'animate.css';


import SegmentIntegration from "../ProjectSettings/IntegrationSettings/Segment";

import { ErrorBoundary } from "react-error-boundary";
import JavascriptSDK from "../ProjectSettings/SDKSettings/JavascriptSDK";

const IntegrationProviderData = [
  {
    name: 'Segment',
    desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon: 'Segment_ads',
    kbLink:false,
  },
];



const IntegrationCard = ({ item, index }) => {
  const [showForm, setShowForm] = useState(false);
  const [isActive, setIsActive] = useState(false);
  const [toggle, setToggle] = useState(false);

  const loadIntegrationForm = (item) => {
    switch (item?.name) {
      case 'Segment': return <SegmentIntegration kbLink={item.kbLink} setIsActive={setIsActive} />;
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
  },[activeProject]);

  return (
    <>
    <ErrorBoundary fallback={<FaErrorComp size={'medium'} title={'Integrations Error'} subtitle={'We are facing some issues with the integrations. Drop us a message on the in-app chat.'} />} onError={FaErrorLog}>
      <div className={'animate__animated animate__fadeInUpBig mb-10 pl-4'}>
        <Row gutter={[24, 24]} justify={'space-between'} className={'mt-0 pl-3'}>
          <Col span={17}>
            <Text type={'title'} level={5} weight={'bold'} extraClass={'pb-2 m-0'}>Connect with your website data</Text>
          </Col>
          <Col>
            <Text type={'title'} size={8} color={'grey'} extraClass={'m-0'}>LEARN MORE</Text>
          </Col>
        </Row>
        <JavascriptSDK />

        <Text type={'title'} level={5} weight={'bold'} align={'center'} color={'grey'} extraClass={'pb-2 m-0'}>OR</Text>

        <Row gutter={[24, 24]} justify={'space-between'} className={'pt-4 pb-2 mt-0 pl-4'}>
          <Col span={17}>
            <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>With your Customer Data Platform</Text>
          </Col>
          <Col>
            <Text type={'title'} size={8} color={'grey'} extraClass={'m-0'}>LEARN MORE</Text>
          </Col>
        </Row>
        <Row className={'mt-4 pl-4'}>
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

        {/* <Row gutter={[24, 24]} justify={'space-between'} className={'pt-8 pb-2 mt-0 '}>
          <Col span={17}>
            <Text type={'title'} level={5} weight={'bold'} extraClass={'m-0'}>Connect with your website data</Text>
          </Col>
          <Col>
            <Text type={'title'} size={8} color={'grey'} extraClass={'m-0'}>LEARN MORE</Text>
          </Col>
        </Row> */}
        {/* <JavascriptSDK /> */}
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
