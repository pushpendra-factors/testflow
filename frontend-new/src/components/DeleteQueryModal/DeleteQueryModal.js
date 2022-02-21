import React from 'react';
import PropTypes from 'prop-types';
import { Button } from 'antd';
import AppModal from '../AppModal';
import { SVG, Text } from '../factorsComponents';
import { BUTTON_TYPES } from '../../utils/buttons.constants';

const DeleteQueryModal = ({ visible, onDelete, toggleModal, isLoading }) => {
  return (
    <AppModal
      visible={visible}
      closable={true}
      footer={null}
      onCancel={toggleModal}
      width={500}
    >
      <div className='flex gap-x-2'>
        <div className='mt-2'>
          <SVG color='#ea6262' name='infoCircle' />
        </div>

        <div className='flex flex-col gap-y-4'>
          <Text
            color='grey-8'
            extraClass='mb-0 text-2xl'
            weight='bold'
            type='title'
          >
            Delete Report?
          </Text>
          <div className='flex flex-col'>
            <Text
              extraClass='text-lg'
              color='grey-2'
              weight='bold'
              type='paragraph'
            >
              Are you sure you want to delete the report?
            </Text>
            <Text
              extraClass='text-lg'
              color='grey-2'
              weight='bold'
              type='paragraph'
            >
              You can't undo this action.
            </Text>
          </div>
          <div className='flex justify-end gap-x-2'>
            <Button type={BUTTON_TYPES.SECONDARY} onClick={toggleModal}>
              Cancel
            </Button>
            <Button
              className='flex items-center'
              onClick={onDelete}
              type={BUTTON_TYPES.PRIMARY}
              danger
              loading={isLoading}
              icon={<SVG name={'delete'} size={20} color={'#fff'} />}
            >
              {'Delete'}
            </Button>
          </div>
        </div>
      </div>
    </AppModal>
  );
};

export default DeleteQueryModal;

DeleteQueryModal.propTypes = {
  visible: PropTypes.bool,
  onDelete: PropTypes.func,
  toggleModal: PropTypes.func,
  isLoading: PropTypes.bool
};

DeleteQueryModal.defaultProps = {
  savedQueryId: false,
  onDelete: _.noop,
  toggleModal: _.noop,
  isLoading: false
};
