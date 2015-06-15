import socket
import sys
sys.path.append('./gen-py/')

from hello import Hello
from thrift import Thrift
from thrift.transport import TSocket
from thrift.transport import TTransport
from thrift.protocol import TBinaryProtocol

try:
    transport = TSocket.TSocket('localhost', 19090)
    transport = TTransport.TBufferedTransport(transport)
    protocol = TBinaryProtocol.TBinaryProtocol(transport)
    client = Hello.Client(protocol)
    transport.open()

    print "client - helloString"
    msg = client.helloString("lalalalalla")
    print "server - " + msg


    transport.close()

except Thrift.TException, ex:
    print "%s" % (ex.message)