import React from 'react';
import { Text } from '../factorsComponents';
import { charts_legend_length } from '../../utils/constants';

function TopLegends({
  parentClassName = 'flex justify-center py-3',
  cardSize,
  legends,
}) {
  return (
    <div className={parentClassName}>
      <div className='flex items-center'>
        <div
          style={{
            backgroundColor: 'rgb(77, 125, 180)',
            width: '16px',
            height: '16px',
            borderRadius: '8px',
          }}
        ></div>
        <div className='px-2'>
          <Text mini type='paragraph'>
            {legends[0].length > charts_legend_length[cardSize] &&
            cardSize !== 1
              ? legends[0].substr(0, charts_legend_length[cardSize]) + '...'
              : legends[0]}
          </Text>
        </div>
      </div>
      {cardSize !== 2 ? (
        <div className='flex items-center'>
          <div
            style={{
              backgroundColor: 'rgb(212, 120, 125)',
              width: '16px',
              height: '16px',
              borderRadius: '8px',
            }}
          ></div>
          <div className='px-2'>
            <Text mini type='paragraph'>
              {legends[1]}
            </Text>
          </div>
        </div>
      ) : null}
    </div>
  );
}

export default TopLegends;
