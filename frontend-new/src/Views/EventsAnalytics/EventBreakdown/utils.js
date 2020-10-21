export const formatData = (data) => {
  const resultInObjFormat = {};
  data.rows.forEach(d => {
    const date = d[0];
    const str = d.slice(1, d.length - 1).join(',');
    if (resultInObjFormat[str]) {
      resultInObjFormat[str].datewise.push({
        date,
        value: d[d.length - 1]
      });
      resultInObjFormat[str].value += d[d.length - 1];
    } else {
      resultInObjFormat[str] = {
        value: d[d.length - 1],
        datewise: [{
          date,
          value: d[d.length - 1]
        }]
      };
    }
  });
  const result = [];
  let idx = 0;
  for (const key in resultInObjFormat) {
    result.push({
      ...resultInObjFormat[key],
      label: key,
      index: idx
    });
    idx++;
  }
  result.sort((a, b) => {
    return parseInt(a.value) <= parseInt(b.value) ? 1 : -1;
  });
  return result;
};
