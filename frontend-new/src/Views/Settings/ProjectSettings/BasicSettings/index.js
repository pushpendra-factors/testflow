import React, { useState } from 'react';
import ViewBasicSettings from './ViewBasicSettings';
import EditBasicSettings from './EditBasicSettings';

function BasicSettings() {
  const [editMode, setEditMode] = useState(false);
  return (
    <>
    {editMode ? <EditBasicSettings setEditMode={setEditMode} /> : <ViewBasicSettings setEditMode={setEditMode} /> }
    </>

  );
}

export default BasicSettings;
