
import os
import sys

import  random
import  struct
import  socket
import  time
import datetime
import  struct

import gevent
from gevent import monkey
gevent.monkey.patch_all()
svr_ip="10.133.146.250"
svr_port=11111

writetimes=0
readtimes=0
prewrite=0
preread=0

def qps():
	while True:
		global writetimes
		global prewrite
		global readtimes
		global preread

		time.sleep(10)
		q = (writetimes-prewrite)/10	
		print "write qps="+str(q)+"\n"
		prewrite=writetimes
		
		q = (readtimes-preread)/10	
		print "read qps="+str(q)+"\n"
		preread=readtimes

def readthread(socket):
	global readtimes
	while True:
		recvdata=socket.recv(5000)
		readtimes = readtimes+1
		if len(recvdata)==0:
			print "closed by peer"

def clientthread():
	global writetimes
	#while True:
	sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
	sock.connect((svr_ip, svr_port))
	#print "client"
	
	gevent.spawn(readthread,sock)
	#sendstring = struct.pack("!2H2I4s",1025,0,0,4,'hhhh')
#	hehe = "hehe"
	s = "aaaaaaaaaaaaaaaaaaaa"
	sendstring = ""
	for i in range(0,2,1):
		sendstring = sendstring+s	


	print 'connect finished'
	while True:
		sock.send(sendstring)
		writetimes = writetimes+1

		
		
if __name__=="__main__":
	print "hello world"
	eventlist=[]
	for i in range(0,1,1):
		ev=gevent.spawn(clientthread)
		eventlist.append(ev)
		#print 'spawn'
	gevent.spawn(qps)
	#time.sleep(100000)
	print 'spawn finished'
	gevent.joinall(eventlist)
		
	
	
	
	
	
	
	
	
	
	
	
	
	
	
