import sys
import json
import traceback

"""
Any python program that is called by a go module should call setup_error_handler()
This will ensure that any error that is caught in python will become a json message
on stderr.
"""

def setup_error_handler():
    def exception_handler(exc_type, exc_value, exc_traceback):
        error_info = {
            "status": 500,
            "error": str(exc_value),
            "message": exc_type.__name__,
            "trace": ''.join(traceback.format_exception(exc_type, exc_value, exc_traceback))
        }
        print(json.dumps(error_info), file=sys.stderr, flush=True)
        sys.exit(1)
    sys.excepthook = exception_handler