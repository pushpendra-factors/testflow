import React from 'react';
import AppModal from 'Components/AppModal/AppModal';
import { Text } from 'Components/factorsComponents';

interface Props {
  segmentName: string;
  visible: boolean;
  isLoading: boolean;
  onOk: () => void;
  onCancel: () => void;
}

const DeleteSegmentModal = ({
  visible,
  isLoading,
  segmentName,
  onOk,
  onCancel
}: Props) => {
  return (
    <AppModal
      okText='Confirm'
      visible={visible}
      onOk={onOk}
      onCancel={onCancel}
      isLoading={isLoading}
      width={504}
    >
      <Text
        type='title'
        color='character-primary'
        weight='bold'
        extraClass='mb-0'
        level={5}
      >
        {`Are you sure you want to delete ${segmentName}?`}
      </Text>
    </AppModal>
  );
};

export default DeleteSegmentModal;
