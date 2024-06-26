const AlertTemplateToTheme = {
  Account_Executives: {
    icon: 'UserTie',
    backgroundColor: '#FFF7E6',
    color: '#D46B08'
  },
  SDRs: {
    icon: 'Headset',
    backgroundColor: '#E6FFFB',
    color: '#08979C'
  },
  Marketing: {
    icon: 'SponsorShip',
    backgroundColor: '#F9F0FF',
    color: '#722ED1'
  },
  Customer_Success: {
    icon: 'Handshake',
    backgroundColor: '#F0F5FF',
    color: '#2F54EB'
  }
};

export const TemplateIDs = {
  FACTORS_HUBSPOT_COMPANY:
    window.document.domain === 'app.factors.ai' ? 1000000 : 4000000,
  FACTORS_APOLLO_HUBSPOT_CONTACTS:
    window.document.domain === 'app.factors.ai' ? 1000001 : 4000002,
  FACTORS_SALESFORCE_COMPANY:
    window.document.domain === 'app.factors.ai' ? 1000002 : 4000003,
  FACTORS_APOLLO_SALESFORCE_CONTACTS:
    window.document.domain === 'app.factors.ai' ? 1000003 : 4000004,
  FACTORS_LINKEDIN_CAPI:
    window.document.domain === 'app.factors.ai' ? 1000004 : 4000005
};

export const getAlertTemplatesTransformation = (data) => {
  return data
    .filter((e) => !e.is_deleted)
    .map((each) => {
      const {
        alert_name,
        alert_message,
        title,
        description,
        payload_props,
        prepopulate
      } = each?.alert;
      // const {question, required_integrations} = each?.template_constants;
      let { categories } = each?.template_constants;
      // This might never happen if we maintain the structure of templates
      if (categories && Array.isArray(categories)) {
        categories = categories.map((e) => e.replace('_', ' '));
      }
      let { icon, color, backgroundColor } =
        AlertTemplateToTheme[(categories && categories[0]) || 'SDRs'] ||
        AlertTemplateToTheme['Account_Executives'];
      return {
        ...each,
        alert_name,
        alert_message,
        title,
        description,
        payload_props,
        prepopulate,
        // question,
        // required_integrations,
        categories,
        icon,
        color,
        backgroundColor
      };
    });
};


export const templateThumbnailImage = (templateID) =>{
  switch(templateID){
    case 4000000: //hubspot companies
      return 'hubspot-company-thumbnail.png';
      case 4000002: //hubspot contacts
        return 'hubspot-company-thumbnail.png';
        case 4000003: //salesforce accounts
          return 'salesforce-company-thumbnail.png';
          case 4000004: //salesforce contacts
            return 'salesforce-company-thumbnail.png';
            case 4000005: //Linkedin CAPI contacts
              return 'linkedin-capi-thumbnail.png';
              default: return 'hubspot-company-thumbnail.png';
  }
}