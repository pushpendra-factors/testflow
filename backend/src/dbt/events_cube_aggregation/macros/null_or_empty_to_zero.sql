
{% macro null_or_empty_to_zero(column_name) %}
    CASE WHEN {{ column_name }} IS NULL THEN 0
    ELSE {{ column_name }} END
{% endmacro %}
