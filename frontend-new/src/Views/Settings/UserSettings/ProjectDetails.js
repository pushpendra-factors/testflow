import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton, Tooltip
} from 'antd';
import { Text } from 'factorsComponents';
import { fetchProjects } from 'Reducers/agentActions';
import { connect } from 'react-redux';

function ProjectDetails({ fetchProjects, projects }) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
    fetchProjects().then((res) => {
      setDataLoading(false);
      console.log('res->>', res);
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
                        <Avatar size={60} shape={'square'} />
                        <div className="flex justify-start flex-col ml-4" >
                          <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>{item.name}</Text>
                          <Text type={'title'} level={7} weight={'regular'} extraClass={'m-0 mt-1'}>{isAdmin ? 'Admin' : 'User'}</Text>
                        </div>
                      </div>
                      <div>

                      <Tooltip placement="top" trigger={'hover'} title={isAdmin ? 'Admin can\'t remove himself' : null}>
                          <Button size={'large'} disabled={isAdmin} type="text">Leave Project</Button>
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
    projects: state.agent.projects
  }
  );
};
export default connect(mapStateToProps, { fetchProjects })(ProjectDetails);
