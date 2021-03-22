import csv


class CsvUtil:

    @staticmethod
    def csv_to_dict_list(headers, rows):
        resp_rows = []

        for row in rows:
            resp = {}
            i = 0

            for col in row:
                col_striped = col.strip()
                if col_striped != "--":
                    resp[headers[i]] = col_striped
                i = i + 1
            
            if len(resp) > 0:
                resp_rows.append(resp)

        return resp_rows

    # Note: This can cause errors if the first record doesnt have all the fields. Requires file object.
    @staticmethod
    def write_map_to_file(array_of_maps, data_file):
        csv_writer = csv.writer(data_file)
        header_written = False

        for object in array_of_maps:
            if header_written == False:
                # Writing headers of CSV file.
                header = object.keys() 
                csv_writer.writerow(header)
                header_written = True
            csv_writer.writerow(object.values())
        data_file.close()

    @staticmethod
    def read_csv(file):
        reader = csv.reader(file)
        return [row for row in reader]

    @staticmethod
    def unmarshall(csv_string):
        if csv_string is None:
            return []
        lines = csv_string.splitlines()
        reader = csv.reader(lines)
        parsed_csv = list(reader)
        return parsed_csv