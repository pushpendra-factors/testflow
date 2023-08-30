//Array of Objects.
export const kpiItemsgroupedByCategoryProperty = (selGroup) => {
  if (!selGroup || !selGroup.properties) {
    return {};
  }
  return (
    selGroup?.properties?.reduce((result, kpiItem) => {
      const category = kpiItem.category;
      if (!category) {
        return result;
      }
      if (!result[category]) {
        result[category] = [];
      }
      const propertyLabel = kpiItem.display_name
        ? kpiItem.display_name
        : kpiItem.name;
      const propertyCategoryType =
        selGroup?.category == 'channels'
          ? kpiItem.object_type
          : kpiItem.entity
          ? kpiItem.entity
          : kpiItem.object_type;

      const propertyDataType = kpiItem.data_type;
      const propertyValueName = kpiItem.name;
      result[category].push([
        propertyLabel,
        propertyValueName,
        propertyDataType,
        propertyCategoryType
      ]);
      return result;
    }, {}) || {}
  );
};
