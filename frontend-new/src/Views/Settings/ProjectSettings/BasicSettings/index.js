import React, { useState, useEffect } from 'react';
import ViewBasicSettings from './ViewBasicSettings';
import EditBasicSettings from './EditBasicSettings';
import { connect } from 'react-redux';
import { fetchProjectAgents, fetchAgentInfo } from 'Reducers/agentActions';
import { fetchProjects } from 'Reducers/global';
import { Row, Col } from 'antd';

function BasicSettings({
  fetchProjectAgents,
  fetchAgentInfo,
  fetchProjects,
  activeProject,
}) {
  const [editMode, setEditMode] = useState(false);

  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
      await fetchProjects();
      await fetchProjectAgents(activeProject.id);
    };
    getData();
  }, [activeProject]);
  return (
    <div className={'fa-container mt-32 mb-12 min-h-screen'}>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={18}>
          {editMode ? (
            <EditBasicSettings setEditMode={setEditMode} />
          ) : (
            <ViewBasicSettings setEditMode={setEditMode} />
          )}
        </Col>
      </Row>
    </div>
  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project,
});
export default connect(mapStateToProps, {
  fetchProjectAgents,
  fetchAgentInfo,
  fetchProjects,
})(BasicSettings);
