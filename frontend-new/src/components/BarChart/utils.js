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
  if (margin >= 50 && margin <= 70) {
    return 50;
  } else if (margin >= 80 && margin <= 100) {
    return 70;
  } else if (margin >= 110 && margin <= 130) {
    return 90;
  } else if (margin >= 140) {
    return 110;
  } else {
    return margin;
  }
};
