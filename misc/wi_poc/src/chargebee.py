def get_chargebee_events(augment = True):
    """
    Gets a set of base events that could be used to select base users.

    Attributes
    ----------
    augment : bool
        whether or not to augment the base events with a $-prefix
        and/or a /(forward-slash)-suffix
    """
    chargebee_events = ["www.chargebee.com/pricing",
                        "www.chargebee.com/trial-signup/",
                        "www.chargebee.com/schedule-a-demo/",
                        "www.chargebee.com/subscription-management/",
                        "www.chargebee.com/recurring-billing-invoicing/",
                        "www.chargebee.com/recurring-payments/",
                        "www.chargebee.com/saas-accounting-and-taxes/",
                        "www.chargebee.com/saas-reporting/",
                        "www.chargebee.com/integrations/",
                        "www.chargebee.com/payment-gateways/",
                        "www.chargebee.com/for-education/",
                        "www.chargebee.com/for-self-service-subscription-business",
                        "www.chargebee.com/for-sales-driven-saas/",
                        "www.chargebee.com/for-subscription-finance-operations/",
                        "www.chargebee.com/enterprise-subscription-billing/",
                        "www.chargebee.com/customers",
                        "form_submitted"]
    if augment:
        chargebee_events_aug = ["$"+x for x in chargebee_events] + chargebee_events
        chargebee_events = [x[:-1] if x.endswith('/') else (x+'/') for x in chargebee_events_aug] + chargebee_events_aug
    chargebee_events = set(chargebee_events)
    return chargebee_events
