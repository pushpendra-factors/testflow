export const generateData = (data) => {
    let result = data.map(elem=>{
        let values = Object.values(elem.data);
        return [elem.name, ...values];
    });
    return result;
}

export const generateColors = (data) => {
    let result = {};
    data.forEach(elem=>{
        result[elem.name] = elem.color;
    });
    return result;
}

export const generateCategories = (data) => {
    let cat_names = Object.keys(data[0].data)
    let result = cat_names.map(elem=>{
        return {
            name: elem,
            conversion_rate: data[data.length-1].data[elem]+"%"
        }
    });
    return result;
}