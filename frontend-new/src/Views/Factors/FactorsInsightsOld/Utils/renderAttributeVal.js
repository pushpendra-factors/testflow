export const renderAttributeValue = (data) => {
    let attributeValue = '';
    if (_.isEmpty(data?.factors_insights_attribute[0]?.factors_attribute_value)) {
        let attributeBoundKey = '';
        if (data?.factors_insights_attribute[0]?.factors_attribute_use_bound == 'OnlyUpper') {
            attributeBoundKey = `<= ${data?.factors_insights_attribute[0]?.factors_attribute_upper_bound}`
        }
        if (data?.factors_insights_attribute[0]?.factors_attribute_use_bound == 'OnlyLower') {
            attributeBoundKey = `>= ${data?.factors_insights_attribute[0]?.factors_attribute_lower_bound}`
        }
        if (data?.factors_insights_attribute[0]?.factors_attribute_use_bound == 'Both') {
            attributeBoundKey = `>= ${data?.factors_insights_attribute[0]?.factors_attribute_lower_bound} and <= ${data?.factors_insights_attribute[0]?.factors_attribute_upper_bound}`
        }
        attributeValue = attributeBoundKey;
    }
    else {
        attributeValue = `= ${data?.factors_insights_attribute[0]?.factors_attribute_value}`;
    }
    return attributeValue
}