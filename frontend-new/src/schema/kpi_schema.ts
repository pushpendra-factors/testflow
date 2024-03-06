import * as yup from 'yup';

const FilterSchema = yup.object({
  extra: yup.array().of(yup.string()).optional(),
  prNa: yup.string().required('prNa is required'),
  prDaTy: yup.string().required('prDaTy is required'),
  co: yup.string().required('co is required'),
  lOp: yup.string().required('lOp is required'),
  en: yup.string().required('en is required'),
  objTy: yup.string().optional(),
  va: yup.string().optional(),
  isPrMa: yup.boolean().optional()
});

const QueryGroup = yup.object({
  ca: yup.string().required('ca is required'),
  pgUrl: yup.string().optional(),
  dc: yup.string().required('dc is required'),
  me: yup.array().of(yup.string().required('me is required')),
  fil: yup.array().of(FilterSchema).optional(),
  gBy: yup.array().optional(),
  fr: yup.number().positive('fr must be positive'),
  to: yup.number().positive('to must be positive'),
  tz: yup.string().required('tz is required'),
  qt: yup.string().required('qt is required'),
  an: yup.string().optional(),
  gbt: yup.string().optional()
});

const QueryGroupBy = yup.object({
  gr: yup.string().optional(),
  prNa: yup.string().required('prNa is required'),
  prDaTy: yup.string().required('prDaTy is required'),
  en: yup.string().required('en is required'),
  objTy: yup.string().optional(),
  dpNa: yup.string().required('dpNa is required'),
  isPrMa: yup.boolean().optional()
});

const KpiSchema = yup.object({
  cl: yup.string().required('cl is required'),
  qG: yup.array().of(QueryGroup),
  gGBy: yup.array().of(QueryGroupBy).optional(),
  gFil: yup.array().of(FilterSchema).optional()
});
export default KpiSchema;
