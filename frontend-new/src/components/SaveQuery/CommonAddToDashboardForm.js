import React, { useCallback } from 'react';
import PropTypes from 'prop-types';
import { useSelector } from 'react-redux';
import { Select, Radio } from 'antd';
import { map, noop } from 'lodash';
import { DASHBOARD_TYPES } from 'Utils/constants';
import { EMPTY_STRING, EMPTY_ARRAY } from 'Utils/global';

import styles from './index.module.scss';
import {
  DASHBOARD_PRESENTATION_KEYS,
  DASHBOARD_PRESENTATION_LABELS
} from './saveQuery.constants';

function AddToDashboardForm({
  selectedDashboards,
  setSelectedDashboards,
  dashboardPresentation,
  setDashboardPresentation
}) {
  const { dashboards } = useSelector((state) => state.dashboard);

  const handlePresentationChange = (e) => {
    setDashboardPresentation(e.target.value);
  };

  const handleSelectChange = useCallback(
    (value) => {
      const resp = value.map(
        (v) => dashboards.data.find((d) => d.name === v).id
      );
      setSelectedDashboards(resp);
    },
    [dashboards.data]
  );

  const getSelectedDashboards = useCallback(
    () =>
      selectedDashboards.map(
        (s) => dashboards.data.find((d) => d.id === s).name
      ),
    [dashboards.data, selectedDashboards]
  );

  const dashboardPresentationOptions = (
    <Radio.Group
      value={dashboardPresentation}
      onChange={handlePresentationChange}
      className={styles.radioGroup}
    >
      {map(DASHBOARD_PRESENTATION_KEYS, (pKey) => (
        <Radio key={pKey} value={pKey}>
          {DASHBOARD_PRESENTATION_LABELS[pKey]}
        </Radio>
      ))}
    </Radio.Group>
  );

  const dashboardList = (
    <Select
      mode='multiple'
      style={{ width: '100%' }}
      placeholder='Please Select'
      onChange={handleSelectChange}
      className={styles.multiSelectStyles}
      value={getSelectedDashboards()}
    >
      {dashboards.data
        .filter((d) => d.class === DASHBOARD_TYPES.USER_CREATED)
        .map((d) => (
          <Select.Option value={d.name} key={d.id}>
            {d.name}
          </Select.Option>
        ))}
    </Select>
  );

  return (
    <>
      {dashboardList}
      {dashboardPresentationOptions}
    </>
  );
}

export default AddToDashboardForm;

AddToDashboardForm.propTypes = {
  selectedDashboards: PropTypes.arrayOf(PropTypes.string),
  setSelectedDashboards: PropTypes.func,
  dashboardPresentation: PropTypes.string,
  setDashboardPresentation: PropTypes.func
};

AddToDashboardForm.defaultProps = {
  selectedDashboards: EMPTY_ARRAY,
  setSelectedDashboards: noop,
  dashboardPresentation: EMPTY_STRING,
  setDashboardPresentation: noop
};
