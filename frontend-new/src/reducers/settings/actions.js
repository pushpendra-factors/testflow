export const FETCH_SMART_PROPERTIES = 'FETCH_SMART_PROPERTIES';
export const FETCH_PROPERTY_CONFIG = 'FETCH_PROPERTY_CONFIG';
export const CREATE_SMART_PROPERTY = 'CREATE_SMART_PROPERTY';
export const UPDATE_SMART_PROPERTY = 'UPDATE_SMART_PROPERTY';

export const fetchSmartPropertiesAction = (smartProperties, status = 'started') => {
    return { type: FETCH_SMART_PROPERTIES, payload:  smartProperties};
};

export const fetchSmartPropertyConfigAction = (config) => {
    return {type: FETCH_PROPERTY_CONFIG, payload: config}
}

export const createSmartPropertyAction = (smartProperty) => {
    return {type: CREATE_SMART_PROPERTY, payload: smartProperty}
}

export const updateSmartPropertyAction = (smartProperty) => {
    return {type: UPDATE_SMART_PROPERTY, payload: smartProperty}
}