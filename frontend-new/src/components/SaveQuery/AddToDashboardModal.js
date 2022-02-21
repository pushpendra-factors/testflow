import React, { useState, useCallback } from 'react';
import PropTypes from 'prop-types';
import { useSelector } from 'react-redux';
import { Select, Radio } from 'antd';
import { Text } from 'factorsComponents';
import {
  apiChartAnnotations,
  CHART_TYPE_TABLE,
  DASHBOARD_TYPES,
} from 'Utils/constants';
import { EMPTY_STRING, EMPTY_OBJECT, EMPTY_ARRAY } from 'Utils/global';
import { getSaveChartOptions } from '../../Views/CoreQuery/utils';
import styles from './index.module.scss';
import AppModal from '../AppModal';

const AddToDashboardModal = ({
  visible,
  isLoading,
  queryType,
  requestQuery,
  onSubmit,
  toggleModalVisibility,
}) => {
  const { dashboards } = useSelector((state) => state.dashboard);

  const [selectedDashboards, setSelectedDashboards] = useState([]);
  const [dashboardPresentation, setDashboardPresentation] = useState(
    apiChartAnnotations[CHART_TYPE_TABLE]
  );

  const resetModalState = useCallback(() => {
    setSelectedDashboards(EMPTY_ARRAY);
    setDashboardPresentation(apiChartAnnotations[CHART_TYPE_TABLE]);
    toggleModalVisibility();
  }, [toggleModalVisibility]);

  const handleCancel = useCallback(() => {
    if (!isLoading) {
      resetModalState();
    }
  }, [resetModalState, isLoading]);

  const handlePresentationChange = (e) => {
    setDashboardPresentation(e.target.value);
  };

  const handleSelectChange = useCallback(
    (value) => {
      const resp = value.map((v) => {
        return dashboards.data.find((d) => d.name === v).id;
      });
      setSelectedDashboards(resp);
    },
    [dashboards.data]
  );

  const getSelectedDashboards = useCallback(() => {
    return selectedDashboards.map((s) => {
      return dashboards.data.find((d) => d.id === s).name;
    });
  }, [dashboards.data, selectedDashboards]);

  const handleSubmit = () => {
    onSubmit({
      selectedDashboards,
      dashboardPresentation,
      onSuccess: () => {
        resetModalState();
      },
    });
  };

  const isSaveBtnDisabled = () => {
    return !selectedDashboards.length;
  };

  const chartOptions = (
    <Radio.Group
      value={dashboardPresentation}
      onChange={handlePresentationChange}
      className={styles.radioGroup}
    >
      {getSaveChartOptions(queryType, requestQuery)}
    </Radio.Group>
  );

  const dashboardList = (
    <Select
      mode='multiple'
      style={{ width: '100%' }}
      placeholder={'Please Select'}
      onChange={handleSelectChange}
      className={styles.multiSelectStyles}
      value={getSelectedDashboards()}
    >
      {dashboards.data
        .filter((d) => d.class === DASHBOARD_TYPES.USER_CREATED)
        .map((d) => {
          return (
            <Select.Option value={d.name} key={d.id}>
              {d.name}
            </Select.Option>
          );
        })}
    </Select>
  );

  return (
    <AppModal
      okText='Save'
      visible={visible}
      onOk={handleSubmit}
      onCancel={handleCancel}
      isLoading={isLoading}
      okButtonProps={{ disabled: isSaveBtnDisabled() }}
    >
      <div className='flex flex-col gap-y-5'>
        <div className='flex flex-col gap-y-2'>
          <Text
            color='grey-6'
            extraClass='mb-0'
            type={'title'}
            level={5}
            weight={'bold'}
          >
            Add to Dashboard
          </Text>
          <Text color='grey-2' extraClass='mb-0' type={'paragraph'}>
            This widget will appear on the following dashboards:
          </Text>
        </div>
        {dashboardList}
        {chartOptions}
      </div>
    </AppModal>
  );
};

export default AddToDashboardModal;

AddToDashboardModal.propTypes = {
  visible: PropTypes.bool,
  isLoading: PropTypes.bool,
  modalTitle: PropTypes.string,
  onSubmit: PropTypes.func,
  toggleModalVisibility: PropTypes.func,
  queryType: PropTypes.string,
  requestQuery: PropTypes.oneOfType([PropTypes.object, PropTypes.array]),
};

AddToDashboardModal.defaultProps = {
  visible: false,
  isLoading: false,
  modalTitle: EMPTY_STRING,
  onSubmit: _.noop,
  toggleModalVisibility: _.noop,
  queryType: EMPTY_STRING,
  requestQuery: EMPTY_OBJECT,
};
