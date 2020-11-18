import { SortData } from '../../Views/CoreQuery/utils';

export const cardClassNames = {
  'full-page': 'w-full',
  'half-page': 'w-1/2',
  'oneThird-page': 'w-1/3'
};

export const getRearrangedData = (units, dashboard) => {
  let result;
  if (!dashboard.units_position) {
    result = units.map((u, index) => {
      return {
        ...u,
        position: index,
        className: 'w-full',
        cardSize: 'full-page'
      };
    });
  } else {
    const unitsPosition = dashboard.units_position;
    const nonPositionedUnits = units.filter(u => !unitsPosition[u.id]);
    const positionedUnits = units.filter(u => unitsPosition[u.id]);
    const result1 = nonPositionedUnits.map((u, index) => {
      return {
        ...u,
        position: index,
        className: 'w-full',
        cardSize: 'full-page'
      };
    });
    const startingPosition = nonPositionedUnits.length;
    let result2 = positionedUnits.map((u) => {
      return {
        ...u,
        position: unitsPosition[u.id].position + startingPosition,
        className: cardClassNames[unitsPosition[u.id].size],
        cardSize: unitsPosition[u.id].size
      };
    });
    result2 = SortData(result2, 'position', 'ascend');
    result = [...result1, ...result2];
  }
  return result;
};
