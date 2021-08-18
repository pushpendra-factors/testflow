import React from 'react';
import { Text } from '../factorsComponents';
import { legend_counts } from '../../utils/constants';
import { useSelector } from 'react-redux';

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
  showAllLegends = false,
  showFullLengthLegends = false,
}) {
  const itemsCount = showAllLegends ? legends.length : legend_counts[cardSize];

  const { eventNames } = useSelector((state) => state.coreQuery);

  const displayLegend = (legend) => {
    const sanitisedLegend = eventNames[legend] || legend;
    return (
      <Text mini type='paragraph'>
        {sanitisedLegend.length > legend_length[cardSize] &&
        !showFullLengthLegends
          ? sanitisedLegend.substr(0, legend_length[cardSize]) + '...'
          : sanitisedLegend}
      </Text>
    );
  };

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
            <div className='px-2'>{displayLegend(legend)}</div>
          </div>
        );
      })}
    </div>
  );
}

export default TopLegends;
