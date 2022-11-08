import React from 'react';
import NumberFormat from 'react-number-format';
import { isArray } from 'lodash';
import { abbreviateNumber } from '../../utils/global';

function Number({
  number,
  className,
  shortHand = false,
  suffix = '',
  prefix = ''
}) {
  const finalVal = shortHand
    ? abbreviateNumber(number)
    : isArray(number)
    ? number[0]
    : number;

  return (
    <span className={className}>
      {shortHand ? (
        `${prefix}${abbreviateNumber(number)}${suffix}`
      ) : (
        <NumberFormat
          displayType='text'
          value={finalVal}
          thousandSeparator
          decimalScale={finalVal < 10 ? 2 : 1}
          suffix={suffix}
          prefix={prefix}
        />
      )}
    </span>
  );
}

export default Number;
