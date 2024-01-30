import sys

import tornado.web
from tornado.ioloop import IOLoop
from tornado.log import logging as log

import logging
import app
from app_config import AppConfig
from app_parser import AppParser
from lib.data_services.factors_data_service import FactorsDataService
from lib.data_services.sentry.sentry_data_service import SentryDataService
from routes import ROUTES
logging.basicConfig(level=logging.INFO)

if __name__ == "__main__":
    input_args, rem = AppParser(sys.argv[1::]).parse()
    AppConfig.build(input_args)
    app.CONFIG = AppConfig
    FactorsDataService.init(app.CONFIG.ADWORDS_APP.get_data_service_path())
    SentryDataService.init(app.CONFIG.SENTRY)

    application = tornado.web.Application(ROUTES)
    application.listen(int(input_args.port))

    log.warning("Listening on port %d..", int(input_args.port))
    IOLoop.instance().start()
    # application.setup()