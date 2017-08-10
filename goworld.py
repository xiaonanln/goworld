#!/usr/bin/env python
# -*- coding: utf8 -*-

import os
import sys
import signal
import psutil
import getopt
import ConfigParser
import time

DISPATCHER_EXE = "dispatcher"
GATE_EXE = "gate"

if os.name == 'nt':
	DISPATCHER_EXE = DISPATCHER_EXE + ".exe"
	GATE_EXE = GATE_EXE + ".exe"

goworldPath = ''
gateids = []
gameids = []
gameName = ''
gamePath = ''
loglevel = "info"

def main():
	opts, args = getopt.getopt(sys.argv[1:], "", ["log="])
	global loglevel
	for opt, val in opts:
		if opt == "--log":
			loglevel = val

	if len(args) == 0:
		showUsage()
		exit(1)

	cmd = args[0].lower()
	verifyExecutionEnv(cmd)

	config = ConfigParser.SafeConfigParser()
	config.read("goworld.ini")
	analyzeConfig(config)

	cmd = args[0].lower()
	if cmd == 'status':
		showStatus()
	elif cmd == "start":
		global gameName
		try:
			gameName = args[1]
		except:
			showUsage()
			exit(1)

		global gamePath
		gamePath = detectGamePath(gameName)
		if not os.path.exists(gamePath):
			print >>sys.stderr, "! %s is not found, goworld.py build %s first" % (gamePath, gameName)
			exit(2)

		startServer()
	elif cmd == 'stop':
		stopServer()
	elif cmd == 'build':
		buildTargets = args[1:]
		if not buildTargets: buildTargets = ['engine']

		for buildTarget in buildTargets:
			build(buildTarget)
	else:
		print >>sys.stderr, "invalid command: %s" % cmd
		showUsage()
		exit(1)

def verifyExecutionEnv(cmd):
	global goworldPath
	goworldPath = os.getcwd()
	print >>sys.stderr, '> Detect goworld path:', goworldPath
	dir = os.path.basename(goworldPath)
	if dir != 'goworld':
		print >>sys.stderr, "must run in goworld directory!"
		exit(2)

	if cmd != 'build':
		if not os.path.exists(getDispatcherExe()):
			print >>sys.stderr, "%s is not found, goworld.py build engine first" % getDispatcherExe()
			exit(2)

		if not os.path.exists(getGateExe()):
			print >> sys.stderr, "%s is not found, goworld.py build engine first" % getGateExe()
			exit(2)

def detectGamePath(gameId, needExe=True):
	dir, gameName = os.path.split(gameId)
	if dir == '':
		dirs = [f for f in os.listdir(".") if os.path.isdir(f) and f not in ('components', 'engine')]
	else:
		dirs = [dir]

	for dir in dirs:
		gameDir = os.path.join(dir, gameName)
		if not os.path.isdir(gameDir):
			continue

		gamePath = os.path.join(gameDir, gameName)
		if os.name == 'nt':
			gamePath += ".exe"

		# if not os.path.exists(gamePath):
		# 	print >>sys.stderr, "! %s is not found, use goworld.py build first" % gamePath
		# 	exit(2)

		return gamePath

	# game not found
	print >>sys.stderr, "! game %s is not found, wrong name?" % gameId
	exit(2)

def showUsage():
	print >>sys.stderr, """Usage:
	goworld.py status - show server status
	goworld.py build engine|<game-name> - build server engine / game
	goworld.py start <game-name> - start game server
	goworld.py stop - stop game server
	"""

def build(target):
	if target == 'dispatcher':
		buildDispatcher()
	elif target == 'gate':
		buildGate()
	elif target == 'engine':
		buildEngine()
	else:
		buildGame(target)

def buildEngine():
	buildDispatcher()
	buildGate()

def buildDispatcher():
	print >>sys.stderr, '> building dispatcher ...',
	os.system('cd "%s" && go build' % os.path.join("components", "dispatcher"))
	print >>sys.stderr, 'OK'

def buildGate():
	print >>sys.stderr, '> building gate ...',
	os.system('cd "%s" && go build' % os.path.join("components", "gate"))
	print >>sys.stderr, 'OK'

def buildGame(gameId):
	gamePath = detectGamePath(gameId)
	gameDir = os.path.dirname(gamePath)
	print >>sys.stderr, '> building %s ...' % gameDir,
	os.system('cd "%s" && go build' % gameDir)
	print >>sys.stderr, 'OK'

def visitProcs():
	dispatcherProcs = []
	gateProcs = []
	gameProcs = []
	for p in psutil.process_iter():
		try:
			if isDispatcherProcess(p):
				dispatcherProcs.append(p)
			elif isGateProcess(p):
				gateProcs.append(p)
			elif isGameProcess(p):
				gameProcs.append(p)
		except psutil.AccessDenied:
			continue

	return dispatcherProcs, gateProcs, gameProcs

