export function getReadableChannelMetricValue(key, value, meta) {
  if (value == null || value == undefined) return 0;
  if (typeof(value) != "number") return value;

  let rValue = value;
  let isFloat = (value % 1) > 0
  if (isFloat) rValue = value.toFixed(0);

  if (meta && meta.currency && key.toLowerCase().indexOf('cost') > -1)
    rValue = rValue + ' ' + meta.currency;

  return rValue;
}