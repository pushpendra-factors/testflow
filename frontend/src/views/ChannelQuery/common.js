export function getReadableChannelMetricValue(key, value, meta) {
  if (value == null || value == undefined) return 0;
  if (typeof(value) != "number") return value;

  let rValue = value;
  let isFloat = (value % 1) > 0
  // no decimal points for value >= 1 and 2 decimal points < 1.
  if (isFloat) rValue = value >= 1 ? value.toFixed(0) : value.toFixed(2);

  if (meta && meta.currency && key.toLowerCase().indexOf('cost') > -1)
    rValue = rValue + ' ' + meta.currency;

  return rValue;
}