import { fetchKPIFilterValues } from 'Reducers/kpi';
import {
  fetchEventsAction,
  fetchEventPropertiesAction,
  fetchUserPropertiesAction,
  setGroupByAction,
  delGroupByAction,
  deleteGroupByEventAction,
  setEventGoalAction,
  setMarketingTouchpointsAction,
  setAttributionModelsAction,
  setAttributionWindowAction,
  setAttrLinkEventsAction,
  setCampChannelAction,
  setMeasuresAction,
  getCampaignConfigAction,
  setCampFiltersAction,
  setCampGroupByAction,
  setAttrDateRangeAction,
  setCampDateRangeAction,
  setDefaultStateAction,
  setTouchPointFiltersAction,
  setAttributionQueryTypeAction,
  setTacticOfferTypeAction,
  setEventsDisplayAction,
  setUserPropertiesNamesAction,
  setEventPropertiesNamesAction,
  setGroupPropertiesNamesAction,
  fetchGroupPropertiesAction,
  resetGroupByAction,
  fetchEventsMapAction,
  fetchEventUserPropertiesAction,
  setButtonClicksPropertiesNamesAction,
  setPageViewsPropertiesNamesAction,
  FETCH_PROPERTY_VALUES_LOADING,
  FETCH_PROPERTY_VALUES_LOADED,
  fetchUserPropertiesActionV2,
  fetchEventUserPropertiesActionV2,
  fetchEventPropertiesActionV2
} from './actions';
import {
  getEventNames,
  fetchEventProperties,
  fetchEventPropertiesV2,
  fetchUserProperties,
  fetchGroupProperties,
  fetchCampaignConfig,
  fetchEventPropertyValues,
  fetchGroupPropertyValues,
  fetchUserPropertyValues,
  fetchButtonClicksPropertyValues,
  fetchPageViewsPropertyValues,
  fetchUserPropertiesV2
} from './services';
import {
  convertToEventOptions,
  convertPropsToOptions,
  convertCampaignConfig,
  convertCustomEventCategoryToOptions,
  convertEventsPropsToOptions,
  convertUserPropsToOptions
} from './utils';

export const fetchEventNames = (projectId) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      getEventNames(dispatch, projectId)
        .then((response) => {
          dispatch(fetchEventsMapAction(response.data.event_names));
          const options = convertToEventOptions(
            response.data.event_names,
            response.data.display_names
          );
          dispatch(setEventsDisplayAction(response.data.display_names));
          resolve(dispatch(fetchEventsAction(options)));
        })
        .catch((err) => {
          resolve(dispatch(fetchEventsAction([])));
        });
    });
  };
};

export const getGroupProperties = (projectId, groupName) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchGroupProperties(projectId, groupName)
        .then((response) => {
          const options = convertPropsToOptions(
            response.data?.properties,
            response.data?.display_names
          );
          resolve(
            dispatch(
              setGroupPropertiesNamesAction(response.data?.display_names)
            )
          );
          resolve(dispatch(fetchGroupPropertiesAction(options, groupName)));
        })
        .catch((err) => {
          resolve(dispatch(fetchGroupPropertiesAction({})));
        });
    });
  };
};
export const getUserPropertiesV2 = (projectId, queryType = '') => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchUserPropertiesV2(projectId, queryType)
        .then((response) => {
          const options = convertUserPropsToOptions(
            response.data?.properties,
            response.data?.display_names,
            response.data?.disabled_event_user_properties
          );
          resolve(
            dispatch(setUserPropertiesNamesAction(response.data?.display_names))
          );
          resolve(dispatch(fetchUserPropertiesActionV2(options.userOptions)));
          resolve(
            dispatch(fetchEventUserPropertiesActionV2(options.eventUserOptions))
          );
        })
        .catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
};
export const getEventPropertiesV2 = (projectId, eventName) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchEventPropertiesV2(projectId, eventName)
        .then((response) => {
          const options = convertEventsPropsToOptions(
            response.data.properties,
            response.data?.display_names
          );
          resolve(
            dispatch(
              setEventPropertiesNamesAction(response.data?.display_names)
            )
          );
          resolve(dispatch(fetchEventPropertiesActionV2(options, eventName)));
        })
        .catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
};

export const getButtonClickProperties = (projectId) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchButtonClicksPropertyValues(projectId)
        .then((response) => {
          const transformedData = convertCustomEventCategoryToOptions(
            response.data
          );
          resolve(
            dispatch(setButtonClicksPropertiesNamesAction(transformedData))
          );
        })
        .catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
};
export const getPageViewsProperties = (projectId) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchPageViewsPropertyValues(projectId)
        .then((response) => {
          const transformedData = convertCustomEventCategoryToOptions(
            response.data
          );
          resolve(dispatch(setPageViewsPropertiesNamesAction(transformedData)));
        })
        .catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
};

export const setGroupBy = (groupByType, groupBy, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setGroupByAction(groupByType, groupBy, index)));
    });
  };
};
export const resetGroupBy = () => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(resetGroupByAction()));
    });
  };
};

