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
    result = clonedUnits.map((u, index) => ({
      ...u,
      position: index,
      className: 'w-full',
      cardSize: 1
    }));
  } else {
    const unitsPosition = dashboard.units_position.position;
    const nonPositionedUnits = units.filter(
      (u) =>
        !Object.prototype.hasOwnProperty.call(unitsPosition, u.id || u.inter_id)
    );
    const positionedUnits = units.filter((u) =>
      Object.prototype.hasOwnProperty.call(unitsPosition, u.id || u.inter_id)
    );
    const result1 = nonPositionedUnits.map((u, index) => ({
      ...u,
      position: index,
      className: 'w-full',
      cardSize: 1
    }));
    const startingPosition = nonPositionedUnits.length;
    let result2 = positionedUnits.map((u) => ({
      ...u,
      position: unitsPosition[u.id || u.inter_id] + startingPosition,
      className:
        cardClassNames[dashboard.units_position.size[u.id || u.inter_id]],
      cardSize: dashboard.units_position.size[u.id || u.inter_id]
    }));
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
    dashboards: { ...defaultState.dashboards, data: payload, completed: true }
  };
  if (lastSelectedDashboardID) {
    const lastSelectedDashboard = _.find(
      payload,
      (db) => db.id === lastSelectedDashboardID
    );
    if (lastSelectedDashboard) {
      return {
        ...fixedState,
        activeDashboard: lastSelectedDashboard
      };
    }
  }
  const data = payload?.[0]?.class !== 'predefined' ? payload[0] : payload[1];
  return {
    ...fixedState,
    activeDashboard: data
  };
};

export const getAllBoardsFolderId = (foldersList) => {
  const allBoardsFolderId = foldersList.find(
    (folder) => folder.name === 'All Boards'
  )?.id;
  return allBoardsFolderId;
};

export const getFoldersListWithDashboardIds = (dashboardsList, foldersList) => {
  const allBoardsFolderId = getAllBoardsFolderId(foldersList);
  const folderIdsMapper = dashboardsList.reduce((prev, curr) => {
    const { folder_id } = curr;

    if (Boolean(folder_id) === true) {
      if (prev[folder_id] != null) {
        return {
          ...prev,
          [folder_id]: [...prev[folder_id], curr.id]
        };
      }
      return {
        ...prev,
        [folder_id]: [curr.id]
      };
    }
    if (prev[allBoardsFolderId] != null) {
      return {
        ...prev,
        [allBoardsFolderId]: [...prev[allBoardsFolderId], curr.id]
      };
    }
    return {
      ...prev,
      [allBoardsFolderId]: [curr.id]
    };
  }, {});

  return foldersList.map((folder) => ({
    ...folder,
    dashboardIds: folderIdsMapper[folder.id] ?? []
  }));
};
