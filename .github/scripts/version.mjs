import { execFileSync } from "node:child_process";
import { appendFileSync, writeFileSync } from "node:fs";
import { pathToFileURL } from "node:url";

const titlePattern = /^(feat|fix|perf|docs|style|refactor|test|build|ci|chore|revert)(\([^)]+\))?!?: \S.*$/;

export function validTitle(title) {
  return titlePattern.test(title);
}

export function releaseVersion(previous, messages) {
  let bump = 0;
  for (const message of messages) {
    const title = message.split("\n")[0];
    if (/^[a-z]+(\([^)]+\))?!: /.test(title) || /^BREAKING[ -]CHANGE: /m.test(message)) {
      bump = Math.max(bump, 3);
    } else if (/^feat(\([^)]+\))?: /.test(title)) {
      bump = Math.max(bump, 2);
    } else if (/^(fix|perf|revert)(\([^)]+\))?: /.test(title)) {
      bump = Math.max(bump, 1);
    }
  }
  if (!bump) return "";
  if (!previous) return "v0.1.0";
  const match = /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.exec(previous);
  if (!match) throw new Error(`Invalid stable version: ${previous}`);
  let [major, minor, patch] = match.slice(1).map(Number);
  if (bump === 3 && major > 0) {
    major++;
    minor = patch = 0;
  } else if (bump >= 2) {
    minor++;
    patch = 0;
  } else {
    patch++;
  }
  return `v${major}.${minor}.${patch}`;
}

function main() {
  if (process.argv.includes("--validate-title")) {
    if (!validTitle(process.env.PR_TITLE ?? "")) {
      throw new Error("PR title must be a Conventional Commit, e.g. feat(watch): add live previews");
    }
    return;
  }
  const git = (...args) => execFileSync("git", args, { encoding: "utf8" }).trim();
  const tags = git("tag", "--merged", "HEAD", "--sort=-version:refname").split("\n");
  const previous = tags.find(tag => /^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.test(tag)) ?? "";
  const messages = git("log", "--format=%B%x00", previous ? `${previous}..HEAD` : "HEAD")
    .split("\0").map(message => message.trim()).filter(Boolean);
  const version = releaseVersion(previous, messages);
  if (!process.env.GITHUB_OUTPUT) throw new Error("GITHUB_OUTPUT is required for release planning");
  appendFileSync(process.env.GITHUB_OUTPUT, `version=${version}\n`);
  if (!version) {
    console.log("No release-worthy commits since the last stable tag.");
    return;
  }
  const notes = messages.map(message => `- ${message.split("\n")[0]}`).join("\n");
  writeFileSync("release-notes.md", `# ${version}\n\n${notes}\n\nCommit: ${git("rev-parse", "HEAD")}\n`);
  console.log(`${previous || "initial release"} -> ${version}`);
}

if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) main();
