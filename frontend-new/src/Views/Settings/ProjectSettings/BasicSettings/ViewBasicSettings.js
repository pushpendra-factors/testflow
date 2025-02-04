import React, { useState, useEffect } from 'react';
import { Row, Col, Button, Avatar, Skeleton, Tooltip } from 'antd';
import { Text } from 'factorsComponents';
import { connect } from 'react-redux';
import { getTimeZoneNameFromCity } from 'Utils/constants';
import { Currency } from 'Utils/currency';
import _ from 'lodash';
import CommonSettingsHeader from 'Components/GenericComponents/CommonSettingsHeader';

function ViewBasicSettings({
  activeProject,
  setEditMode,
  agents,
  currentAgent,
  currentProjectSettings
}) {
  const [dataLoading, setDataLoading] = useState(true);
  const [enableEdit, setEnableEdit] = useState(false);

  useEffect(() => {
    setEnableEdit(false);
    agents &&
      currentAgent &&
      agents.map((agent) => {
        if (agent.uuid === currentAgent.uuid) {
          if (agent.role === 1) {
            setEnableEdit(true);
          }
        }
      });
    setDataLoading(false);
  }, [activeProject, agents, currentAgent, currentProjectSettings]);

  return (
    <div className='mb-10'>
      <CommonSettingsHeader
        title='General'
        description='Personalize your project by updating its name, timezone, logo, and other key details.'
        actionsNode={
          <div className='flex justify-end'>
            <Tooltip
              placement='top'
              trigger='hover'
              title={enableEdit ? 'Only Admin can edit' : null}
            >
              <Button
                size='large'
                disabled={dataLoading || enableEdit}
                onClick={() => setEditMode(true)}
              >
                Edit Details
              </Button>
            </Tooltip>
          </div>
        }
      />

      <Row className='mt-2'>
        <Col>
          {dataLoading ? (
            <Skeleton.Avatar active size={104} shape='square' />
          ) : activeProject?.profile_picture ? (
            <img
              src={activeProject?.profile_picture}
              alt='avatar'
              style={{ width: '105px', borderRadius: '4px' }}
            />
          ) : (
            <Avatar
              size={104}
              shape='square'
              style={{
                color: '#fff',
                backgroundColor: '#52BE95',
                fontSize: '42px',
                textTransform: 'uppercase',
                fontWeight: '400',
                borderRadius: '4px'
              }}
            >{`${activeProject?.name?.charAt(0)}`}</Avatar>
          )}
          <Text type='paragraph' mini extraClass='m-0 mt-1' color='grey'>
            A logo helps personalise your Project
          </Text>
        </Col>
      </Row>
      <Row className='mt-6'>
        <Col span={12}>
          <Text type='title' level={7} extraClass='m-0'>
            Project Name
          </Text>
          {dataLoading ? (
            <Skeleton.Input style={{ width: 200 }} active size='small' />
          ) : (
            <Text type='title' level={6} extraClass='m-0' weight='bold'>
              {activeProject.name ? activeProject.name : '---'}
            </Text>
          )}
        </Col>
        <Col span={12}>
          <Text type='title' level={7} extraClass='m-0'>
            Project URL
          </Text>
          {dataLoading ? (
            <Skeleton.Input style={{ width: 200 }} active size='small' />
          ) : (
            <Text type='title' level={6} extraClass='m-0' weight='bold'>
              {activeProject.project_uri ? activeProject.project_uri : '---'}
            </Text>
          )}
        </Col>
      </Row>
      {/* <Row className={'mt-6'}>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Date Format
            </Text>
            {dataLoading ? (
              <Skeleton.Input
                style={{ width: 200 }}
                active={true}
                size={'small'}
              />
            ) : (
              <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>
                {activeProject.date_format ? activeProject.date_format : '---'}
              </Text>
            )}
          </Col>
          <Col span={12}>
            <Text type={'title'} level={7} extraClass={'m-0'}>
              Time Format
            </Text>
            {dataLoading ? (
              <Skeleton.Input
                style={{ width: 200 }}
                active={true}
                size={'small'}
              />
            ) : (
              <Text type={'title'} level={6} extraClass={'m-0'} weight={'bold'}>
                {activeProject.time_format ? activeProject.time_format : '---'}
              </Text>
            )}
          </Col>
        </Row> */}
      <Row className='mt-6 mb-10'>
        <Col span={12}>
          <Text type='title' level={7} extraClass='m-0'>
            Time Zone
          </Text>
          {dataLoading ? (
            <Skeleton.Input style={{ width: 200 }} active size='small' />
          ) : (
            <Text type='title' level={6} extraClass='m-0' weight='bold'>
              {!_.isEmpty(activeProject?.time_zone)
                ? `${getTimeZoneNameFromCity(activeProject?.time_zone)?.text}`
                : '---'}
            </Text>
          )}
        </Col>
        <Col span={12}>
          <Text type='title' level={7} extraClass='m-0'>
            Currency
          </Text>
          {dataLoading ? (
            <Skeleton.Input style={{ width: 200 }} active size='small' />
          ) : (
            <Text type='title' level={6} extraClass='m-0' weight='bold'>
              {currentProjectSettings?.currency
                ? `${Currency[currentProjectSettings?.currency]} (${
                    currentProjectSettings?.currency
                  })`
                : '---'}
            </Text>
          )}
        </Col>
      </Row>
    </div>
  );
}

const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
  agents: state.agent.agents,
  projects: state.global.projects,
  currentAgent: state.agent.agent_details,
  currentProjectSettings: state.global.currentProjectSettings
});

export default connect(mapStateToProps)(ViewBasicSettings);
