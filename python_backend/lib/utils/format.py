
class FormatUtil:

    # adwords gives in format of %, <10%, >90%
    @staticmethod
    def get_numeric_from_percentage_string(string_with_percent):
        if string_with_percent is None:
            return 0.0

        string_without_percent = string_with_percent.replace("%", "").replace("<", "").replace(">", "").strip()
        if string_without_percent == "":
           return 0.0
        return float(string_without_percent)/100