import React from 'react';
import { Text } from '../factorsComponents';
import { legend_counts } from '../../utils/constants';

const legend_length = {
  0: 15,
  1: 20,
  2: 10,
};

function TopLegends({
  colors,
  legends,
  parentClassName = 'flex flex-wrap justify-center py-3',
  cardSize,
  showFullLegends,
}) {
  let itemsCount = legend_counts[cardSize];
  return (
    <div className={parentClassName}>
      {legends.slice(0, itemsCount).map((legend, index) => {
        return (
          <div key={legend + index} className='flex items-center'>
            <div
              style={{
                backgroundColor: colors[index],
                width: '16px',
                height: '16px',
                borderRadius: '8px',
              }}
            ></div>
            <div className='px-2'>
              <Text mini type='paragraph'>
                {legend.length > legend_length[cardSize] && !showFullLegends
                  ? legend.substr(0, legend_length[cardSize]) + '...'
                  : legend}
              </Text>
            </div>
          </div>
        );
      })}
    </div>
  );
}

export default TopLegends;
