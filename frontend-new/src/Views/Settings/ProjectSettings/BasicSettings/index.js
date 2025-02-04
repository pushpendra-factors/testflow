import React, { useState, useEffect } from 'react';
import { connect } from 'react-redux';
import { fetchProjectAgents, fetchAgentInfo } from 'Reducers/agentActions';
import { fetchProjectsList } from 'Reducers/global';
import { Row, Col } from 'antd';
import EditBasicSettings from './EditBasicSettings';
import ViewBasicSettings from './ViewBasicSettings';

function BasicSettings({
  fetchProjectAgents,
  fetchAgentInfo,
  fetchProjectsList,
  activeProject
}) {
  const [editMode, setEditMode] = useState(false);

  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
      await fetchProjectsList();
      await fetchProjectAgents(activeProject.id);
    };
    getData();
  }, [activeProject]);
  return (
    <div className='fa-container'>
      <Row gutter={[24, 24]} justify='center'>
        <Col span={24}>
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
  activeProject: state.global.active_project
});
export default connect(mapStateToProps, {
  fetchProjectAgents,
  fetchAgentInfo,
  fetchProjectsList
})(BasicSettings);
