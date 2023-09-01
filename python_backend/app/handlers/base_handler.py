import tornado.web
# from raven.contrib.tornado import SentryMixin
from lib.adwords.cors import Cors


class BaseHandler(tornado.web.RequestHandler):

    def set_default_headers(self):
        request_headers = self.request.headers
        origin = request_headers.get('Origin')
        self.set_header("Access-Control-Allow-Headers", "x-requested-with, Origin, Content-Type")
        self.set_header("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
        self.set_header("Access-Control-Allow-Credentials", "true")
        allowed_origin = Cors.get_cors_allowed_origin(origin)
        if allowed_origin != None:
            self.set_header("Access-Control-Allow-Origin", allowed_origin)
        self.clear_header("Server")

# def write_error(self, status_code, **kwargs):
#     if status_code == 500:
#         exception = kwargs["exc_info"][1]
#         tb = kwargs["exc_info"][2]
#         stack = traceback.extract_tb(tb)
#         clean_stack = [i for i in stack if i[0][-6:] != "gen.py" and i[0][-13:] != "concurrent.py"]
#         error_msg = "{}\n  Exception: {}".format("".join(traceback.format_list(clean_stack)), exception)
#
#         logging.error(error_msg)  # do something with your error...
#         self.captureException()
