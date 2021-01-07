/* eslint-disable */

import { 
  fetchEventsAction, fetchEventPropertiesAction, 
  fetchUserPropertiesAction, setGroupByAction, 
  delGroupByAction, deleteGroupByEventAction, 
  setEventGoalAction, setMarketingTouchpointsAction, 
  setAttributionModelsAction, setAttributionWindowAction, 
  setAttrLinkEventsAction, setCampChannelAction, 
  setMeasuresAction, getCampaignConfigAction, 
  setCampFiltersAction, setCampGroupByAction
} from './actions';
import { getEventNames, fetchEventProperties, fetchUserProperties, fetchCampaignConfig } from './services';
import { convertToEventOptions, convertPropsToOptions, convertCampaignConfig } from './utils';

export const fetchEventNames = (projectId) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      getEventNames(dispatch, projectId)
        .then((response) => {
          const options = convertToEventOptions(response.data.event_names);
          resolve(dispatch(fetchEventsAction(options)));
        }).catch((err) => {
          resolve(dispatch(fetchEventsAction([])));
        });
    });
  };
};
 
export const getUserProperties = (projectId, queryType) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchUserProperties(projectId, queryType).then((response) => {
        const options = convertPropsToOptions(response.data);
        resolve(dispatch(fetchUserPropertiesAction(options)));
      }).catch((err) => {
        // resolve(dispatch(fetchEventPropertiesAction({})));
      })
    })
  }
}

export const getEventProperties = (projectId, eventName) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchEventProperties(projectId, eventName)
        .then((response) => {
          const options = convertPropsToOptions(response.data);
          resolve(dispatch(fetchEventPropertiesAction(options, eventName)));
        }).catch((err) => {
          // resolve(dispatch(fetchEventPropertiesAction({})));
        });
    });
  };
}

export const setGroupBy = (type, groupBy, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setGroupByAction(type, groupBy, index)))
    })
  }
}

export const delGroupBy = (type, payload, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(delGroupByAction(type, payload, index)))
    })
  }
}

export const deleteGroupByForEvent = (ev, index) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(deleteGroupByEventAction(ev, index)))
    })
  }
}

export const setGoalEvent = (goalEvent) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setEventGoalAction(goalEvent)))
    })
  }
}

export const setTouchPoint = (touchpoint) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setMarketingTouchpointsAction(touchpoint)))
    })
  }
}

export const setModels = (models) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttributionModelsAction(models)))
    })
  }
}

export const setWindow = (window) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttributionWindowAction(window)))
    })
  }
}

export const setLinkedEvents = (linkedEvents) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setAttrLinkEventsAction(linkedEvents)))
    })
  }
}

export const setCampChannel = (channel) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampChannelAction(channel)))
    })
  }
}

export const setCampMeasures = (measures) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setMeasuresAction(measures)))
    })
  }
}

export const setCampFilters = (filters) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampFiltersAction(filters)))
    })
  }
}

export const setCampGroupBy = (groupBy) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      resolve(dispatch(setCampGroupByAction(groupBy)))
    })
  }
}

export const getCampaignConfigData = (projectId, channel) => {
  return (dispatch) => {
    return new Promise((resolve, reject) => {
      fetchCampaignConfig(projectId, channel).then(res => {
          const payload = convertCampaignConfig(res.data.result);
          resolve(dispatch(getCampaignConfigAction(payload)));
        }).catch((err) => {
          console.log(err);
        });
    })
  }
}
