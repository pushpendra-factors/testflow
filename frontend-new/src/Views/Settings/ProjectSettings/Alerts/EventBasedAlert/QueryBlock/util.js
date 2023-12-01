
export const  removeDuplicateAndEmptyKeys = (obj) => {
    const uniqueKeys = {};
    //blacklisted groups
    let removeGroupList = ["Company identification"];
    for (const key in obj) {
      //remove duplicate keys
      if (!uniqueKeys.hasOwnProperty(key)) {
        //remove blacklisted keys
        if(!key.includes(removeGroupList)){
          //remove empty keys
          if(!_.isEmpty(obj[key])){
          uniqueKeys[key] = obj[key]; 
          }
        }
        
      }     
    }
    return uniqueKeys;
  }

  export const blackListedCategories = [
    'Hubspot Contacts',
    'Salesforce Users',
    'LeadSquared Person',
    'Marketo Person',
    'Hubspot Companies',
    'Hubspot Deals',
    'Salesforce Accounts',
    'Salesforce Opportunities',
  ]