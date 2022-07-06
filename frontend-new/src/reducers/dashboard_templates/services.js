import { get, getHostUrl, post, del, put } from '../../utils/request';
import {
  TEMPLATES_LOADED,
  // TEMPLATES_UNITS_LOADING_FAILED,
  TEMPLATES_LOADING,
  TEMPLATES_LOADING_FAILED,
  // TEMPLATE_UNITS_LOADING,
  // TEMPLATE_UNITS_LOADED,
} from '../types';

const host = getHostUrl();

export const fetchTemplates = () => {
  return async function (dispatch) {
    try {
      dispatch({ type: TEMPLATES_LOADING });
      const url = host + 'common/dashboard_templates';
      const res = await get(null, url);
      dispatch({ type: TEMPLATES_LOADED, payload: res.data });
    } catch (err) {
      console.log(err);
      dispatch({ type: TEMPLATES_LOADING_FAILED });
    }
  };
};

export const createDashboardFromTemplate = async (projectId,templateId)=>{
  const url = host + 'projects/' + projectId + '/dashboard_template/' + templateId+ '/trigger';
  return post(null,url);
}
