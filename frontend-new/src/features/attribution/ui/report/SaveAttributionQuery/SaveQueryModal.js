import React, { useState, useCallback, useEffect } from 'react';
import AppModal from 'Components/AppModal';
import { Text, SVG } from 'factorsComponents';
import { Button, Input, Menu, Dropdown } from 'antd';
import styles from './index.module.scss';
import {
  DEFAULT_DASHBOARD_PRESENTATION,
  DASHBOARD_PRESENTATION_KEYS
} from './constants';
import useAutoFocus from 'hooks/useAutoFocus';
import { ACTION_TYPES } from './constants';

const SaveQueryModal = ({
  toggleSaveModalVisibility,
  visibility,
  isLoading,
  onSubmit,
  activeAction,
  savedQueryPresentation,
  queryTitle
}) => {
  const { TextArea } = Input;
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [dashboardPresentation, setDashboardPresentation] = useState(
    DEFAULT_DASHBOARD_PRESENTATION
  );
  const inputRef = useAutoFocus(visibility);

  const handleTitleChange = (e) => {
    setTitle(e.target.value);
  };

  const handleDescriptionChange = (e) => {
    setDescription(e.target.value);
  };

  const handleCancel = useCallback(() => {
    if (!isLoading) {
      setTitle('');
      setDashboardPresentation(DEFAULT_DASHBOARD_PRESENTATION);
      toggleSaveModalVisibility();
    }
  }, [isLoading]);

  const handleSubmit = () => {
    onSubmit({
      title,
      description,
      selectedDashboards: [],
      dashboardPresentation: dashboardPresentation.value,
      onSuccess: () => {
        handleCancel();
      }
    });
  };

  const footer = (title) => {
    return (
      <div className='flex justify-between align-center px-3'>
        <div className='flex align-center justify-start gap-1'>
          <div className='self-center'>
            <Text
              color='black'
              extraClass='m-0'
              type='title'
              level={7}
              weight='bold'
            >
              Save as
            </Text>
          </div>

          <Dropdown
            overlay={dashboardPresentationMenu}
            trigger={['click']}
            overlayStyle={{ zIndex: 100001 }}
            placement='topRight'
          >
            <Button type='text' className='flex items-center'>
              {dashboardPresentation.label}
              <SVG name='caretDown' size={16} extraClass='ml-1' />
            </Button>
          </Dropdown>
        </div>
        <div>
          <Button onClick={handleCancel}>Cancel</Button>
          <Button type='primary' disabled={!title} onClick={handleSubmit}>
            Save
          </Button>
        </div>
      </div>
    );
  };

  const dashboardPresentationMenu = (
    <Menu>
      {DASHBOARD_PRESENTATION_KEYS.map((option, index) => (
        <Menu.Item
          key={option.value}
          onClick={() => {
            setDashboardPresentation(DASHBOARD_PRESENTATION_KEYS[index]);
          }}
        >
          <div className='flex items-center'>
            <span className='mr-3'>
              {DASHBOARD_PRESENTATION_KEYS[index].label}
            </span>
          </div>
        </Menu.Item>
      ))}
    </Menu>
  );

  useEffect(() => {
    if (visibility) {
      if (activeAction === ACTION_TYPES.EDIT) {
        if (queryTitle) setTitle(queryTitle);
        if (savedQueryPresentation) {
          const presentation = DASHBOARD_PRESENTATION_KEYS.find(
            (option) => option.value === savedQueryPresentation
          );
          if (presentation) setDashboardPresentation(presentation);
        }
      }
    }
  }, [activeAction, queryTitle, visibility, savedQueryPresentation]);

  return (
    <AppModal
      visible={visibility}
      closable
      footer={footer(title)}
      onCancel={handleCancel}
      isLoading={isLoading}
    >
      <div className='flex flex-col gap-y-10'>
        <Text
          color='black'
          extraClass='m-0'
          type='title'
          level={5}
          weight='bold'
        >
          Save this Report
        </Text>
        <div className='flex flex-col gap-y-8'>
          <Input
            onChange={handleTitleChange}
            value={title}
            className={`fa-input ${styles.input}`}
            size='large'
            placeholder='Name'
            ref={inputRef}
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
