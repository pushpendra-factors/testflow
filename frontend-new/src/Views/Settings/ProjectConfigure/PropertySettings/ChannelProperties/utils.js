export class FilterClass {
    name = '';
    property = '';
    condition = '';
    logical_operator = '';
    value = '';

    constructor(name, property, condition, logical_operator, value) {
        this.name = name;
        this.property = property;
        this.condition = condition;
        this.logical_operator = logical_operator;
        this.value = value;
    }

    setFilter(filterObj, combOper) {
        this.condition = filterObj.operator;
        this.name = filterObj.prop.type;
        this.property = filterObj.prop.name;
        this.value = filterObj.values;
        this.logical_operator = combOper;
    }

    getFilter() {
        return {
            name: this.name,
            property: this.property,
            condition: this.condition,
            logical_operator: this.logical_operator,
            value: this.value
        }
    }
}
export class SmartPropertyClass  {
    id = ''; 
    name = '';
    description = '';
    type_alias = '';
    rules = [];

    constructor(id, name, description, type, rules) {
        this.id = id; 
        this.name = name;
        this.description = description;
        this.type_alias = type;
        this.rules = rules;
    }
}

export class PropertyRule {
    value = "";
    source = "";
    filters = [];

    constructor(value, source, filters) {
        this.value = value;
        this.source = source;
        this.filters = filters;
    }

}