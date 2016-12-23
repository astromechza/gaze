#!/usr/bin/env python

import BaseHTTPServer
import logging
import json
import argparse

"""
example_python_receiver.py 

This script is an example HTTP server that you can run in order to test the 'web' behaviour type for 
gaze. 

Launch it as so:

```
$ ./example_python_receiver.py -port 8080
```

And then use `http://127.0.0.1:8080` as the url field for a web behaviour in your gaze config.
"""


log = logging.getLogger(__name__)


class ExampleHandler(BaseHTTPServer.BaseHTTPRequestHandler):

    def log_stuff(self):
        log.info("Incoming %s request on %s", self.command, self.path)
        for k, v in self.headers.items():
            log.info("Header '%s' -> '%s'", k, v)

    def read_all(self):
        return self.rfile.read(int(self.headers.getheader('Content-Length')))

    def do_POST(self):
        self.log_stuff()
        try:
            content = self.read_all().strip()
            log.info("Content: %s", json.dumps(json.loads(content), indent=2))
            self.send_response(204)
            self.end_headers()
        except Exception:
            log.exception("something happened")
            self.send_response(500)
            self.end_headers()
        finally:
            self.wfile.flush()

    def do_PUT(self):
        return self.do_POST()


def main():
    h = logging.StreamHandler()
    h.setLevel(logging.DEBUG)
    h.setFormatter(logging.Formatter("%(asctime)s : %(levelname)s : %(message)s"))
    logging.root.addHandler(h)
    logging.root.setLevel(logging.DEBUG)

    p = argparse.ArgumentParser()
    p.add_argument('-p', '--port', default=8080)
    args = p.parse_args()

    server_address = ('', args.port)
    httpd = BaseHTTPServer.HTTPServer(server_address, ExampleHandler)
    log.info("Starting example server at: %s...", server_address)

    try:
        httpd.serve_forever()
    except KeyboardInterrupt:
        pass
    httpd.server_close()

if __name__ == '__main__':
    main()
