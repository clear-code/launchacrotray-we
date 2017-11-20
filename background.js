(async () => {
  await configs.$load();
  await applyMCDConfigs();
  await setDefaultPath();
  await setAcrotrayMonitor();
  configs.$addObserver(onConfigUpdated);
})();

async function applyMCDConfigs() {
  try {
    var response = await send({ command: 'read-mcd-configs' });
    log('loaded MCD configs: ', response);
    Object.keys(response).forEach((aKey) => {
      configs[aKey] = response[aKey];
      configs.$lock(aKey);
    });
  }
  catch(aError) {
    log('Failed to read MCD configs: ', aError);
  }
}

async function setDefaultPath() {
  if (configs.acrotrayapp)
    return;
  try {
    let response = await send({ command: 'get-acrotray-path' });
    if (response) {
      log('Received: ', response);
      if (response.path)
        configs.acrotrayapp = response.path;
    }
  }
  catch(aError) {
    log('Error: ', aError);
  }
}

async function setAcrotrayMonitor() {
  log('setup Acrotray.exe monitor');
  window.setInterval(launch, 5000);
}

function onConfigUpdated(aKey) {
  switch (aKey) {
    case 'contextMenu':
      if (configs.contextMenu) {
        installMenuItems();
      }
      else {
        browser.contextMenus.removeAll();
      }
      break;

    case 'forceielist':
      uninstallBlocker();
      if (!configs.disableForce)
        installBlocker();
      break;

    case 'disableForce':
      if (configs.disableForce) {
        uninstallBlocker();
      }
      else {
        installBlocker();
      }
      break;
  }
}

async function launch() {
  if (!configs.acrotrayapp && !configs.acrotrayargs)
    return;

  let message = {
    command: 'launch',
    params: {
      path: configs.acrotrayapp,
      args: configs.acrotrayargs.trim().split(/\s+/).filter((aItem) => !!aItem),
      url:  aURL
    }
  };
  try{
    let response = await send(message);
    log('Received: ', response);
  }
  catch(aError) {
    log('Error: ', aError);
  }
}

function send(aMessage) {
  log('Sending: ', aMessage);
  if (configs.debug)
    aMessage.logging = true;
  return browser.runtime.sendNativeMessage('com.clear_code.launchacrotray_we_host', aMessage);
}
