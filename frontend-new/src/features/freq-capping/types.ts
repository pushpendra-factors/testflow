import { toNumber } from 'lodash';

export enum ComponentStates {
    EMPTY,
    LOADING,
    LIST,
    VIEW
}

export interface RuleQueryParams {
    rule_id: string;
}

export class AdvanceRuleFilters {
    filters: any[] = [];

    impression_threshold = 1000;

    click_threshold = 5;
}

export class FrequencyCap {
    id = '';

    project_id = 0;

    object_type = 'campaign';

    name = '';

    display_name = '';

    status = 'active';

    description = '';

    object_ids: string[] = [];

    granularity = 'monthly';

    impression_threshold = 1000;

    click_threshold = 5;

    is_advanced_rule_enabled = false;

    advanced_rule_type = 'account';

    advanced_rules: AdvanceRuleFilters[] = [];

    constructor(project_id: string) {
        this.project_id = toNumber(project_id);
    }
}
