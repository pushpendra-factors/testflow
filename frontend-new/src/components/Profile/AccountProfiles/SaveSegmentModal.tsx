import React, { useCallback, useEffect, useState } from 'react';
import cx from 'classnames';
import AppModal from 'Components/AppModal/AppModal';
import { Text } from 'Components/factorsComponents';
import { Input } from 'antd';
import styles from './index.module.scss';

interface Props {
  visible: boolean;
  handleSubmit: (name: string) => void;
  handleCancel: () => void;
  isLoading: boolean;
}

const SaveSegmentModal = ({
  visible,
  handleSubmit,
  handleCancel,
  isLoading
}: Props) => {
  const [newSegmentName, setNewSegmentName] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value } = e.target;
    setNewSegmentName(value);
  };

  const onSubmit = useCallback(() => {
    handleSubmit(newSegmentName);
  }, [handleSubmit, newSegmentName]);

  useEffect(() => {
    setNewSegmentName('');
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
      <div className='flex flex-col gap-y-5'>
        <Text
          type='title'
          level={4}
          color='character-primary'
          extraClass='mb-0'
          weight='bold'
        >
          New Segment
        </Text>
        <div className='flex flex-col gap-y-2'>
          <Text type='title' color='character-primary' extraClass='mb-0'>
            Segment name
          </Text>
          <Input
            onChange={handleInputChange}
            value={newSegmentName}
            className={cx('fa-input', styles.input)}
            size='large'
            placeholder='Eg- Paid search visitors'
          />
        </div>
      </div>
    </AppModal>
  );
};

export default SaveSegmentModal;
