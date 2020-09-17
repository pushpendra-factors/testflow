import React from 'react';
import { Switch } from 'antd';

function FiltersInfo({ setGrouping, grouping }) {
  return (
    <Switch checked={grouping} checkedChildren="grouped" unCheckedChildren="ungrouped" onChange={setGrouping} />
  );
}

export default FiltersInfo;
