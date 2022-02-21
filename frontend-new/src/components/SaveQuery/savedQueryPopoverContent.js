import React, { memo } from 'react';
import PropTypes from 'prop-types';
import { Button } from 'antd';

import { SVG, Text } from 'factorsComponents';

import { BUTTON_TYPES } from '../../utils/buttons.constants';

const SavedQueryPopoverContent = ({ onCancel, onOk }) => {
  return (
    <div className='flex gap-x-2 p-4'>
      <SVG name={'infocircle'} />
      <div className='flex flex-col gap-y-4'>
        <div className='flex flex-col gap-y-1'>
          <Text mini color='grey-2' type='paragraph'>
            Any changes made to the saved report cannot be saved.
          </Text>

          <Text mini color='grey-2' type='paragraph'>
            Do you want to save this as a new report?
          </Text>
        </div>
        <div className='flex gap-x-2 justify-end'>
          <Button
            onClick={onCancel}
            type={BUTTON_TYPES.SECONDARY}
            size={'large'}
          >
            {'Cancel'}
          </Button>
          <Button onClick={onOk} type={BUTTON_TYPES.PRIMARY} size={'large'}>
            {'Yes'}
          </Button>
        </div>
      </div>
    </div>
  );
};

export default memo(SavedQueryPopoverContent);

SavedQueryPopoverContent.propTypes = {
  onCancel: PropTypes.func,
  onOk: PropTypes.func,
};

SavedQueryPopoverContent.defaultProps = {
  onCancel: _.noop,
  onOk: _.noop,
};
