import type { Reporter, TestResult } from '@rstest/core';

import path from 'node:path';
import { createInterface } from 'node:readline';

// rstest wraps every file in a synthetic suite with this name. It isn't
// exported, and it would otherwise show up as a test in every package.
const ROOT_SUITE_NAME = 'Rstest:_internal_root_suite';

// Emits rstest results as Go test2json events so gotestsum can render frontend
// tests alongside `go test -json`. Each spec file becomes a "package", describe
// blocks become parent tests and cases become subtests, mirroring Go's
// Test/subtest naming. Events are written to stdout, rstest's default reporter
// is replaced by this one so nothing else of rstest's ends up there.
//
// When forwardStdin is set, test2json lines piped into rstest's stdin (go test
// -json, from cmake/scripts/UnifiedTests.cmake) are written out through the same
// writer. That lets both run in parallel without two processes sharing a pipe,
// where large events could get split and interleaved mid-line.
export class Test2JsonReporter implements Reporter {
  flushOutputStreams = false;

  private readonly root: string;
  private readonly forwardStdin: boolean;
  private stdinClosed: Promise<void> = Promise.resolve();
  // Tests that have already reported a result. Console output that shows up
  // after a test finished (like jsdom complaining from a timer) is attached to
  // the package instead, gotestsum treats output for a finished test as a new
  // test that never completed and fails it.
  private readonly finished = new Set<string>();

  constructor(root: string, forwardStdin: boolean) {
    this.root = root;
    this.forwardStdin = forwardStdin;
  }

  private pkg(testPath: string): string {
    return path.relative(this.root, testPath).replace(/\.spec\.tsx?$/, '');
  }

  private name(parents: Array<string> | undefined, name: string): string {
    return [...(parents ?? []), name].join('/').replaceAll(' ', '_');
  }

  private emit(event: Record<string, unknown>): void {
    process.stdout.write(`${JSON.stringify({ Time: new Date().toISOString(), ...event })}\n`);
  }

  private finish(result: TestResult, isPackage: boolean): void {
    const base = {
      Package: this.pkg(result.testPath),
      ...(isPackage ? {} : { Test: this.name(result.parentNames, result.name) }),
    };
    if (!isPackage) {
      this.finished.add(`${base.Package}\0${base.Test}`);
    }
    for (const error of result.errors ?? []) {
      this.emit({ ...base, Action: 'output', Output: `${error.stack ?? error.message}\n` });
    }
    this.emit({
      ...base,
      Action: result.status === 'todo' ? 'skip' : result.status,
      Elapsed: (result.duration ?? 0) / 1000,
    });
  }

  onTestRunStart: Reporter['onTestRunStart'] = () => {
    if (!this.forwardStdin) {
      return;
    }
    const lines = createInterface({ input: process.stdin, crlfDelay: Number.POSITIVE_INFINITY });
    lines.on('line', line => process.stdout.write(`${line}\n`));
    this.stdinClosed = new Promise(resolve => lines.on('close', resolve));
  };

  // Keep rstest alive until whatever is piped in has finished too, and until
  // stdout has drained, since it can be asynchronous when it's a pipe.
  onTestRunEnd: Reporter['onTestRunEnd'] = async () => {
    await this.stdinClosed;
    await new Promise<void>(resolve => process.stdout.write('', () => resolve()));
  };

  onTestFileStart: Reporter['onTestFileStart'] = file => {
    this.emit({ Action: 'start', Package: this.pkg(file.testPath) });
  };

  onTestSuiteStart: Reporter['onTestSuiteStart'] = suite => {
    if (suite.name === ROOT_SUITE_NAME) {
      return;
    }
    this.emit({ Action: 'run', Package: this.pkg(suite.testPath), Test: this.name(suite.parentNames, suite.name) });
  };

  onTestSuiteResult: Reporter['onTestSuiteResult'] = result => {
    if (result.name !== ROOT_SUITE_NAME) {
      this.finish(result, false);
    }
  };

  onTestCaseStart: Reporter['onTestCaseStart'] = test => {
    this.emit({ Action: 'run', Package: this.pkg(test.testPath), Test: this.name(test.parentNames, test.name) });
  };

  onTestCaseResult: Reporter['onTestCaseResult'] = result => this.finish(result, false);

  onTestFileResult: Reporter['onTestFileResult'] = result => this.finish(result, true);

  onUserConsoleLog: Reporter['onUserConsoleLog'] = log => {
    const pkg = this.pkg(log.testPath);
    const test = log.taskName ? this.name(log.taskParentNames, log.taskName) : undefined;
    const attach = test !== undefined && !this.finished.has(`${pkg}\0${test}`);
    this.emit({
      Action: 'output',
      Package: pkg,
      ...(attach ? { Test: test } : {}),
      Output: log.content.endsWith('\n') ? log.content : `${log.content}\n`,
    });
  };
}
