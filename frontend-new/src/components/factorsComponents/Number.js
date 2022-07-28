import React from 'react';
import NumberFormat from 'react-number-format';
import { isArray } from 'lodash';
import { abbreviateNumber } from '../../utils/dataFormatter';

const Number = ({
  type,
  number,
  className,
  shortHand = false,
  suffix = '',
  prefix = ''
}) => {
  const finalVal = shortHand ? abbreviateNumber(number) : number;

  return (
    <span className={className}>
      {shortHand ? (
        `${prefix}${abbreviateNumber(number)}${suffix}`
      ) : (
        <NumberFormat
          displayType={'text'}
          value={isArray(finalVal) ? finalVal[0] : finalVal}
          thousandSeparator={true}
          decimalScale={1}
          suffix={suffix}
          prefix={prefix}
        />
      )}
    </span>
  );
};

export default Number;