def showStatus():
	_showStatus(1, len(gateids), len(gameids))

def _showStatus(expectDispatcherCount, expectGateCount, expectGameCount):
	dispatcherProcs, gateProcs, gameProcs = visitProcs()
	print >>sys.stderr, "%-16s expect %d found %d %s" % ("dispatcher", expectDispatcherCount, len(dispatcherProcs), "GOOD" if len(dispatcherProcs) == expectDispatcherCount else "BAD!")
	print >>sys.stderr, "%-16s expect %d found %d %s" % ("gate", expectGateCount, len(gateProcs), "GOOD" if expectGateCount == len(gateProcs) else "BAD!")
	print >>sys.stderr, "%-16s expect %d found %d %s" % ("game", expectGameCount, len(gameProcs), "GOOD" if expectGameCount == len(gameProcs) else "BAD!")

def startServer():
	dispatcherProcs, gateProcs, gameProcs = visitProcs()
	if dispatcherProcs or gateProcs or gameProcs:
		print >>sys.stderr, "goworld is already running ..."
		_showStatus(1, len(gateids), len(gameids))
		exit(2)

	# now the system is clear, start server processes ...
	print >>sys.stderr, "Start dispatcher ...",
	dispatcherProc = psutil.Popen([getDispatcherExe()])
	print >>sys.stderr, dispatcherProc.status()

	for gateid in gateids:
		print >>sys.stderr, "Start gate%d ..." % gateid,
		gateProc = psutil.Popen([getGateExe(), "-gid=%d" % gateid, "-log", loglevel])
		print >>sys.stderr, gateProc.status()

	for gameid in gameids:
		print >> sys.stderr, "Start game%d ..." % gameid,
		gameProc = psutil.Popen([getGameExe(), "-gid=%d" % gameid, "-log", loglevel])
		print >> sys.stderr, gameProc.status()

	_showStatus(1, len(gateids), len(gameids))

def stopServer():
	dispatcherProcs, gateProcs, gameProcs = visitProcs()
	if not dispatcherProcs and not gateProcs and not gameProcs:
		_showStatus(1, len(gateids), len(gameids))
		print >>sys.stderr, "! goworld is not running"
		exit(2)

	# Close gates first to shutdown clients
	for proc in gateProcs:
		killProc(proc)

	print >>sys.stderr, "Waiting for gate processes to terminate ...",
	waitProcsToTerminate( isGateProcess )
	print >>sys.stderr, 'OK'

	for proc in gameProcs:
		proc.send_signal(signal.SIGTERM)

	print >>sys.stderr, "Waiting for game processes to terminate ...",
	waitProcsToTerminate(isGameProcess)
	print >>sys.stderr, 'OK'

	for proc in dispatcherProcs:
		killProc(proc)

	print >>sys.stderr, "Waiting for game processes to terminate ...",
	waitProcsToTerminate(isDispatcherProcess)
	print >>sys.stderr, 'OK'

	_showStatus(0, 0, 0)

def killProc(p):
	try:
		p.kill()
	except psutil.NoSuchProcess:
		pass

def waitProcsToTerminate(filter):
	while True:
		exists = False
		for p in psutil.process_iter():
			if filter(p):
				exists = True
				break

		if not exists:
			break

		time.sleep(0.1)

def isDispatcherProcess(p):
	try: return p.name() == DISPATCHER_EXE
	except psutil.Error: return False

def isGameProcess(p):
	try:
		return p.name() != GATE_EXE and isExeContains(p, "goworld") and isCmdContains(p, "-gid=")
	except psutil.Error:
		return False

def isGateProcess(p):
	try:
		return p.name() == GATE_EXE and isExeContains(p, "goworld") and isCmdContains(p, "-gid=")
	except psutil.Error:
		return False

def isCmdContains(p, opt):
	for cmdopt in p.cmdline():
		if opt in cmdopt:
			return True
	return False

def isExeContains(p, s):
	return s in p.exe()

def getDispatcherExe():
	return os.path.join("components", "dispatcher", DISPATCHER_EXE)

def getGateExe():
	return os.path.join("components", "gate", GATE_EXE)

def getGameExe():
	global gamePath
	return gamePath

def analyzeConfig(config):
	for sec in config.sections():
		if sec[:4] == "game" and sec != "game_common": # game config
			gameid = int(sec[4:])
			gameids.append(gameid)
		elif sec[:4] == "gate" and sec != "gate_common": # gate config
			gateid = int(sec[4:])
			gateids.append(gateid)

	gameids.sort()
	gateids.sort()
	print >>sys.stderr, "> Found %d games and %d gates in goworld.ini" % (len(gameids), len(gateids))

if __name__ == '__main__':
	main()
