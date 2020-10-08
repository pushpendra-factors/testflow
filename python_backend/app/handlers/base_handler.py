import tornado.web
# from raven.contrib.tornado import SentryMixin


class BaseHandler(tornado.web.RequestHandler):
    pass
# def write_error(self, status_code, **kwargs):
#     if status_code == 500:
#         exception = kwargs['exc_info'][1]
#         tb = kwargs['exc_info'][2]
#         stack = traceback.extract_tb(tb)
#         clean_stack = [i for i in stack if i[0][-6:] != 'gen.py' and i[0][-13:] != 'concurrent.py']
#         error_msg = '{}\n  Exception: {}'.format(''.join(traceback.format_list(clean_stack)), exception)
#
#         logging.error(error_msg)  # do something with your error...
#         self.captureException()
