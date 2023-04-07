import { Col, Row } from 'antd';
import { SVG, Text } from 'Components/factorsComponents';
import React, { useState } from 'react';
import JavascriptSDK from 'Views/Settings/ProjectSettings/SDKSettings/JavascriptSDK';
import styles from './index.module.scss';
const OnBoard1 = () => {
  return (
    <div className={styles['onBoardContainer']}>
      <JavascriptSDK isOnBoardFlow={true} />
      {/* <Row
        style={{
          width: 'min-content',
          display: 'flex',
          flexWrap: 'nowrap',
          padding: '15px 15px 15px 0',
          border: '1px solid #f5f5f5',
          borderRadius: '5px',
          margin: '10px 15px',
          cursor: 'pointer'
        }}
      >
        <Col
          style={{ display: 'flex', alignItems: 'center', padding: '0 10px' }}
        >
          {' '}
          <SVG name='sendArrow' />{' '}
        </Col>
        <Col style={{ whiteSpace: 'nowrap' }}>
          <Row>
            {' '}
            <Text type={'title'} level={7} weight={'bold'} extraClass={'m-0'}>
              Send your snippet and instructions to your teammate
            </Text>
          </Row>
          <Row>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Nice, tidy instructions for your favourite engineer.
            </Text>
          </Row>
        </Col>
      </Row> */}
    </div>
  );
};

export default OnBoard1;
