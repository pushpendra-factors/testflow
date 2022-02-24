import React, { useCallback } from 'react';
import PropTypes from 'prop-types';
import { SVG } from 'factorsComponents';
import { Button, Tooltip, Dropdown, Menu } from 'antd';
import { BUTTON_TYPES } from '../../utils/buttons.constants';
import ControlledComponent from '../ControlledComponent';
import SavedQueryPopoverContent from './savedQueryPopoverContent';
import styles from './index.module.scss';
import DeleteQueryModal from '../DeleteQueryModal';
import { useDispatch, useSelector } from 'react-redux';
import { Wait } from '../../utils/dataFormatter';

const QueryActions = ({
  savedQueryId,
  handleSaveClick,
  handleEditClick,
  handleDeleteReport,
  toggleAddToDashboardModal,
}) => {
  const dispatch = useDispatch();
  const { active_project } = useSelector((state) => state.global);
  const [showSavedQueryPopover, toggleSavedQueryPopover] = useToggle(false);
  const [showDeleteModal, toggleDeleteModal] = useToggle(false);

  const handleDelete = useCallback(async () => {
    toggleDeleteModal();
    // for modal animation
    await Wait(500);
    handleDeleteReport();
  }, [dispatch, active_project, savedQueryId]);

  const handlePopoverOkClick = useCallback(() => {
    toggleSavedQueryPopover();
    handleSaveClick();
  }, [handleSaveClick]);

  const handleDeleteClick = useCallback(() => {
    toggleDeleteModal();
  }, []);

  const getActionsMenu = () => {
    return (
      <Menu className={styles['more-actions-menu']}>
        <Menu.Item key='0'>
          <a onClick={handleEditClick} href='#!'>
            <SVG name='edit' />
            Edit Details
          </a>
        </Menu.Item>
        <Menu.Item key='1'>
          <a onClick={handleDeleteClick} href='#!'>
            <SVG name='delete1' />
            Delete
          </a>
        </Menu.Item>
      </Menu>
    );
  };

  return (
    <div className='flex gap-x-6 items-center'>
      <ControlledComponent controller={!savedQueryId}>
        <Button
          onClick={handleSaveClick}
          type={BUTTON_TYPES.PRIMARY}
          size={'large'}
          icon={<SVG name={'save'} size={20} color={'white'} />}
        >
          {'Save'}
        </Button>
      </ControlledComponent>

      <ControlledComponent controller={!!savedQueryId}>
        <Popover
          placement='bottom'
          visible={showSavedQueryPopover}
          content={
            <SavedQueryPopoverContent
              onCancel={toggleSavedQueryPopover}
              onOk={handlePopoverOkClick}
            />
          }
        >
          <Button
            onClick={toggleSavedQueryPopover}
            type={BUTTON_TYPES.SECONDARY}
            size={'large'}
            icon={<SVG name={'save'} size={20} color={'#8692A3'} />}
          >
            {'Save'}
          </Button>
        </Popover>
        <Tooltip placement='bottom' title='Save as New'>
          <Button
            onClick={handleSaveClick}
            size='large'
            type='text'
            icon={<SVG name={'pluscopy'} />}
          ></Button>
        </Tooltip>
        <Tooltip placement='bottom' title='Add to Dashboard'>
          <Button
            onClick={toggleAddToDashboardModal}
            size='large'
            type='text'
            icon={<SVG name={'addtodash'} />}
          ></Button>
        </Tooltip>
        <Dropdown overlay={getActionsMenu()} trigger={['hover']}>
          <Button
            size='large'
            type='text'
            icon={<SVG name={'threedot'} />}
          ></Button>
        </Dropdown>

        <DeleteQueryModal
          visible={showDeleteModal}
          onDelete={handleDelete}
          toggleModal={toggleDeleteModal}
        />
      </ControlledComponent>
    </div>
  );
};

export default QueryActions;

QueryActions.propTypes = {
  savedQueryId: PropTypes.oneOfType([
    PropTypes.number,
    PropTypes.instanceOf(null),
  ]),
  handleSaveClick: PropTypes.func,
  handleEditClick: PropTypes.func,
  handleDeleteReport: PropTypes.func,
  toggleAddToDashboardModal: PropTypes.func,
};

QueryActions.defaultProps = {
  savedQueryId: null,
  handleSaveClick: _.noop,
  handleEditClick: _.noop,
  handleDeleteReport: _.noop,
  toggleAddToDashboardModal: _.noop,
};
