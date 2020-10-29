import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton
} from 'antd';
import { Text } from 'factorsComponents';
import { fetchProjects } from 'Reducers/agentActions';
import { connect } from 'react-redux';

function ProjectDetails({ fetchProjects, projects }) {
  const [dataLoading, setDataLoading] = useState(true);

  useEffect(() => {
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
                  return (
                    <div key={index} className="flex justify-between items-center border-bottom--thin-2 py-5" >
                      <div className="flex justify-start items-center" >
                        <Avatar size={60} shape={'square'} />
                        <div className="flex justify-start flex-col ml-4" >
                          <Text type={'title'} level={6} weight={'bold'} extraClass={'m-0'}>{item.name}</Text>
                          <Text type={'title'} level={7} weight={'regular'} extraClass={'m-0 mt-1'}>Owner</Text>
                        </div>
                      </div>
                      <div>
                        <Button size={'large'} type="text">Leave Project</Button>
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
    projects: state.global.projects
  }
  );
};
export default connect(mapStateToProps, { fetchProjects })(ProjectDetails);
