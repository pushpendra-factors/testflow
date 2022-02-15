import React, { memo } from 'react';
import PropTypes from 'prop-types';

const ControlledComponent = ({ controller, children }) => {
  if (!controller) return null;
  return (
    <>
      {children}
    </>
  )
}

export default memo(ControlledComponent);

ControlledComponent.propTypes = {
  controller: PropTypes.oneOfType([PropTypes.string, PropTypes.number, PropTypes.bool]).isRequired
}