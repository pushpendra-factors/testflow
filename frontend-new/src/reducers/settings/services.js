import { get, getHostUrl, post, del, put } from "../../utils/request";

const host = getHostUrl();

export const getSmartProperties = (dispatch, projectId) => {
  return get(dispatch, host + "projects/" + projectId + "/v1/smart_properties/rules", {});
};

export const getPropertyMapping = (dispatch, projectId) => {
  return get(dispatch, host + "projects/" + projectId + "/v1/kpi/property_mappings", {});
};

export const getSmartPropertiesConfig = (dispatch, projectId, type) => {
    return get(dispatch, host + "projects/" + projectId + "/v1/smart_properties/config/" + type, {});
};

export const createSmartProperty = (dispatch, projectId, smartProperty) => {
    return post(dispatch, host + "projects/" + projectId + "/v1/smart_properties/rules", smartProperty);
}

export const createPropertyMapping = (dispatch, projectId, property) => {
    return post(dispatch, host + "projects/" + projectId + "/v1/kpi/property_mappings", property);
}
export const deletePropertyMapping = (dispatch, projectId, propertyId) => {
    return del(dispatch, host + "projects/" + projectId + "/v1/kpi/property_mappings/"+propertyId);
}

export const modifySmartProperty = (dispatch, projectId, smartProperty) => {
  return put(dispatch, host + "projects/" + projectId + "/v1/smart_properties/rules/" + smartProperty.id, smartProperty);
}

export const removeSmartProperty = (dispatch, projectId, id) => {
  return del(dispatch, host + "projects/" + projectId + "/v1/smart_properties/rules/" + id);
}

export const getClickableElements = (dispatch, projectId) => {
  return get(dispatch, host + "projects/" + projectId + "/clickable_elements", {});
}

export const enableOrDisableClickableElement = (dispatch, projectId, id) => {
  return get(dispatch, host + "projects/" + projectId + "/clickable_elements/"+ id +"/toggle", {});
}

