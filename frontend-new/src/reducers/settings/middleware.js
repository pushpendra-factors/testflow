
import {  
    fetchSmartPropertiesAction, 
    fetchSmartPropertyConfigAction, 
    fetchClickableElementsAction, 
    toggleClickableElementAction,
} from './actions';
import {
    getSmartProperties, 
    getSmartPropertiesConfig, 
    createSmartProperty, 
    modifySmartProperty, 
    removeSmartProperty,
    getClickableElements,
    enableOrDisableClickableElement,
} from './services';


export const fetchSmartProperties = (projectId) => {
    return (dispatch) => {
      return new Promise((resolve, reject) => {
        getSmartProperties(dispatch, projectId).then((response) => {
            const options = [...response.data];
            resolve(dispatch(fetchSmartPropertiesAction(options)));
          }).catch((err) => {
            resolve(dispatch(fetchSmartPropertiesAction([])));
          });
      });
    };
  };

export const fetchSmartPropertiesConfig = (projectId, type) => {
    return (dispatch) => {
        return new Promise((resolve, reject) => {
            getSmartPropertiesConfig(dispatch, projectId, type).then((res) => {
                if(res?.data) {
                    resolve(dispatch(fetchSmartPropertyConfigAction(res.data)));
                }
            }).catch((err) => {
                resolve(dispatch(fetchSmartPropertyConfigAction({})));
            })
        })
    }
}

export const addSmartProperty = (projectId, smartProperty) => {
    return (dispatch) => {
        return new Promise((resolve, reject) => {
            createSmartProperty(dispatch, projectId, smartProperty).then((res) => {
                resolve(res);
            }).catch((err) => {
                reject(err);
                // resolve(dispatch(fetchSmartPropertyConfigAction({})));
            })
        })
    }
}

export const updateSmartProperty = (projectId, smartProperty) => {
    return (dispatch) => {
        return new Promise((resolve, reject) => {
            modifySmartProperty(dispatch, projectId, smartProperty).then((res) => {
                resolve(res);
            }).catch((err) => {
                reject(err);
            })
        })
    }
}

export const deleteSmartProperty = (projectId, id) => {
    return (dispatch) => {
        return new Promise((resolve, reject) => {
            removeSmartProperty(dispatch, projectId, id).then((res) => {
                resolve(res);
            }).catch((err) => {
                reject(err);
            })
        })
    }
}

export const fetchClickableElements = (projectId) => {
    return (dispatch) => {
        return new Promise((resolve, reject) => {
            getClickableElements(dispatch, projectId).then((res) => {
                resolve(dispatch(fetchClickableElementsAction(res.data)));
            }).catch((err) => {
                resolve(dispatch(fetchClickableElementsAction([])));
            });
        });
    }
}

export const toggleClickableElement = (projectId, id, currentState) => {
    return (dispatch) => {
        return new Promise((resolve, reject) => {
            enableOrDisableClickableElement(dispatch, projectId, id).then(() => {
                // Set toggled state if toggle API is successful
                resolve(dispatch(toggleClickableElementAction({projectId: projectId, id: id, enabled: !currentState})));
            }).catch(() => {
                resolve(dispatch(toggleClickableElementAction({projectId: projectId, id: id, enabled: currentState})));
            });
        });
    }
}