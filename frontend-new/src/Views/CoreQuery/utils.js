export const initialResultState = [1, 2, 3, 4].map(() => {
  return { loading: false, error: false, data: null };
});

export const calculateFrequencyData = (eventData, userData) => {
  const rows = eventData.result_group[0].rows.map((elem, index) => {
    const eventVals = elem.slice(1).map((e, idx) => {
      if (!e) return e;
      const eVal = e / userData.result_group[0].rows[index][idx + 1];
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [elem[0], ...eventVals];
  });
  const result = { result_group: [{ ...eventData.result_group[0], rows }] };
  return result;
};

export const calculateActiveUsersData = (userData, sessionData) => {
  const rows = userData.result_group[0].rows.map((elem) => {
    const eventVals = elem.slice(1).map((e) => {
      if (!e) return e;
      const eVal = sessionData.result_group[0].rows[0] / e;
      return eVal % 1 !== 0 ? parseFloat(eVal.toFixed(2)) : eVal;
    });
    return [elem[0], ...eventVals];
  });
  const result = { result_group: [{ ...userData.result_group[0], rows }] };
  return result;
};
