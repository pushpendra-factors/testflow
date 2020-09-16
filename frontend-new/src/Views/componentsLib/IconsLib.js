import React from 'react';
import {
  Breadcrumb, Row, Col, Divider
} from 'antd';
import { SVG, Text } from 'factorsComponents';

const iconList = [
  'Brand',
  'Delete',
  'Home',
  'Corequery',
  'Key',
  'Bug',
  'Report',
  'Notify',
  'Hexagon',
  'Plus',
  'Mouseevent',
  'Teamfeed',
  'Play',
  'Fav',
  'Virtual',
  'Events_cq',
  'Funnels_cq',
  'Campaigns_cq',
  'Attributions_cq',
  'Templates_cq',
  'Help',
  'Filter',
  'Trash',
  'Times',
  'Calendar',
  'Groupby',
  'Sortdown'
];

class CheckBoxLib extends React.Component {
  render() {
    return (
      <>
        <div className="mt-20 mb-8">
          <Divider orientation="left">
            <Breadcrumb>
              <Breadcrumb.Item> Components </Breadcrumb.Item>
              <Breadcrumb.Item> Icons </Breadcrumb.Item>
            </Breadcrumb>
          </Divider>
        </div>

        <Row>
          <Col span={20}>
            <div className={'flex justify-start items-center flex-wrap'}>
              {iconList.map((icon, index) => {
                return (
                  <div key={index} className={'fa-icon--container m-0 mr-4 mb-4'}>
                    <SVG name={icon} color={'purple'} />
                    <p>{icon}</p>
                  </div>
                );
              })}
            </div>
          </Col>
        </Row>

      </>

    );
  }
}

export default CheckBoxLib;
