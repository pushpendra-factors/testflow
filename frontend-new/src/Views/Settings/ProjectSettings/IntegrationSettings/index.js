import React, { useState, useEffect } from 'react';
import {
  Row, Col, Switch, Avatar, Skeleton
} from 'antd';
import { Text, SVG } from 'factorsComponents';
import { PictureOutlined } from '@ant-design/icons';


const IntegrationProviderData = [
  {
    name: 'Segment',
    desc: 'Segment is a Customer Data Platform (CDP) that simplifies collecting and using data from the users of your digital properties and SaaS applications',
    icon:'Segment_ads'
  },
  {
    name: 'Hubspot',
    desc: 'Sync your Contact, Company and Deal objects with Factors on a daily basis',
    icon:'Hubspot_ads'
  },
  {
    name: 'Salesforce',
    desc: 'Sync your Leads, Contact, Account, Opportunity and Campaign objects with Factors on a daily basis',
    icon:'Salesforce_ads'
  },
  {
    name: 'Google',
    desc: 'Integrate reporting from Google Search, Youtube and Display Network',
    icon:'Google_ads'
  },
  {
    name: 'Facebook',
    desc: 'Pull in reports from Facebook, Instagram and Facebook Audience Network',
    icon:'Facebook_ads'
  },
  {
    name: 'LinkedIn',
    desc: 'Sync LinkedIn ads reports with Factors for performance reporting',
    icon:'Linkedin_ads'
  },
  {
    name: 'Drift',
    desc: 'Track events and conversions from Drift’s chat solution on the website',
    icon:'DriftLogo'
  },
];

const IntegrationCard = ({ item, index }) => { 
  return (
        <div key={index} className="fa-intergration-card">
            <div className={'flex flex-col'}>
                <div className={'flex justify-between items-center'}>
                    <div className={'flex justify-between items-center'}>
                        <Avatar size={46} shape={'square'} icon={ <SVG name={item.icon} size={46} color={'purple'} />} style={{backgroundColor: '#F5F6F8'}} />
                        <div className={'flex flex-col justify-start items-start ml-4'}>
                            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>{item.name}</Text>
                            <Text type={'title'} level={7} weight={'thin'} color={'grey'} extraClass={'m-0'}>By Factors</Text>
                        </div>
                    </div>
                    <div><span size={'small'} style={{ width: '50px' }}><Switch checkedChildren="On" unCheckedChildren="OFF" 
                    // defaultChecked={(index / 2 === 0)} 
                    /></span> </div>
                </div>
                <Text type={'paragraph'} mini extraClass={'m-0 mt-4'} color={'grey'} lineHeight={'medium'}>{item.desc}</Text>
            </div>
        </div>
  );
};
function IntegrationSettings() {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 500);
  });

  return (
    <>
        <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Integrations</Text>
          </Col>
        </Row>
        <Row className={'mt-4'}>
            <Col span={24}>
            { dataLoading ? <Skeleton active paragraph={{ rows: 4 }}/>
              : IntegrationProviderData.map((item, index) => {
                return (
                        <IntegrationCard item={item} index={index} key={index} />
                );
              })
            }
            </Col>
        </Row>

      </div>
    </>

  );
}

export default IntegrationSettings;