export const delGroupBy = (type, payload, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(delGroupByAction(type, payload, index)));
    });
  };
};

export const deleteGroupByForEvent = (ev, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(deleteGroupByEventAction(ev, index)));
    });
  };
};

export const setGoalEvent = (goalEvent) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setEventGoalAction(goalEvent)));
    });
  };
};

export const setTouchPoint = (touchpoint) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setMarketingTouchpointsAction(touchpoint)));
    });
  };
};

export const setTouchPointFilters = (touchPointFilters) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setTouchPointFiltersAction(touchPointFilters)));
    });
  };
};

export const setattrQueryType = (attrQueryType) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttributionQueryTypeAction(attrQueryType)));
    });
  };
};

export const setTacticOfferType = (tacticOfferType) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setTacticOfferTypeAction(tacticOfferType)));
    });
  };
};

export const setAttrDateRange = (dateRange) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttrDateRangeAction(dateRange)));
    });
  };
};

export const setModels = (models) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttributionModelsAction(models)));
    });
  };
};

export const setWindow = (window) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttributionWindowAction(window)));
    });
  };
};

export const setLinkedEvents = (linkedEvents) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttrLinkEventsAction(linkedEvents)));
    });
  };
};

export const setCampChannel = (channel) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampChannelAction(channel)));
    });
  };
};

export const setCampMeasures = (measures) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setMeasuresAction(measures)));
    });
  };
};

export const setCampFilters = (filters) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampFiltersAction(filters)));
    });
  };
};

export const setCampGroupBy = (groupBy) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampGroupByAction(groupBy)));
    });
  };
};

export const setCampDateRange = (dateRange) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampDateRangeAction(dateRange)));
    });
  };
};

export const getCampaignConfigData = (projectId, channel) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchCampaignConfig(projectId, channel)
        .then((res) => {
          const payload = convertCampaignConfig(res.data.result);
          resolve(dispatch(getCampaignConfigAction(payload)));
        })
        .catch((err) => {
          console.log(err);
        });
    });
  };
};

export const resetState = () => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      setDefaultStateAction();
    }).catch((err) => {
      console.log(err);
    });
  };
};

export const getUserPropertyValues =
  (projectId, propertyName) => (dispatch) => {
    return new Promise((resolve, reject) => {
      dispatch({ type: FETCH_PROPERTY_VALUES_LOADING });
      fetchUserPropertyValues(projectId, propertyName)
        .then((response) => {
          resolve(
            dispatch({
              type: FETCH_PROPERTY_VALUES_LOADED,
              payload: response.data,
              propName: propertyName
            })
          );
        })
        .catch((err) => {
          console.log(err);
          resolve(
            dispatch({
              type: FETCH_PROPERTY_VALUES_LOADED,
              payload: {},
              propName: propertyName
            })
          );
        });
    });
  };

export const getEventPropertyValues =
  (projectId, eventName, propertyName) => (dispatch) => {
    return new Promise((resolve, reject) => {
      dispatch({ type: FETCH_PROPERTY_VALUES_LOADING });
      fetchEventPropertyValues(projectId, eventName, propertyName)
        .then((response) => {
          resolve(
            dispatch({
              type: FETCH_PROPERTY_VALUES_LOADED,
              payload: response.data,
              propName: propertyName
            })
          );
        })
        .catch((err) => {
          console.log(err);
          resolve(
            dispatch({
              type: FETCH_PROPERTY_VALUES_LOADED,
              payload: {},
              propName: propertyName
            })
          );
        });
    });
  };

export const getGroupPropertyValues =
  (projectId, groupName, propertyName) => (dispatch) => {
    return new Promise((resolve, reject) => {
      dispatch({ type: FETCH_PROPERTY_VALUES_LOADING });
      fetchGroupPropertyValues(projectId, groupName, propertyName)
        .then((response) => {
          resolve(
            dispatch({
              type: FETCH_PROPERTY_VALUES_LOADED,
              payload: response.data,
              propName: propertyName
            })
          );
        })
        .catch((err) => {
          console.log(err);
          resolve(
            dispatch({
              type: FETCH_PROPERTY_VALUES_LOADED,
              payload: {},
              propName: propertyName
            })
          );
        });
    });
  };

export const getKPIPropertyValues = (projectId, data) => (dispatch) => {
  return new Promise((resolve, reject) => {
    dispatch({ type: FETCH_PROPERTY_VALUES_LOADING });
    fetchKPIFilterValues(projectId, data)
      .then((response) => {
        resolve(
          dispatch({
            type: FETCH_PROPERTY_VALUES_LOADED,
            payload: response.data,
            propName: data?.property_name
          })
        );
      })
      .catch((err) => {
        console.log(err);
        resolve(
          dispatch({
            type: FETCH_PROPERTY_VALUES_LOADED,
            payload: {},
            propName: data?.property_name
          })
        );
      });
  });
};
