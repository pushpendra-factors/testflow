import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton
} from 'antd';
import { Text } from 'factorsComponents';
import { UserOutlined } from '@ant-design/icons';

function EditUserDetails(props) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    setTimeout(() => {
      setDataLoading(false);
    }, 2000);
  });

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Profile</Text>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col>
            { dataLoading ? <Skeleton.Avatar active={true} size={104} shape={'square'} />
              : <Avatar size={104} shape={'square'} icon={<UserOutlined />} />
            }
            <Text type={'paragraph'} mini extraClass={'m-0 mt-1'} color={'grey'} >A photo helps personalise your account</Text>
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Name</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>Vishnu Baliga</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Email</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>baliga@factors.ai</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Mobile</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>+91-8123456789</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Password</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>&#8226; &#8226; &#8226; &#8226; &#8226; &#8226;</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col className={'flex justify-start items-center'}>
            <Button disabled={dataLoading} onClick={props.editDetails}>Edit Details</Button>
            <Button disabled={dataLoading} className={'ml-4'} onClick={props.editPassword} >Change Password</Button>
          </Col>
        </Row>
      </div>

    </>

  );
}

export default EditUserDetails;
