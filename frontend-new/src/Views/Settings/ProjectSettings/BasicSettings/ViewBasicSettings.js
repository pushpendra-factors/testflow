import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton
} from 'antd';
import { Text } from 'factorsComponents';
import { UserOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';
import { fetchAgentInfo } from 'Reducers/agentActions';

function ViewBasicSettings({ activeProject, setEditMode, fetchAgentInfo }) {
  const [dataLoading, setDataLoading] = useState(true);
  // const [activeProject, setActiveProject] = useState(null);

  useEffect(() => {
    fetchAgentInfo().then(() => {
      setDataLoading(false);
    });
  }, [activeProject]);

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Basic Details</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} disabled={dataLoading} onClick={() => setEditMode(true)}>Edit Details</Button>
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
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{activeProject.name ? activeProject.name : '---'}</Text>
            }
          </Col>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Project URL</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{activeProject.project_uri ? activeProject.project_uri : '---'}</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Date Format</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{activeProject.date_format ? activeProject.date_format : '---'}</Text>
            }
          </Col>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Time Format</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{activeProject.time_format ? activeProject.time_format : '---' }</Text>
            }
          </Col>
        </Row>
        <Row className={'mt-6'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>Time Zone</Text>
            { dataLoading ? <Skeleton.Input style={{ width: 200 }} active={true} size={'small'} />
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{activeProject.time_zone ? activeProject.time_zone : '---'}</Text>
            }
          </Col>
        </Row>
      </div>

    </>

  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});

export default connect(mapStateToProps, { fetchAgentInfo })(ViewBasicSettings);
