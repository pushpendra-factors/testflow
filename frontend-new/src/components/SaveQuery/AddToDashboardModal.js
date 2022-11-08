import React, { useState, useCallback, memo } from 'react';
import PropTypes from 'prop-types';
import { noop } from 'lodash';
import { Text } from 'factorsComponents';
import { EMPTY_ARRAY } from 'Utils/global';
import AppModal from '../AppModal';
import { DEFAULT_DASHBOARD_PRESENTATION } from './saveQuery.constants';
import AddToDashboardForm from './CommonAddToDashboardForm';

function AddToDashboardModal({
  visible,
  isLoading,
  onSubmit,
  toggleModalVisibility
}) {
  const [selectedDashboards, setSelectedDashboards] = useState([]);
  const [dashboardPresentation, setDashboardPresentation] = useState(
    DEFAULT_DASHBOARD_PRESENTATION
  );

  const resetModalState = useCallback(() => {
    setSelectedDashboards(EMPTY_ARRAY);
    setDashboardPresentation(DEFAULT_DASHBOARD_PRESENTATION);
    toggleModalVisibility();
  }, [toggleModalVisibility]);

  const handleCancel = useCallback(() => {
    if (!isLoading) {
      resetModalState();
    }
  }, [resetModalState, isLoading]);

  const handleSubmit = () => {
    onSubmit({
      selectedDashboards,
      dashboardPresentation,
      onSuccess: () => {
        resetModalState();
      }
    });
  };

  const isSaveBtnDisabled = () => !selectedDashboards.length;

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
            type='title'
            level={5}
            weight='bold'
          >
            Add to Dashboard
          </Text>
          <Text color='grey-2' extraClass='mb-0' type='paragraph'>
            This widget will appear on the following dashboards:
          </Text>
        </div>
        <AddToDashboardForm
          selectedDashboards={selectedDashboards}
          setSelectedDashboards={setSelectedDashboards}
          dashboardPresentation={dashboardPresentation}
          setDashboardPresentation={setDashboardPresentation}
        />
      </div>
    </AppModal>
  );
}

export default memo(AddToDashboardModal);

AddToDashboardModal.propTypes = {
  visible: PropTypes.bool,
  isLoading: PropTypes.bool,
  onSubmit: PropTypes.func,
  toggleModalVisibility: PropTypes.func
};

AddToDashboardModal.defaultProps = {
  visible: false,
  isLoading: false,
  onSubmit: noop,
  toggleModalVisibility: noop
};
