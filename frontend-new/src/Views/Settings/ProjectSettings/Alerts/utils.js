export const getErrorMsg = (errorItem, string) => {
    let index = errorItem?.failed_at?.indexOf(string);
    let msg = errorItem?.details[index] ? errorItem?.details[index] : null
    return msg
}

export const SLACK = "Slack";
export const WEBHOOK = "WH";
export const MS_TEAMS = "Teams";

export const convertObjectToKeyValuePairArray = (obj = {}) => {
    return Object.keys(obj).map((key) => [key, obj[key]]);
}

export const getMsgPayloadMapping = (groupBy) => { 
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
        var obj = {}
        groupBy.map((item) => {
            obj[item.property] = "${Property Value}"
        })
        return obj
    }
    else return null
}

export  const getMsgPayloadMappingWebhook = (groupBy, matchEventName, dummyPayloadValue) => { 
    if (groupBy && groupBy.length && groupBy[0] && groupBy[0].property) {
        var arr = []
        groupBy.map((item) => { 
            let obj =  {
                'DisplayName':  matchEventName(item?.property),
                'PropValue': dummyPayloadValue[item?.property] ? dummyPayloadValue[item?.property] : '${Property Value}'
            }
            arr.push(obj)
        })
        return arr
    }
    else return null
}

