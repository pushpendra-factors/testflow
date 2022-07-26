import React from 'react';
import { useSelector } from 'react-redux';
import { Tooltip } from 'antd';
import { Text } from '../factorsComponents';
import { legend_counts } from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';

const legend_length = {
  0: 15,
  1: 20,
  2: 10,
};

function TopLegends({
  colors,
  legends,
  parentClassName = 'flex flex-wrap justify-center col-gap-3 row-gap-3',
  cardSize,
  showAllLegends = false,
  showFullLengthLegends = false,
}) {
  const itemsCount = showAllLegends ? legends.length : legend_counts[cardSize];

  const { eventNames } = useSelector((state) => state.coreQuery);

  const displayLegend = (legend) => {
    if (!legend) return null;
    return (
      <Text mini type='paragraph'>
        {legend.length > legend_length[cardSize] && !showFullLengthLegends
          ? legend.substr(0, legend_length[cardSize]) + '...'
          : legend}
      </Text>
    );
  };

  return (
    <div className={parentClassName}>
      {legends.slice(0, itemsCount).map((legend, index) => {
        const sanitisedLegend = eventNames[legend] || legend;
        return (
          <Tooltip key={legend + index} title={sanitisedLegend}>
            <div className='flex items-center col-gap-2'>
              <LegendsCircle color={colors[index]} />
              {displayLegend(sanitisedLegend)}
            </div>
          </Tooltip>
        );
      })}
    </div>
  );
}

export default TopLegends;
