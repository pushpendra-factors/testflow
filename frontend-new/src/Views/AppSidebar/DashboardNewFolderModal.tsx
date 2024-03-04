import React, { useCallback, useEffect, useState } from 'react';
import cx from 'classnames';
import { Input } from 'antd';
import AppModal from 'Components/AppModal/AppModal';
import { Text } from 'Components/factorsComponents';
import styles from './index.module.scss';

interface Props {
  visible: boolean;
  handleSubmit: (name: string) => void;
  handleCancel: () => void;
  isLoading: boolean;
  renameFolder: boolean;
}

const DashboardNewFolderModal = ({
  visible,
  handleSubmit,
  handleCancel,
  isLoading,
  renameFolder = false
}: Props) => {
  const [newFolderName, setNewFolderName] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value } = e.target;
    setNewFolderName(value);
  };

  const onSubmit = useCallback(() => {
    handleSubmit(newFolderName);
  }, [handleSubmit, newFolderName]);

  useEffect(() => {
    setNewFolderName('');
  }, [visible]);

  return (
    <AppModal
      okText='Save'
      visible={visible}
      onOk={onSubmit}
      onCancel={handleCancel}
      isLoading={isLoading}
      width={542}
    >
      <div className='flex flex-col row-gap-5'>
        <Text
          type='title'
          level={4}
          color='character-primary'
          extraClass='mb-0'
          weight='bold'
        >
          {renameFolder ? 'Rename folder' : 'Create new folder'}
        </Text>
        <div className='flex flex-col row-gap-2'>
          <Text type='title' color='character-primary' extraClass='mb-0'>
            Folder name
          </Text>
          <Input
            onChange={handleInputChange}
            value={newFolderName}
            className={cx('fa-input', styles.input)}
            size='large'
            placeholder='Eg- Folder 1'
          />
        </div>
      </div>
    </AppModal>
  );
};

export default DashboardNewFolderModal;
