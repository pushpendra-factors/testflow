import _ from 'lodash';
import { SortData } from '../../utils/dataFormatter';
import { defaultState } from './constants';
import { DASHBOARD_KEYS } from '../../constants/localStorage.constants';
import { getItemFromLocalStorage } from '../../utils/localStorage.helpers';

export const cardClassNames = {
  1: 'w-full',
  0: 'w-1/2',
  2: 'w-1/4'
};

export const getRearrangedData = (units, dashboard) => {
  let result;
  if (!dashboard.units_position || !dashboard.units_position.position) {
    const clonedUnits = [...units];
    result = clonedUnits.reverse().map((u, index) => {
      return {
        ...u,
        position: index,
        className: 'w-full',
        cardSize: 1
      };
    });
  } else {
    const unitsPosition = dashboard.units_position.position;
    const nonPositionedUnits = units.filter((u) => {
      return !Object.prototype.hasOwnProperty.call(unitsPosition, u.id);
    });
    const positionedUnits = units.filter((u) => {
      return Object.prototype.hasOwnProperty.call(unitsPosition, u.id);
    });
    const result1 = nonPositionedUnits.reverse().map((u, index) => {
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
    result = [...result2, ...result1];
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

export const getUpdateStateOnDashboardsLoaded = ({ payload }) => {
  const lastSelectedDashboardID = getItemFromLocalStorage(
    DASHBOARD_KEYS.ACTIVE_DASHBOARD_ID
  );
  const fixedState = {
    ...defaultState,
    dashboards: { ...defaultState.dashboards, data: payload }
  };
  if (lastSelectedDashboardID) {
    const lastSelectedDashboard = _.find(
      payload,
      (db) => db.id === +lastSelectedDashboardID
    );
    if (lastSelectedDashboard) {
      return {
        ...fixedState,
        activeDashboard: lastSelectedDashboard
      };
    }
  }
  return {
    ...fixedState,
    activeDashboard: payload[0]
  };
};
