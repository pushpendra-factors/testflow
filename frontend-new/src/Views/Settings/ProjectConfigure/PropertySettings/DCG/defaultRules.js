const defaultRules =
[
    {
       "channel":"Paid Search",
       "conditions":[
          {
             "value":"$none",
             "property":"$gclid",
             "condition":"notEqual",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Paid Search",
       "conditions":[
          {
             "value":"google",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"bing",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"adwords",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"youtube",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"paid",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"cpc",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"ppc",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"adwords",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"display",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"cpm",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          }
       ]
    },
    {
       "channel":"Paid Search",
       "conditions":[
          {
             "value":"google.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"AND"
          },
          {
             "value":"bing.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"duckduckgo.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"yahoo.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"yandex.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"baidu.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"$none",
             "property":"$campaign",
             "condition":"notEqual",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Paid Social",
       "conditions":[
          {
             "value":"$none",
             "property":"$fbclid",
             "condition":"notEqual",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Paid Social",
       "conditions":[
          {
             "value":"facebook",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"fb",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"linkedin",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"twitter",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"quora",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"pinterest",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"snapchat",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"instagram",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"paid",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"cpc",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"ppc",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"cpm",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          }
       ]
    },
    {
       "channel":"Paid Social",
       "conditions":[
          {
             "value":"paidsocial",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Paid Social",
       "conditions":[
          {
             "value":"paidsocial",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Paid Social",
       "conditions":[
          {
             "value":"paid",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"cpc",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"ppc",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"cpm",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"OR"
          },
          {
             "value":"facebook.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"AND"
          },
          {
             "value":"linkedin.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"quora.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"pinterest.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"twitter.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"snapchat.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"instagram.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          }
       ]
    },
    {
       "channel":"Organic Social",
       "conditions":[
          {
             "value":"$none",
             "property":"$fbclid",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"paid",
             "property":"$medium",
             "condition":"notEqual",
             "logical_operator":"AND"
          },
          {
             "value":"cpc",
             "property":"$medium",
             "condition":"notEqual",
             "logical_operator":"OR"
          },
          {
             "value":"ppc",
             "property":"$medium",
             "condition":"notEqual",
             "logical_operator":"OR"
          },
          {
             "value":"cpm",
             "property":"$medium",
             "condition":"notEqual",
             "logical_operator":"OR"
          },
          {
             "value":"facebook.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"AND"
          },
          {
             "value":"linkedin.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"quora.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"pinterest.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"twitter.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"snapchat.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"youtube.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"instagram.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          }
       ]
    },
    {
       "channel":"Organic Search",
       "conditions":[
          {
             "value":"$none",
             "property":"$gclid",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"$none",
             "property":"$fbclid",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"$none",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"$none",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"$none",
             "property":"$campaign",
             "condition":"equals",
             "logical_operator":"AND"
          },
          {
             "value":"google.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"AND"
          },
          {
             "value":"bing.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"duckduckgo.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"yahoo.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"yandex.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          },
          {
             "value":"baidu.",
             "property":"$initial_referrer_domain",
             "condition":"contains",
             "logical_operator":"OR"
          }
       ]
    },
    {
       "channel":"Email",
       "conditions":[
          {
             "value":"email",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Email",
       "conditions":[
          {
             "value":"email",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Affiliate",
       "conditions":[
          {
             "value":"affiliate",
             "property":"$source",
             "condition":"equals",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Affiliate",
       "conditions":[
          {
             "value":"affiliate",
             "property":"$medium",
             "condition":"equals",
             "logical_operator":"AND"
          }
       ]
    },
    {
       "channel":"Other Campaigns",
       "conditions":[
          {
             "value":"$none",
             "property":"$campaign",
             "condition":"notEqual",
             "logical_operator":"AND"
          }
       ]
    },
    {
      "channel":"Internal",
      "conditions":[
         {
            "value":"$none",
            "property":"$initial_referrer_domain",
            "condition":"notEqual",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$initial_page_domain",
            "condition":"notEqual",
            "logical_operator":"AND"
         },
         {
            "value":"$initial_page_domain",
            "property":"$initial_referrer_domain",
            "condition":"equals",
            "logical_operator":"AND"
         }
      ]
    },
    {
       "channel":"Referral",
       "conditions":[
          {
             "value":"$none",
             "property":"$initial_referrer_domain",
             "condition":"notEqual",
             "logical_operator":"AND"
          }
       ]
    },
    {
      "channel":"Direct",
      "conditions":[
         {
            "value":"$none",
            "property":"$source",
            "condition":"equals",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$medium",
            "condition":"equals",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$initial_referrer",
            "condition":"equals",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$initial_referrer_domain",
            "condition":"equals",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$gclid",
            "condition":"equals",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$fbclid",
            "condition":"equals",
            "logical_operator":"AND"
         },
         {
            "value":"$none",
            "property":"$campaign",
            "condition":"equals",
            "logical_operator":"AND"
         }
      ]
   }
 ]
export default defaultRules

