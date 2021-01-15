export const getFirstDayOfLastWeek = () => {
    const d = new Date();
    const first = d.getDate() - d.getDay() - 7;
    return new Date(d.setDate(first));
}

export const getLastDayOfLastWeek = () => {
    const d = new Date();
    const last = d.getDate() - d.getDay() - 1;
    return new Date(d.setDate(last));
}

export const getFirstDayOfLastMonth = () => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth() - 1, 1);
}
  
export const getLastDayOfLastMonth = () => {
    const d = new Date();
    return new Date(d.getFullYear(), d.getMonth(), 0);
}