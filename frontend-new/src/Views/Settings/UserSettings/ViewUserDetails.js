import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton
} from 'antd';
import { Text } from '../../../components/factorsComponents';
import { UserOutlined } from '@ant-design/icons';
import { fetchAgentInfo } from '../../../reducers/agentActions';
import { connect } from 'react-redux';

function ViewUserDetails({
  fetchAgentInfo, editDetails, editPassword, agent
}) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchAgentInfo().then(() => {
      setDataLoading(false);
    });
  }, [fetchAgentInfo]);

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
            {dataLoading ? <Skeleton.Avatar active={true} size={104} shape={'square'}  />
              : <Avatar size={104} style={{ color: '#f56a00', backgroundColor: '#fde3cf', fontSize: '42px', textTransform: 'uppercase', fontWeight:'400' }}>{`${agent?.first_name?.charAt(0)}${agent?.last_name?.charAt(0)}`}</Avatar>
            }
            <Text type={'paragraph'} mini extraClass={'m-0 mt-1'} color={'grey'} >A photo helps personalise your account</Text>
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Name</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{`${agent?.first_name} ${agent?.last_name}`}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Email</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{agent?.email}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Mobile</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{agent?.phone}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col>
            <Text type={'title'} level={7} extraClass={'m-0'}>Password</Text>
            {dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>&#8226; &#8226; &#8226; &#8226; &#8226; &#8226;</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col className={'flex justify-start items-center'}>
            <Button size={'large'} disabled={dataLoading} onClick={editDetails}>Edit Details</Button>
            <Button size={'large'} disabled={dataLoading} className={'ml-4'} onClick={editPassword} >Change Password</Button>
          </Col>
        </Row>
      </div>

    </>

  );
}

const mapStatesToProps = (state) => {
  return {
    agent: state.agent.agent_details
  };
};
export default connect(mapStatesToProps, { fetchAgentInfo })(ViewUserDetails);
