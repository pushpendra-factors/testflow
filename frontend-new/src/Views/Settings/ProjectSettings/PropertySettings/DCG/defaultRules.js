const defaultRules =
    [
        {
            "channel": "Direct",
            "conditions": [
                {
                    "value": "$none",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$initial_referrer",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$initial_referrer_domain",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$gclid",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$fbclid",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$campaign",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                }
            ]
        },
        {
            "channel": "Paid Search",
            "conditions": [
                {
                    "value": "$none",
                    "property": "$gclid",
                    "condition": "NOT EQUAL",
                    "logical_operator": "AND"
                }
            ]
        },
        {
            "channel": "Paid Search",
            "conditions": [
                {
                    "value": "google",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "bing",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "adwords",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "youtube",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "paid",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "cpc",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "ppc",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "adwords",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "display",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "cpm",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                }
            ]
        },
        {
            "channel": "Paid Search",
            "conditions": [
                {
                    "value": "google.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "AND"
                },
                {
                    "value": "bing.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "duckduckgo.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "yahoo.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "yandex.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "baidu.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "$none",
                    "property": "$campaign",
                    "condition": "NOT EQUAL",
                    "logical_operator": "AND"
                }
            ]
        },
        {
            "channel": "Paid Social",
            "conditions": [
                {
                    "value": "$none",
                    "property": "$fbclid",
                    "condition": "NOT EQUAL",
                    "logical_operator": "AND"
                }
            ]
        },
        {
            "channel": "Paid Social",
            "conditions": [
                {
                    "value": "facebook",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "fb",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "linkedin",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "twitter",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "quora",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "pinterest",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "snapchat",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "paid",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "cpc",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "ppc",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "cpm",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                }
            ]
        },
        {
            "channel": "Paid Social",
            "conditions": [
                {
                    "value": "paidsocial",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                }
            ]
        },
        {
            "channel": "Paid Social",
            "conditions": [
                {
                    "value": "paidsocial",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                }
            ]
        },
        {
            "channel": "Paid Social",
            "conditions": [
                {
                    "value": "paid",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "cpc",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "ppc",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "cpm",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "OR"
                },
                {
                    "value": "facebook.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "AND"
                },
                {
                    "value": "linkedin.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "quora.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "pinterest.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "twitter.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "snapchat.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                }
            ]
        },
        {
            "channel": "Organic Social",
            "conditions": [
                {
                    "value": "$none",
                    "property": "$fbclid",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "paid",
                    "property": "$medium",
                    "condition": "NOT EQUAL",
                    "logical_operator": "AND"
                },
                {
                    "value": "cpc",
                    "property": "$medium",
                    "condition": "NOT EQUAL",
                    "logical_operator": "OR"
                },
                {
                    "value": "ppc",
                    "property": "$medium",
                    "condition": "NOT EQUAL",
                    "logical_operator": "OR"
                },
                {
                    "value": "cpm",
                    "property": "$medium",
                    "condition": "NOT EQUAL",
                    "logical_operator": "OR"
                },
                {
                    "value": "facebook.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "AND"
                },
                {
                    "value": "linkedin.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "quora.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "pinterest.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "twitter.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "snapchat.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "youtube.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                }
            ]
        },
        {
            "channel": "Organic Search",
            "conditions": [
                {
                    "value": "$none",
                    "property": "$gclid",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$fbclid",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$source",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$medium",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "$none",
                    "property": "$campaign",
                    "condition": "EQUALS",
                    "logical_operator": "AND"
                },
                {
                    "value": "google.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "AND"
                },
                {
                    "value": "bing.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "duckduckgo.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "yahoo.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "yandex.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                },
                {
                    "value": "baidu.",
                    "property": "$initial_referrer_domain",
                    "condition": "CONTAINS",
                    "logical_operator": "OR"
                }
            ]
        },
            {
                "channel": "Email",
                "conditions": [
                    {
                        "value": "email",
                        "property": "$source",
                        "condition": "EQUALS",
                        "logical_operator": "AND"
                    }
                ]
            },
            {
                "channel": "Email",
                "conditions": [
                    {
                        "value": "email",
                        "property": "$medium",
                        "condition": "EQUALS",
                        "logical_operator": "AND"
                    }
                ]
            },
            {
                "channel": "Affiliate",
                "conditions": [
                    {
                        "value": "affiliate",
                        "property": "$source",
                        "condition": "EQUALS",
                        "logical_operator": "AND"
                    }
                ]
            },
            {
                "channel": "Affiliate",
                "conditions": [
                    {
                        "value": "affiliate",
                        "property": "$medium",
                        "condition": "EQUALS",
                        "logical_operator": "AND"
                    }
                ]
            },
        {
            "channel": "Referral",
            "conditions": [
                {
                    "value": "$none",
                    "property": "$initial_referrer_domain",
                    "condition": "NOT EQUAL",
                    "logical_operator": "AND"
                }
            ]
        }
    ]

export default defaultRules