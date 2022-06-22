import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton, Tooltip, message, Modal
} from 'antd';
import { Text } from 'factorsComponents';
import { projectAgentRemove, fetchAgentInfo } from 'Reducers/agentActions';
import { connect } from 'react-redux';
import { ExclamationCircleOutlined } from '@ant-design/icons';
import { fetchProjects } from "Reducers/global";

const { confirm } = Modal;

function ProjectDetails({
  fetchProjects, projects, projectAgentRemove, fetchAgentInfo, currentAgent
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const leaveProject = (projectId, agentUUID, projectName) => {
    confirm({
      title: 'Are you sure you want to leave project?',
      icon: <ExclamationCircleOutlined />,
      content: 'Please confirm to proceed',
      okText: 'Yes',
      onOk() {
        projectAgentRemove(projectId, agentUUID).then(() => {
          message.success(`Left project ${projectName}`);
        }).catch((err) => {
          message.error(err?.data?.error);
        });
      }
    });
  };
  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
    };
    getData();
    fetchProjects().then(() => {
      setDataLoading(false);
    });
  }, [fetchProjects]);
  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Your Projects</Text>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col span={24}>
            { dataLoading ? <Skeleton avatar active paragraph={{ rows: 4 }}/>
              : <>
                {projects.map((item, index) => {
                  const isAdmin = (item.role === 2);
                  return (
                    <div key={index} className="flex justify-between items-center border-bottom--thin-2 py-5" >
                      <div className="flex justify-start items-center" >
                        <Avatar size={60} shape={'square'}  src="assets/avatar/company-logo.png" />
                        <div className="flex justify-start flex-col ml-4" >
                          <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>{item.name}</Text>
                          <Text type={'title'} level={7} weight={'regular'} extraClass={'m-0 mt-1'}>{isAdmin ? 'Admin' : 'User'}</Text>
                        </div>
                      </div>
                      <div>

                      <Tooltip placement="top" trigger={'hover'} title={isAdmin ? 'Admin can\'t remove himself' : null}>
                          <Button onClick={() => leaveProject(item.id, currentAgent.uuid, item.name)} size={'large'} disabled={isAdmin} type="text">Leave Project</Button>
                      </Tooltip>

                      </div>
                    </div>
                  );
                })}
              </>
            }
          </Col>
        </Row>
      </div>

    </>

  );
}

const mapStateToProps = (state) => {
  return ({
    projects: state.global.projects,
    currentAgent: state.agent.agent_details

  }
  );
};
export default connect(mapStateToProps, { fetchProjects, projectAgentRemove, fetchAgentInfo })(ProjectDetails);
