import {$, ProcessOutput} from 'zx';

class CommandInfo {
  name: string;
  installPath: string;
  version: string;

  constructor(data: Partial<CommandInfo>) {
    Object.assign(this, {
      ...data,
    });
  }

  isInstalled(): boolean {
    return this.installPath !== '';
  }
}


async function getInstallInfo(command: string): Promise<CommandInfo> {
  const result = await $`which ${command}`;

  return new CommandInfo({
    name: command,
    installPath: result.stdout.trim(),
  });
}

async function EXEC(command: string | CommandInfo, ...args: Array<string>): Promise<ProcessOutput> {
  let commandString = '';
  if (typeof command === 'string') {
    commandString = command;
  } else if (command instanceof CommandInfo) {
    commandString = command.installPath;
  } else {
    throw new Error('invalid command provided to EXEC');
  }

  commandString = `${commandString} ${ args.join(' ') }`.trim();

  console.debug('DEBUG:', commandString);
  return $`${commandString}`;
}



const pwd = (await $`pwd`).stdout.trim();
console.info(`working in: ${ pwd }`);

const go = (await getInstallInfo('go'));

if (go.isInstalled()) {
  console.info(`go is installed at: ${ go.installPath }`);
} else {
  console.warn(`go is not installed!`);
}

EXEC(go, 'version');
// EXEC(go, 'list', '-m');
