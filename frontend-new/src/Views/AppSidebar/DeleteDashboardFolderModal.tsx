import React from 'react';
import AppModal from 'Components/AppModal/AppModal';
import { Text } from 'Components/factorsComponents';

interface Props {
  folderName: string;
  visible: boolean;
  isLoading: boolean;
  onOk: () => void;
  onCancel: () => void;
}

const DeleteDashboardFolderModal = ({
  visible,
  isLoading,
  folderName,
  onOk,
  onCancel
}: Props) => (
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
      {`Are you sure you want to delete "${folderName}"?`}
    </Text>
  </AppModal>
);

export default DeleteDashboardFolderModal;
