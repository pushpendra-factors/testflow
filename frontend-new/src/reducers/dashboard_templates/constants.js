import { EMPTY_OBJECT, EMPTY_ARRAY } from '../../utils/global';

export const defaultState = {
  templates: {
    loading: false,
    error: false,
    data: EMPTY_ARRAY
  },
  activeTemplate: EMPTY_OBJECT,
  activeTemplateUnits: {
    loading: false,
    error: false,
    data: EMPTY_ARRAY
  }
};
