# Launchacrotray WE

Provides ability to watch acrotray.exe and launches it automatically if it isn't executed.

*IMPORTANT NOTE: This Firefox extension works only on Windows.*

# Steps to install

 1. Download a zip package of the native messaging host from the [releases page](https://github.com/clear-code/launchacrotray-we/releases).
 2. Unzip downloaded file.
 3. Double-click the batch file named `install.bat`.
 4. Install "Launchacrotray WE" Firefox addon from its xpi package.

# Steps to uninstall

 1. Uninstall "Launchacrotray WE" Firefox addon via the addon manager.
 2. Double-click the batch file named `uninstall.bat`.

# How to build the native messaging host

```bash
$ make host
```

# How to customize default addon behavior

This addon supports to customize by MCD.

Here is the preference keys.

* `extensions.launchacrotray.acrotrayapp`
  * Specify path to `acrotray.exe`. The default value is filled in by registry. For example, `C:\Program Files (x86)\Adobe\Acrobat 2017\Acrobat\acrotray.exe`.
* `extensions.launchacrotray.acrotrayargs`
  * Specify command arguments for `acrotray.exe`. The default value is empty.
* `extensions.launchacrotray.watchinterval`
  * Specify interval to check `acrotray.exe` process. The default interval is 15 seconds.
* `extensions.launchacrotray.debug`
  * Specify whether debug mode is enabled. The default value is `false`.

# License

MPL 2.0
