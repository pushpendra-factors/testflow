import React, { useState, useEffect } from 'react';
import ViewBasicSettings from './ViewBasicSettings';
import EditBasicSettings from './EditBasicSettings';
import { connect } from 'react-redux';
import { fetchProjectAgents, fetchAgentInfo, fetchProjects } from 'Reducers/agentActions';

function BasicSettings({
  fetchProjectAgents, fetchAgentInfo, fetchProjects, activeProject
}) {
  const [editMode, setEditMode] = useState(false);

  useEffect(() => {
    const getData = async () => {
      await fetchAgentInfo();
      await fetchProjects();
      await fetchProjectAgents(activeProject.id);
    };
    getData();
  });
  return (
    <>
    {editMode ? <EditBasicSettings setEditMode={setEditMode} /> : <ViewBasicSettings setEditMode={setEditMode} /> }
    </>

  );
}
const mapStateToProps = (state) => ({
  activeProject: state.global.active_project
});
export default connect(mapStateToProps, { fetchProjectAgents, fetchAgentInfo, fetchProjects })(BasicSettings);
