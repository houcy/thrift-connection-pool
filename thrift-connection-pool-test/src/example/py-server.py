# -*- coding: UTF-8 -*-

#!/usr/bin/env python
 
import socket
import sys
sys.path.append('./gen-py/')
 
from hello import Hello
from hello.ttypes import *
 
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol
from thrift.server import TServer
 
class HelloHandler:
    def helloString(self, msg):
        ret = "helloString Received: " + msg
        print ret
        return msg

 
handler = HelloHandler()
processor = Hello.Processor(handler)
transport = TSocket.TServerSocket("127.0.0.1", 19090)
tfactory = TTransport.TBufferedTransportFactory()
pfactory = TBinaryProtocol.TBinaryProtocolFactory()
 
server = TServer.TSimpleServer(processor, transport, tfactory, pfactory)
 
print "Starting thrift server in python..."
server.serve()