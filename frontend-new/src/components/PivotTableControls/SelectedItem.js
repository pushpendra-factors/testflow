import React from 'react';
import PropTypes from 'prop-types';
import cx from 'classnames';

import { SVG, Text } from 'factorsComponents';
import { EMPTY_STRING } from 'Utils/global';

import styles from './PivotTableControls.module.scss';
import ControlledComponent from '../ControlledComponent';

const SelectedItem = ({ label, showRemoveBtn, onRemove }) => {
  const handleRemove = () => {
    onRemove(label);
  };

  return (
    <div
      className={cx(
        'py-2 px-3 flex items-center justify-between',
        styles.selectedBtn
      )}
    >
      <Text
        color='grey-2'
        extraClass='m-0'
        type={'title'}
        level={6}
        weight={'medium'}
      >
        {label}
      </Text>
      <ControlledComponent controller={showRemoveBtn}>
        <div className='cursor-pointer' onClick={handleRemove}>
          <SVG onClick={handleRemove} size={12} name={'close'} />
        </div>
      </ControlledComponent>
    </div>
  );
};

export default SelectedItem;

SelectedItem.propTypes = {
  label: PropTypes.string,
  showRemoveBtn: PropTypes.bool,
  onRemove: PropTypes.func,
};

SelectedItem.defaultProps = {
  label: EMPTY_STRING,
  showRemoveBtn: true,
  onRemove: _.noop,
};
