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
currentGameId = ''
nohup = False

def main():
	opts, args = getopt.getopt(sys.argv[1:], "", ["log=", "nohup"])
	global loglevel
	for opt, val in opts:
		if opt == "--log":
			loglevel = val
		elif opt == "--nohup":
			nohup = True
			print >>sys.stderr, "> Using nohup"

	if len(args) == 0:
		showUsage()
		exit(1)

	cmd = args[0].lower()
	verifyExecutionEnv(cmd)

	config = ConfigParser.SafeConfigParser()
	config.read("goworld.ini")
	analyzeConfig(config)
	detectCurrentGameId()

	for cmd, cmdArgs in parseArguments(args):
		if cmd == 'status':
			showStatus()
		elif cmd in ("start", "restore"):
			global gameName
			try:
				gameName = cmdArgs[0]
			except:
				showUsage()
				exit(1)

			global gamePath
			gamePath = detectGamePath(gameName)
			if not os.path.exists(gamePath):
				print >>sys.stderr, "! %s is not found, goworld.py build %s first" % (gamePath, gameName)
				exit(2)

			if cmd == "start":
				startServer()
			elif cmd == "restore":
				restoreGames()

		elif cmd == 'stop':
			stopServer()
		elif cmd == 'kill':
			stopServer(kill=True)
		elif cmd == 'build':
			buildTargets = cmdArgs
			if not buildTargets: buildTargets = ['engine']

			for buildTarget in buildTargets:
				build(buildTarget)
		elif cmd == 'freeze':
			freezeGames()
		elif cmd == 'sleep':
			sleepTime = float(cmdArgs[0])
			time.sleep(sleepTime)
		else:
			print >>sys.stderr, "invalid command: %s" % cmd
			showUsage()
			exit(1)

	print >>sys.stderr, '> %s %s OK' % (sys.argv[0], ' '.join(args))

def parseArguments(args):
	cmds = []
	i = 0
	while i < len(args):
		cmd = args[i]
		if cmd in ('start', 'restore'):
			args = (args[i+1],) if i+1<len(args) else ()
			i += 1
		elif cmd in ('build'):
			args = args[i+1:]
			i = len(args)
		else:
			args = ()

		if cmd == 'reload': # reload == freeze + restore
			if not currentGameId:
				_showStatus(1, len(gateids), len(gameids))
				print >>sys.stderr, '! can not detect current game, not running ?'
				exit(2)

			print >>sys.stderr, '> Detected game: %s for reload' % currentGameId
			cmds.append(('freeze', ()))
			# cmds.append(('sleep', (1, )))
			cmds.append(('restore', (currentGameId, )))
		else:
			cmds.append( (cmd, args) )

	return cmds


def detectCurrentGameId():
	global currentGameId

	if currentGameId != '':
		return

	_, _, gameProcs = visitProcs()
	if not gameProcs:
		return

	gameExe = None
	for proc in gameProcs:
		if gameExe is None: gameExe = proc.exe()
		elif gameExe != proc.exe():
			print >>sys.stderr, '! found multiple game processes with different exe: %s & %s' % gameExe, proc.exe()
			return

	if gameExe == '':
		print >>sys.stderr, '! get process exe failed'
		return

	gameExe = os.path.relpath(gameExe, goworldPath)
	print >>sys.stderr, '> Found game exe: %s' % gameExe
	if os.name == 'nt' and gameExe.endswith('.exe'): # strip .exe if necessary
		gameExe = gameExe[:-4]

	currentGameId = os.path.dirname(gameExe)

def verifyExecutionEnv(cmd):
	global goworldPath
	goworldPath = os.getcwd()
	print >>sys.stderr, '> Detected goworld path:', goworldPath
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
	goworld.py kill - kill game server processes
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
	if os.system('cd "%s" && go build' % os.path.join("components", "dispatcher")) != 0:
		exit(2)
	print >>sys.stderr, 'OK'

def buildGate():
	print >>sys.stderr, '> building gate ...',
	if os.system('cd "%s" && go build' % os.path.join("components", "gate")) != 0:
		exit(2)
	print >>sys.stderr, 'OK'

def buildGame(gameId):
	gamePath = detectGamePath(gameId)
	gameDir = os.path.dirname(gamePath)
	print >>sys.stderr, '> building %s ...' % gameDir,
	if os.system('cd "%s" && go build' % gameDir) != 0:
		exit(2)
	print >>sys.stderr, 'OK'

def freezeGames():
	_, _, gameProcs = visitProcs()
	if not gameProcs:
		print >>sys.stderr, "! game process is not found"
		exit(2)

	for proc in gameProcs:
		proc.send_signal(signal.SIGINT)

	print >>sys.stderr, "Waiting for game processes to terminate ...",
	waitProcsToTerminate(isGameProcess)
	print >>sys.stderr, 'OK'

	_showStatus(1, len(gateids), 0)

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
	gameName = "game (unknown)" if not currentGameId else "game (%s)" % currentGameId
	print >>sys.stderr, "%-32s expect %d found %d %s" % ("dispatcher", expectDispatcherCount, len(dispatcherProcs), "GOOD" if len(dispatcherProcs) == expectDispatcherCount else "BAD!")
	print >>sys.stderr, "%-32s expect %d found %d %s" % ("gate", expectGateCount, len(gateProcs), "GOOD" if expectGateCount == len(gateProcs) else "BAD!")
	print >>sys.stderr, "%-32s expect %d found %d %s" % (gameName, expectGameCount, len(gameProcs), "GOOD" if expectGameCount == len(gameProcs) else "BAD!")

def restoreGames():
	dispatcherProcs, _, gameProcs = visitProcs()
	if len(dispatcherProcs) != 1 or gameProcs:
		print >>sys.stderr, "! wrong process status"
		_showStatus(1, len(gateids), 0)
		exit(2)

	global nohup
	nohupArgs = ['nohup'] if nohup else []

	for gameid in gameids:
		print >> sys.stderr, "Restore game%d ..." % gameid,
		gameProc = psutil.Popen(nohupArgs+[getGameExe(), "-gid=%d" % gameid, "-log", loglevel, '-restore'])
		print >> sys.stderr, gameProc.status()

	_showStatus(1, len(gateids), len(gameids))

def startServer():
	dispatcherProcs, gateProcs, gameProcs = visitProcs()
	if dispatcherProcs or gateProcs or gameProcs:
		print >>sys.stderr, "goworld is already running ..."
		_showStatus(1, len(gateids), len(gameids))
		exit(2)

	# now the system is clear, start server processes ...
	global nohup
	nohupArgs = ['nohup'] if nohup else []
	print >>sys.stderr, "Start dispatcher ...",
	dispatcherProc = psutil.Popen(nohupArgs+[getDispatcherExe()])
	print >>sys.stderr, dispatcherProc.status()

	for gameid in gameids:
		print >> sys.stderr, "Start game%d ..." % gameid,
		gameProc = psutil.Popen(nohupArgs+[getGameExe(), "-gid=%d" % gameid, "-log", loglevel])
		print >> sys.stderr, gameProc.status()

	for gateid in gateids:
		print >>sys.stderr, "Start gate%d ..." % gateid,
		gateProc = psutil.Popen(nohupArgs+[getGateExe(), "-gid=%d" % gateid, "-log", loglevel])
		print >>sys.stderr, gateProc.status()

	_showStatus(1, len(gateids), len(gameids))

def stopServer(kill=False):
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
		if not kill:
			proc.send_signal(signal.SIGTERM)
		else:
			killProc(proc)

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
