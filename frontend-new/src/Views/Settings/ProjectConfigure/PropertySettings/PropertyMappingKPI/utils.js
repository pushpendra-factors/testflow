import { get, map } from 'lodash';

export const EMPTY_ARRAY = [];


export const getPropertyDisplayName = (configArr, categoryName, propertyName ) => { 
  let mainItem = configArr?.find((item)=>{
     if(item?.display_category == categoryName){
       return item
     }
  }) 
  let SubItem = mainItem?.properties?.find((item)=>{
     if(item?.name == propertyName){
       return item    
      }
  }) 
  return SubItem?.display_name
}


export const getPropertiesDetails = (queries) => {
 const propertiesDetails =  queries?.map((item)=>{
  return {
          "ca": item?.category,
          "dc": item?.group, 
          "name": item?.name,
          "da_ty": item?.objType,
          "en": item?.entity, 
        }
})
return propertiesDetails
}

  export const getNormalizedKpi = ({ kpi }) => {
    const propertiesValues = kpi?.properties.map((props) => {
      return [
        props?.display_name,
        props?.name,
        props?.data_type ? props?.data_type : "",
        props?.object_type ? props?.object_type : '',
        props?.entity,
      ];
    });
    return {
      label: get(kpi, 'display_category'),
      group: get(kpi, 'display_category'),
      category: get(kpi, 'category'),
      icon: 'custom_events',
      values: propertiesValues
    };
  };



