import React, { useState, useEffect } from 'react';
import {
  Row, Col, Modal, Button, Avatar, Skeleton
} from 'antd';
import { Text } from 'factorsComponents';
import { UserOutlined } from '@ant-design/icons';

function EditUserDetails() {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 2000);
  });

  const screenNotComplete = () => {
    Modal.error({
      title: 'Work in progress',
      content: 'Edit screen not updated in Figma.'
    });
  };

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Basic Details</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button disabled={dataLoading} onClick={screenNotComplete}>Edit Details</Button>
            </div>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col>
            { dataLoading ? <Skeleton.Avatar active={true} size={104} shape={'square'} />
              : <Avatar size={104} shape={'square'} icon={<UserOutlined />} />
            }
            <Text type={'paragraph'} mini extraClass={'m-0 mt-1'} color={'grey'} >A logo helps personalise your Project</Text>
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Project Name</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>FactorsAI</Text>
            }
          </Col>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Project URL</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>http://www.factors.ai</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Date Format</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>MM-DD-YY</Text>
            }
          </Col>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Time Format</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>12 Hour</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Time Zone</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>IST -- UTC +5:30 India and Sri Lanka</Text>
            }
          </Col>
        </Row>
      </div>

    </>

  );
}

export default EditUserDetails;
