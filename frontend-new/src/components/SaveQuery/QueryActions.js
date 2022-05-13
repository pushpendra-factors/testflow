import React, { useState } from 'react';
import _ from 'lodash';
import PropTypes from 'prop-types';
import { SVG } from 'factorsComponents';
import { Button, Tooltip } from 'antd';
import { BUTTON_TYPES } from '../../constants/buttons.constants';
import ControlledComponent from '../ControlledComponent';
import styles from './index.module.scss';
import { QUERY_TYPE_PROFILE } from '../../utils/constants';
import FaSelect from 'Components/FaSelect';

const QueryActions = ({
  queryType,
  savedQueryId,
  handleSaveClick,
  handleEditClick,
  handleDeleteClick,
  toggleAddToDashboardModal
}) => {
  const [options, setOptions] = useState(false);

  const setActions = (opt) => {
    if (opt[1] === 'edit') {
      handleEditClick();
    } else if (opt[1] === 'trash') {
      handleDeleteClick();
    }
    setOptions(false);
  };

  const getActionsMenu = () => {
    return options ? (
      <FaSelect
        extraClass={styles.additionalops}
        options={[
          ['Edit Details', 'edit'],
          ['Delete', 'trash']
        ]}
        optionClick={(val) => setActions(val)}
        onClickOutside={() => setOptions(false)}
        posRight={true}
      ></FaSelect>
    ) : null;
  };

  return (
    <div className="flex gap-x-6 items-center">
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
        <Tooltip placement="bottom" title="Save as New">
          <Button
            onClick={handleSaveClick}
            size="large"
            type="text"
            icon={<SVG name={'pluscopy'} />}
          ></Button>
        </Tooltip>
        <ControlledComponent controller={queryType !== QUERY_TYPE_PROFILE}>
          <Tooltip placement="bottom" title="Add to Dashboard">
            <Button
              onClick={toggleAddToDashboardModal}
              size="large"
              type="text"
              icon={<SVG name={'addtodash'} />}
            ></Button>
          </Tooltip>
        </ControlledComponent>
        <div className={'relative'}>
          <Button
            size="large"
            type="text"
            icon={<SVG name={'threedot'} />}
            onClick={() => setOptions(!options)}
          ></Button>
          {getActionsMenu()}
        </div>
      </ControlledComponent>
    </div>
  );
};

export default QueryActions;

QueryActions.propTypes = {
  savedQueryId: PropTypes.oneOfType([
    PropTypes.number,
    PropTypes.instanceOf(null)
  ]),
  handleSaveClick: PropTypes.func,
  handleEditClick: PropTypes.func,
  handleDeleteReport: PropTypes.func,
  toggleAddToDashboardModal: PropTypes.func
};

QueryActions.defaultProps = {
  savedQueryId: null,
  handleSaveClick: _.noop,
  handleEditClick: _.noop,
  handleDeleteReport: _.noop,
  toggleAddToDashboardModal: _.noop
};
