import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton
} from 'antd';
import { Text } from 'factorsComponents';
import { UserOutlined } from '@ant-design/icons';
import { connect } from 'react-redux';

function ViewBasicSettings(props) {
  const [dataLoading, setDataLoading] = useState(true);
  const [activeProject, setActiveProject] = useState(null);

  useEffect(() => {
    if (props.project) {
      setActiveProject(props.project);
      setDataLoading(false);
    }
  });

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Basic Details</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              <Button size={'large'} disabled={dataLoading} onClick={() => props.setEditMode(true)}>Edit Details</Button>
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
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{activeProject?.name}</Text>
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

const mapStateToProps = (state) => ({
  project: state.global.active_project
});

export default connect(mapStateToProps)(ViewBasicSettings);
