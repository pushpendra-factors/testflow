import React, { useCallback, useState } from 'react';
import cx from 'classnames';
import { noop } from 'lodash';
import AppModal from 'Components/AppModal/AppModal';
import { SVG, Text } from 'Components/factorsComponents';
import styles from './index.module.scss';
import { Input, notification } from 'antd';
import ControlledComponent from 'Components/ControlledComponent/ControlledComponent';

const UpdateSegmentModal = ({
  visible = false,
  onCreate = noop,
  onUpdate = noop,
  onCancel = noop,
  isLoading = false
}) => {
  const [saveMode, setSaveMode] = useState(null); // create/update
  const [newSegmentName, setNewSegmentName] = useState('');

  const handleNameChange = (e) => {
    setNewSegmentName(e.target.value);
  };

  const handleSubmit = useCallback(() => {
    if (saveMode === null) {
      notification.error({
        message: 'Error',
        description: 'Please choose an option.',
        duration: 2
      });
      return;
    }
    if (saveMode === 'create' && Boolean(newSegmentName) === false) {
      notification.error({
        message: 'Error',
        description: 'Please enter new segment name.',
        duration: 2
      });
      return;
    }
    if (saveMode === 'create') {
      onCreate(newSegmentName);
      return;
    }
    onUpdate();
  }, [newSegmentName, onCreate, onUpdate, saveMode]);

  return (
    <AppModal
      okText={saveMode === 'create' ? 'Save new segment' : 'Save segment'}
      visible={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
      isLoading={isLoading}
      width={542}
      cancelText='Discard Changes'
    >
      <div className='flex flex-col row-gap-5'>
        <Text
          type='title'
          level={4}
          color='character-primary'
          extraClass='mb-0'
          weight='bold'
        >
          Save Segment
        </Text>
        <div className='flex flex-col row-gap-4'>
          <div
            className={cx(
              'p-4 flex justify-between items-center rounded-lg',
              styles['update-option-container']
            )}
            role='button'
            onClick={() => setSaveMode('update')}
          >
            <div className='flex col-gap-4 items-center'>
              <SVG name='save' size={24} color='#595959' />
              <div className='flex flex-col'>
                <Text
                  type='title'
                  extraClass='mb-0'
                  weight='medium'
                  color='character-primary'
                  level={6}
                >
                  Save changes
                </Text>
                <Text
                  type='title'
                  extraClass='mb-0'
                  weight='medium'
                  color='character-secondary'
                >
                  Overwrite changes on the current segment
                </Text>
              </div>
            </div>
            <ControlledComponent controller={saveMode === 'update'}>
              <SVG size={24} name='checkCircle' color='#1890FF' />
            </ControlledComponent>
          </div>
          <div
            className={cx('rounded-lg', styles['update-option-container'])}
            role='button'
            onClick={() => setSaveMode('create')}
          >
            <div
              className={cx('flex justify-between p-4 items-center', {
                'border-b border-gray-300': saveMode === 'create'
              })}
            >
              <div className='flex col-gap-4 items-center'>
                <SVG name='pieChart' size={24} color='#595959' />
                <div className='flex flex-col'>
                  <Text
                    type='title'
                    extraClass='mb-0'
                    weight='medium'
                    color='character-primary'
                    level={6}
                  >
                    Save as new segment
                  </Text>
                  <Text
                    type='title'
                    extraClass='mb-0'
                    weight='medium'
                    color='character-secondary'
                  >
                    Create a new segment with these changes.
                  </Text>
                </div>
              </div>
              <ControlledComponent controller={saveMode === 'create'}>
                <SVG size={24} name='checkCircle' color='#1890FF' />
              </ControlledComponent>
            </div>
            <ControlledComponent controller={saveMode === 'create'}>
              <div
                className={cx(
                  'pr-4 py-3 flex flex-col row-gap-2',
                  styles['new-segment-input-container']
                )}
              >
                <Text
                  type='title'
                  color='character-secondary'
                  extraClass='mb-0'
                >
                  Enter new segment name
                </Text>
                <Input
                  value={newSegmentName}
                  onChange={handleNameChange}
                  className={styles['create-mode-input-box']}
                  placeholder='Eg: Paid Search Visitors'
                />
              </div>
            </ControlledComponent>
          </div>
        </div>
      </div>
    </AppModal>
  );
};

export default UpdateSegmentModal;
