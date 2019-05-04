export default function reducer(state={
  dashboards: [],
  units: [],
}, action) {

  switch (action.type) { 
    case "FETCH_DASHBOARDS_FULFILLED": {
      return { ...state, dashboards: action.payload }
    }
    case "CREATE_DASHBOARD_FULFILLED": {
      let _state = { ...state };
      _state.dashboards = [ ...state.dashboards ];
      _state.dashboards.push(action.payload);
      return _state;
    }
    case "FETCH_DASHBOARD_UNITS_FULFILLED": {
      return {
        ...state,
        units: action.payload
      }
    }
    case "DELETE_DASHBOARD_UNIT_FULFILLED": {
      let _state = { ...state };
      _state.units = [ ...state.units ];
      let delUnit = action.payload;

      // Get unit index to delete from store.
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
