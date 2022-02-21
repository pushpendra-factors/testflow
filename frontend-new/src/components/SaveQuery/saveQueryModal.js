import React, { useState, useCallback, useEffect } from 'react';
import PropTypes from 'prop-types';
import { Input } from 'antd';
import { Text } from 'factorsComponents';
import { EMPTY_STRING, EMPTY_OBJECT, isStringLengthValid } from 'Utils/global';
import styles from './index.module.scss';
import { ACTION_TYPES } from './saveQuery.constants';
import AppModal from '../AppModal';

const SaveQueryModal = ({
  visible,
  isLoading,
  modalTitle,
  onSubmit,
  toggleModalVisibility,
  activeAction,
  queryTitle,
}) => {
  const { TextArea } = Input;

  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');

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

  const handleDescriptionChange = (e) => {
    setDescription(e.target.value);
  };

  const resetModalState = useCallback(() => {
    setTitle('');
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
      onSuccess: () => {
        resetModalState();
      },
    });
  };

  const isSaveBtnDisabled = () => {
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
          type={'title'}
          level={3}
          weight={'bold'}
        >
          {modalTitle}
        </Text>
        <div className='flex flex-col gap-y-8'>
          <Input
            onChange={handleTitleChange}
            value={title}
            className={'fa-input'}
            size={'large'}
            placeholder='Name'
            className={styles.input}
          />
          <TextArea
            className={styles.input}
            onChange={handleDescriptionChange}
            value={description}
            placeholder='Description (Optional)'
          />
        </div>
      </div>
    </AppModal>
  );
};

export default SaveQueryModal;

SaveQueryModal.propTypes = {
  visible: PropTypes.bool,
  isLoading: PropTypes.bool,
  modalTitle: PropTypes.string,
  queryType: PropTypes.string,
  requestQuery: PropTypes.oneOfType([PropTypes.object, PropTypes.array]),
  onSubmit: PropTypes.func,
  toggleModalVisibility: PropTypes.func,
  activeAction: PropTypes.string,
  queryTitle: PropTypes.string,
};

SaveQueryModal.defaultProps = {
  visible: false,
  isLoading: false,
  modalTitle: EMPTY_STRING,
  queryType: EMPTY_STRING,
  requestQuery: EMPTY_OBJECT,
  onSubmit: _.noop,
  toggleModalVisibility: _.noop,
  activeAction: EMPTY_STRING,
  queryTitle: EMPTY_STRING,
};
