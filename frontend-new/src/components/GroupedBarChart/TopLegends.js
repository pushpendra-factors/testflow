import React from 'react';
import { useSelector } from 'react-redux';
import { Tooltip } from 'antd';
import { Text } from '../factorsComponents';
import { legend_counts } from '../../utils/constants';
import LegendsCircle from '../../styles/components/LegendsCircle';
import truncateURL from 'Utils/truncateURL';

const legend_length = {
  0: 15,
  1: 20,
  2: 10
};

function TopLegends({
  colors,
  legends,
  parentClassName = 'flex flex-wrap justify-center gap-x-3 gap-y-3',
  cardSize,
  showAllLegends = false,
  showFullLengthLegends = false
}) {
  const itemsCount = showAllLegends ? legends.length : legend_counts[cardSize];

  const { eventNames } = useSelector((state) => state.coreQuery);
  const { projectDomainsList } = useSelector((state) => state.global);

  const displayLegend = (legend) => {
    if (!legend) return null;
    let urlTruncatedlegend = truncateURL(legend, projectDomainsList);
    return (
      <Text mini type='paragraph'>
        {urlTruncatedlegend.length > legend_length[cardSize] &&
        !showFullLengthLegends
          ? urlTruncatedlegend.slice(0, legend_length[cardSize]) + '...'
          : urlTruncatedlegend}
      </Text>
    );
  };

  return (
    <div className={parentClassName}>
      {legends.slice(0, itemsCount).map((legend, index) => {
        const sanitisedLegend = eventNames[legend] || legend;
        return (
          <Tooltip key={legend + index} title={sanitisedLegend}>
            <div className='flex items-center gap-x-2'>
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
