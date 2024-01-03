
{% macro null_or_empty_to_none(column_name) %}
    CASE WHEN {{ column_name }} IS NULL THEN '$none' WHEN {{ column_name }} = '' THEN '$none'
    ELSE {{ column_name }} END
{% endmacro %}
