
import { fetchSmartPropertiesAction, fetchSmartPropertyConfigAction } from './actions';
import {getSmartProperties, getSmartPropertiesConfig, 
    createSmartProperty, modifySmartProperty, 
    removeSmartProperty} from './services';


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