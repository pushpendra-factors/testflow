import React from 'react';
import PropTypes from 'prop-types';

import { Text } from 'factorsComponents';

const ColumnsHeading = ({ heading }) => {
  return (
    <Text
      color='black'
      extraClass='m-0'
      type={'title'}
      level={6}
      weight={'bold'}
    >
      {heading}
    </Text>
  );
};

export default ColumnsHeading;

ColumnsHeading.propTypes = {
  heading: PropTypes.string.isRequired,
};
