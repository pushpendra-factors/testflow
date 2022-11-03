import React, { useState, useCallback, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Input, Checkbox } from 'antd';
import { noop } from 'lodash';
import { Text } from 'factorsComponents';
import { EMPTY_STRING, isStringLengthValid, EMPTY_ARRAY } from 'Utils/global';
import styles from './index.module.scss';
import AppModal from '../AppModal';
import {
  DEFAULT_DASHBOARD_PRESENTATION,
  ACTION_TYPES
} from './saveQuery.constants';
import AddToDashboardForm from './CommonAddToDashboardForm';

function SaveQueryModal({
  visible,
  isLoading,
  modalTitle,
  onSubmit,
  toggleModalVisibility,
  activeAction,
  queryTitle
}) {
  const { TextArea } = Input;

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [showAddToDashboard, setShowAddToDashboard] = useState(false);
  const [selectedDashboards, setSelectedDashboards] = useState([]);
  const [dashboardPresentation, setDashboardPresentation] = useState(
    DEFAULT_DASHBOARD_PRESENTATION
  );

  useEffect(() => {
    if (visible && queryTitle) {
      if (activeAction === ACTION_TYPES.EDIT) {
        setTitle(queryTitle);
      } else {
        setTitle(`${queryTitle} - Copy`);
      }
    }
  }, [activeAction, queryTitle, visible]);

  const handleTitleChange = (e) => {
    setTitle(e.target.value);
  };

  const handleAddToDashboardChange = (e) => {
    setShowAddToDashboard(e.target.checked);
  };

  const handleDescriptionChange = (e) => {
    setDescription(e.target.value);
  };

  const resetModalState = useCallback(() => {
    setTitle('');
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
      title,
      addToDashboard: showAddToDashboard,
      selectedDashboards,
      dashboardPresentation,
      onSuccess: () => {
        resetModalState();
      }
    });
  };

  const isSaveBtnDisabled = () => {
    if (showAddToDashboard)
      return !isStringLengthValid(title) || !selectedDashboards.length;
    return !isStringLengthValid(title);
  };

  return (
    <AppModal
      okText='Save'
      visible={visible}
      onOk={handleSubmit}
      onCancel={handleCancel}
      isLoading={isLoading}
      okButtonProps={{ disabled: isSaveBtnDisabled() }}
    >
      <div className='flex flex-col gap-y-10'>
        <Text
          color='black'
          extraClass='m-0'
          type='title'
          level={3}
          weight='bold'
        >
          {modalTitle}
        </Text>
        <div className='flex flex-col gap-y-8'>
          <Input
            onChange={handleTitleChange}
            value={title}
            className={`fa-input ${styles.input}`}
            size='large'
            placeholder='Name'
          />
          <TextArea
            className={styles.input}
            onChange={handleDescriptionChange}
            value={description}
            placeholder='Description (Optional)'
          />
        </div>
        <div>
          <Checkbox onChange={handleAddToDashboardChange}>
            Add to Dashboard
          </Checkbox>
        </div>
        {showAddToDashboard && (
          <AddToDashboardForm
            selectedDashboards={selectedDashboards}
            setSelectedDashboards={setSelectedDashboards}
            dashboardPresentation={dashboardPresentation}
            setDashboardPresentation={setDashboardPresentation}
          />
        )}
      </div>
    </AppModal>
  );
}

export default SaveQueryModal;

SaveQueryModal.propTypes = {
  visible: PropTypes.bool,
  isLoading: PropTypes.bool,
  modalTitle: PropTypes.string,
  onSubmit: PropTypes.func,
  toggleModalVisibility: PropTypes.func,
  activeAction: PropTypes.string,
  queryTitle: PropTypes.string
};

SaveQueryModal.defaultProps = {
  visible: false,
  isLoading: false,
  modalTitle: EMPTY_STRING,
  onSubmit: noop,
  toggleModalVisibility: noop,
  activeAction: EMPTY_STRING,
  queryTitle: EMPTY_STRING
};
