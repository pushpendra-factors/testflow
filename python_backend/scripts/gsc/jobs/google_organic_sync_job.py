from .reports_fetch_job import ReportsFetch


# Note: If the number of custom paths exceed 7 in the subClasses. Move it to strategic pattern.
class GetSearchConsoleDataJob(ReportsFetch):
    DIMENSIONS = ["query", "page", "country", "device"]

    def __init__(self, next_info):
        super().__init__(next_info)

    # using transform to dedup.
    def transform_entities(self, rows):
        transformed_rows = []
        for row in rows:
            transformed_row = {}
            for key in row:
                if key == "keys":
                    for index in range(len(self.DIMENSIONS)):
                        transformed_row[self.DIMENSIONS[index]] = row["keys"][index]
                else:
                    transformed_row[key] = row[key]
            transformed_rows.append(transformed_row)
        
        return transformed_rows