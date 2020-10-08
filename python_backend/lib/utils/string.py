import re


class StringUtil:

    @staticmethod
    def first_letter_to_lower(s):
        if len(s) == 0:
            return ''

        f = s[0].lower()
        if len(s) == 1:
            return f

        return f + s[1:]

    @staticmethod
    def is_valid_value_type(s):
        return isinstance(s, str) or isinstance(s, int) or isinstance(s, float) or isinstance(s, bool)

    @staticmethod
    def snake_to_pascal_case(fields):
        pascals = []
        for f in fields:
            p = ''.join(x.capitalize() or '_' for x in f.split('_'))
            pascals.append(p)

        return pascals

    @staticmethod
    def camel_case_to_snake_case(s):
        s1 = re.sub('(.)([A-Z][a-z]+)', r'\1_\2', s)
        return re.sub('([a-z0-9])([A-Z])', r'\1_\2', s1).lower()
