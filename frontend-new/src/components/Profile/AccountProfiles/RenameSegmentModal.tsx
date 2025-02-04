import React, { useCallback, useEffect, useMemo, useState } from 'react';
import cx from 'classnames';
import AppModal from 'Components/AppModal/AppModal';
import { Text } from 'Components/factorsComponents';
import { Input } from 'antd';
import styles from './index.module.scss';

interface Props {
  visible: boolean;
  isLoading: boolean;
  segmentName: string;
  handleSubmit: (name: string) => void;
  onCancel: () => void;
}

const RenameSegmentModal = ({
  visible,
  isLoading,
  segmentName,
  handleSubmit,
  onCancel
}: Props) => {
  const [segmentNewName, setSegmentNewName] = useState('');

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const { value } = e.target;
    setSegmentNewName(value);
  };

  const onSubmit = useCallback(() => {
    handleSubmit(segmentNewName);
  }, [handleSubmit, segmentNewName]);

  useEffect(() => {
    setSegmentNewName(segmentName);
  }, [segmentName]);

  return (
    <AppModal
      okText='Save'
      visible={visible}
      okButtonProps={{
        disabled: segmentNewName?.trim() === segmentName?.trim()
      }}
      onOk={onSubmit}
      onCancel={onCancel}
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
          Rename Segment
        </Text>
        <div className='flex flex-col gap-y-2'>
          <Text type='title' color='character-primary' extraClass='mb-0'>
            Enter new segment name
          </Text>
          <Input
            onChange={handleInputChange}
            value={segmentNewName}
            className={cx('fa-input', styles.input)}
            size='large'
            placeholder='Name'
          />
        </div>
      </div>
    </AppModal>
  );
};

export default RenameSegmentModal;
