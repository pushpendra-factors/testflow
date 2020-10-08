import csv


class CsvUtil:

    @staticmethod
    def csv_to_dict_list(headers, csv_list):
        resp_rows = []

        rows = csv.reader(csv_list)
        for row in rows:
            resp = {}
            i = 0

            for col in row:
                col_striped = col.strip()
                if col_striped != '--':
                    resp[headers[i]] = col_striped
                i = i + 1
            
            if len(resp) > 0:
                resp_rows.append(resp)

        return resp_rows
