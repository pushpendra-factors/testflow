import React, { Fragment, memo } from 'react';
import PropTypes from 'prop-types';

const ControlledComponent = ({ controller, children }) => {
  if (!controller) return null;
  return <Fragment>{children}</Fragment>;
};

export default memo(ControlledComponent);

ControlledComponent.propTypes = {
  controller: PropTypes.oneOfType([
    PropTypes.string,
    PropTypes.number,
    PropTypes.bool
  ]).isRequired
};