export const dummyPayloadValue = {
    "$6Signal_name": "Acme Inc",
    "$company ": "Acme Inc",
    "$6Signal_domain": "acme.com",
    "$6Signal_annual_revenue ": "5,744,000",
    "$6Signal_revenue_range": "$5M - $10M",
    "$6Signal_employee_count": "75",
    "$6Signal_employee_range": "50 - 99",
    "$6Signal_industry": "Software and Technology",
    "$6Signal_address": "92 Pleasant Street SE",
    "$6Signal_naics": "33324",
    "$6Signal_naics_description": "Industrial Machinery Manufacturing",
    "$6Signal_sic": "7372",
    "$6Signal_sic_description": "Prepackaged Software",
    "$6Signal_phone": "49 2734 57520",
    "$6Signal_city": "Boston",
    "$6Signal_country": "United States",
    "$6Signal_state": "Massachusetts",
    "$6Signal_region ": "Northern America",
    "$6Signal_country_iso_code": "US",
    "$6Signal_zip": "2210",
    "$city": "Boston",
    "$country": "United States",
    "$postal_code": "33626",
    "$continent": "Northern America",
    "$region": "Northern America",
    "$email": "jhondoe@acme.com",
    "$phone": "49 2734 57520",
    "$user_id": "jhondoe@acme.com",
    "$name": "John Doe",
    "$first_name": "John",
    "$last_name": "Doe",
    "$joinTime": "31-Dec-2023",
    "$day_of_first_event": "Sunday",
    "$hour_of_first_event": "6",
    "$browser": "Chrome",
    "$browserVersion | $browser_version": "12.51",
    "$platform": "Web",
    "$device_brand": "Apple",
    "$screen_width": "1,366",
    "$screen_height": "768",
    "$os": "iOS",
    "$os_version": "17.1.2",
    "$device_family": "Phones",
    "$device_manufacturer": "Apple",
    "$device_model": "Iphone",
    "$device_name": "Iphone",
    "$device_type": "Desktop",
    "$initial_creative": "meme_V1",
    "$initial_fbclid": "cJMIi3lNO46fJX928Sk",
    "$initial_gclid": "cJMIi3lNO46fJX928Sk",
    "$initial_keyword": "Who is visiting",
    "$initial_keyword_match_type": "B",
    "$initial_referrer": "google.com",
    "$initial_referrer_url": "google.com/images",
    "$initial_referrer_domain ": "google.com",
    "$initial_source": "google",
    "$initial_medium": "paid",
    "$initial_term": "Who is visiting",
    "$initial_adgroup": "competitor_US",
    "$initial_adgroup_id": "128912",
    "$initial_campaign": "deanon_generic",
    "$initial_campaign_id": "281925",
    "$initial_content": "blog",
    "$initial_channel": "Paid search",
    "$latest_creative": "meme_V1",
    "$latest_fbclid": "cJMIi3lNO46fJX928Sk",
    "$latest_gclid": "cJMIi3lNO46fJX928Sk",
    "$latest_keyword": "Who is visiting",
    "$latest_keyword_match_type": "B",
    "$latest_referrer": "google.com",
    "$latest_referrer_url": "google.com/images",
    "$latest_referrer_domain ": "google.com",
    "$latest_source": "google",
    "$latest_medium": "paid",
    "$latest_term": "Who is visiting",
    "$latest_adgroup": "competitor_US",
    "$latest_adgroup_id": "128912",
    "$latest_campaign": "deanon_generic",
    "$latest_campaign_id": "281925",
    "$latest_content": "blog",
    "$latest_channel": "Paid search",
    "$creative": "meme_V1",
    "$fbclid": "cJMIi3lNO46fJX928Sk",
    "$gclid": "cJMIi3lNO46fJX928Sk",
    "$keyword": "Who is visiting",
    "$keyword_match_type": "B",
    "$referrer": "google.com",
    "$referrer_url": "google.com/images",
    "$referrer_domain ": "google.com",
    "$source": "google",
    "$medium": "paid",
    "$term": "Who is visiting",
    "$adgroup": "competitor_US",
    "$adgroup_id": "128912",
    "$campaign": "deanon_generic",
    "$campaign_id": "281925",
    "$content": "blog",
    "$channel": "Paid search",
    "$initial_page_domain": "www.acme.com",
    "$initial_page_url": "www.acme.com/pricing",
    "$initial_page_raw_url": "https://acme.com/pricing?utm_source=google",
    "$initial_page_load_time": "3 sec",
    "$initial_page_scroll_percent": "78",
    "$initial_page_spent_time": "23 sec",
    "$session_latest_page_url": "www.acme.com",
    "$session_latest_page_raw_url": "www.acme.com/pricing",
    "$session_spent_time": "29 sec",
    "$session_time": "29 sec",
    "$page_count": "4",
    "$is_first_session": "TRUE",
    "$session_count": "6",
    "$timestamp": "24-Jan-2024",
    "$day_of_week": "Saturday",
    "$hour_of_day": "6",
    "$browserVersion": "12.21",
    "$page_domain": "www.acme.com",
    "$page_url": "www.acme.com/pricing",
    "$page_raw_url": "https://acme.com/pricing?utm_source=google",
    "$page_title": "Acme Inc. Pricing",
    "$page_scroll_percent": "78",
    "$page_load_time": "3 sec",
    "$is_page_view": "true",
    "$latest_page_domain": "www.acme.com",
    "$latest_page_url": "www.acme.com/pricing",
    "$latest_page_raw_url": "https://acme.com/pricing?utm_source=google",
    "$latest_page_load_time": "3 sec",
    "$latest_page_scroll_percent": "78",
    "$latest_page_spent_time": "29 sec",
    "$domain_name": "acme.com",
    "page_url": "www.acme.com/pricing",
    "$li_domain": "https://acme.com",
    "$li_localized_name": "Acme Inc",
    "$li_org_id": "14534978",
    "$li_preferred_country": "US",
    "$li_vanity_name": "acme-inc",
    "$g2_product_ids": "factorsai",
    "$g2_tag": "products.reviews",
    "$g2_visitor_city": "Boston",
    "$g2_visitor_country": "United States",
    "$g2_visitor_state": "Massachusetts",
    "$g2_company_id": "28191",
    "$g2_country": "United States",
    "$g2_domain:": "acme.com",
    "$g2_employees": "415",
    "$g2_employees_range": "251 - 500",
    "$g2_legal_name": "Acme Inc",
    "$g2_name": "Acme",
    "Session referrer domain" : "google.com",
    "Session browser version" : "12.51",
    "Company" : "Acme Inc",
    "Company region" : "Northern America",
    "Company annual revenue" : "5,744,000",

}
const AlertTemplateToTheme = {
    'Account_Executives': {
        icon: 'UserTie',
        backgroundColor: '#FFF7E6',
        color: '#D46B08',
    },
    'SDRs': {
        icon: 'Headset',
        backgroundColor: '#E6FFFB',
        color: '#08979C',
    },
    'Marketing': {
        icon: 'SponsorShip',
        backgroundColor: '#F9F0FF',
        color: '#722ED1',
    },
    'Customer_Success': {
        icon: 'Handshake',
        backgroundColor: '#F0F5FF',
        color: '#2F54EB',
    }
}
export const getAlertTemplatesTransformation = (data)=>{
    return data
            .filter(e=>!e.is_deleted)
            .map((each)=>{
                
                const {alert_name, alert_message, title, description, payload_props, prepopulate} = each?.alert;
                const {question, required_integrations} = each?.template_constants;
                let {categories} = each?.template_constants;
                // This might never happen if we maintain the structure of templates
                if(categories && Array.isArray(categories)){
                    categories = categories.map((e)=>e.replace('_',' '))
                }
                let {icon, color, backgroundColor} = AlertTemplateToTheme[ (categories && categories[0]) || 'SDRs'] || AlertTemplateToTheme['Account_Executives']
                return {
                    ...each, 
                    alert_name,
                    alert_message,
                    title,
                    description, 
                    payload_props,
                    prepopulate,
                    question,
                    required_integrations,
                    categories,
                    icon, color, backgroundColor
                }
            });
}