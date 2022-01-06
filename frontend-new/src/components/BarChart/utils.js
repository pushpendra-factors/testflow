export const getMaxYpoint = (maxVal) => {
  let it = 1;
  const mxVal = isNaN(maxVal) ? 0 : maxVal;
  while (true) {
    if (Math.pow(10, it) < mxVal) {
      it++;
    } else {
      break;
    }
  }
  const pow10 = Math.pow(10, it - 1);
  it = 2;
  while (true) {
    if (pow10 * it > mxVal) {
      return pow10 * it;
    } else {
      it = it + 2;
    }
  }
};

export const getBarChartLeftMargin = (maxVal) => {
  const margin = maxVal.toString().length * 10;
  if (margin >= 50) {
    return margin - ((margin - 50) / 10) * 4;
  }
  if (margin < 25) {
    return 25;
  }
  return margin;
};
