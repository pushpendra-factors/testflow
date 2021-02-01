export const getMaxYpoint = (maxVal) => {
  let it = 1;
  while (true) {
    if (Math.pow(10, it) < maxVal) {
      it++;
    } else {
      break;
    }
  }
  const pow10 = Math.pow(10, it - 1);
  it = 2;
  while (true) {
    if (pow10 * it > maxVal) {
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
  if(margin < 20) {
    return 20;
  }
  return margin;
};
