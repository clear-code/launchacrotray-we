.PHONY: xpi host managed

xpi: extlib/webextensions-lib-configs/Configs.js extlib/webextensions-lib-options/Options.js extlib/webextensions-lib-l10n/l10n.js
	git submodule update
	cp extlib/webextensions-lib-configs/Configs.js common/
	cp extlib/webextensions-lib-options/Options.js options/
	cp extlib/webextensions-lib-l10n/l10n.js options/
	rm -f launchacrotray-we.xpi
	zip -r -0 launchacrotray-we.xpi *.json *.js _locales common options

extlib/webextensions-lib-configs/Configs.js:
	git submodule update --init

extlib/webextensions-lib-options/Options.js:
	git submodule update --init

extlib/webextensions-lib-l10n/l10n.js:
	git submodule update --init

host:
	host/build.sh
	rm -f launchacrotray-we-host.zip
	cd host && zip -r -9 ../launchacrotray-we-host.zip 386 amd64 *.bat *.json

all: host xpi

