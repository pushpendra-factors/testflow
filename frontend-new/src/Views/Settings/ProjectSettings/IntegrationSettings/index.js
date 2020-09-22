import React, { useState, useEffect } from 'react';
import {
  Row, Col, Switch, Avatar, Skeleton
} from 'antd';
import { Text } from 'factorsComponents';
import { PictureOutlined } from '@ant-design/icons';
const IntegrationCard = ({ item, index }) => {
  console.log('key', item);
  return (
        <div key={index} className="fa-intergration-card">
            <div className={'flex flex-col'}>
                <div className={'flex justify-between items-center'}>
                    <div className={'flex justify-between items-center'}>
                        <Avatar size={46} shape={'square'} icon={<PictureOutlined />} />
                        <div className={'flex flex-col justify-start items-start ml-4'}>
                            <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>{item}</Text>
                            <Text type={'title'} level={7} weight={'thin'} color={'grey'} extraClass={'m-0'}>By Factors</Text>
                        </div>
                    </div>
                    <div><span size={'small'} style={{ width: '50px' }}><Switch checkedChildren="On" unCheckedChildren="OFF" defaultChecked={(index / 2 === 0)} /></span> </div>
                </div>
                <Text type={'paragraph'} mini extraClass={'m-0 mt-4'} color={'grey'} lineHeight={'medium'}>Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. </Text>
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
              : ['Segment', 'Slack', 'Mailchimp', 'Hubspot'].map((item, index) => {
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
