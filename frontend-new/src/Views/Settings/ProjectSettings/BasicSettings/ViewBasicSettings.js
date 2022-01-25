import React, { useState, useEffect } from 'react';
import {
  Row, Col, Button, Avatar, Skeleton, Tooltip
} from 'antd';
import { Text } from 'factorsComponents'; 
import { connect } from 'react-redux'; 
import { TimeZoneOffsetValues } from 'Utils/constants';


const getKeyByValue = (obj, value) =>  Object.keys(obj).find(key => obj[key]?.city === value);

const TimeZoneName = 
{
  "IST":'IST',
  "PT" :'PT (Pacific Time)',
  "CT" :'CT (Central Time)',
  "ET" :'ET (Eastern Time)',
  "GMT" :'GMT',
  "AEST" :'AEST (Australia Eastern Standard Time)', 
}


function ViewBasicSettings({
  activeProject,
  setEditMode,
  agents,
  currentAgent
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [enableEdit, setEnableEdit] = useState(false);

  useEffect(() => {
    setEnableEdit(false);
    agents && currentAgent && agents.map((agent) => {
      if (agent.uuid === currentAgent.uuid) {
        if (agent.role === 1) {
          setEnableEdit(true);
        }
      }
    });
    setDataLoading(false);
  }, [activeProject, agents, currentAgent]);

  return (
    <>
      <div className={'mb-10 pl-4'}>
        <Row>
          <Col span={12}>
            <Text type={'title'} level={3} weight={'bold'} extraClass={'m-0'}>Basic Details</Text>
          </Col>
          <Col span={12}>
            <div className={'flex justify-end'}>
              {
                <Tooltip placement="top" trigger={'hover'} title={enableEdit ? 'Only Admin can edit' : null}>
                  <Button size={'large'} disabled={dataLoading || enableEdit} onClick={() => setEditMode(true)}>Edit Details</Button>
                </Tooltip>
              }
            </div>
          </Col>
        </Row>
        <Row className={'mt-2'}>
          <Col>
            { dataLoading ? <Skeleton.Avatar active={true} size={104} shape={'square'} />
              : activeProject?.profile_picture ? <img src={activeProject?.profile_picture} alt="avatar" style={{width:'105px'}} />:<Avatar size={104} shape={'square'} style={{ color: '#fff', backgroundColor: '#52BE95', fontSize: '42px', textTransform: 'uppercase', fontWeight:'400' }}>{`${activeProject?.name?.charAt(0)}`}</Avatar>

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
              : <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>{(activeProject?.time_zone) ? TimeZoneName[getKeyByValue(TimeZoneOffsetValues,activeProject.time_zone)] : '---'}</Text>
            }
          </Col>
        </Row>
      </div>

    </>

  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  agents: state.agent.agents,
  projects: state.global.projects,
  currentAgent: state.agent.agent_details
});

export default connect(mapStateToProps)(ViewBasicSettings);
