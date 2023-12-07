const Company_identification = "Company identification";

export const  removeDuplicateAndEmptyKeys = (obj) => {
    const uniqueKeys = {}; 
    for (const key in obj) {
      //remove duplicate keys
      if (!uniqueKeys.hasOwnProperty(key)) { 
          //remove empty keys
          if(!_.isEmpty(obj[key])){
          uniqueKeys[key] = obj[key]; 
          } 
      }     
    }
    //remove 'Company identification' from user properties since the same duplicate properties is available in $6_signal
    if(uniqueKeys.hasOwnProperty('user')){
      if(uniqueKeys?.user?.hasOwnProperty(Company_identification)){
        delete uniqueKeys?.user[Company_identification];
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