import assert from "node:assert/strict";
import test from "node:test";
import { execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import { fileURLToPath } from "node:url";
import { releaseVersion, validTitle } from "./version.mjs";

test("Conventional PR titles", () => {
  for (const title of ["feat(watch): add previews", "fix: resolve crash", "feat!: change schema", "fix: x"]) {
    assert.equal(validTitle(title), true);
  }
  for (const title of ["Update stuff", "feat: ", "unknown: change behavior"]) {
    assert.equal(validTitle(title), false);
  }
});

test("release planner uses reachable stable tags and writes outputs", () => {
  const root = mkdtempSync(join(tmpdir(), "guardrail-version-"));
  try {
    const git = (...args) => execFileSync("git", args, { cwd: root, stdio: "pipe" });
    git("init", "--quiet");
    git("config", "user.email", "test@example.invalid");
    git("config", "user.name", "Version Test");
    git("commit", "--allow-empty", "-m", "feat: initial feature");
    const output = join(root, "output");
    const script = fileURLToPath(new URL("./version.mjs", import.meta.url));
    const plan = () => {
      writeFileSync(output, "");
      execFileSync(process.execPath, [script], {
        cwd: root, env: { ...process.env, GITHUB_OUTPUT: output }, stdio: "pipe"
      });
      return readFileSync(output, "utf8").trim();
    };
    assert.equal(plan(), "version=v0.1.0");
    git("tag", "v0.1.0");
    assert.equal(plan(), "version=");
    git("commit", "--allow-empty", "-m", "docs: improve docs");
    assert.equal(plan(), "version=");
    git("commit", "--allow-empty", "-m", "fix: repair watch");
    assert.equal(plan(), "version=v0.1.1");
    git("tag", "v0.2.0-beta.1");
    git("commit", "--allow-empty", "-m", "feat!: change contract");
    assert.equal(plan(), "version=v0.2.0");
    assert.match(readFileSync(join(root, "release-notes.md"), "utf8"), /change contract/);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

test("initial and non-release changes", () => {
  assert.equal(releaseVersion("", ["feat: add watch"]), "v0.1.0");
  assert.equal(releaseVersion("", ["docs: update guide"]), "");
  assert.equal(releaseVersion("v0.1.0", ["ci: add workflow", "test: add tests"]), "");
  assert.equal(releaseVersion("v0.1.0", []), "");
});

test("highest semantic bump wins", () => {
  assert.equal(releaseVersion("v1.2.3", ["fix: bug", "perf: improve speed"]), "v1.2.4");
  assert.equal(releaseVersion("v1.2.3", ["fix: bug", "feat: new parser"]), "v1.3.0");
  assert.equal(releaseVersion("v1.2.3", ["feat!: new contract"]), "v2.0.0");
  assert.equal(releaseVersion("v1.2.3", ["refactor: contract\n\nBREAKING CHANGE: new input"]), "v2.0.0");
  assert.equal(releaseVersion("v0.1.9", ["feat!: new contract"]), "v0.2.0");
  assert.equal(releaseVersion("v0.1.9", ["fix: bug"]), "v0.1.10");
  assert.equal(releaseVersion("v1.0.0", ["revert: undo change"]), "v1.0.1");
  assert.throws(() => releaseVersion("v1.0.0-beta", ["fix: bug"]), /Invalid stable version/);
});
