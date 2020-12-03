import { SortData } from "../../utils/dataFormatter";

export const cardClassNames = {
  1: 'w-full',
  0: 'w-1/2',
  2: 'w-1/3'
};

export const getRearrangedData = (units, dashboard) => {
  let result;
  if (!dashboard.units_position || !dashboard.units_position.position) {
    result = units.map((u, index) => {
      return {
        ...u,
        position: index,
        className: 'w-full',
        cardSize: 1
      };
    });
  } else {
    const unitsPosition = dashboard.units_position.position;
    const nonPositionedUnits = units.filter(u => {
      return !Object.prototype.hasOwnProperty.call(unitsPosition, u.id);
    });
    const positionedUnits = units.filter(u => {
      return Object.prototype.hasOwnProperty.call(unitsPosition, u.id);
    });
    const result1 = nonPositionedUnits.map((u, index) => {
      return {
        ...u,
        position: index,
        className: 'w-full',
        cardSize: 1
      };
    });
    const startingPosition = nonPositionedUnits.length;
    let result2 = positionedUnits.map((u) => {
      return {
        ...u,
        position: unitsPosition[u.id] + startingPosition,
        className: cardClassNames[dashboard.units_position.size[u.id]],
        cardSize: dashboard.units_position.size[u.id]
      };
    });
    result2 = SortData(result2, 'position', 'ascend');
    result = [...result1, ...result2];
  }
  return result;
};

export const getRequestForNewState = (newState) => {
  const body = {
    position: {},
    size: {}
  };
  newState.forEach((elem, index) => {
    body.position[elem.id] = index;
    body.size[elem.id] = elem.cardSize;
  });
  return body;
};
