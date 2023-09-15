import React, { useCallback } from 'react';
import AppModal from 'Components/AppModal/AppModal';
import { Text } from 'Components/factorsComponents';
import { useDispatch, useSelector } from 'react-redux';
import {
  setExitConfirmationModalAction,
  setFiltersDirtyAction
} from 'Reducers/accountProfilesView/actions';

const ExitConfirmationModal = () => {
  const { showExitConfirmationModal, routingCallback } = useSelector(
    (state) => state.accountProfilesView
  );

  const dispatch = useDispatch();

  const handleOk = useCallback(() => {
    if (routingCallback != null) {
      dispatch(setFiltersDirtyAction(false));
      setTimeout(() => {
        routingCallback();
      }, 250);
    }
    dispatch(setExitConfirmationModalAction(false));
  }, [dispatch, routingCallback]);

  const handleCancel = useCallback(() => {
    dispatch(setExitConfirmationModalAction(false));
  }, [dispatch]);

  return (
    <AppModal
      okText='Confirm'
      visible={showExitConfirmationModal}
      onOk={handleOk}
      onCancel={handleCancel}
      width={504}
    >
      <div className='flex flex-col row-gap-2'>
        <Text
          type='title'
          level={4}
          color='character-primary'
          extraClass='mb-0'
          weight='bold'
        >
          Are you sure you want to leave?
        </Text>
        <div className='flex flex-col row-gap-2'>
          <Text type='title' color='character-secondary' extraClass='mb-0'>
            You have unsaved filter condition, and will be lost unless you save
            it
          </Text>
        </div>
      </div>
    </AppModal>
  );
};

export default ExitConfirmationModal;
