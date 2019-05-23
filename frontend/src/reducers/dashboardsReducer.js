export default function reducer(state={
  dashboards: [],
  units: [],
}, action) {

  switch (action.type) { 
    case "FETCH_DASHBOARDS_FULFILLED": {
      let _state = { ...state }
      _state.dashboards = action.payload;

      // reset units.
      if (_state.dashboards.length == 0)
        _state.units = [];

      return _state;
    }
    case "CREATE_DASHBOARD_FULFILLED": {
      let _state = { ...state };
      _state.dashboards = [ ...state.dashboards ];
      _state.dashboards.push(action.payload);
      return _state;
    }
    case "UPDATE_DASHBOARD_FULFILLED": {
      let _state = { ...state };
      _state.dashboards = [ ...state.dashboards ];
      let updateDashboard = action.payload;
      
      let updateIndex = -1;
      for (let i in _state.dashboards) {
        let dashboard = _state.dashboards[i];
        if(dashboard.project_id == updateDashboard.project_id
          && dashboard.id == updateDashboard.id) {
            updateIndex = i;
        }
      }
      
      if (updateIndex != -1) {
        _state.dashboards[updateIndex] = { 
          ..._state.dashboards[updateIndex],
          ...updateDashboard,
        }
      }

      return _state;
    }
    case "FETCH_DASHBOARD_UNITS_FULFILLED": {
      return {
        ...state,
        units: action.payload,
      }
    }
    case "UPDATE_DASHBOARD_UNIT_FULFILLED": {
      let _state = { ...state };
      _state.units = [ ...state.units ];
      let updateUnit = action.payload;

      // Get unit index to update, from store.
      let updateIndex = -1;
      for (let i in _state.units) {
        let unit = _state.units[i];
        if (unit.project_id == updateUnit.project_id
          && unit.dashboard_id == updateUnit.dashboard_id
          && unit.id == updateUnit.id) {
            updateIndex = i;
          }
      }

      if (updateIndex != -1) {
        _state.units[updateIndex] = { 
          ..._state.units[updateIndex],
          ...updateUnit,
        }
      }

      return _state;
    }
    case "DELETE_DASHBOARD_UNIT_FULFILLED": {
      let _state = { ...state };
      _state.units = [ ...state.units ];
      let delUnit = action.payload;

      // Get unit index to delete, from store.
      let delIndex = -1;
      for (let i in _state.units) {
        let unit = _state.units[i];
        if (unit.project_id == delUnit.project_id
          && unit.dashboard_id == delUnit.dashboard_id
          && unit.id == delUnit.id) {
            delIndex = i;
          }
      }

      if (delIndex != -1) 
        _state.units.splice(delIndex, 1);
      
      return _state;
    }
  }
  return state
}
