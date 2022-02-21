import React from 'react';
import PropTypes from 'prop-types';
import { SVG } from 'factorsComponents';
import { Button, Tooltip, Dropdown, Menu } from 'antd';
import { BUTTON_TYPES } from '../../utils/buttons.constants';
import ControlledComponent from '../ControlledComponent';
import styles from './index.module.scss';

const QueryActions = ({
  savedQueryId,
  handleSaveClick,
  handleEditClick,
  handleDeleteClick,
  toggleAddToDashboardModal,
}) => {
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
        {/* <Popover
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
        </Popover> */}
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
