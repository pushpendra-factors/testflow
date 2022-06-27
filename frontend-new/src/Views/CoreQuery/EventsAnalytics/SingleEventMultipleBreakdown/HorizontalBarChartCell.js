import React, { memo, useState } from 'react';
import HorizontalBarChart from '../../../../components/HorizontalBarChart';
import { Text } from '../../../../components/factorsComponents';

function HorizontalBarChartCell({
  series,
  cardSize,
  isDashboardWidget,
  ...rest
}) {
  const [showAll, setShowAll] = useState(false);

  const displayedSeries = [
    {
      ...series[0],
      data: showAll ? series[0].data.slice(0, 20) : series[0].data.slice(0, 10)
    }
  ];

  const height =
    40 * displayedSeries[0].data.length > 75
      ? 40 * displayedSeries[0].data.length
      : 75;

  return (
    <>
      <HorizontalBarChart
        cardSize={cardSize}
        series={displayedSeries}
        height={height}
        {...rest}
      />
      {!isDashboardWidget && (
        <>
          {!showAll && series[0].data.length > 10 && (
            <div
              className="cursor-pointer"
              onClick={setShowAll.bind(null, true)}
            >
              <Text
                color="brand-6"
                type="title"
                weight="bold"
                extraClass="mb-0"
              >
                Show More
              </Text>
            </div>
          )}
          {showAll && (
            <div
              className="cursor-pointer"
              onClick={setShowAll.bind(null, false)}
            >
              <Text
                color="brand-6"
                type="title"
                weight="bold"
                extraClass="mb-0"
              >
                Show Less
              </Text>
            </div>
          )}
        </>
      )}
    </>
  );
}

export default memo(HorizontalBarChartCell);
