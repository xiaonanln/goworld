import os

for path, dirs, files in os.walk(".."):
	try: dirs.remove('.git')
	except: pass
	try: dirs.remove('.svn')
	except: pass
	try: dirs.remove('.idea')
	except: pass
	for fn in files:
		if fn.endswith(".go"):
			fp = os.path.join(path, fn)
			print fp
			with os.popen('gofmt -s "%s"' % fp) as fd:
				fmtout = fd.read()

			with open("fp", 'wb') as fd:
				fd.write(fmtout)
