export const initialResultState = [1, 2, 3, 4].map(elem => {
    return { loading: false, error: false, data: null }
});

export const calculateFrequencyData = (eventData, userData) => {
    const rows = eventData.result_group[0].rows.map((elem, index) => {
        const eventVals = elem.slice(1).map((e, idx) => {
            if (!e) return e;
            const e_val = e / userData.result_group[0].rows[index][idx + 1];
            return e_val % 1 !== 0 ? parseFloat(e_val.toFixed(2)) : e_val;
        });
        return [elem[0], ...eventVals];
    });
    const result = { result_group: [{ ...eventData.result_group[0], rows }] }
    return result;
}

export const calculateActiveUsersData = (userData, sessionData) => {
    const rows = userData.result_group[0].rows.map((elem) => {
        const eventVals = elem.slice(1).map((e) => {
            if (!e) return e;
            const e_val = sessionData.result_group[0].rows[0] / e;
            return e_val % 1 !== 0 ? parseFloat(e_val.toFixed(2)) : e_val;
        });
        return [elem[0], ...eventVals];
    });
    const result = { result_group: [{ ...userData.result_group[0], rows }] }
    console.log(result);
    return result;
}